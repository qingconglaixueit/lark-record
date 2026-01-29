package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"lark-record/models"
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

// CreateTaskFromFieldValues 从字段值创建任务
func (s *LarkTaskService) CreateTaskFromFieldValues(tableConfig models.TableConfig, fieldValues map[string]interface{}) error {
	// 获取任务配置
	taskConfig := tableConfig.Task

	// 检查是否启用任务创建
	if !taskConfig.Enabled {
		// 检查旧版本配置兼容性
		if !tableConfig.CreateTask {
			return nil
		}
		// 使用旧版本配置
		return s.createTaskFromOldConfig(tableConfig, fieldValues)
	}

	// 提取任务信息
	taskTitle, dueTimestamp, isAllDay, assignees, err := s.extractTaskInfo(taskConfig, fieldValues)
	if err != nil {
		return err
	}

	// 创建任务
	return s.CreateTask(taskTitle, dueTimestamp, isAllDay, assignees)
}

// createTaskFromOldConfig 从旧版本配置创建任务（向后兼容）
func (s *LarkTaskService) createTaskFromOldConfig(tableConfig models.TableConfig, fieldValues map[string]interface{}) error {
	// 构建临时任务配置
	taskConfig := models.TaskConfig{
		Enabled:        true,
		SummaryField:   tableConfig.TaskSummaryField,
		DueField:       tableConfig.TaskDueField,
		AssigneeField:  tableConfig.TaskAssigneeField,
		DefaultSummary: "来自多维表格的任务",
		DefaultDueDays: 1,
	}

	// 提取任务信息
	taskTitle, dueTimestamp, isAllDay, assignees, err := s.extractTaskInfo(taskConfig, fieldValues)
	if err != nil {
		return err
	}

	// 创建任务
	return s.CreateTask(taskTitle, dueTimestamp, isAllDay, assignees)
}

// extractTaskInfo 从字段值中提取任务信息
func (s *LarkTaskService) extractTaskInfo(taskConfig models.TaskConfig, fieldValues map[string]interface{}) (string, int64, bool, []map[string]interface{}, error) {
	// 提取任务标题
	taskTitle := s.extractFieldValue(fieldValues, taskConfig.SummaryField)
	if taskTitle == "" {
		taskTitle = taskConfig.DefaultSummary
		if taskTitle == "" {
			taskTitle = "来自多维表格的任务"
		}
	}

	// 提取任务截止时间
	dueTimestamp := s.extractDueTimestamp(fieldValues, taskConfig.DueField, taskConfig.DefaultDueDays)

	// 提取任务负责人
	assignees := s.extractAssignees(fieldValues, taskConfig.AssigneeField)
	if len(assignees) == 0 {
		// 尝试自动查找用户字段
		assignees = s.findUserFields(fieldValues)
	}

	if len(assignees) == 0 {
		return "", 0, false, nil, fmt.Errorf("未找到任务负责人信息")
	}

	return taskTitle, dueTimestamp, true, assignees, nil
}

// extractFieldValue 从字段值中提取单个字段的值
func (s *LarkTaskService) extractFieldValue(fieldValues map[string]interface{}, fieldName string) string {
	if fieldName == "" {
		return ""
	}

	value, exists := fieldValues[fieldName]
	if !exists {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%v", v)
	case []interface{}:
		// 处理数组类型的值
		var result string
		for i, item := range v {
			if i > 0 {
				result += ", "
			}
			result += fmt.Sprintf("%v", item)
		}
		return result
	default:
		return fmt.Sprintf("%v", v)
	}
}

// extractDueTimestamp 从字段值中提取任务截止时间
func (s *LarkTaskService) extractDueTimestamp(fieldValues map[string]interface{}, fieldName string, defaultDueDays int) int64 {
	if fieldName == "" {
		// 使用默认截止时间
		return time.Now().Add(time.Duration(defaultDueDays) * 24 * time.Hour).UnixMilli()
	}

	value, exists := fieldValues[fieldName]
	if !exists {
		// 使用默认截止时间
		return time.Now().Add(time.Duration(defaultDueDays) * 24 * time.Hour).UnixMilli()
	}

	// 处理时间戳，支持int64和float64两种类型
	var timestamp int64
	switch v := value.(type) {
	case int64:
		timestamp = v
	case float64:
		timestamp = int64(v)
	default:
		// 使用默认截止时间
		return time.Now().Add(time.Duration(defaultDueDays) * 24 * time.Hour).UnixMilli()
	}

	// 检查时间戳是否有效（大于0且小于2100年的毫秒时间戳）
	if timestamp > 0 && timestamp < 3250368000000 {
		return timestamp
	}

	// 使用默认截止时间
	return time.Now().Add(time.Duration(defaultDueDays) * 24 * time.Hour).UnixMilli()
}

// extractAssignees 从字段值中提取任务负责人
func (s *LarkTaskService) extractAssignees(fieldValues map[string]interface{}, fieldName string) []map[string]interface{} {
	if fieldName == "" {
		return nil
	}

	value, exists := fieldValues[fieldName]
	if !exists {
		return nil
	}

	var assignees []map[string]interface{}

	// 处理单个用户
	if userMap, ok := value.(map[string]interface{}); ok {
		if id, ok := userMap["id"].(string); ok {
			assignees = append(assignees, map[string]interface{}{
				"id": id,
			})
		}
	} else if userArray, ok := value.([]interface{}); ok {
		// 处理用户数组
		for _, userItem := range userArray {
			if userMap, ok := userItem.(map[string]interface{}); ok {
				if id, ok := userMap["id"].(string); ok {
					assignees = append(assignees, map[string]interface{}{
						"id": id,
					})
				}
			}
		}
	}

	return assignees
}

// findUserFields 自动查找用户类型的字段
func (s *LarkTaskService) findUserFields(fieldValues map[string]interface{}) []map[string]interface{} {
	for _, value := range fieldValues {
		// 处理单个用户
		if userMap, ok := value.(map[string]interface{}); ok {
			if id, ok := userMap["id"].(string); ok {
				return []map[string]interface{}{{
					"id": id,
				}}
			}
		} else if userArray, ok := value.([]interface{}); ok {
			// 处理用户数组
			for _, userItem := range userArray {
				if userMap, ok := userItem.(map[string]interface{}); ok {
					if id, ok := userMap["id"].(string); ok {
						return []map[string]interface{}{{
							"id": id,
						}}
					}
				}
			}
		}
	}

	return nil
}