package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/HeisenbergV/repoinsight/api"
	"github.com/HeisenbergV/repoinsight/pkg/crawler"
	"github.com/HeisenbergV/repoinsight/pkg/logger"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
	} `yaml:"database"`
	API struct {
		Github struct {
			Token string `yaml:"token"`
		} `yaml:"github"`
		Deepseek struct {
			APIKey string `yaml:"api_key"`
		} `yaml:"deepseek"`
	} `yaml:"api"`
	App struct {
		SearchKeyword   string `yaml:"search_keyword"`
		IntervalHours   int    `yaml:"interval_hours"`
		MaxReposPerPage int    `yaml:"max_repos_per_page"`
		Port            int    `yaml:"port"`
	} `yaml:"app"`
	Log struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
		Output string `yaml:"output"`
		File   struct {
			Path       string `yaml:"path"`
			Name       string `yaml:"name"`
			MaxSize    int    `yaml:"max_size"`
			MaxBackups int    `yaml:"max_backups"`
			MaxAge     int    `yaml:"max_age"`
			Compress   bool   `yaml:"compress"`
		} `yaml:"file"`
	} `yaml:"log"`
}

func loadConfig() (*Config, error) {
	data, err := os.ReadFile("config.yml")
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 从环境变量读取 GitHub Token
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		config.API.Github.Token = token
	}

	if config.API.Github.Token == "" {
		return nil, fmt.Errorf("GitHub Token 未设置，请设置 GITHUB_TOKEN 环境变量或在配置文件中设置")
	}

	return &config, nil
}

func main() {
	// 加载配置
	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// 初始化日志
	logConfig := logger.Config{
		Level:  config.Log.Level,
		Format: config.Log.Format,
		Output: config.Log.Output,
		FileConfig: struct {
			Path       string `yaml:"path"`
			Name       string `yaml:"name"`
			MaxSize    int    `yaml:"max_size"`
			MaxBackups int    `yaml:"max_backups"`
			MaxAge     int    `yaml:"max_age"`
			Compress   bool   `yaml:"compress"`
		}{
			Path:       config.Log.File.Path,
			Name:       config.Log.File.Name,
			MaxSize:    config.Log.File.MaxSize,
			MaxBackups: config.Log.File.MaxBackups,
			MaxAge:     config.Log.File.MaxAge,
			Compress:   config.Log.File.Compress,
		},
	}
	if err := logger.Init(logConfig); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	// 连接数据库
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		config.Database.Host,
		config.Database.User,
		config.Database.Password,
		config.Database.Name,
		config.Database.Port,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 自动迁移数据库表
	if err := db.AutoMigrate(&api.Repository{}); err != nil {
		log.Fatalf("迁移数据库失败: %v", err)
	}

	// 创建爬虫配置
	crawlerConfig := &crawler.Config{
		Token:           config.API.Github.Token,
		SearchKeyword:   config.App.SearchKeyword,
		MaxReposPerPage: config.App.MaxReposPerPage,
	}

	// 创建爬虫实例
	c := crawler.NewCrawler(db, crawlerConfig)

	// 创建 API 处理器
	handler := api.NewHandler(db)

	// 设置路由
	router := api.SetupRouter(handler)

	// 创建等待组
	var wg sync.WaitGroup
	wg.Add(2)

	// 创建退出通道
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 创建服务器实例
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.App.Port),
		Handler: router,
	}

	// 创建爬虫定时器
	ticker := time.NewTicker(time.Duration(config.App.IntervalHours) * time.Hour)
	defer ticker.Stop()

	// 启动 API 服务
	go func() {
		defer wg.Done()
		log.Println("启动 API 服务...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("API 服务错误: %v", err)
		}
	}()

	// 启动爬虫服务
	go func() {
		defer wg.Done()
		log.Println("启动爬虫服务...")

		// 立即执行一次
		if err := c.Start(); err != nil {
			log.Printf("爬取失败: %v", err)
		}

		for {
			select {
			case <-ticker.C:
				if err := c.Start(); err != nil {
					log.Printf("爬取失败: %v", err)
				}
			case <-quit:
				log.Println("爬虫服务正在关闭...")
				return
			}
		}
	}()

	// 等待退出信号
	<-quit
	log.Println("正在关闭服务...")

	// 停止爬虫定时器
	ticker.Stop()

	// 设置关闭超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭 HTTP 服务器
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("服务器强制关闭: %v", err)
	}

	// 等待所有服务关闭
	wg.Wait()
	log.Println("服务已关闭")
}
