package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	"gorm.io/gorm/schema"
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

func initConfig() (*Config, error) {
	data, err := os.ReadFile("config.yml")
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	if config.API.Github.Token == "" {
		return nil, fmt.Errorf("github token 未设置，请设置 GITHUB_TOKEN 环境变量或在配置文件中设置")
	}

	if config.API.Deepseek.APIKey == "" {
		return nil, fmt.Errorf("deepseek API key 未设置，请设置 DEEPSEEK_API_KEY 环境变量或在配置文件中设置")
	}

	return &config, nil
}

func initDB(config *Config) (*gorm.DB, error) {
	var baseDB *gorm.DB

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
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, // 使用单数表名
			},
			PrepareStmt: true, // 启用预处理语句缓存
		})
		if err == nil {
			baseDB = db
			break
		}
		fmt.Printf("数据库连接失败，正在重试 (%d/%d): %v\n", i+1, maxRetries, err)
		time.Sleep(retryInterval)
	}

	if baseDB == nil {
		return nil, fmt.Errorf("连接数据库失败")
	}

	// 配置连接池
	sqlDB, err := baseDB.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %v", err)
	}

	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetMaxOpenConns(200)
	sqlDB.SetConnMaxLifetime(24 * time.Hour)
	sqlDB.SetConnMaxIdleTime(12 * time.Hour)

	// 读取 schema.sql
	schemaSQL, err := os.ReadFile("schema.sql")
	if err != nil {
		return nil, fmt.Errorf("读取 schema.sql 失败: %v", err)
	}

	// 单独提取 CREATE FUNCTION 语句
	createFuncStart := strings.Index(string(schemaSQL), "CREATE OR REPLACE FUNCTION")
	createFuncEnd := strings.Index(string(schemaSQL), "$$ language 'plpgsql';")
	if createFuncStart != -1 && createFuncEnd != -1 {
		createFuncEnd += len("$$ language 'plpgsql';")
		createFunc := string(schemaSQL)[createFuncStart:createFuncEnd]
		if err := baseDB.Exec(createFunc).Error; err != nil {
			return nil, fmt.Errorf("执行函数创建失败: %v", err)
		}
		// 去掉 function 语句部分
		schemaSQL = append(schemaSQL[:createFuncStart], schemaSQL[createFuncEnd:]...)
	}

	// 其余 SQL 语句分号分割后逐条执行
	stmts := strings.Split(string(schemaSQL), ";")
	for _, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if err := baseDB.Exec(stmt).Error; err != nil {
			return nil, fmt.Errorf("执行数据库迁移失败: %v\nSQL: %s", err, stmt)
		}
	}

	fmt.Println("数据库迁移完成")

	return baseDB, nil
}

func initLog(config *Config) error {
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
		return fmt.Errorf("初始化日志失败: %v", err)
	}
	return nil
}

func main() {
	config, err := initConfig()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	if err := initLog(config); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}

	db, err := initDB(config)
	if err != nil {
		fmt.Printf("连接数据库失败: %v\n", err)
		os.Exit(1)
	}

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

	crawler := crawler.NewCrawler(db, crawlerConfig)
	aiAnalyzer := ai.NewAnalyzer(db, analyzerConfig)
	handler := api.NewHandler(db)

	router := api.SetupRouter(handler)

	var wg sync.WaitGroup
	wg.Add(3)

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
		if err := crawler.Start(); err != nil {
			fmt.Printf("爬取失败: %v\n", err)
		}

		for {
			select {
			case <-crawlerTicker.C:
				fmt.Printf("开始新一轮爬取...\n")
				if err := crawler.Start(); err != nil {
					fmt.Printf("爬取失败: %v\n", err)
				}
			case <-quit:

				fmt.Printf("爬虫服务正在关闭...\n")
				return
			}
		}
	}()

	// 启动 AI 分析器
	go func() {
		defer wg.Done()
		fmt.Printf("启动 AI 分析服务...\n")
		if err := aiAnalyzer.Start(); err != nil {
			fmt.Printf("启动 AI 分析器失败: %v\n", err)
		}
		<-quit
		fmt.Printf("AI 分析服务正在关闭...\n")
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

	// 关闭数据库连接
	sqlDB, err := db.DB()
	if err != nil {
		fmt.Printf("获取数据库连接失败: %v\n", err)
	} else {
		if err := sqlDB.Close(); err != nil {
			fmt.Printf("关闭数据库连接失败: %v\n", err)
		}
	}

	// 等待所有服务完成
	wg.Wait()
	fmt.Printf("所有服务已关闭\n")
}
