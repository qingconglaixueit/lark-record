package models

// WriteField 待写入字段配置
type WriteField struct {
	FieldName string `json:"field_name"` // 字段名
	Default   string `json:"default"`    // 默认值
	UiType    string `json:"ui_type"`    // 字段UI类型
}

// SiliconFlowConfig SiliconFlow API配置
type SiliconFlowConfig struct {
	ApiKey        string `json:"api_key"`        // API密钥
	Model         string `json:"model"`          // 使用的模型
	DefaultPrompt string `json:"default_prompt"` // 默认提示词
}

// AIParseConfig AI解析配置
type AIParseConfig struct {
	Enabled     bool     `json:"enabled"`      // 是否启用AI解析
	BaseField   []string `json:"base_field"`   // 基于的字段
	ResultField string   `json:"result_field"` // 结果字段
	Prompt      string   `json:"prompt"`       // 提示词
}

// Config 飞书配置
type Config struct {
	AppID       string            `json:"app_id"`
	AppSecret   string            `json:"app_secret"`
	Tables      []TableConfig     `json:"tables"`        // 多个表格配置
	GroupChatID string            `json:"group_chat_id"` // 消息发送群ID
	SiliconFlow SiliconFlowConfig `json:"silicon_flow"`  // SiliconFlow API配置

	// 向后兼容旧版本配置
	TableID     string       `json:"table_id,omitempty"`
	WriteFields []WriteField `json:"write_fields,omitempty"`
	CheckFields []string     `json:"check_fields,omitempty"`
}

// TaskConfig 任务配置
type TaskConfig struct {
	Enabled           bool   `json:"enabled"`           // 是否启用任务创建
	SummaryField      string `json:"summary_field"`      // 任务标题字段
	DueField          string `json:"due_field"`          // 任务截止日期字段
	AssigneeField     string `json:"assignee_field"`     // 任务负责人字段
	DefaultSummary    string `json:"default_summary"`    // 默认任务标题
	DefaultDueDays    int    `json:"default_due_days"`   // 默认截止天数
}

// TableConfig 单个表格的配置
type TableConfig struct {
	URL               string        `json:"url"`                 // 飞书多维表格URL
	AppToken          string        `json:"app_token"`           // 从URL解析的app_token
	TableID           string        `json:"table_id"`            // 数据表ID
	Name              string        `json:"name"`                // 表格名称
	WriteFields       []WriteField  `json:"write_fields"`        // 待写入的字段
	CheckFields       []string      `json:"check_fields"`        // 需要检测是否有值的字段
	Task              TaskConfig    `json:"task"`                // 任务配置
	AIParse           AIParseConfig `json:"ai_parse"`            // AI解析配置

	// 向后兼容旧版本配置
	CreateTask        bool   `json:"create_task,omitempty"`         // 是否创建任务
	TaskSummaryField  string `json:"task_summary_field,omitempty"`  // 任务标题字段
	TaskDueField      string `json:"task_due_field,omitempty"`      // 任务截止日期字段
	TaskAssigneeField string `json:"task_assignee_field,omitempty"` // 任务负责人字段
}

// Bitable 飞书多维表格
type Bitable struct {
	AppToken string `json:"app_token"`
	TableID  string `json:"table_id"`
	Name     string `json:"name"`
}

// TableInfo 数据表信息
type TableInfo struct {
	TableID string `json:"table_id"`
	Name    string `json:"name"`
}

// Field 表格字段
type Field struct {
	FieldName string `json:"field_name"`
	FieldType string `json:"field_type"`
	FieldID   string `json:"field_id"`
	IsPrimary bool   `json:"is_primary"`
	UiType    string `json:"ui_type"`
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