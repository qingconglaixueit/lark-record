package handlers

import (
	"fmt"
	"lark-record/models"
	"lark-record/services"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// 存储配置信息的缓存
var configCache models.Config
var cacheMutex sync.RWMutex

// SaveConfig 保存配置
func SaveConfig(c *gin.Context) {
	var config models.Config
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 测试配置是否有效 - 验证凭证
	larkService := services.NewLarkService(config.AppID, config.AppSecret)
	err := larkService.ValidateCredentials()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "飞书配置无效: " + err.Error()})
		return
	}

	// 保存配置到缓存
	cacheMutex.Lock()
	configCache = config
	cacheMutex.Unlock()

	c.JSON(http.StatusOK, gin.H{"message": "配置保存成功"})
}

// GetConfig 获取配置
func GetConfig(c *gin.Context) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	if configCache.AppID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "未配置"})
		return
	}

	c.JSON(http.StatusOK, configCache)
}

// GetBitables 获取多维表格列表
func GetBitables(c *gin.Context) {
	cacheMutex.RLock()
	config := configCache
	cacheMutex.RUnlock()

	if config.AppID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先配置飞书应用信息"})
		return
	}

	larkService := services.NewLarkService(config.AppID, config.AppSecret)
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
	cacheMutex.RLock()
	config := configCache
	cacheMutex.RUnlock()

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

	larkService := services.NewLarkService(config.AppID, config.AppSecret)
	tables, err := larkService.GetBitableTables(appToken, isWiki)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tables)
}

// GetTableFields 获取数据表的字段列表
func GetTableFields(c *gin.Context) {
	cacheMutex.RLock()
	config := configCache
	cacheMutex.RUnlock()

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

	larkService := services.NewLarkService(config.AppID, config.AppSecret)
	fields, err := larkService.GetTableFields(appToken, tableID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, fields)
}

// AddRecord 新增记录
func AddRecord(c *gin.Context) {
	cacheMutex.RLock()
	config := configCache
	cacheMutex.RUnlock()

	if config.AppID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先配置飞书应用信息"})
		return
	}
	
	// 输出当前配置信息（用于调试）
	fmt.Printf("当前配置信息：\n")
	fmt.Printf("- AppID: %s\n", config.AppID)
	fmt.Printf("- GroupChatID: %s\n", config.GroupChatID)
	fmt.Printf("- Tables配置数量: %d\n", len(config.Tables))

	var req models.AddRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	larkService := services.NewLarkService(config.AppID, config.AppSecret)
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
				
				// 持续检测，直到所有指定字段都有数据
				for {
					completed, fieldValues, err := larkService.CheckFieldsCompleted(req.AppToken, req.TableID, recordID, checkFields)
				if err != nil {
					fmt.Printf("❌ 检查字段状态失败: %v\n", err)
					// 等待一段时间后重试
					time.Sleep(5 * time.Second)
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
						if table.AppToken == req.AppToken && table.TableID == req.TableID && table.CreateTask {
							fmt.Printf("🔄 开始创建任务...\n")
							
							// 查找记录人字段（ui_type为User的字段）
							var assigneeID string
							var recordTime int64
							
							// 先从fieldValues中查找记录人
							// 先从表格配置的write_fields中查找user类型的字段
							for _, writeField := range table.WriteFields {
								// 获取该字段的值
								if value, exists := fieldValues[writeField.FieldName]; exists {
									// 检查是否是用户类型字段
									if userMap, ok := value.(map[string]interface{}); ok {
										if id, ok := userMap["id"].(string); ok {
											assigneeID = id
											fmt.Printf("👤 找到记录人: %s\n", assigneeID)
											break
										}
									}
								}
							}

							// 如果没有找到记录人，尝试从所有字段中查找
							if assigneeID == "" {
								for _, value := range fieldValues {
									if userMap, ok := value.(map[string]interface{}); ok {
										if id, ok := userMap["id"].(string); ok {
											assigneeID = id
											fmt.Printf("👤 从所有字段中找到记录人: %s\n", assigneeID)
											break
										}
									}
								}
							}
							
							// 如果仍然没有找到记录人，重新获取记录的所有字段
							if assigneeID == "" {
								fmt.Printf("🔍 尝试重新获取记录的所有字段...\n")
								// 重新获取记录的所有字段
								recordFields, err := larkService.GetRecord(req.AppToken, req.TableID, recordID)
								if err != nil {
									fmt.Printf("❌ 重新获取记录失败: %v\n", err)
								} else {
									// 从所有字段中查找记录人
									for fieldName, value := range recordFields {
										// 检查是否为单个用户格式
										if userMap, ok := value.(map[string]interface{}); ok {
											if id, ok := userMap["id"].(string); ok {
												assigneeID = id
												fmt.Printf("👤 从字段 '%s' 中找到记录人: %s\n", fieldName, assigneeID)
												break
											}
										}
										// 检查是否为用户数组格式
										if userArray, ok := value.([]interface{}); ok && len(userArray) > 0 {
											if firstUser, ok := userArray[0].(map[string]interface{}); ok {
												if id, ok := firstUser["id"].(string); ok {
													assigneeID = id
													fmt.Printf("👤 从字段 '%s' 的用户数组中找到记录人: %s\n", fieldName, assigneeID)
													break
												}
											}
										}
									}
								}
							}

							// 获取任务标题
							var taskTitle string
							if summaryField := table.TaskSummaryField; summaryField != "" {
								if value, exists := fieldValues[summaryField]; exists {
									switch v := value.(type) {
									case string:
										taskTitle = v
									case float64:
										taskTitle = fmt.Sprintf("%v", v)
									case []interface{}:
										// 处理数组类型的值
										for i, item := range v {
											if i > 0 {
												taskTitle += ", "
											}
											taskTitle += fmt.Sprintf("%v", item)
										}
									default:
										taskTitle = fmt.Sprintf("%v", v)
									}
								}
							}

							// 如果没有找到任务标题，使用默认标题
							if taskTitle == "" {
								taskTitle = "来自多维表格的任务"
							}

							// 如果找到记录人，创建任务
							if assigneeID != "" {
								// 设置默认截止时间为当前时间
								defaultDue := time.Now().UnixMilli()
								
								// 尝试从字段值中获取截止时间
								for fieldName, value := range fieldValues {
									// 处理时间戳，支持int64和float64两种类型
									var timestamp int64
									switch v := value.(type) {
									case int64:
										timestamp = v
									case float64:
										timestamp = int64(v)
									default:
										continue
									}
										
									if timestamp > 0 && timestamp < 3250368000000 {
										// 这看起来是一个有效的时间戳
										recordTime = timestamp
										fmt.Printf("⏰ 从字段 '%s' 中获取到截止时间：%d", fieldName, timestamp)
										// 转换为东八区时间以便显示
										t := time.Unix(timestamp/1000, 0).In(time.FixedZone("Asia/Shanghai", 8*3600))
										fmt.Printf("📅 格式化时间：%s", t.Format("2006-01-02 15:04:05"))
										break
									}
								}

								// 如果没有找到有效的截止时间，使用默认值
								dueTime := recordTime
								if dueTime == 0 {
									dueTime = defaultDue
								}

								// 创建任务
								taskID, err := larkService.CreateTask(assigneeID, taskTitle, dueTime, false)
								if err != nil {
									fmt.Printf("❌ 创建任务失败: %v\n", err)
								} else {
									fmt.Printf("✅ 任务创建成功！任务ID: %s\n", taskID)
								}
							} else {
								fmt.Printf("⚠️ 未找到记录人信息，无法创建任务\n")
							}
							
							break
						}
					}

					break
				} else {
					// 还有字段没有数据，继续检测
					fmt.Printf("⏳ 记录ID %s 的指定字段尚未全部有数据，继续检测...\n", recordID)
					// 等待一段时间后重试
					time.Sleep(5 * time.Second)
				}
			}
		}()
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "记录添加成功",
		"recordID": recordID,
	})
}

// CheckRecordStatus 检查记录状态
func CheckRecordStatus(c *gin.Context) {
	cacheMutex.RLock()
	config := configCache
	cacheMutex.RUnlock()

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