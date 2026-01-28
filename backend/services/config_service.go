package services

import (
	"encoding/json"
	"fmt"
	"lark-record/models"
	"os"
	"sync"
)

// ConfigService 配置管理服务
type ConfigService struct {
	configPath string
	config     *models.Config
	mutex      sync.RWMutex
}

// NewConfigService 创建新的配置服务
func NewConfigService(configPath string) *ConfigService {
	if configPath == "" {
		configPath = "./config.json"
	}

	service := &ConfigService{
		configPath: configPath,
		config:     &models.Config{},
	}

	// 初始化时加载配置
	service.loadConfig()

	return service
}

// GetConfig 获取配置
func (s *ConfigService) GetConfig() *models.Config {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 返回配置的副本，避免外部直接修改
	configCopy := *s.config
	return &configCopy
}

// SetConfig 设置配置
func (s *ConfigService) SetConfig(config *models.Config) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 更新配置
	*s.config = *config

	// 保存到文件
	return s.saveConfig()
}

// UpdateConfig 增量更新配置
func (s *ConfigService) UpdateConfig(newConfig *models.Config) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 更新基础配置
	if newConfig.AppID != "" {
		s.config.AppID = newConfig.AppID
	}
	if newConfig.AppSecret != "" {
		s.config.AppSecret = newConfig.AppSecret
	}
	if newConfig.GroupChatID != "" {
		s.config.GroupChatID = newConfig.GroupChatID
	}

	// 更新SiliconFlow配置
	if newConfig.SiliconFlow.ApiKey != "" {
		s.config.SiliconFlow = newConfig.SiliconFlow
	}

	// 更新表格配置
	if newConfig.Tables != nil && len(newConfig.Tables) > 0 {
		// 创建一个map用于快速查找现有表格
		existingTables := make(map[string]int) // key: appToken_tableID, value: index
		for i, table := range s.config.Tables {
			key := fmt.Sprintf("%s_%s", table.AppToken, table.TableID)
			existingTables[key] = i
		}

		// 处理新表格配置
		for _, newTable := range newConfig.Tables {
			key := fmt.Sprintf("%s_%s", newTable.AppToken, newTable.TableID)
			if index, exists := existingTables[key]; exists {
				// 更新现有表格
				s.config.Tables[index] = newTable
				logInfo("更新表格配置: %s", newTable.Name)
			} else {
				// 添加新表格
				s.config.Tables = append(s.config.Tables, newTable)
				logInfo("添加新表格配置: %s", newTable.Name)
			}
		}
	}

	// 保存到文件
	return s.saveConfig()
}

// IsConfigured 检查是否已配置
func (s *ConfigService) IsConfigured() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.config.AppID != "" && s.config.AppSecret != ""
}

// loadConfig 从文件加载配置
func (s *ConfigService) loadConfig() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查文件是否存在
	if _, err := os.Stat(s.configPath); os.IsNotExist(err) {
		logInfo("配置文件不存在，将使用默认配置")
		return
	}

	// 读取文件内容
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		logError("读取配置文件失败: %v", err)
		return
	}

	// 解析JSON
	if err := json.Unmarshal(data, s.config); err != nil {
		logError("解析配置文件失败: %v", err)
		return
	}

	logInfo("配置文件加载成功")
}

// saveConfig 保存配置到文件
func (s *ConfigService) saveConfig() error {
	// 转换为JSON
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return fmt.Errorf("转换配置为JSON失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(s.configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	logInfo("配置已保存到文件")
	return nil
}