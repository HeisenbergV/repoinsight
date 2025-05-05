package models

import (
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	FullName      string         `gorm:"size:255;uniqueIndex" json:"full_name"`
	Name          string         `gorm:"size:255" json:"name"`
	Owner         string         `gorm:"size:255" json:"owner"`
	Description   string         `gorm:"type:text" json:"description"`
	URL           string         `gorm:"size:255" json:"url"`
	Stars         int            `json:"stars"`
	Forks         int            `json:"forks"`
	Language      string         `gorm:"size:50" json:"language"`
	Topics        string         `gorm:"type:text" json:"topics"`
	Readme        string         `gorm:"type:text" json:"readme"`
	AIAnalysis    string         `gorm:"type:text" json:"ai_analysis"`
	LastPushedAt  *time.Time     `json:"last_pushed_at"`
	IsArchived    bool           `json:"is_archived"`
	License       string         `gorm:"size:100" json:"license"`
	DefaultBranch string         `gorm:"size:100" json:"default_branch"`
	OpenIssues    int            `json:"open_issues"`
	Watchers      int            `json:"watchers"`
	Size          int            `json:"size"`
	HasIssues     bool           `json:"has_issues"`
	HasProjects   bool           `json:"has_projects"`
	HasWiki       bool           `json:"has_wiki"`
	HasPages      bool           `json:"has_pages"`
	HasDownloads  bool           `json:"has_downloads"`
	IsTemplate    bool           `json:"is_template"`
}
