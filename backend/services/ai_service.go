package services

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"lark-record/models"
)

// 常量定义
const (
	// AI解析缓存有效期
	AIParseCacheExpireTime = 1 * time.Hour  // AI解析结果缓存有效期
)

// 定期清理过期缓存的函数
func (s *AIService) cleanExpiredCache() {
	for {
		// 每10分钟清理一次缓存
		time.Sleep(10 * time.Minute)
		
		now := time.Now()
		
		// 清理parseCache
		s.parseCacheTime.Range(func(key, value interface{}) bool {
			if now.After(value.(time.Time)) {
				s.parseCache.Delete(key)
				s.parseCacheTime.Delete(key)
			}
			return true
		})
	}
}

// AIService AI服务
type AIService struct {
	BaseService
	config *models.SiliconFlowConfig
	// AI解析结果缓存
	parseCache     sync.Map
	parseCacheTime sync.Map
}

// NewAIService 创建AI服务实例
func NewAIService(config *models.SiliconFlowConfig) *AIService {
	// 创建基础服务实例
	baseService := NewBaseService("", "")
	
	// 创建服务实例
	aiService := &AIService{
		BaseService:    baseService,
		config:        config,
		parseCache:    sync.Map{},
		parseCacheTime: sync.Map{},
	}
	
	// 启动定期清理缓存的goroutine
	go aiService.cleanExpiredCache()
	
	return aiService
}

// SiliconFlowRequest SiliconFlow API请求
type SiliconFlowRequest struct {
	Model    string        `json:"model"`
	Messages []interface{} `json:"messages"`
	Stream   bool          `json:"stream"`
}

// SiliconFlowResponse SiliconFlow API响应
type SiliconFlowResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// SiliconFlowModel SiliconFlow模型信息
type SiliconFlowModel struct {
	ID string `json:"id"`
}

// SiliconFlowModelsResponse SiliconFlow模型列表响应
type SiliconFlowModelsResponse struct {
	Data []SiliconFlowModel `json:"data"`
}

// GetModels 获取可用的AI模型列表
func (s *AIService) GetModels() ([]string, error) {
	if s.config.ApiKey == "" {
		return nil, fmt.Errorf("SiliconFlow API key not configured")
	}

	// 使用BaseService的handleHTTPRequest方法发送请求
	resp, bodyBytes, err := s.handleHTTPRequest("GET", "https://api.siliconflow.cn/v1/models", s.config.ApiKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response SiliconFlowModelsResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// 提取模型ID列表
	models := make([]string, 0, len(response.Data))
	for _, model := range response.Data {
		models = append(models, model.ID)
	}

	return models, nil
}

// getCacheKey 生成缓存键
func (s *AIService) getCacheKey(content string, prompt string) string {
	data := fmt.Sprintf("%s:%s:%s", s.config.Model, content, prompt)
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

// ParseWithAI 使用AI解析内容（带缓存和重试机制）
func (s *AIService) ParseWithAI(content string, prompt string) (string, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		logInfo("[AI解析] 总耗时: %v", elapsed)
	}()

	logInfo("[AI解析] 输入内容: %s", content)
	logInfo("[AI解析] 提示词: %s", prompt)

	if s.config.ApiKey == "" {
		return "", fmt.Errorf("SiliconFlow API key not configured")
	}

	// 如果没有提供提示词，使用默认提示词
	if prompt == "" {
		prompt = s.config.DefaultPrompt
	}

	// 处理提示词和内容的组合
	if strings.Contains(prompt, "{content}") {
		// 如果提示词中包含{content}占位符，替换它
		prompt = strings.Replace(prompt, "{content}", content, -1)
	} else {
		// 否则将内容放在提示词的后面
		prompt += "\n\n" + content
	}

	// 生成缓存键
	cacheKey := s.getCacheKey(content, prompt)
	
	// 检查缓存
	if cachedResult, ok := s.parseCache.Load(cacheKey); ok {
		if cachedTime, ok := s.parseCacheTime.Load(cacheKey); ok {
			// 缓存有效期
			if time.Since(cachedTime.(time.Time)) < AIParseCacheExpireTime {
				logInfo("[AI解析] 使用缓存结果")
				return cachedResult.(string), nil
			}
		}
	}

	// 构建请求体
	requestBody := SiliconFlowRequest{
		Model: s.config.Model,
		Messages: []interface{}{
			map[string]string{
				"role":    "system",
				"content": "You are a helpful AI assistant.",
			},
			map[string]string{
				"role":    "user",
				"content": prompt,
			},
		},
		Stream: false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	// 发送请求（带重试机制）
	var resp *http.Response
	var responseBody []byte
	retryDelay := InitialRetryDelay
	
	for i := 0; i < MaxRetries; i++ {
		// 使用BaseService的handleHTTPRequest方法发送请求
		resp, responseBody, err = s.handleHTTPRequest("POST", "https://api.siliconflow.cn/v1/chat/completions", s.config.ApiKey, jsonData)
		
		if err == nil && resp.StatusCode == http.StatusOK {
			break // 请求成功
		}
		
		if err != nil {
			logError("[AI解析] 请求失败: %v", err)
		} else {
			logError("[AI解析] 请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(responseBody))
		}
		
		if i < MaxRetries-1 {
			// 重试间隔：指数退避策略
			logInfo("[AI解析] %v后重试... (第%d/%d次)", retryDelay, i+2, MaxRetries)
			time.Sleep(retryDelay)
			retryDelay *= 2 // 指数退避
		}
	}

	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// 解析响应
	var response SiliconFlowResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// 提取解析结果
	if len(response.Choices) > 0 && len(response.Choices[0].Message.Content) > 0 {
		result := response.Choices[0].Message.Content
		
		// 缓存结果
		s.parseCache.Store(cacheKey, result)
		s.parseCacheTime.Store(cacheKey, time.Now())
		
		return result, nil
	}

	return "", fmt.Errorf("no valid response from AI service")
}