package models

// Config 飞书配置
type Config struct {
	AppID      string `json:"app_id"`
	AppSecret  string `json:"app_secret"`
	TableID    string `json:"table_id"`
	WriteFields []string `json:"write_fields"` // 待写入的字段
	CheckFields []string `json:"check_fields"` // 需要检测是否有值的字段
	GroupChatID string `json:"group_chat_id"`  // 消息发送群ID
}

// Bitable 飞书多维表格
type Bitable struct {
	AppToken string `json:"app_token"`
	TableID  string `json:"table_id"`
	Name     string `json:"name"`
}

// Field 表格字段
type Field struct {
	FieldName string `json:"field_name"`
	FieldType string `json:"field_type"`
	FieldID   string `json:"field_id"`
}

// Record 记录数据
type Record struct {
	Fields map[string]interface{} `json:"fields"`
}

// AddRecordRequest 新增记录请求
type AddRecordRequest struct {
	AppToken string                 `json:"app_token"`
	TableID  string                 `json:"table_id"`
	Fields   map[string]interface{} `json:"fields"`
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	GroupChatID string `json:"group_chat_id"`
	Message     string `json:"message"`
}