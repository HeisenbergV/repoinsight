package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/HeisenbergV/repoinsight/api"
	"github.com/HeisenbergV/repoinsight/pkg/ai"
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
			APIKey   string `yaml:"api_key"`
			BaseURL  string `yaml:"base_url"`
			Interval int    `yaml:"interval"`
		} `yaml:"deepseek"`
		Wechat struct {
			AppID        string `yaml:"app_id"`
			AppSecret    string `yaml:"app_secret"`
			TemplateID   string `yaml:"template_id"`
			PushInterval int    `yaml:"push_interval"`
		} `yaml:"wechat"`
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

	// 从环境变量读取 Deepseek API Key
	if apiKey := os.Getenv("DEEPSEEK_API_KEY"); apiKey != "" {
		config.API.Deepseek.APIKey = apiKey
	}

	if config.API.Github.Token == "" {
		return nil, fmt.Errorf("GitHub Token 未设置，请设置 GITHUB_TOKEN 环境变量或在配置文件中设置")
	}

	if config.API.Deepseek.APIKey == "" {
		return nil, fmt.Errorf("Deepseek API Key 未设置，请设置 DEEPSEEK_API_KEY 环境变量或在配置文件中设置")
	}

	return &config, nil
}

func main() {
	// 加载配置
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
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
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("正在连接数据库...\n")
	// 连接数据库，添加重试机制
	var db *gorm.DB
	maxRetries := 10
	retryInterval := 5 * time.Second

	for i := 0; i < maxRetries; i++ {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			config.Database.Host,
			config.Database.User,
			config.Database.Password,
			config.Database.Name,
			config.Database.Port,
		)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		fmt.Printf("数据库连接失败，正在重试 (%d/%d): %v\n", i+1, maxRetries, err)
		time.Sleep(retryInterval)
	}

	if err != nil {
		fmt.Printf("连接数据库失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("数据库连接成功\n")

	// 自动迁移数据库表
	fmt.Printf("正在迁移数据库表...\n")
	if err := db.AutoMigrate(&api.Repository{}); err != nil {
		fmt.Printf("迁移数据库失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("数据库表迁移完成\n")

	// 创建爬虫配置
	crawlerConfig := &crawler.Config{
		Token:           config.API.Github.Token,
		SearchKeyword:   config.App.SearchKeyword,
		MaxReposPerPage: config.App.MaxReposPerPage,
	}

	// 创建 AI 分析器配置
	analyzerConfig := &ai.Config{
		APIKey:     config.API.Deepseek.APIKey,
		APIBaseURL: config.API.Deepseek.BaseURL,
		BatchSize:  10,
		Interval:   time.Duration(config.API.Deepseek.Interval) * time.Minute,
	}

	// 创建爬虫实例
	c := crawler.NewCrawler(db, crawlerConfig)

	// 创建 AI 分析器实例
	a := ai.NewAnalyzer(db, analyzerConfig)

	// 启动 AI 分析器
	go func() {
		fmt.Printf("启动 AI 分析服务...\n")
		if err := a.Start(); err != nil {
			fmt.Printf("启动 AI 分析器失败: %v\n", err)
		}
	}()

	// 创建 API 处理器
	handler := api.NewHandler(db)

	// 设置路由
	router := api.SetupRouter(handler)
	// 创建等待组
	var wg sync.WaitGroup
	wg.Add(4) // 增加到 4，因为现在有四个服务

	// 创建退出通道
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 创建服务器实例
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.App.Port),
		Handler: router,
	}

	// 创建爬虫定时器
	crawlerTicker := time.NewTicker(time.Duration(config.App.IntervalHours) * time.Hour)
	defer crawlerTicker.Stop()

	// 启动 API 服务
	go func() {
		defer wg.Done()
		fmt.Printf("启动 API 服务...\n")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("API 服务错误: %v\n", err)
		}
	}()

	// 启动爬虫服务
	go func() {
		defer wg.Done()
		fmt.Printf("启动爬虫服务...\n")

		// 立即执行一次
		if err := c.Start(); err != nil {
			fmt.Printf("爬取失败: %v\n", err)
		}

		for {
			select {
			case <-crawlerTicker.C:
				fmt.Printf("开始新一轮爬取...\n")
				if err := c.Start(); err != nil {
					fmt.Printf("爬取失败: %v\n", err)
				}
			case <-quit:
				fmt.Printf("爬虫服务正在关闭...\n")
				return
			}
		}
	}()

	// 等待退出信号
	<-quit
	fmt.Printf("正在关闭服务...\n")

	// 创建一个带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 优雅地关闭 HTTP 服务器
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("服务器关闭出错: %v\n", err)
	}

	// 等待所有服务完成
	wg.Wait()
	fmt.Printf("所有服务已关闭\n")
}
