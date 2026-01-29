package handlers

import (
	"fmt"
	"lark-record/models"
	"lark-record/services"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// serviceManager 全局服务管理器
var serviceManager *services.ServiceManager

// SetServiceManager 设置服务管理器
func SetServiceManager(manager *services.ServiceManager) {
	serviceManager = manager
}

// configService 全局配置服务
var configService *services.ConfigService

// SetConfigService 设置配置服务
func SetConfigService(configSvc *services.ConfigService) {
	configService = configSvc
}

// min 辅助函数，返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 定义日志接口类型
type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

// logger 全局日志实例
var logger Logger

// SetLogger 设置日志实例
func SetLogger(log Logger) {
	logger = log
}

// 日志输出函数
func logInfo(format string, v ...interface{}) {
	if logger != nil {
		logger.Printf("[INFO] "+format, v...)
	} else {
		fmt.Printf("[INFO] "+format+"\n", v...)
	}
}

func logError(format string, v ...interface{}) {
	if logger != nil {
		logger.Printf("[ERROR] "+format, v...)
	} else {
		fmt.Printf("[ERROR] "+format+"\n", v...)
	}
}

// AIParseRequest AI解析请求
type AIParseRequest struct {
	Content        string `json:"content"`
	BaseFieldValue string `json:"base_field_value"`
	Prompt         string `json:"prompt"`
}

// AIParseResponse AI解析响应
type AIParseResponse struct {
	Result string `json:"result"`
}



// TestConfig 测试配置是否有效（不保存配置）
func TestConfig(c *gin.Context) {
	var testConfig models.Config
	if err := c.ShouldBindJSON(&testConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 测试配置是否有效 - 验证凭证
	larkService := services.NewLarkService(testConfig.AppID, testConfig.AppSecret)
	err := larkService.ValidateCredentials()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "飞书配置无效: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "配置有效！"})
}

// SaveConfig 保存配置
func SaveConfig(c *gin.Context) {
	var newConfig models.Config
	if err := c.ShouldBindJSON(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 测试配置是否有效 - 验证凭证
	larkService := services.NewLarkService(newConfig.AppID, newConfig.AppSecret)
	err := larkService.ValidateCredentials()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "飞书配置无效: " + err.Error()})
		return
	}

	// 使用配置服务更新配置
	if err := configService.SetConfig(&newConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存配置失败: " + err.Error()})
		return
	}

	// 获取更新后的配置
	config := configService.GetConfig()

	c.JSON(http.StatusOK, gin.H{"message": "配置保存成功", "config": config})
}

// GetConfig 获取配置
func GetConfig(c *gin.Context) {
	if configService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "配置服务未初始化"})
		return
	}

	config := configService.GetConfig()
	if config.AppID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "未配置"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// GetBitables 获取多维表格列表
func GetBitables(c *gin.Context) {
	if configService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "配置服务未初始化"})
		return
	}

	// 获取配置
	config := configService.GetConfig()

	if config.AppID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先配置飞书应用信息"})
		return
	}

	larkService := serviceManager.GetLarkService(config.AppID, config.AppSecret)
	bitables, err := larkService.GetBitables()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 确保返回空数组而不是null
	if bitables == nil {
		bitables = []models.Bitable{}
	}

	c.JSON(http.StatusOK, bitables)
}

// GetBitableTables 获取多维表格中的数据表列表
func GetBitableTables(c *gin.Context) {
	if configService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "配置服务未初始化"})
		return
	}

	// 获取配置
	config := configService.GetConfig()

	if config.AppID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先配置飞书应用信息"})
		return
	}

	appToken := c.Query("app_token")
	if appToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少app_token参数"})
		return
	}

	isWiki := c.Query("is_wiki") == "true"

	larkService := serviceManager.GetLarkService(config.AppID, config.AppSecret)
	tables, err := larkService.GetBitableTables(appToken, isWiki)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tables)
}

// GetTableFields 获取数据表的字段列表
func GetTableFields(c *gin.Context) {
	if configService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "配置服务未初始化"})
		return
	}

	// 获取配置
	config := configService.GetConfig()

	if config.AppID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先配置飞书应用信息"})
		return
	}

	appToken := c.Query("app_token")
	tableID := c.Query("table_id")

	if appToken == "" || tableID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必要参数"})
		return
	}

	larkService := serviceManager.GetLarkService(config.AppID, config.AppSecret)
	fields, err := larkService.GetTableFields(appToken, tableID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, fields)
}

// AddRecord 添加记录
func AddRecord(c *gin.Context) {
	if configService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "配置服务未初始化"})
		return
	}

	// 获取配置
	config := configService.GetConfig()

	if config.AppID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先配置飞书应用信息"})
		return
	}

	// 输出当前配置信息（用于调试）
	logInfo("当前配置信息：")
	logInfo("- AppID: %s", config.AppID)
	logInfo("- GroupChatID: %s", config.GroupChatID)
	logInfo("- Tables配置数量: %d", len(config.Tables))

	var req models.AddRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	larkService := serviceManager.GetLarkService(config.AppID, config.AppSecret)
	recordID, err := larkService.AddRecord(req.AppToken, req.TableID, req.Fields)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// TODO: 暂时关闭初始添加记录后的消息发送功能，只保留检测字段后的消息发送功能
	// if config.GroupChatID != "" {
	// 	go func() {
	// 		// 拼接字段值到消息中
	// 		message := fmt.Sprintf("✅ 记录已添加！\n\n记录ID: %s\n\n记录字段值：\n", recordID)
	// 		for fieldName, value := range req.Fields {
	// 			// 处理不同类型的值，确保消息格式清晰
	// 			switch v := value.(type) {
	// 			case string:
	// 				message += fmt.Sprintf("%s: %s\n", fieldName, v)
	// 			case []interface{}:
	// 				// 处理数组类型的值（如多选）
	// 	message += fmt.Sprintf("%s: ", fieldName)
	// 	for i, item := range v {
	// 		if i > 0 {
	// 		message += ", "
	// 	}
	// 	message += fmt.Sprintf("%v", item)
	// 	}
	// 	message += "\n"
	// default:
	// 	message += fmt.Sprintf("%s: %v\n", fieldName, v)
	// }
	// }
	// message += "\n🔍 系统将持续监测指定字段，完成后会发送通知。"
	// err = larkService.SendMessage(config.GroupChatID, message)
	// if err != nil {
	// 	fmt.Printf("发送初始消息失败: %v\n", err)
	// }
	// }()
	// }

	// 支持新的多表格配置和旧的单表格配置
	var checkFields []string
	var tableName string
	if len(config.Tables) > 0 {
		// 新格式：从对应的表格配置中获取检测字段和表格名称
		for _, table := range config.Tables {
			if table.AppToken == req.AppToken && table.TableID == req.TableID {
				checkFields = table.CheckFields
				tableName = table.Name
				break
			}
		}
	} else {
		// 旧格式：向后兼容
		checkFields = config.CheckFields
		tableName = "未命名表格"
	}

	// 持续检测指定字段是否有数据
	if checkFields != nil && len(checkFields) > 0 {
		go func() {
			fmt.Printf("🔍 开始检测记录ID %s 的字段: %v\n", recordID, checkFields)

			// 等待10秒后开始检测，避免立即检测可能出现的数据同步延迟
			time.Sleep(10 * time.Second)

			// 设置最大检测次数和基础间隔
			maxChecks := 20
			baseInterval := 10 * time.Second
			maxInterval := 5 * time.Minute
			checkCount := 0

			// 持续检测，直到所有指定字段都有数据或达到最大检测次数
			for checkCount < maxChecks {
				completed, fieldValues, err := larkService.CheckFieldsCompleted(req.AppToken, req.TableID, recordID, checkFields)
				if err != nil {
					fmt.Printf("❌ 检查字段状态失败: %v\n", err)
					
					// 检查是否是网络错误或飞书API错误，决定是否重试
					retry := strings.Contains(err.Error(), "network") || strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "API")
					if !retry {
						fmt.Printf("❌ 检查字段状态失败，错误不可重试，停止检测\n")
						break
					}
					
					// 等待一段时间后重试
					// 计算智能轮询间隔：基础间隔 * (2^min(checkCount, 6))，最大不超过maxInterval
					exponentialFactor := 1 << uint(min(checkCount, 6)) // 2的幂，最多64倍
					checkInterval := baseInterval * time.Duration(exponentialFactor)
					if checkInterval > maxInterval {
						checkInterval = maxInterval
					}
					time.Sleep(checkInterval)
					checkCount++
					continue
				}

				if completed {
					// 所有字段都有数据了，打印字段值
					fmt.Printf("✅ 记录ID %s 的指定字段已全部有数据！\n", recordID)
					fmt.Printf("📋 字段数据：\n")

					// 准备发送消息的内容，将表格名称放在第一行
					message := fmt.Sprintf("📊 表格：%s\n\n📢 记录ID %s 的指定字段已全部有数据！\n\n检测字段内容：\n", tableName, recordID)
					for fieldName, value := range fieldValues {
						// 处理不同类型的值，确保消息格式清晰
						switch v := value.(type) {
						case string:
							fmt.Printf("  - %s: %s\n", fieldName, v)
							message += fmt.Sprintf("%s: %s\n", fieldName, v)
						case []interface{}:
							// 处理数组类型的值（如多选）
							fmt.Printf("  - %s: ", fieldName)
							message += fmt.Sprintf("%s: ", fieldName)
							for i, item := range v {
								if i > 0 {
									fmt.Printf(", ")
									message += ", "
								}
								// 检查是否为用户类型
								if userMap, ok := item.(map[string]interface{}); ok {
									// 提取用户信息
									var userInfo string
									if enName, ok := userMap["en_name"].(string); ok && enName != "" {
										userInfo += fmt.Sprintf("en_name:%s", enName)
									}
									if id, ok := userMap["id"].(string); ok && id != "" {
										if userInfo != "" {
											userInfo += " "
										}
										userInfo += fmt.Sprintf("id:%s", id)
									}
									if name, ok := userMap["name"].(string); ok && name != "" {
										if userInfo != "" {
											userInfo += " "
										}
										userInfo += fmt.Sprintf("name:%s", name)
									}
									fmt.Printf("%s", userInfo)
									message += userInfo
								} else {
									fmt.Printf("%v", item)
									message += fmt.Sprintf("%v", item)
								}
							}
							fmt.Printf("\n")
							message += "\n"
						case float64:
							// 尝试将float64值作为时间戳处理
							// 飞书时间戳通常是毫秒级，且在合理的时间范围内（1970年至今）
							timestamp := int64(v)
							if timestamp > 0 && timestamp < 3250368000000 { // 小于2100年的毫秒时间戳
								// 转换为东八区时间
								timestampSec := timestamp / 1000 // 转换为秒级时间戳
								t := time.Unix(timestampSec, 0).In(time.FixedZone("Asia/Shanghai", 8*3600))
								timeStr := t.Format("2006-01-02 15:04:05")
								fmt.Printf("  - %s: %s\n", fieldName, timeStr)
								message += fmt.Sprintf("%s: %s\n", fieldName, timeStr)
							} else {
								// 普通数字类型
								fmt.Printf("  - %s: %v\n", fieldName, v)
								message += fmt.Sprintf("%s: %v\n", fieldName, v)
							}
						case map[string]interface{}:
							// 处理单个用户类型的值
							if (fieldName == "记录人" || strings.Contains(fieldName, "人")) || (v["id"] != nil && (v["name"] != nil || v["en_name"] != nil)) {
								// 提取用户信息
								var userInfo string
								if enName, ok := v["en_name"].(string); ok && enName != "" {
									userInfo += fmt.Sprintf("en_name:%s", enName)
								}
								if id, ok := v["id"].(string); ok && id != "" {
									if userInfo != "" {
										userInfo += " "
									}
									userInfo += fmt.Sprintf("id:%s", id)
								}
								if name, ok := v["name"].(string); ok && name != "" {
									if userInfo != "" {
										userInfo += " "
									}
									userInfo += fmt.Sprintf("name:%s", name)
								}
								if userInfo == "" {
									userInfo = "未知用户"
								}
								fmt.Printf("  - %s: %s\n", fieldName, userInfo)
								message += fmt.Sprintf("%s: %s\n", fieldName, userInfo)
							} else {
								// 其他复杂对象，简化显示
								fmt.Printf("  - %s: %v\n", fieldName, v)
								message += fmt.Sprintf("%s: [复杂对象]\n", fieldName)
							}
						default:
							fmt.Printf("  - %s: %v\n", fieldName, v)
							message += fmt.Sprintf("%s: %v\n", fieldName, v)
						}
					}

					// 发送消息
					if config.GroupChatID != "" {
						err = larkService.SendMessage(config.GroupChatID, message)
						if err != nil {
							fmt.Printf("❌ 发送消息失败: %v\n", err)
						} else {
							fmt.Printf("✅ 消息发送成功！\n")
						}
					}

					// 检查是否需要创建任务
			for _, table := range config.Tables {
				if table.AppToken == req.AppToken && table.TableID == req.TableID {
					// 使用异步方式创建任务，避免阻塞主线程
						go func(tableConfig models.TableConfig) {
							fmt.Printf("🔄 开始创建任务...\n")
							err := larkService.CreateTaskFromFieldValues(tableConfig, fieldValues)
							if err != nil {
								fmt.Printf("❌ 创建任务失败: %v\n", err)
							} else {
								fmt.Printf("✅ 任务创建成功！\n")
							}
						}(table)
					break
				}
			}

					break
				} else {
					// 还有字段没有数据，继续检测
					fmt.Printf("⏳ 记录ID %s 的指定字段尚未全部有数据，继续检测...\n", recordID)
					// 等待一段时间后重试
					// 计算智能轮询间隔：基础间隔 * (2^min(checkCount, 6))，最大不超过maxInterval
					exponentialFactor := 1 << uint(min(checkCount, 6)) // 2的幂，最多64倍
					checkInterval := baseInterval * time.Duration(exponentialFactor)
					if checkInterval > maxInterval {
						checkInterval = maxInterval
					}
					time.Sleep(checkInterval)
					checkCount++
				}
			}

			// 如果达到最大检测次数仍未完成，记录日志
			if checkCount >= maxChecks {
				fmt.Printf("⏰ 记录ID %s 的字段检测已达到最大次数(%d次)，自动停止检测\n", recordID, maxChecks)
			}
		}()
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "记录添加成功",
		"recordID": recordID,
	})
}

// GetAIModels 获取可用的AI模型列表
func GetAIModels(c *gin.Context) {
	if configService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "配置服务未初始化"})
		return
	}

	// 获取配置
	config := configService.GetConfig()

	// 验证配置
	if config.SiliconFlow.ApiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SiliconFlow API key not configured"})
		return
	}

	// 创建AI服务实例
	aiService := services.NewAIService(&config.SiliconFlow)

	// 获取模型列表
	models, err := aiService.GetModels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"models": models})
}

// AIParse 使用AI解析内容
func AIParse(c *gin.Context) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		fmt.Printf("[AIParse] 请求处理总耗时: %v\n", elapsed)
	}()

	if configService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "配置服务未初始化"})
		return
	}

	// 获取配置
	config := configService.GetConfig()

	// 验证配置
	if config.SiliconFlow.ApiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SiliconFlow API key not configured"})
		return
	}

	// 解析请求
	var req AIParseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("[AIParse] 请求解析失败: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("[AIParse] 解析到的请求参数: %+v\n", req)

	// 创建AI服务实例
	aiService := services.NewAIService(&config.SiliconFlow)

	// 调用AI解析
	content := req.Content
	if content == "" {
		content = req.BaseFieldValue
	}

	fmt.Printf("[AIParse] 调用AI服务，输入内容: %s\n", content)
	result, err := aiService.ParseWithAI(content, req.Prompt)
	if err != nil {
		fmt.Printf("[AIParse] AI解析失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("[AIParse] AI解析成功，结果: %s\n", result)
	c.JSON(http.StatusOK, AIParseResponse{Result: result})
}

// CheckRecordStatus 检查记录状态
func CheckRecordStatus(c *gin.Context) {
	if configService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "配置服务未初始化"})
		return
	}

	// 获取配置
	config := configService.GetConfig()

	if config.AppID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先配置飞书应用信息"})
		return
	}

	appToken := c.Query("app_token")
	tableID := c.Query("table_id")
	recordID := c.Query("record_id")

	if appToken == "" || tableID == "" || recordID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必要参数"})
		return
	}

	larkService := services.NewLarkService(config.AppID, config.AppSecret)

	// 支持新的多表格配置和旧的单表格配置
	var checkFields []string
	if len(config.Tables) > 0 {
		// 新格式：从对应的表格配置中获取检测字段
		for _, table := range config.Tables {
			if table.AppToken == appToken && table.TableID == tableID {
				checkFields = table.CheckFields
				break
			}
		}
	} else {
		// 旧格式：向后兼容
		checkFields = config.CheckFields
	}

	completed, _, err := larkService.CheckFieldsCompleted(appToken, tableID, recordID, checkFields)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"completed": completed,
	})
}