package services

import (
	"context"
	"encoding/json"
	"fmt"
	"lark-record/models"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type LarkService struct {
	appID     string
	appSecret string
	client    *lark.Client
}

func NewLarkService(appID, appSecret string) *LarkService {
	return &LarkService{
		appID:     appID,
		appSecret: appSecret,
	}
}

// ValidateCredentials 验证飞书应用凭证是否有效
func (s *LarkService) ValidateCredentials() error {
	s.initClient()
	ctx := context.Background()

	// 基本验证：检查 App ID 和 App Secret 格式
	if len(s.appID) < 10 {
		return fmt.Errorf("App ID 格式不正确")
	}
	if len(s.appSecret) < 10 {
		return fmt.Errorf("App Secret 格式不正确")
	}

	// 尝试简单的 API 调用来验证凭证
	// 使用获取用户信息的 API
	req := larkdrive.NewListFileReqBuilder().
		PageSize(1).
		Build()

	resp, err := s.client.Drive.File.List(ctx, req)
	if err != nil {
		// 网络错误
		return fmt.Errorf("无法连接到飞书API，请检查网络: %v", err)
	}

	// 检查是否是认证错误
	if resp.Code == 99991600 || resp.Code == 99991601 {
		return fmt.Errorf("App ID 或 App Secret 不正确")
	}

	// 如果返回权限错误，说明凭证有效但可能没有 Drive 权限
	if resp.Code == 99991663 {
		// 这不是凭证错误，只是没有文件，凭证应该是有效的
		return nil
	}

	// 其他错误，可能是权限问题，但凭证格式正确
	if !resp.Success() {
		// 只要不是认证错误，就认为凭证有效
		if resp.Code != 99991600 && resp.Code != 99991601 {
			// 凭证有效，但可能缺少某些权限
			fmt.Printf("凭证验证通过，但API返回: %s (Code: %d)\n", resp.Msg, resp.Code)
			return nil
		}
		return fmt.Errorf("凭证验证失败: %s (Code: %d)", resp.Msg, resp.Code)
	}

	return nil
}

// initClient 初始化飞书客户端
func (s *LarkService) initClient() {
	if s.client == nil {
		s.client = lark.NewClient(s.appID, s.appSecret)
	}
}

// GetBitables 获取用户有权限访问的所有多维表格
func (s *LarkService) GetBitables() ([]models.Bitable, error) {
	s.initClient()

	ctx := context.Background()

	// 获取文件列表，设置更大的页面大小并搜索多维表格
	req := larkdrive.NewListFileReqBuilder().
		PageSize(500).
		Build()

	resp, err := s.client.Drive.File.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("获取多维表格列表失败: %v", err)
	}

	if !resp.Success() {
		return nil, fmt.Errorf("获取多维表格列表失败: %s (Code: %d)", resp.Msg, resp.Code)
	}

	var bitables []models.Bitable
	if resp.Data != nil && resp.Data.Files != nil {
		fmt.Printf("获取到 %d 个文件\n", len(resp.Data.Files))
		for _, item := range resp.Data.Files {
			// 过滤出多维表格
			if item.Type != nil {
				name := "未知"
				if item.Name != nil {
					name = *item.Name
				}
				fmt.Printf("文件类型: %s, 名称: %s\n", *item.Type, name)
				if *item.Type == "bitable" {
					var appToken string
					if item.Token != nil {
						appToken = *item.Token
					}
					bitables = append(bitables, models.Bitable{
						AppToken: appToken,
						TableID:  "",
						Name:     name,
					})
					fmt.Printf("  → 找到多维表格: %s (Token: %s)\n", name, appToken)
				}
			}
		}
	} else {
		fmt.Println("未获取到任何文件")
	}

	if len(bitables) == 0 {
		return nil, fmt.Errorf("未找到多维表格\n\n请确保：\n1. 飞书应用已授予 drive:drive 和 drive:drive:readonly 权限\n2. 您的账号有权限访问至少一个多维表格\n3. 多维表格已在飞书中创建\n4. 多维表格在飞书云文档或个人空间中\n\n提示：如果多维表格在飞书群组或知识库中，可能需要先将多维表格添加到个人云文档或知识库根目录")
	}

	return bitables, nil
}

// GetBitableTables 获取多维表格中的所有数据表
func (s *LarkService) GetBitableTables(appToken string) ([]string, error) {
	s.initClient()

	ctx := context.Background()

	req := larkbitable.NewListAppTableReqBuilder().
		AppToken(appToken).
		Build()

	resp, err := s.client.Bitable.AppTable.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("获取数据表列表失败: %v", err)
	}

	if !resp.Success() {
		if resp.Code == 99991663 {
			return nil, fmt.Errorf("无法访问该多维表格\n\n可能的原因：\n1. 飞书应用没有访问该多维表格的权限\n2. 该多维表格在知识库中，需要管理员将应用添加到知识库\n3. App Token 不正确\n\n解决方法：\n• 对于知识库中的表格：请联系知识库管理员，将飞书应用添加为协作者\n• 或者：将该多维表格添加到您的个人云文档")
		}
		return nil, fmt.Errorf("获取数据表列表失败: %s (Code: %d)", resp.Msg, resp.Code)
	}

	var tableIDs []string
	if resp.Data != nil && resp.Data.Items != nil {
		for _, table := range resp.Data.Items {
			if table.TableId != nil {
				tableIDs = append(tableIDs, *table.TableId)
			}
		}
	}

	return tableIDs, nil
}

// GetTableFields 获取数据表的所有字段
func (s *LarkService) GetTableFields(appToken, tableID string) ([]models.Field, error) {
	s.initClient()

	ctx := context.Background()

	req := larkbitable.NewListAppTableFieldReqBuilder().
		AppToken(appToken).
		TableId(tableID).
		Build()

	resp, err := s.client.Bitable.AppTableField.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("获取字段列表失败: %v", err)
	}

	if !resp.Success() {
		return nil, fmt.Errorf("获取字段列表失败: %s", resp.Msg)
	}

	var fields []models.Field
	if resp.Data != nil && resp.Data.Items != nil {
		for _, field := range resp.Data.Items {
			var fieldName, fieldID string
			var fieldType int
			if field.FieldName != nil {
				fieldName = *field.FieldName
			}
			if field.Type != nil {
				fieldType = *field.Type
			}
			if field.FieldId != nil {
				fieldID = *field.FieldId
			}
			fields = append(fields, models.Field{
				FieldName: fieldName,
				FieldType: fmt.Sprintf("%d", fieldType),
				FieldID:   fieldID,
			})
		}
	}

	return fields, nil
}

// AddRecord 新增记录
func (s *LarkService) AddRecord(appToken, tableID string, fields map[string]interface{}) (string, error) {
	s.initClient()

	ctx := context.Background()

	// 构建记录数据
	record := larkbitable.NewAppTableRecordBuilder().
		Fields(fields).
		Build()

	req := larkbitable.NewCreateAppTableRecordReqBuilder().
		AppToken(appToken).
		TableId(tableID).
		AppTableRecord(record).
		Build()

	resp, err := s.client.Bitable.AppTableRecord.Create(ctx, req)
	if err != nil {
		return "", fmt.Errorf("新增记录失败: %v", err)
	}

	if !resp.Success() {
		return "", fmt.Errorf("新增记录失败: %s", resp.Msg)
	}

	if resp.Data != nil && resp.Data.Record != nil && resp.Data.Record.RecordId != nil {
		return *resp.Data.Record.RecordId, nil
	}

	return "", fmt.Errorf("新增记录失败: 未返回记录ID")
}

// GetRecord 获取记录详情
func (s *LarkService) GetRecord(appToken, tableID, recordID string) (map[string]interface{}, error) {
	s.initClient()

	ctx := context.Background()

	req := larkbitable.NewGetAppTableRecordReqBuilder().
		AppToken(appToken).
		TableId(tableID).
		RecordId(recordID).
		Build()

	resp, err := s.client.Bitable.AppTableRecord.Get(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("获取记录失败: %v", err)
	}

	if !resp.Success() {
		return nil, fmt.Errorf("获取记录失败: %s", resp.Msg)
	}

	if resp.Data != nil && resp.Data.Record != nil {
		return resp.Data.Record.Fields, nil
	}

	return nil, fmt.Errorf("获取记录失败: 未返回记录数据")
}

// CheckFieldsCompleted 检查指定字段是否都有值
func (s *LarkService) CheckFieldsCompleted(appToken, tableID, recordID string, checkFields []string) (bool, error) {
	fields, err := s.GetRecord(appToken, tableID, recordID)
	if err != nil {
		return false, err
	}

	for _, fieldName := range checkFields {
		if value, ok := fields[fieldName]; !ok || value == nil || value == "" {
			return false, nil
		}
	}

	return true, nil
}

// SendMessage 发送消息到群聊
func (s *LarkService) SendMessage(groupChatID, message string) error {
	s.initClient()

	ctx := context.Background()

	// 构建消息内容
	msgContent := map[string]string{
		"text": message,
	}
	msgContentBytes, _ := json.Marshal(msgContent)

	// 构建请求体
	body := larkim.NewCreateMessageReqBodyBuilder().
		ReceiveId(groupChatID).
		MsgType("text").
		Content(string(msgContentBytes)).
		Build()

	// 构建请求
	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType("chat_id").
		Body(body).
		Build()

	resp, err := s.client.Im.Message.Create(ctx, req)
	if err != nil {
		return fmt.Errorf("发送消息失败: %v", err)
	}

	if !resp.Success() {
		return fmt.Errorf("发送消息失败: %s", resp.Msg)
	}

	return nil
}