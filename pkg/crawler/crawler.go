package crawler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/HeisenbergV/repoinsight/api"
	"github.com/HeisenbergV/repoinsight/pkg/logger"
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
		return fmt.Errorf("搜索仓库失败: %v", err)
	}

	logger.Infof("找到 %d 个仓库", len(result.Repositories))

	// 处理每个仓库
	for i, repo := range result.Repositories {
		logger.Infof("正在处理第 %d/%d 个仓库: %s", i+1, len(result.Repositories), repo.GetFullName())
		if err := c.processRepository(repo); err != nil {
			logger.Errorf("处理仓库 %s 失败: %v", repo.GetFullName(), err)
			continue
		}
		logger.Infof("成功处理仓库: %s", repo.GetFullName())
	}

	logger.Info("爬取完成")
	return nil
}

func (c *Crawler) processRepository(repo *github.Repository) error {
	// 获取 README
	readme, err := c.getReadme(repo.GetOwner().GetLogin(), repo.GetName())
	if err != nil {
		logger.Warnf("获取仓库 %s 的 README 失败: %v", repo.GetFullName(), err)
	}

	// 获取 AI 分析
	aiAnalysis, err := c.analyzeRepository(repo, readme)
	if err != nil {
		logger.Warnf("分析仓库 %s 失败: %v", repo.GetFullName(), err)
	}

	// 转换标签为 JSON
	topics, _ := json.Marshal(repo.Topics)

	// 创建或更新仓库记录
	repository := &api.Repository{
		FullName:      repo.GetFullName(),
		Name:          repo.GetName(),
		Owner:         repo.GetOwner().GetLogin(),
		Description:   repo.GetDescription(),
		URL:           repo.GetHTMLURL(),
		Stars:         repo.GetStargazersCount(),
		Forks:         repo.GetForksCount(),
		Language:      repo.GetLanguage(),
		Topics:        string(topics),
		Readme:        readme,
		AIAnalysis:    aiAnalysis,
		LastPushedAt:  repo.GetPushedAt().Time,
		CreatedAt:     repo.GetCreatedAt().Time,
		UpdatedAt:     repo.GetUpdatedAt().Time,
		IsArchived:    repo.GetArchived(),
		License:       repo.GetLicense().GetName(),
		DefaultBranch: repo.GetDefaultBranch(),
		OpenIssues:    repo.GetOpenIssuesCount(),
		Watchers:      repo.GetWatchersCount(),
		Size:          repo.GetSize(),
		HasIssues:     repo.GetHasIssues(),
		HasProjects:   repo.GetHasProjects(),
		HasWiki:       repo.GetHasWiki(),
		HasPages:      repo.GetHasPages(),
		HasDownloads:  repo.GetHasDownloads(),
		IsTemplate:    repo.GetIsTemplate(),
	}

	// 检查数据库中是否已存在相同 URL 的记录
	var existingRepo api.Repository
	result := c.db.Where("url = ?", repository.URL).First(&existingRepo)

	if result.Error == nil {
		// 记录已存在，更新数据
		logger.Infof("更新仓库: %s (URL: %s)", repository.FullName, repository.URL)
		return c.db.Model(&existingRepo).Updates(repository).Error
	} else if result.Error == gorm.ErrRecordNotFound {
		// 记录不存在，创建新记录
		logger.Infof("创建新仓库: %s (URL: %s)", repository.FullName, repository.URL)
		return c.db.Create(repository).Error
	} else {
		// 其他错误
		return result.Error
	}
}

func (c *Crawler) getReadme(owner, repo string) (string, error) {
	readme, _, err := c.client.RepositoriesGetReadme(context.Background(), owner, repo, &github.RepositoryContentGetOptions{})
	if err != nil {
		return "", err
	}

	content, err := readme.GetContent()
	if err != nil {
		return "", err
	}

	return content, nil
}

func (c *Crawler) analyzeRepository(repo *github.Repository, readme string) (string, error) {
	// TODO: 实现 AI 分析逻辑
	// 这里需要调用 AI 服务来分析仓库
	// 暂时返回一个简单的分析结果
	return fmt.Sprintf("这是一个 %s 项目，主要使用 %s 语言开发。\n\n项目描述：%s\n\nREADME 内容：%s",
		repo.GetName(),
		repo.GetLanguage(),
		repo.GetDescription(),
		readme[:min(500, len(readme))]), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
