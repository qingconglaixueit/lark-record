package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"lark-record/models"
)

// AIService AI服务
type AIService struct {
	config *models.SiliconFlowConfig
}

// NewAIService 创建AI服务实例
func NewAIService(config *models.SiliconFlowConfig) *AIService {
	return &AIService{
		config: config,
	}
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

	// 创建HTTP请求
	req, err := http.NewRequest("GET", "https://api.siliconflow.cn/v1/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.config.ApiKey))

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// 解析响应
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
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

// ParseWithAI 使用AI解析内容
func (s *AIService) ParseWithAI(content string, prompt string) (string, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		fmt.Printf("[AI解析] 总耗时: %v\n", elapsed)
	}()

	fmt.Printf("[AI解析] 输入内容: %s\n", content)
	fmt.Printf("[AI解析] 提示词: %s\n", prompt)

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

	// 创建HTTP请求
	req, err := http.NewRequest("POST", "https://api.siliconflow.cn/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.config.ApiKey))

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// 解析响应
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	fmt.Printf("[AI解析] API响应: %s\n", string(bodyBytes))

	var response SiliconFlowResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// 提取结果
	if len(response.Choices) > 0 {
		result := response.Choices[0].Message.Content
		fmt.Printf("[AI解析] 解析结果: %s\n", result)
		return result, nil
	}

	return "", fmt.Errorf("no response content")
}
