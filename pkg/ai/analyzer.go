package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/HeisenbergV/repoinsight/api"
	"github.com/HeisenbergV/repoinsight/pkg/logger"
	"github.com/gin-gonic/gin"
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
	return &Analyzer{
		db:     db,
		config: config,
	}
}

func (a *Analyzer) Start() error {
	logger.Info("启动 AI 分析服务...")
	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := a.processUnanalyzedRepositories(); err != nil {
				logger.Errorf("处理未分析的仓库失败: %v", err)
			}
		}
	}
}

func (a *Analyzer) processUnanalyzedRepositories() error {
	var repositories []api.Repository

	// 查找 ai_analysis 为空的仓库
	result := a.db.Where("ai_analysis = '' OR ai_analysis IS NULL").
		Limit(a.config.BatchSize).
		Find(&repositories)

	if result.Error != nil {
		return fmt.Errorf("查询未分析的仓库失败: %v", result.Error)
	}

	logger.Infof("找到 %d 个未分析的仓库", len(repositories))

	for _, repo := range repositories {
		logger.Infof("正在分析仓库: %s", repo.FullName)

		analysis, err := a.analyzeRepository(&repo)
		if err != nil {
			logger.Errorf("分析仓库 %s 失败: %v", repo.FullName, err)
			continue
		}

		// 更新仓库的 AI 分析结果
		if err := a.db.Model(&repo).Update("ai_analysis", analysis).Error; err != nil {
			logger.Errorf("更新仓库 %s 的分析结果失败: %v", repo.FullName, err)
			continue
		}

		logger.Infof("成功分析仓库: %s", repo.FullName)

		// 避免请求过于频繁
		time.Sleep(2 * time.Second)
	}

	return nil
}

func (a *Analyzer) analyzeRepository(repo *api.Repository) (string, error) {
	prompt := fmt.Sprintf(`请用中文分析以下 GitHub 仓库，生成一段 100-500 字的总结，说明这个项目是做什么的，有什么用途和特点。
仓库名称：%s
仓库描述：%s
主要语言：%s
README 内容：
%s

1. **一句话类比**：像"XXX 领域的 Uber"或"XX 界的 ChatGPT"。
2. **解决什么问题**：普通人能遇到的实际问题（例如"帮你自动整理手机照片"）。
3. **谁会用这个**：目标用户（如"适合经常写文档的上班族"）。
4. **不用它的麻烦**：对比手动操作的缺点（如"否则需要手动复制粘贴 100 次"）`, repo.FullName, repo.Description, repo.Language, repo.Readme)

	request := DeepseekRequest{
		Model: "deepseek-chat",
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %v", err)
	}

	req, err := http.NewRequest("POST", a.config.APIBaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.config.APIKey))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API 请求失败，状态码: %d", resp.StatusCode)
	}

	var response DeepseekResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("API 响应中没有内容")
	}

	return response.Choices[0].Message.Content, nil
}

// AnalyzeRepository 分析单个仓库信息并生成摘要
func (a *Analyzer) AnalyzeRepository(c *gin.Context, repoName, description, readme string) (string, error) {
	logger := logger.WithFields(map[string]interface{}{
		"service": "ai_analysis",
		"repo":    repoName,
	})

	// 确保上下文存在
	if c == nil || c.Request == nil {
		return "", fmt.Errorf("无效的上下文")
	}

	// 创建一个带有取消的上下文
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	fmt.Printf("\n[AI分析] 开始处理项目: %s\n", repoName)
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
README 内容：%s`, repoName, description, readme)

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
	req, err := http.NewRequestWithContext(ctx, "POST", a.config.APIBaseURL, bytes.NewBuffer(jsonData))
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
	client := &http.Client{}

	// 记录开始时间
	startTime := time.Now()

	// 发送请求
	fmt.Printf("[AI分析] 正在等待 AI 响应...\n")
	resp, err := client.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			fmt.Printf("[AI分析] 请求被取消\n")
			logger.Info("请求被取消")
			return "", context.Canceled
		default:
			fmt.Printf("[AI分析] 错误: 发送请求到 Deepseek API 失败: %v\n", err)
			logger.WithError(err).Error("发送请求到 Deepseek API 失败")
			return "", fmt.Errorf("发送请求到 Deepseek API 失败: %v", err)
		}
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
