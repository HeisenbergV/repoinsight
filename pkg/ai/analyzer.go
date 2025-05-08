package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/HeisenbergV/repoinsight/pkg/logger"
	"github.com/HeisenbergV/repoinsight/pkg/models"
	"gorm.io/gorm"
)

type Analyzer struct {
	db     *gorm.DB
	config *Config
}

type Config struct {
	APIKey     string
	APIBaseURL string
	BatchSize  int
	Interval   time.Duration
	MaxRetries int
}

type DeepseekRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type DeepseekResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func NewAnalyzer(db *gorm.DB, config *Config) *Analyzer {
	if config.APIBaseURL == "" {
		config.APIBaseURL = "https://api.deepseek.com/v1/chat/completions"
	}
	if config.BatchSize == 0 {
		config.BatchSize = 10
	}
	if config.Interval == 0 {
		config.Interval = 5 * time.Minute
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	return &Analyzer{
		db:     db,
		config: config,
	}
}

func (a *Analyzer) Start() error {
	logger.Info("启动 AI 分析服务...")
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for range ticker.C {
		if err := a.processUnanalyzedRepositories(); err != nil {
			logger.Errorf("处理未分析的仓库失败: %v", err)
		}
	}
	return nil
}

func (a *Analyzer) processUnanalyzedRepositories() error {
	var repositories []models.Repository

	// 查找需要分析的仓库
	result := a.db.Where("analysis_status = ? OR (analysis_status = ? AND updated_at > last_analyzed_at)",
		"pending", "failed").
		Limit(a.config.BatchSize).
		Find(&repositories)

	if result.Error != nil {
		return fmt.Errorf("查询未分析的仓库失败: %v", result.Error)
	}

	logger.Infof("找到 %d 个需要分析的仓库", len(repositories))

	for _, repo := range repositories {
		logger.Infof("正在分析仓库: %s", repo.FullName)

		// 更新状态为分析中
		if err := a.db.Model(&repo).Updates(map[string]interface{}{
			"analysis_status": "analyzing",
		}).Error; err != nil {
			logger.Errorf("更新仓库 %s 的状态失败: %v", repo.FullName, err)
			continue
		}

		// 分析仓库
		analysis, err := a.analyzeRepository(&repo)
		if err != nil {
			// 更新状态为失败
			a.db.Model(&repo).Updates(map[string]interface{}{
				"analysis_status": "failed",
			})
			logger.Errorf("分析仓库 %s 失败: %v", repo.FullName, err)
			continue
		}

		// 保存分析结果到 ai_analysis 表
		aiAnalysis := &models.AIAnalysis{
			URL:          repo.URL,
			Content:      analysis,
			Status:       "completed",
			ModelVersion: "deepseek-chat",
		}

		// 使用事务确保数据一致性
		err = a.db.Transaction(func(tx *gorm.DB) error {
			// 更新或创建分析结果
			if err := tx.Where("url = ?", repo.URL).
				Assign(aiAnalysis).
				FirstOrCreate(aiAnalysis).Error; err != nil {
				return err
			}

			// 更新仓库状态
			if err := tx.Model(&repo).Updates(map[string]interface{}{
				"analysis_status":  "completed",
				"last_analyzed_at": time.Now(),
			}).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			logger.Errorf("保存仓库 %s 的分析结果失败: %v", repo.FullName, err)
			continue
		}

		logger.Infof("成功分析仓库: %s", repo.FullName)

		// 避免请求过于频繁
		time.Sleep(2 * time.Second)
	}

	return nil
}

func (a *Analyzer) analyzeRepository(repo *models.Repository) (string, error) {
	logger := logger.WithFields(map[string]interface{}{
		"service": "ai_analysis",
		"repo":    repo.FullName,
	})

	fmt.Printf("\n[AI分析] 开始处理项目: %s\n", repo.FullName)
	logger.Info("开始分析仓库")

	// 构建提示词
	prompt := fmt.Sprintf(`
	请用通俗语言解释以下 GitHub 项目的作用，避免技术术语。按此结构输出 Markdown：
1. **一句话类比**：像"XXX 领域的 Uber"或"XX 界的 ChatGPT"。
2. **解决什么问题**：普通人能遇到的实际问题（例如"帮你自动整理手机照片"）。
3. **谁会用这个**：目标用户（如"适合经常写文档的上班族"）。
4. **不用它的麻烦**：对比手动操作的缺点（如"否则需要手动复制粘贴 100 次"）。

项目名称：%s
项目描述：%s
README 内容：%s`, repo.FullName, repo.Description, repo.Readme)

	fmt.Printf("[AI分析] 正在生成分析提示词...\n")
	logger.WithField("prompt_length", len(prompt)).Debug("构建提示词完成")

	// 准备请求体
	requestBody := DeepseekRequest{
		Model: "deepseek-chat",
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Printf("[AI分析] 错误: 序列化请求体失败: %v\n", err)
		logger.WithError(err).Error("序列化请求体失败")
		return "", fmt.Errorf("序列化请求体失败: %v", err)
	}

	fmt.Printf("[AI分析] 正在发送请求到 Deepseek API...\n")
	logger.Debug("准备发送请求到 Deepseek API")

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", a.config.APIBaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("[AI分析] 错误: 创建 HTTP 请求失败: %v\n", err)
		logger.WithError(err).Error("创建 HTTP 请求失败")
		return "", fmt.Errorf("创建 HTTP 请求失败: %v", err)
	}

	// 设置请求头
	if a.config.APIKey == "" {
		fmt.Printf("[AI分析] 错误: 未设置 DEEPSEEK_API_KEY 环境变量\n")
		logger.Error("未设置 DEEPSEEK_API_KEY 环境变量")
		return "", fmt.Errorf("未设置 DEEPSEEK_API_KEY 环境变量")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.config.APIKey)

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 记录开始时间
	startTime := time.Now()

	// 发送请求，支持重试
	var resp *http.Response
	var lastErr error
	for i := 0; i < a.config.MaxRetries; i++ {
		resp, lastErr = client.Do(req)
		if lastErr == nil {
			break
		}
		logger.Warnf("第 %d 次请求失败: %v, 准备重试...", i+1, lastErr)
		time.Sleep(time.Second * time.Duration(i+1))
	}

	if lastErr != nil {
		fmt.Printf("[AI分析] 错误: 发送请求到 Deepseek API 失败: %v\n", lastErr)
		logger.WithError(lastErr).Error("发送请求到 Deepseek API 失败")
		return "", fmt.Errorf("发送请求到 Deepseek API 失败: %v", lastErr)
	}
	defer resp.Body.Close()

	requestDuration := time.Since(startTime)
	fmt.Printf("[AI分析] 收到响应 (耗时: %v)\n", requestDuration)
	logger.WithFields(map[string]interface{}{
		"status_code": resp.StatusCode,
		"duration":    requestDuration,
	}).Info("收到 Deepseek API 响应")

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[AI分析] 错误: 读取响应体失败: %v\n", err)
		logger.WithError(err).Error("读取响应体失败")
		return "", fmt.Errorf("读取响应体失败: %v", err)
	}

	logger.WithField("response_length", len(body)).Debug("读取响应体完成")

	// 解析响应
	var deepseekResp DeepseekResponse
	if err := json.Unmarshal(body, &deepseekResp); err != nil {
		fmt.Printf("[AI分析] 错误: 解析响应失败: %v\n", err)
		logger.WithError(err).Error("解析响应失败")
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if len(deepseekResp.Choices) == 0 {
		fmt.Printf("[AI分析] 错误: API 响应中没有生成内容\n")
		logger.Error("API 响应中没有生成内容")
		return "", fmt.Errorf("API 响应中没有生成内容")
	}

	analysis := deepseekResp.Choices[0].Message.Content
	fmt.Printf("\n[AI分析] 分析结果:\n%s\n", analysis)
	logger.WithField("analysis_length", len(analysis)).Info("成功生成分析结果")

	return analysis, nil
}
