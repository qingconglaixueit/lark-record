package handlers

import (
	"fmt"
	"lark-record/models"
	"lark-record/services"
	"net/http"
	"sync"

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

	larkService := services.NewLarkService(config.AppID, config.AppSecret)
	tables, err := larkService.GetBitableTables(appToken)
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

	// 检查是否需要发送消息
	// 支持新的多表格配置和旧的单表格配置
	var checkFields []string
	if len(config.Tables) > 0 {
		// 新格式：从对应的表格配置中获取检测字段
		for _, table := range config.Tables {
			if table.AppToken == req.AppToken && table.TableID == req.TableID {
				checkFields = table.CheckFields
				break
			}
		}
	} else {
		// 旧格式：向后兼容
		checkFields = config.CheckFields
	}

	if checkFields != nil && len(checkFields) > 0 && config.GroupChatID != "" {
		go func() {
			completed, fieldValues, err := larkService.CheckFieldsCompleted(req.AppToken, req.TableID, recordID, checkFields)
			if err != nil {
				fmt.Printf("检查字段状态失败: %v\n", err)
				return
			}

			if completed {
				// 拼接字段值到消息中
				message := fmt.Sprintf("记录已完成！记录ID: %s\n\n检测字段值：\n", recordID)
				for fieldName, value := range fieldValues {
					// 处理不同类型的值，确保消息格式清晰
					switch v := value.(type) {
					case string:
						message += fmt.Sprintf("%s: %s\n", fieldName, v)
					case []interface{}:
						// 处理数组类型的值（如多选）
						message += fmt.Sprintf("%s: ", fieldName)
						for i, item := range v {
							if i > 0 {
								message += ", "
							}
							message += fmt.Sprintf("%v", item)
						}
						message += "\n"
					default:
						message += fmt.Sprintf("%s: %v\n", fieldName, v)
					}
				}
				err = larkService.SendMessage(config.GroupChatID, message)
				if err != nil {
					fmt.Printf("发送消息失败: %v\n", err)
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