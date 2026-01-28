package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// LarkMessageService 处理飞书消息相关功能
type LarkMessageService struct {
	BaseService
}

// NewLarkMessageService 创建一个新的 LarkMessageService 实例
func NewLarkMessageService(appID, appSecret string) *LarkMessageService {
	baseService := NewBaseService(appID, appSecret)
	return &LarkMessageService{
		BaseService: baseService,
	}
}

// SendMessage 发送消息到指定群聊
func (s *LarkMessageService) SendMessage(groupChatID, message string) error {
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

	// 输出发送消息的详细信息
	log.Printf("📤 准备发送消息到群聊 %s", groupChatID)
	log.Printf("📝 消息内容: %s", message)

	resp, err := s.client.Im.Message.Create(ctx, req)
	if err != nil {
		log.Printf("❌ 发送消息失败: %v", err)
		return fmt.Errorf("发送消息失败: %v", err)
	}

	if !resp.Success() {
		log.Printf("❌ 发送消息失败: %s (Code: %d)", resp.Msg, resp.Code)
		// 输出完整的响应信息以帮助诊断
		respBytes, _ := json.Marshal(resp)
		log.Printf("📋 完整响应: %s", string(respBytes))
		return fmt.Errorf("发送消息失败: %s (Code: %d)", resp.Msg, resp.Code)
	}

	// 输出发送成功的信息
	log.Printf("✅ 消息发送成功!")
	if resp.Data != nil && resp.Data.MessageId != nil && *resp.Data.MessageId != "" {
		log.Printf("📄 消息ID: %s", *resp.Data.MessageId)
	}

	return nil
}