package crawler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/HeisenbergV/repoinsight/pkg/logger"
	"github.com/HeisenbergV/repoinsight/pkg/models"
	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type GitHubClient interface {
	SearchRepositories(ctx context.Context, query string, opts *github.SearchOptions) (*github.RepositoriesSearchResult, *github.Response, error)
	RepositoriesGetReadme(ctx context.Context, owner, repo string, opts *github.RepositoryContentGetOptions) (*github.RepositoryContent, *github.Response, error)
}

type Crawler struct {
	client GitHubClient
	db     *gorm.DB
	config *Config
}

type Config struct {
	Token           string
	SearchKeyword   string
	MaxReposPerPage int
}

type githubClientAdapter struct {
	*github.Client
}

func (g *githubClientAdapter) SearchRepositories(ctx context.Context, query string, opts *github.SearchOptions) (*github.RepositoriesSearchResult, *github.Response, error) {
	return g.Client.Search.Repositories(ctx, query, opts)
}

func (g *githubClientAdapter) RepositoriesGetReadme(ctx context.Context, owner, repo string, opts *github.RepositoryContentGetOptions) (*github.RepositoryContent, *github.Response, error) {
	return g.Client.Repositories.GetReadme(ctx, owner, repo, opts)
}

func NewCrawler(db *gorm.DB, config *Config) *Crawler {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &Crawler{
		client: &githubClientAdapter{client},
		db:     db,
		config: config,
	}
}

func (c *Crawler) Start() error {
	logger.Info("开始爬取 GitHub 仓库...")

	// 创建爬取历史记录
	crawlHistory := &models.CrawlHistory{
		Keyword:        c.config.SearchKeyword,
		StartedAt:      time.Now(),
		Status:         "running",
		TotalRepos:     0,
		ProcessedRepos: 0,
	}
	if err := c.db.Create(crawlHistory).Error; err != nil {
		return fmt.Errorf("创建爬取历史记录失败: %v", err)
	}

	// 搜索仓库
	opts := &github.SearchOptions{
		Sort:  "updated",
		Order: "desc",
		ListOptions: github.ListOptions{
			PerPage: c.config.MaxReposPerPage,
		},
	}

	// 执行搜索
	result, _, err := c.client.SearchRepositories(context.Background(), c.config.SearchKeyword, opts)
	if err != nil {
		crawlHistory.Status = "failed"
		crawlHistory.ErrorMessage = err.Error()
		c.db.Save(crawlHistory)
		return fmt.Errorf("搜索仓库失败: %v", err)
	}

	totalRepos := len(result.Repositories)
	logger.Infof("找到 %d 个仓库", totalRepos)

	// 更新爬取历史记录
	crawlHistory.TotalRepos = totalRepos
	c.db.Save(crawlHistory)

	// 处理每个仓库
	for i, repo := range result.Repositories {
		logger.Infof("正在处理第 %d/%d 个仓库: %s", i+1, totalRepos, repo.GetFullName())
		if err := c.processRepository(repo, i+1, crawlHistory); err != nil {
			logger.Errorf("处理仓库 %s 失败: %v", repo.GetFullName(), err)
			continue
		}
		logger.Infof("成功处理仓库: %s", repo.GetFullName())
	}

	// 更新爬取历史记录状态
	crawlHistory.Status = "completed"
	crawlHistory.CompletedAt = time.Now()
	c.db.Save(crawlHistory)

	logger.Info("爬取完成")
	return nil
}

func (c *Crawler) processRepository(repo *github.Repository, rank int, crawlHistory *models.CrawlHistory) error {
	maxRetries := 3
	retryInterval := time.Second

	for i := 0; i < maxRetries; i++ {
		var existingRepo models.Repository
		result := c.db.Where("url = ?", repo.GetHTMLURL()).First(&existingRepo)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				// 转换标签为 JSON
				topics, _ := json.Marshal(repo.Topics)

				// 创建新记录
				newRepo := models.Repository{
					FullName:       repo.GetFullName(),
					Name:           repo.GetName(),
					Owner:          repo.GetOwner().GetLogin(),
					Description:    repo.GetDescription(),
					URL:            repo.GetHTMLURL(),
					Stars:          repo.GetStargazersCount(),
					Forks:          repo.GetForksCount(),
					Language:       repo.GetLanguage(),
					Topics:         string(topics),
					Readme:         "",
					LastPushedAt:   repo.GetPushedAt().Time,
					IsArchived:     repo.GetArchived(),
					License:        "",
					DefaultBranch:  repo.GetDefaultBranch(),
					OpenIssues:     repo.GetOpenIssuesCount(),
					Watchers:       repo.GetWatchersCount(),
					Size:           repo.GetSize(),
					HasIssues:      repo.GetHasIssues(),
					HasProjects:    repo.GetHasProjects(),
					HasWiki:        repo.GetHasWiki(),
					HasPages:       repo.GetHasPages(),
					HasDownloads:   repo.GetHasDownloads(),
					IsTemplate:     repo.GetIsTemplate(),
					SearchKeyword:  c.config.SearchKeyword,
					SearchRank:     rank,
					LastCrawledAt:  time.Now(),
					AnalysisStatus: "pending",
				}

				if err := c.db.Create(&newRepo).Error; err != nil {
					if i < maxRetries-1 {
						logger.Warnf("创建仓库记录失败，正在重试 (%d/%d): %v", i+1, maxRetries, err)
						time.Sleep(retryInterval)
						continue
					}
					return fmt.Errorf("创建仓库记录失败: %v", err)
				}
			} else {
				if i < maxRetries-1 {
					logger.Warnf("查询仓库记录失败，正在重试 (%d/%d): %v", i+1, maxRetries, result.Error)
					time.Sleep(retryInterval)
					continue
				}
				return fmt.Errorf("查询仓库记录失败: %v", result.Error)
			}
		} else {
			// 转换标签为 JSON
			topics, _ := json.Marshal(repo.Topics)

			// 更新现有记录
			updates := map[string]interface{}{
				"name":            repo.GetName(),
				"owner":           repo.GetOwner().GetLogin(),
				"description":     repo.GetDescription(),
				"stars":           repo.GetStargazersCount(),
				"forks":           repo.GetForksCount(),
				"language":        repo.GetLanguage(),
				"topics":          string(topics),
				"readme":          "",
				"last_pushed_at":  repo.GetPushedAt().Time,
				"is_archived":     repo.GetArchived(),
				"license":         "",
				"default_branch":  repo.GetDefaultBranch(),
				"open_issues":     repo.GetOpenIssuesCount(),
				"watchers":        repo.GetWatchersCount(),
				"size":            repo.GetSize(),
				"has_issues":      repo.GetHasIssues(),
				"has_projects":    repo.GetHasProjects(),
				"has_wiki":        repo.GetHasWiki(),
				"has_pages":       repo.GetHasPages(),
				"has_downloads":   repo.GetHasDownloads(),
				"is_template":     repo.GetIsTemplate(),
				"search_keyword":  c.config.SearchKeyword,
				"search_rank":     rank,
				"last_crawled_at": time.Now(),
				"analysis_status": "pending",
			}

			if err := c.db.Model(&existingRepo).Updates(updates).Error; err != nil {
				if i < maxRetries-1 {
					logger.Warnf("更新仓库记录失败，正在重试 (%d/%d): %v", i+1, maxRetries, err)
					time.Sleep(retryInterval)
					continue
				}
				return fmt.Errorf("更新仓库记录失败: %v", err)
			}
		}

		// 更新爬取历史记录的处理进度
		crawlHistory.ProcessedRepos++
		if err := c.db.Save(crawlHistory).Error; err != nil {
			logger.Warnf("更新爬取历史记录失败: %v", err)
		}

		return nil
	}
	return fmt.Errorf("处理仓库失败，已达到最大重试次数")
}
