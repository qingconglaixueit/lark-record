package services

import (
	"encoding/json"
	"fmt"
	"log"
)

// LarkTaskService 处理飞书任务相关功能
type LarkTaskService struct {
	BaseService
}

// NewLarkTaskService 创建一个新的LarkTaskService实例
func NewLarkTaskService(appID, appSecret string) *LarkTaskService {
	baseService := NewBaseService(appID, appSecret)
	return &LarkTaskService{
		BaseService: baseService,
	}
}

// getTenantAccessToken 获取租户访问令牌，使用BaseService中的统一实现
func (s *LarkTaskService) getTenantAccessToken() (string, error) {
	return s.GetTenantAccessToken()
}

// CreateTask 创建一个飞书任务
func (s *LarkTaskService) CreateTask(title string, dueTimestamp int64, isAllDay bool, assignees []map[string]interface{}) error {
	token, err := s.getTenantAccessToken()
	if err != nil {
		return fmt.Errorf("获取访问令牌失败: %w", err)
	}

	// 构建成员列表
	var members []map[string]interface{}
	for _, assignee := range assignees {
		if id, ok := assignee["id"].(string); ok {
			members = append(members, map[string]interface{}{
				"id":   id,
				"type": "user",
				"role": "assignee",
				"name": "",
			})
		}
	}

	if len(members) == 0 {
		return fmt.Errorf("没有有效的负责人ID")
	}

	// 构建请求体，使用用户提供的API格式
	reqBody := map[string]interface{}{
		"summary": title,
		"due": map[string]interface{}{
			"timestamp":  dueTimestamp,
			"is_all_day": isAllDay,
		},
		"members": members,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("请求体序列化失败: %w", err)
	}

	// 使用BaseService的handleHTTPRequest方法发送请求
	_, body, err := s.handleHTTPRequest(
		"POST",
		"https://open.feishu.cn/open-apis/task/v2/tasks?user_id_type=user_id",
		token,
		jsonData,
	)
	if err != nil {
		return fmt.Errorf("创建任务失败: %w", err)
	}

	// 解析响应
	type CreateTaskResponse struct {
		Code int `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Task struct {
				TaskID string `json:"task_id"`
				GUID   string `json:"guid"`
				URL    string `json:"url"`
			} `json:"task"`
		} `json:"data"`
	}

	var result CreateTaskResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 0 {
		log.Printf("📋 创建任务API响应: %s", string(body))
		return fmt.Errorf("创建任务失败: %s (Code: %d)", result.Msg, result.Code)
	}

	// 输出创建成功的信息
	log.Printf("✅ 任务创建成功! 任务ID: %s, 任务GUID: %s", result.Data.Task.TaskID, result.Data.Task.GUID)
	log.Printf("🔗 任务链接: %s", result.Data.Task.URL)

	return nil
}