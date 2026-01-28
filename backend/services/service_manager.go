package services

import (
	"sync"
	"time"
)

// ServiceManager 服务管理器，负责管理LarkService实例的生命周期
type ServiceManager struct {
	mu            sync.RWMutex
	services      map[string]*LarkService // key: appID_appSecret
	serviceExpiry map[string]time.Time    // key: appID_appSecret, value: 过期时间
}

// NewServiceManager 创建新的服务管理器
func NewServiceManager() *ServiceManager {
	manager := &ServiceManager{
		services:      make(map[string]*LarkService),
		serviceExpiry: make(map[string]time.Time),
	}
	
	// 启动定期清理过期服务的goroutine
	go manager.cleanExpiredServices()
	
	return manager
}

// GetLarkService 获取或创建LarkService实例
func (m *ServiceManager) GetLarkService(appID, appSecret string) *LarkService {
	if appID == "" || appSecret == "" {
		return nil
	}
	
	key := appID + "_" + appSecret
	
	// 尝试从缓存中获取服务实例
	m.mu.RLock()
	service, exists := m.services[key]
	m.mu.RUnlock()
	
	// 如果服务存在且未过期，直接返回
	if exists {
		m.mu.RLock()
		expiry, _ := m.serviceExpiry[key]
		m.mu.RUnlock()
		
		if time.Now().Before(expiry) {
			return service
		}
	}
	
	// 创建或更新服务实例
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 双重检查，避免竞态条件
	service, exists = m.services[key]
	if exists {
		expiry, _ := m.serviceExpiry[key]
		if time.Now().Before(expiry) {
			return service
		}
	}
	
	// 创建新的服务实例
	service = NewLarkService(appID, appSecret)
	m.services[key] = service
	m.serviceExpiry[key] = time.Now().Add(24 * time.Hour) // 服务实例有效期24小时
	
	return service
}

// cleanExpiredServices 定期清理过期的服务实例
func (m *ServiceManager) cleanExpiredServices() {
	for {
		// 每1小时清理一次
		time.Sleep(1 * time.Hour)
		
		now := time.Now()
		m.mu.Lock()
		
		for key, expiry := range m.serviceExpiry {
			if now.After(expiry) {
				delete(m.services, key)
				delete(m.serviceExpiry, key)
			}
		}
		
		m.mu.Unlock()
	}
}