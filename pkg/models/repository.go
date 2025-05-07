package models

import (
	"time"

	"gorm.io/gorm"
)

// Response 表示 API 响应
type Response struct {
	Data interface{} `json:"data"`
	Meta Meta        `json:"meta"`
}

// Meta 表示分页元数据
type Meta struct {
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// SystemStatus 表示系统状态
type SystemStatus struct {
	TotalRepositories int64     `json:"total_repositories"`
	LastUpdated       time.Time `json:"last_updated"`
	Status            string    `json:"status"`
}

// AIAnalysis 表示仓库的 AI 分析结果
type AIAnalysis struct {
	ID           uint   `gorm:"primaryKey"`
	URL          string `gorm:"uniqueIndex;not null"`
	Content      string `gorm:"type:text"`
	Status       string `gorm:"type:varchar(20);not null;default:'pending'"`
	ModelVersion string `gorm:"type:varchar(50)"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Repository struct {
	ID            uint           `json:"id" gorm:"primarykey"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
	FullName      string         `json:"full_name" gorm:"uniqueIndex"` // 项目全名，如 "owner/repo"
	Name          string         `json:"name"`                         // 项目名称
	Owner         string         `json:"owner"`                        // 项目所有者
	Description   string         `json:"description"`                  // 项目描述
	URL           string         `json:"url"`                          // GitHub 项目地址
	Stars         int            `json:"stars"`                        // 星标数
	Forks         int            `json:"forks"`                        // 分支数
	Language      string         `json:"language"`                     // 主要编程语言
	Topics        string         `json:"topics" gorm:"type:text"`      // 项目标签，JSON 格式存储
	Readme        string         `json:"readme" gorm:"type:text"`      // 原始 README 内容
	LastPushedAt  time.Time      `json:"last_pushed_at"`               // 最后推送时间
	IsArchived    bool           `json:"is_archived"`                  // 是否已归档
	License       string         `json:"license"`                      // 许可证
	DefaultBranch string         `json:"default_branch"`               // 默认分支
	OpenIssues    int            `json:"open_issues"`                  // 开放问题数
	Watchers      int            `json:"watchers"`                     // 关注者数
	Size          int            `json:"size"`                         // 仓库大小（KB）
	HasIssues     bool           `json:"has_issues"`                   // 是否启用问题跟踪
	HasProjects   bool           `json:"has_projects"`                 // 是否启用项目
	HasWiki       bool           `json:"has_wiki"`                     // 是否启用 Wiki
	HasPages      bool           `json:"has_pages"`                    // 是否启用 Pages
	HasDownloads  bool           `json:"has_downloads"`                // 是否启用下载
	IsTemplate    bool           `json:"is_template"`                  // 是否是模板仓库
	// 新增字段
	LastAnalyzedAt *time.Time `json:"last_analyzed_at"` // 最后分析时间
	AnalysisStatus string     `json:"analysis_status"`  // 分析状态
	SearchKeyword  string     `json:"search_keyword"`   // 搜索关键词
	SearchRank     int        `json:"search_rank"`      // 搜索结果排名
	LastCrawledAt  time.Time  `json:"last_crawled_at"`  // 最后爬取时间
}

type CrawlHistory struct {
	ID             uint      `json:"id" gorm:"primarykey;tableName:crawl_history"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Keyword        string    `json:"keyword"`         // 搜索关键词
	StartedAt      time.Time `json:"started_at"`      // 开始时间
	CompletedAt    time.Time `json:"completed_at"`    // 完成时间
	TotalRepos     int       `json:"total_repos"`     // 找到的仓库总数
	ProcessedRepos int       `json:"processed_repos"` // 已处理的仓库数
	Status         string    `json:"status"`          // 爬取状态
	ErrorMessage   string    `json:"error_message"`   // 错误信息
}
