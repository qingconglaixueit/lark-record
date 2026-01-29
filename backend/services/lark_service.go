package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"lark-record/models"
	"net/http"
	"strings"
	"sync"
	"time"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
)

// 常量定义
const (
	// 缓存有效期
	TenantAccessTokenExpireTime = 1 * time.Hour  // 租户访问令牌有效期
	WikiTokenCacheExpireTime    = 1 * time.Hour  // Wiki Token缓存有效期
	FieldsCacheExpireTime       = 1 * time.Hour  // 字段缓存有效期（延长至1小时）
	BitablesCacheExpireTime     = 1 * time.Hour  // 多维表格列表缓存有效期
	TablesCacheExpireTime       = 1 * time.Hour  // 数据表列表缓存有效期
	
	// 重试配置
	MaxRetries     = 3              // 最大重试次数
	InitialRetryDelay = 1 * time.Second // 初始重试间隔
)

// 定期清理过期缓存的函数
func (s *LarkService) cleanExpiredCache() {
	for {
		// 每10分钟清理一次缓存
		time.Sleep(10 * time.Minute)
		
		now := time.Now()
		
		// 清理wikiTokenCache
		s.wikiTokenCache.Range(func(key, value interface{}) bool {
			// 在原始代码中wikiTokenCache只存储bool值，需要先修改为存储结构体
			// 这里我们需要先检查是否已经是结构体类型
			if cacheItem, ok := value.(struct {
				isWiki  bool
				expires time.Time
			}); ok {
				if now.After(cacheItem.expires) {
					s.wikiTokenCache.Delete(key)
				}
			}
			return true
		})
		
		// 清理fieldsCache
		s.fieldsCacheTime.Range(func(key, value interface{}) bool {
			if now.After(value.(time.Time)) {
				s.fieldsCache.Delete(key)
				s.fieldsCacheTime.Delete(key)
			}
			return true
		})
		
		// 清理bitablesCache
		s.bitablesCacheTime.Range(func(key, value interface{}) bool {
			if now.After(value.(time.Time)) {
				s.bitablesCache.Delete(key)
				s.bitablesCacheTime.Delete(key)
			}
			return true
		})
		
		// 清理tablesCache
		s.tablesCacheTime.Range(func(key, value interface{}) bool {
			if now.After(value.(time.Time)) {
				s.tablesCache.Delete(key)
				s.tablesCacheTime.Delete(key)
			}
			return true
		})
	}
}

// WikiNodeResponse 知识库节点响应
type WikiNodeResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Node struct {
			NodeToken string `json:"node_token"`
			ObjToken  string `json:"obj_token"`
			Title     string `json:"title"`
			ObjType   string `json:"obj_type"`
		} `json:"node"`
	} `json:"data"`
}

// WikiNode 知识库节点
type WikiNode struct {
	NodeToken string `json:"node_token"`
	Title     string `json:"title"`
	ObjToken  string `json:"obj_token"`
	ObjType   string `json:"obj_type"`
	HasChild  bool   `json:"has_child"`
}

// WikiNodesResponse 知识库节点列表响应
type WikiNodesResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Items   []WikiNode `json:"items"`
		HasMore bool       `json:"has_more"`
		Token   string     `json:"page_token"`
	} `json:"data"`
}

// BaseService 基础服务结构，提供共享功能
type BaseService struct {
	appID            string
	appSecret        string
	client           *lark.Client
	httpClient       *http.Client
	// 访问令牌缓存
	tenantAccessToken string
	tokenExpireTime   time.Time
	tokenMutex        sync.RWMutex
}

// handleHTTPRequest 通用HTTP请求处理函数
// 提供通用的HTTP请求构建和响应处理逻辑
func (s *BaseService) handleHTTPRequest(method, url, token string, body []byte) (*http.Response, []byte, error) {
	// 创建HTTP请求
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 如果有请求体，设置请求体
	if body != nil && len(body) > 0 {
		req.Body = io.NopCloser(bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}

	// 设置请求头部
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("发送请求失败: %w", err)
	}

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return resp, respBody, nil
}

// LarkService 飞书API服务
// 处理飞书API调用的核心服务
// 实现了令牌管理、多维表格操作、字段管理等功能
type LarkService struct {
	BaseService
	// Wiki转换结果缓存
	wikiConvertCache sync.Map
	// 字段缓存
	fieldsCache     sync.Map
	fieldsCacheTime sync.Map
	// Wiki Token缓存
	wikiTokenCache  sync.Map
	// 多维表格列表缓存
	bitablesCache     sync.Map
	bitablesCacheTime sync.Map
	// 数据表列表缓存
	tablesCache     sync.Map
	tablesCacheTime sync.Map
	// 拆分的服务
	bitableService  *LarkBitableService
	messageService  *LarkMessageService
	taskService     *LarkTaskService
}

// NewBaseService 创建基础服务实例
func NewBaseService(appID, appSecret string) BaseService {
	return BaseService{
		appID:     appID,
		appSecret: appSecret,
		client:    lark.NewClient(appID, appSecret),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// NewLarkService 创建一个新的LarkService实例
// 初始化服务并启动定期清理缓存的goroutine
func NewLarkService(appID, appSecret string) *LarkService {
	baseService := NewBaseService(appID, appSecret)
	
	larkService := &LarkService{
		BaseService: baseService,
	}
	
	// 启动定期清理缓存的goroutine
	go larkService.cleanExpiredCache()
	
	// 初始化拆分的服务，共享同一个BaseService实例的引用
	larkService.bitableService = NewLarkBitableService(appID, appSecret)
	larkService.bitableService.BaseService = baseService
	
	larkService.messageService = NewLarkMessageService(appID, appSecret)
	larkService.messageService.BaseService = baseService
	
	larkService.taskService = NewLarkTaskService(appID, appSecret)
	larkService.taskService.BaseService = baseService
	
	return larkService
}



// 原有的CreateTask方法已删除，使用后续的HTTP实现版本

// GetClient 获取飞书客户端
// 初始化并返回飞书客户端实例
func (s *BaseService) GetClient() *lark.Client {
	if s.client == nil {
		s.client = lark.NewClient(s.appID, s.appSecret)
	}
	return s.client
}

// GetBitables 获取多维表格列表（转发到bitableService）
func (s *LarkService) GetBitables() ([]models.Bitable, error) {
	return s.bitableService.GetBitables()
}

// GetBitableTables 获取多维表格中的数据表列表（转发到bitableService）
func (s *LarkService) GetBitableTables(appToken string, isWiki bool) ([]models.TableInfo, error) {
	return s.bitableService.GetBitableTables(appToken, isWiki)
}

// CreateTaskFromFieldValues 从字段值创建任务
// 该方法将调用taskService的同名方法
func (s *LarkService) CreateTaskFromFieldValues(tableConfig models.TableConfig, fieldValues map[string]interface{}) error {
	return s.taskService.CreateTaskFromFieldValues(tableConfig, fieldValues)
}

// GetTenantAccessToken 获取租户访问令牌
// 使用双重检查锁定模式确保并发安全
// 缓存令牌并在过期前自动刷新
func (s *BaseService) GetTenantAccessToken() (string, error) {
	// 快速检查令牌是否有效
	s.tokenMutex.RLock()
	if s.tenantAccessToken != "" && s.tokenExpireTime.After(time.Now()) {
		token := s.tenantAccessToken
		s.tokenMutex.RUnlock()
		return token, nil
	}
	s.tokenMutex.RUnlock()

	// 需要获取新令牌，加写锁
	s.tokenMutex.Lock()
	defer s.tokenMutex.Unlock()

	// 再次检查，防止在等待锁的过程中已有其他协程获取了新令牌
	if s.tenantAccessToken != "" && s.tokenExpireTime.After(time.Now()) {
		return s.tenantAccessToken, nil
	}

	// 调用飞书API获取租户访问令牌
	url := "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal"
	reqBody := map[string]string{
		"app_id":     s.appID,
		"app_secret": s.appSecret,
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("构建令牌请求失败: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return "", fmt.Errorf("创建令牌请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("发送令牌请求失败: %w", err)
	}
	defer httpResp.Body.Close()

	httpBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return "", fmt.Errorf("读取令牌响应失败: %w", err)
	}

	type TokenResponse struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int64  `json:"expire"`
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(httpBody, &tokenResp); err != nil {
		return "", fmt.Errorf("解析令牌响应失败: %w", err)
	}

	if tokenResp.Code != 0 {
		return "", fmt.Errorf("获取租户访问令牌失败: %s (Code: %d)", tokenResp.Msg, tokenResp.Code)
	}

	// 更新令牌和过期时间
	s.tenantAccessToken = tokenResp.TenantAccessToken
	s.tokenExpireTime = time.Now().Add(time.Duration(tokenResp.Expire) * time.Second)

	return tokenResp.TenantAccessToken, nil
}

// 原有的initClient方法已替换为BaseService的GetClient方法



// fetchWikiTablesDirectly 直接通过HTTP API获取Wiki节点关联的数据表

// fetchBitableTables 获取指定bitable的所有数据表
func (s *LarkService) fetchBitableTables(bitableToken, bitableName, accessToken string) ([]models.TableInfo, error) {
	fmt.Printf("✅ 找到Bitable节点: 标题=%s, ObjToken=%s\n", bitableName, bitableToken)

	// 尝试获取这个bitable的数据表列表
	tablesURL := fmt.Sprintf("https://open.feishu.cn/open-apis/bitable/v1/apps/%s/tables", bitableToken)
	tablesReq, err := http.NewRequest("GET", tablesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建数据列表请求失败: %w", err)
	}
	tablesReq.Header.Set("Authorization", "Bearer "+accessToken)

	tablesResp, err := s.httpClient.Do(tablesReq)
	if err != nil {
		return nil, fmt.Errorf("获取数据表列表失败: %w", err)
	}
	defer tablesResp.Body.Close()

	tablesBody, err := io.ReadAll(tablesResp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取数据表列表响应失败: %w", err)
	}

	type TablesResponse struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Items []struct {
				TableID string `json:"table_id"`
				Name    string `json:"name"`
			} `json:"items"`
		} `json:"data"`
	}

	var tablesResult TablesResponse
	if err := json.Unmarshal(tablesBody, &tablesResult); err != nil {
		return nil, fmt.Errorf("解析数据表列表失败: %w", err)
	}

	if tablesResult.Code != 0 {
		return nil, fmt.Errorf("获取数据表列表失败: %s (Code: %d)", tablesResult.Msg, tablesResult.Code)
	}

	if len(tablesResult.Data.Items) > 0 {
		fmt.Printf("  - 在 '%s' 中找到 %d 个数据表\n", bitableName, len(tablesResult.Data.Items))

		var tables []models.TableInfo
		for _, table := range tablesResult.Data.Items {
			tables = append(tables, models.TableInfo{
				TableID: table.TableID,
				Name:    table.Name,
			})
			fmt.Printf("    * 表格: %s (%s)\n", table.Name, table.TableID)
		}
		return tables, nil
	}

	return []models.TableInfo{}, nil
}


// fetchNodeTablesDirectly 直接从节点获取table信息
func (s *LarkService) fetchNodeTablesDirectly(nodeToken, accessToken, targetWikiToken string) ([]models.TableInfo, error) {
	var allTables []models.TableInfo

	// 获取节点信息
	nodeURL := fmt.Sprintf("https://open.feishu.cn/open-apis/wiki/v2/nodes/%s?user_id_type=user_id", nodeToken)
	nodeReq, err := http.NewRequest("GET", nodeURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建节点请求失败: %w", err)
	}
	nodeReq.Header.Set("Authorization", "Bearer "+accessToken)

	nodeResp, err := s.httpClient.Do(nodeReq)
	if err != nil {
		return nil, fmt.Errorf("获取节点信息失败: %w", err)
	}
	defer nodeResp.Body.Close()

	nodeBody, err := io.ReadAll(nodeResp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取节点响应失败: %w", err)
	}

	if nodeResp.StatusCode != 200 {
		fmt.Printf("    ⚠️  节点 %s 返回HTTP %d\n", nodeToken, nodeResp.StatusCode)
		return []models.TableInfo{}, nil
	}

	type NodeDetailResponse struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Node struct {
				NodeToken string `json:"node_token"`
				ObjToken  string `json:"obj_token"`
				ObjType   string `json:"obj_type"`
				Title     string `json:"title"`
				HasChild  bool   `json:"has_child"`
			} `json:"node"`
		} `json:"data"`
	}

	var nodeDetail NodeDetailResponse
	if err := json.Unmarshal(nodeBody, &nodeDetail); err != nil {
		fmt.Printf("    ⚠️  解析节点信息失败: %v\n", err)
		return []models.TableInfo{}, nil
	}

	node := nodeDetail.Data.Node
	fmt.Printf("      ↳ 子节点: %s (%s), 类型: %s\n", node.Title, node.NodeToken, node.ObjType)

	// 如果这个节点是bitable，获取其数据表
	if node.ObjType == "bitable" && node.ObjToken != "" {
		fmt.Printf("        ✅ 找到Bitable: %s\n", node.Title)
		tables, err := s.fetchBitableTables(node.ObjToken, node.Title, accessToken)
		if err != nil {
			fmt.Printf("        ⚠️  获取 '%s' 的数据表失败: %v\n", node.Title, err)
		} else {
			allTables = append(allTables, tables...)
		}
	}

	return allTables, nil
}



// WikiTokenInfo 存储Wiki Token的相关信息
type WikiTokenInfo struct {
	IsWiki    bool   // 是否为Wiki Token
	ObjToken  string // 如果是Wiki Token，对应的ObjToken
	ObjType   string // 如果是Wiki Token，对应的ObjType
	Title     string // 如果是Wiki Token，对应的标题
}

// GetWikiTokenInfo 获取Wiki Token的相关信息
func (s *LarkService) GetWikiTokenInfo(appToken, token string) WikiTokenInfo {
	// 检查缓存
	if info, ok := s.wikiTokenCache.Load(appToken); ok {
		return info.(WikiTokenInfo)
	}

	// 默认返回结果
	result := WikiTokenInfo{
		IsWiki: false,
	}

	// 调用飞书 Wiki API 检查是否为有效的 Wiki Token
	// 这里使用 GET /wiki/v2/spaces/get_node 接口，如果返回成功则说明是 Wiki Token
	getNodeURL := fmt.Sprintf("https://open.feishu.cn/open-apis/wiki/v2/spaces/get_node?user_id_type=user_id&token=%s", appToken)
	nodeReq, err := http.NewRequest("GET", getNodeURL, nil)
	if err != nil {
		// 如果创建请求失败，使用简单的前缀检查
		result.IsWiki = strings.HasPrefix(appToken, "BEsNwa") || strings.HasPrefix(appToken, "Bdsqwa") || strings.HasPrefix(appToken, "BdUswa")
		// 缓存结果
		s.wikiTokenCache.Store(appToken, result)
		return result
	}

	nodeReq.Header.Set("Authorization", "Bearer "+token)
	nodeResp, nodeErr := s.httpClient.Do(nodeReq)
	if nodeErr != nil {
		// 如果请求失败，使用简单的前缀检查
		result.IsWiki = strings.HasPrefix(appToken, "BEsNwa") || strings.HasPrefix(appToken, "Bdsqwa") || strings.HasPrefix(appToken, "BdUswa")
		// 缓存结果
		s.wikiTokenCache.Store(appToken, result)
		return result
	}
	defer nodeResp.Body.Close()

	nodeBody, _ := io.ReadAll(nodeResp.Body)

	type GetNodeResponse struct {
		Code int `json:"code"`
		Data struct {
			Node struct {
				ObjToken string `json:"obj_token"`
				ObjType  string `json:"obj_type"`
				Title    string `json:"title"`
			} `json:"node"`
		} `json:"data"`
	}

	var nodeResult GetNodeResponse
	if err := json.Unmarshal(nodeBody, &nodeResult); err != nil {
		// 如果解析失败，使用简单的前缀检查
		result.IsWiki = strings.HasPrefix(appToken, "BEsNwa") || strings.HasPrefix(appToken, "Bdsqwa") || strings.HasPrefix(appToken, "BdUswa")
		// 缓存结果
		s.wikiTokenCache.Store(appToken, result)
		return result
	}

	// 如果返回成功，则是有效的 Wiki Token
	result.IsWiki = nodeResult.Code == 0
	if result.IsWiki {
		result.ObjToken = nodeResult.Data.Node.ObjToken
		result.ObjType = nodeResult.Data.Node.ObjType
		result.Title = nodeResult.Data.Node.Title
	}
	// 缓存结果（有效期1小时）
	s.wikiTokenCache.Store(appToken, result)
	return result
}

// IsWikiToken 检查 appToken 是否是 Wiki Token（兼容旧接口）
func (s *LarkService) IsWikiToken(appToken, token string) bool {
	info := s.GetWikiTokenInfo(appToken, token)
	return info.IsWiki
}

// IsWikiTokenOld 判断是否为Wiki Token (旧版本，兼容原有调用)
// 内部调用新版本IsWikiToken，避免重复获取令牌
func (s *LarkService) IsWikiTokenOld(appToken string) bool {
	token, err := s.GetTenantAccessToken()
	if err != nil {
		// 如果获取token失败，使用旧的简单检查方法
		return strings.HasPrefix(appToken, "BEsNwa") || strings.HasPrefix(appToken, "Bdsqwa") || strings.HasPrefix(appToken, "BdUswa")
	}
	return s.IsWikiToken(appToken, token)
}

// GetTableFieldsWithToken 获取数据表字段（带token参数）
// 避免重复获取访问令牌
func (s *LarkService) GetTableFieldsWithToken(appToken, tableID, token string) ([]models.Field, error) {
	// 缓存键
	cacheKey := fmt.Sprintf("%s:%s", appToken, tableID)
	
	// 检查缓存
	if cachedFields, ok := s.fieldsCache.Load(cacheKey); ok {
		if cachedTime, ok := s.fieldsCacheTime.Load(cacheKey); ok {
			// 缓存有效期
			if time.Since(cachedTime.(time.Time)) < FieldsCacheExpireTime {
				return cachedFields.([]models.Field), nil
			}
		}
	}

	// 检查是否为 Wiki Token
	realAppToken := appToken
	wikiInfo := s.GetWikiTokenInfo(appToken, token)
	if wikiInfo.IsWiki {
		if wikiInfo.ObjType == "bitable" && wikiInfo.ObjToken != "" {
			fmt.Printf("✅ 检测到 Wiki Token，获取到 ObjToken: %s\n", wikiInfo.ObjToken)
			realAppToken = wikiInfo.ObjToken
		}
	}

	// 使用实际的 appToken 获取字段
	fieldsURL := fmt.Sprintf("https://open.feishu.cn/open-apis/bitable/v1/apps/%s/tables/%s/fields?user_id_type=user_id", realAppToken, tableID)
	fieldsReq, err := http.NewRequest("GET", fieldsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建字段请求失败: %w", err)
	}
	fieldsReq.Header.Set("Authorization", "Bearer "+token)

	fieldsResp, err := s.httpClient.Do(fieldsReq)
	if err != nil {
		return nil, fmt.Errorf("获取字段列表失败: %w", err)
	}
	defer fieldsResp.Body.Close()

	fieldsBody, err := io.ReadAll(fieldsResp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取字段响应失败: %w", err)
	}

	type FieldsResponse struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Items []struct {
				FieldName string `json:"field_name"`
				Type      int    `json:"type"`
				FieldId   string `json:"field_id"`
				Property  *struct {
					IsPrimary *bool `json:"is_primary"`
				} `json:"property,omitempty"`
				UiType string `json:"ui_type"`
			} `json:"items"`
		} `json:"data"`
	}

	var fieldsResult FieldsResponse
	if err := json.Unmarshal(fieldsBody, &fieldsResult); err != nil {
		return nil, fmt.Errorf("解析字段响应失败: %w", err)
	}

	if fieldsResult.Code != 0 {
		fmt.Printf("📋 字段API响应: %s\n", string(fieldsBody))
		return nil, fmt.Errorf("获取字段列表失败: %s (Code: %d)", fieldsResult.Msg, fieldsResult.Code)
	}

	var fields []models.Field
	for _, field := range fieldsResult.Data.Items {
		isPrimary := false
		if field.Property != nil && field.Property.IsPrimary != nil {
			isPrimary = *field.Property.IsPrimary
		}
		fields = append(fields, models.Field{
			FieldName: field.FieldName,
			FieldType: fmt.Sprintf("%d", field.Type),
			FieldID:   field.FieldId,
			IsPrimary: isPrimary,
			UiType:    field.UiType,
		})
	}

	// 缓存字段结果
	s.fieldsCache.Store(cacheKey, fields)
	s.fieldsCacheTime.Store(cacheKey, time.Now())

	return fields, nil
}

// getTableFieldsViaHTTP 通过HTTP API获取数据表的所有字段
// 优化：使用统一的Wiki Token处理函数和通用HTTP请求处理函数
func (s *LarkService) getTableFieldsViaHTTP(appToken, tableID string) ([]models.Field, error) {
	// 获取访问令牌
	token, err := s.GetTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("获取访问令牌失败: %w", err)
	}

	// 检查是否为 wiki token
	realAppToken := appToken
	isWiki, objType, objToken, wikiErr := s.getWikiTokenInfo(appToken, token)
	if wikiErr != nil {
		fmt.Printf("⚠️ Wiki Token处理警告: %v\n", wikiErr)
	}

	if isWiki && objType == "bitable" && objToken != "" {
		fmt.Printf("✅ 检测到 Wiki Token，获取到 ObjToken: %s\n", objToken)
		realAppToken = objToken
	}

	// 使用实际的 appToken 获取字段
	fieldsURL := fmt.Sprintf("https://open.feishu.cn/open-apis/bitable/v1/apps/%s/tables/%s/fields?user_id_type=user_id", realAppToken, tableID)

	// 使用通用HTTP请求处理函数
	fieldsResp, fieldsBody, err := s.handleHTTPRequest("GET", fieldsURL, token, nil)
	if err != nil {
		return nil, fmt.Errorf("获取字段列表失败: %w", err)
	}
	defer fieldsResp.Body.Close()

	type FieldsResponse struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Items []struct {
				FieldName string `json:"field_name"`
				Type      int    `json:"type"`
				FieldId   string `json:"field_id"`
				Property  *struct {
					IsPrimary *bool `json:"is_primary"`
				} `json:"property,omitempty"`
				UiType string `json:"ui_type"`
			} `json:"items"`
		} `json:"data"`
	}

	var fieldsResult FieldsResponse
	if err := json.Unmarshal(fieldsBody, &fieldsResult); err != nil {
		return nil, fmt.Errorf("解析字段响应失败: %w", err)
	}

	if fieldsResult.Code != 0 {
		return nil, fmt.Errorf("获取字段列表失败: %s (Code: %d)", fieldsResult.Msg, fieldsResult.Code)
	}

	var fields []models.Field
	for _, field := range fieldsResult.Data.Items {
		isPrimary := false
		if field.Property != nil && field.Property.IsPrimary != nil {
			isPrimary = *field.Property.IsPrimary
		}
		fields = append(fields, models.Field{
			FieldName: field.FieldName,
			FieldType: fmt.Sprintf("%d", field.Type),
			FieldID:   field.FieldId,
			IsPrimary: isPrimary,
			UiType:    field.UiType,
		})
	}

	return fields, nil
}

// GetTableFields 获取数据表的所有字段（带缓存）
func (s *LarkService) GetTableFields(appToken, tableID string) ([]models.Field, error) {
	// 获取访问令牌
	token, err := s.GetTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("获取访问令牌失败: %w", err)
	}

	// 调用带token参数的版本
	return s.GetTableFieldsWithToken(appToken, tableID, token)
}



// AddRecord 新增记录
func (s *LarkService) AddRecord(appToken, tableID string, fields map[string]interface{}) (string, error) {
	// 获取访问令牌
	token, err := s.GetTenantAccessToken()
	if err != nil {
		return "", fmt.Errorf("获取访问令牌失败: %w", err)
	}

	// 使用通用的 getWikiTokenInfo 函数处理 Wiki Token
	isWiki, objType, objToken, wikiErr := s.getWikiTokenInfo(appToken, token)
	if wikiErr != nil {
		fmt.Printf("⚠️ Wiki Token处理警告: %v\n", wikiErr)
	}

	// 设置实际的AppToken
	realAppToken := appToken
	if isWiki && objType == "bitable" && objToken != "" {
		fmt.Printf("✅ 检测到 Wiki Token，获取到 ObjToken: %s\n", objToken)
		realAppToken = objToken
	}

	// 首先尝试使用SDK添加记录
	record := larkbitable.NewAppTableRecordBuilder().
		Fields(fields).
		Build()

	req := larkbitable.NewCreateAppTableRecordReqBuilder().
		AppToken(realAppToken).
		TableId(tableID).
		AppTableRecord(record).
		Build()

	resp, err := s.GetClient().Bitable.AppTableRecord.Create(context.Background(), req)
	if err == nil && resp.Success() {
		if resp.Data != nil && resp.Data.Record != nil && resp.Data.Record.RecordId != nil {
			return *resp.Data.Record.RecordId, nil
		}
		return "", fmt.Errorf("新增记录失败: 未获取到记录ID")
	}

	// 如果获取失败，尝试使用HTTP API直接添加记录
	fmt.Println("🔍 SDK添加记录失败，尝试使用HTTP API...")

	// 使用实际的 appToken 添加记录
	fieldsURL := fmt.Sprintf("https://open.feishu.cn/open-apis/bitable/v1/apps/%s/tables/%s/records?user_id_type=user_id", realAppToken, tableID)

	// 添加调试日志
	fmt.Printf("📋 准备添加记录 - AppToken: %s, TableID: %s\n", realAppToken, tableID)
	fmt.Printf("📋 Fields数据: %+v\n", fields)

	// 获取表格字段信息，用于验证
	fmt.Println("🔍 获取表格字段信息，用于验证...")
	tableFields, err := s.GetTableFieldsWithToken(realAppToken, tableID, token)
	if err != nil {
		fmt.Printf("⚠️ 获取表格字段失败: %v\n", err)
	} else {
		fmt.Printf("✅ 表格字段信息: %d 个字段\n", len(tableFields))
		for _, field := range tableFields {
			fmt.Printf("  - 字段名: %s, 类型: %s, ID: %s\n", field.FieldName, field.FieldType, field.FieldID)
		}

		// 检查必填字段是否都已提供
		fmt.Println("🔍 检查必填字段是否都已提供...")
		for _, field := range tableFields {
			// 检查是否为必填字段（通常字段ID以 "opt" 开头的是可选字段，其他可能是必填）
			isRequired := !strings.HasPrefix(field.FieldID, "opt")
			if isRequired {
				if _, exists := fields[field.FieldName]; !exists {
					fmt.Printf("⚠️ 必填字段缺失: %s (ID: %s)\n", field.FieldName, field.FieldID)
				} else {
					fmt.Printf("✅ 必填字段已提供: %s\n", field.FieldName)
				}
			}
		}

		// 检查字段类型是否匹配并格式化字段值
		fmt.Println("🔍 检查字段类型是否匹配并格式化字段值...")
		for fieldName, fieldValue := range fields {
			// 查找对应的字段定义
			var fieldDef *models.Field
			for _, field := range tableFields {
				if field.FieldName == fieldName {
					fieldDef = &field
					break
				}
			}

			if fieldDef != nil {
				// 格式化字段值，特别是User类型字段
				if fieldValue != nil && fieldValue != "" {
					// 处理User类型字段（ui_type为User或field_type为11）
					if (fieldDef.UiType == "User" || fieldDef.FieldType == "11") && !strings.Contains(fmt.Sprintf("%T", fieldValue), "[]") {
						// 将普通字符串转换为User类型需要的格式: [{"id": "用户ID"}]
						userId := fmt.Sprintf("%v", fieldValue)
						fields[fieldName] = []interface{}{map[string]interface{}{"id": userId}}
						fmt.Printf("✅ User类型字段 '%s' 的值已格式化: %+v\n", fieldName, fields[fieldName])
					}
				}

				// 根据字段类型检查值
				switch fieldDef.FieldType {
				case "1": // 文本
					if fieldValue != nil && fmt.Sprintf("%v", fieldValue) == "" {
						fmt.Printf("⚠️ 文本字段 '%s' 的值为空\n", fieldName)
					}
				case "2": // 数字
					if _, ok := fieldValue.(float64); !ok && fieldValue != nil {
						fmt.Printf("⚠️ 数字字段 '%s' 的值类型不匹配，期望数字，实际: %T\n", fieldName, fieldValue)
					}
				case "3": // 单选
					if fieldValue == nil || fmt.Sprintf("%v", fieldValue) == "" {
						fmt.Printf("⚠️ 单选字段 '%s' 的值为空\n", fieldName)
					}
				case "4": // 多选
					if _, ok := fieldValue.([]interface{}); !ok && fieldValue != nil {
						fmt.Printf("⚠️ 多选字段 '%s' 的值类型不匹配，期望数组，实际: %T\n", fieldName, fieldValue)
					}
				case "5": // 日期
					if _, ok := fieldValue.(int64); !ok && fieldValue != nil {
						fmt.Printf("⚠️ 日期字段 '%s' 的值类型不匹配，期望时间戳，实际: %T\n", fieldName, fieldValue)
					}
				case "11": // 人员
					if _, ok := fieldValue.([]interface{}); !ok && fieldValue != nil {
						fmt.Printf("⚠️ 人员字段 '%s' 的值类型不匹配，期望数组，实际: %T\n", fieldName, fieldValue)
					}
				case "13": // 附件
					if _, ok := fieldValue.([]interface{}); !ok && fieldValue != nil {
						fmt.Printf("⚠️ 附件字段 '%s' 的值类型不匹配，期望数组，实际: %T\n", fieldName, fieldValue)
					}
				case "15": // 复选框
					if _, ok := fieldValue.(bool); !ok && fieldValue != nil {
						fmt.Printf("⚠️ 复选框字段 '%s' 的值类型不匹配，期望布尔值，实际: %T\n", fieldName, fieldValue)
					}
				case "17": // URL
					if fieldValue != nil && fmt.Sprintf("%v", fieldValue) == "" {
						fmt.Printf("⚠️ URL字段 '%s' 的值为空\n", fieldName)
					}
				case "18": // 邮箱
					if fieldValue != nil && fmt.Sprintf("%v", fieldValue) == "" {
						fmt.Printf("⚠️ 邮箱字段 '%s' 的值为空\n", fieldName)
					}
				case "19": // 电话
					if fieldValue != nil && fmt.Sprintf("%v", fieldValue) == "" {
						fmt.Printf("⚠️ 电话字段 '%s' 的值为空\n", fieldName)
					}
				case "20": // 进度
					if _, ok := fieldValue.(float64); !ok && fieldValue != nil {
						fmt.Printf("⚠️ 进度字段 '%s' 的值类型不匹配，期望数字，实际: %T\n", fieldName, fieldValue)
					}
				case "21": // 评分
					if _, ok := fieldValue.(float64); !ok && fieldValue != nil {
						fmt.Printf("⚠️ 评分字段 '%s' 的值类型不匹配，期望数字，实际: %T\n", fieldName, fieldValue)
					}
				case "23": // 货币
					if _, ok := fieldValue.(float64); !ok && fieldValue != nil {
						fmt.Printf("⚠️ 货币字段 '%s' 的值类型不匹配，期望数字，实际: %T\n", fieldName, fieldValue)
					}
				default:
					fmt.Printf("⚠️ 未知字段类型 '%s' 的字段 '%s'\n", fieldDef.FieldType, fieldName)
				}
			} else {
				fmt.Printf("⚠️ 未找到字段 '%s' 的定义\n", fieldName)
			}
		}
	}

	// 确保fields不为空
	if fields == nil {
		fields = make(map[string]interface{})
	}

	// 构建请求体 - 使用单条记录格式
	reqBody := map[string]interface{}{
		"fields": fields,
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("构建请求体失败: %w", err)
	}

	// 添加请求体调试日志
	fmt.Printf("📋 请求体: %s\n", string(reqBodyBytes))

	httpReq, err := http.NewRequest("POST", fieldsURL, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("添加记录失败: %w", err)
	}
	defer httpResp.Body.Close()

	httpBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	type AddRecordResponse struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Record struct {
				RecordID string `json:"record_id"`
			} `json:"record"`
		} `json:"data"`
	}

	var addResult AddRecordResponse
	if err := json.Unmarshal(httpBody, &addResult); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if addResult.Code != 0 {
		fmt.Printf("📋 添加记录API响应: %s\n", string(httpBody))

		// 尝试解析更详细的错误信息
		type ErrorResponse struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
			Data struct {
				ErrorDetails []struct {
					Field   string `json:"field"`
					Message string `json:"message"`
				} `json:"error_details,omitempty"`
			} `json:"data"`
		}

		var errorResp ErrorResponse
		if json.Unmarshal(httpBody, &errorResp) == nil {
			if len(errorResp.Data.ErrorDetails) > 0 {
				errorDetails := ""
				for _, detail := range errorResp.Data.ErrorDetails {
					errorDetails += fmt.Sprintf("字段 '%s': %s; ", detail.Field, detail.Message)
				}
				return "", fmt.Errorf("新增记录失败: %s (Code: %d). 详细错误: %s", addResult.Msg, addResult.Code, errorDetails)
			}
		}

		return "", fmt.Errorf("新增记录失败: %s (Code: %d)", addResult.Msg, addResult.Code)
	}

	if addResult.Data.Record.RecordID != "" {
		return addResult.Data.Record.RecordID, nil
	}

	return "", fmt.Errorf("新增记录失败: 未获取到记录ID")
}

// CheckFieldsCompleted 检查记录中的指定字段是否已完成，并返回字段值
// 优化：使用统一的Wiki Token处理函数，改进错误处理
func (s *LarkService) CheckFieldsCompleted(appToken, tableID, recordID string, checkFields []string) (bool, map[string]interface{}, error) {
	// 直接使用HTTP API获取记录，确保指定user_id_type=user_id
	token, err := s.GetTenantAccessToken()
	if err != nil {
		return false, nil, fmt.Errorf("获取访问令牌失败: %w", err)
	}

	// 检查 appToken 是否是 wiki token，如果是需要先获取 obj_token
	realAppToken := appToken
	isWiki, objType, objToken, wikiErr := s.getWikiTokenInfo(appToken, token)
	if wikiErr != nil {
		fmt.Printf("⚠️ Wiki Token处理警告: %v\n", wikiErr)
	}

	if isWiki {
		if objType == "bitable" && objToken != "" {
			fmt.Printf("✅ 检测到 Wiki Token，获取到 ObjToken: %s\n", objToken)
			realAppToken = objToken
		}
	}

	// 使用实际的 appToken 获取记录，确保使用user_id_type=user_id
	recordURL := fmt.Sprintf("https://open.feishu.cn/open-apis/bitable/v1/apps/%s/tables/%s/records/%s?user_id_type=user_id", realAppToken, tableID, recordID)

	httpReq, err := http.NewRequest("GET", recordURL, nil)
	if err != nil {
		return false, nil, fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return false, nil, fmt.Errorf("获取记录失败: %w", err)
	}
	defer httpResp.Body.Close()

	httpBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return false, nil, fmt.Errorf("读取响应失败: %w", err)
	}

	type GetRecordResponse struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Record struct {
				Fields map[string]interface{} `json:"fields"`
			} `json:"record"`
		} `json:"data"`
	}

	var getResult GetRecordResponse
	if err := json.Unmarshal(httpBody, &getResult); err != nil {
		return false, nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if getResult.Code != 0 {
		fmt.Printf("📋 获取记录API响应: %s\n", string(httpBody))
		return false, nil, fmt.Errorf("获取记录失败: %s (Code: %d)", getResult.Msg, getResult.Code)
	}

	// 检查字段是否都已完成，并收集字段值
	fieldValues := make(map[string]interface{})
	allCompleted := true

	for _, fieldName := range checkFields {
		value := getResult.Data.Record.Fields[fieldName]
		if value == nil || value == "" {
			allCompleted = false
			break
		}
		fieldValues[fieldName] = value
	}

	return allCompleted, fieldValues, nil
}

// getWikiTokenInfo 获取Wiki Token的实际AppToken信息
// 新增：统一处理Wiki Token的函数，避免重复代码
func (s *LarkService) getWikiTokenInfo(appToken, token string) (isWiki bool, objType string, objToken string, err error) {
	isWiki = s.IsWikiToken(appToken, token)
	if !isWiki {
		return false, "", "", nil
	}

	// 调用飞书Wiki API获取obj_token
	getNodeURL := fmt.Sprintf("https://open.feishu.cn/open-apis/wiki/v2/spaces/get_node?user_id_type=user_id&token=%s", appToken)
	_, nodeBody, err := s.handleHTTPRequest("GET", getNodeURL, token, nil)
	if err != nil {
		return true, "", "", fmt.Errorf("获取Wiki节点信息失败: %w", err)
	}

	type GetNodeResponse struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Node struct {
				ObjToken string `json:"obj_token"`
				ObjType  string `json:"obj_type"`
				Title    string `json:"title"`
			} `json:"node"`
		} `json:"data"`
	}

	var nodeResult GetNodeResponse
	if err := json.Unmarshal(nodeBody, &nodeResult); err != nil {
		return true, "", "", fmt.Errorf("解析Wiki节点响应失败: %w", err)
	}

	if nodeResult.Code != 0 {
		return true, "", "", fmt.Errorf("获取Wiki节点信息失败: %s (Code: %d)", nodeResult.Msg, nodeResult.Code)
	}

	return true, nodeResult.Data.Node.ObjType, nodeResult.Data.Node.ObjToken, nil
}



// GetRecord 获取记录的所有字段
// 优化：使用统一的Wiki Token处理函数，改进错误处理
func (s *LarkService) GetRecord(appToken, tableID, recordID string) (map[string]interface{}, error) {
	// 获取访问令牌
	token, err := s.GetTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("获取访问令牌失败: %w", err)
	}

	// 检查 appToken 是否是 wiki token，如果是需要先获取 obj_token
	realAppToken := appToken
	isWiki, objType, objToken, wikiErr := s.getWikiTokenInfo(appToken, token)
	if wikiErr != nil {
		fmt.Printf("⚠️ Wiki Token处理警告: %v\n", wikiErr)
	}

	if isWiki && objType == "bitable" && objToken != "" {
		fmt.Printf("✅ 检测到 Wiki Token，获取到 ObjToken: %s\n", objToken)
		realAppToken = objToken
	}

	// 首先尝试使用SDK获取记录
	req := larkbitable.NewGetAppTableRecordReqBuilder().
		AppToken(realAppToken).
		TableId(tableID).
		RecordId(recordID).
		Build()

	resp, err := s.GetClient().Bitable.AppTableRecord.Get(context.Background(), req)
	if err == nil && resp.Success() {
		if resp.Data != nil && resp.Data.Record != nil && resp.Data.Record.Fields != nil {
			return resp.Data.Record.Fields, nil
		}
		return nil, fmt.Errorf("获取记录失败: 未获取到记录数据")
	}

	// 如果SDK失败，使用HTTP API获取记录，确保指定user_id_type=user_id
	recordURL := fmt.Sprintf("https://open.feishu.cn/open-apis/bitable/v1/apps/%s/tables/%s/records/%s?user_id_type=user_id", realAppToken, tableID, recordID)

	// 使用通用HTTP请求处理函数
	httpResp, httpBody, err := s.handleHTTPRequest("GET", recordURL, token, nil)
	if err != nil {
		return nil, fmt.Errorf("获取记录失败: %w", err)
	}
	defer httpResp.Body.Close()

	type GetRecordResponse struct {
		Code int `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Record struct {
				Fields map[string]interface{} `json:"fields"`
			} `json:"record"`
		} `json:"data"`
	}

	var getResult GetRecordResponse
	if err := json.Unmarshal(httpBody, &getResult); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if getResult.Code != 0 {
		fmt.Printf("📋 获取记录API响应: %s\n", string(httpBody))
		return nil, fmt.Errorf("获取记录失败: %s (Code: %d)", getResult.Msg, getResult.Code)
	}

	return getResult.Data.Record.Fields, nil
}

// SendMessage 发送消息到指定群聊
func (s *LarkService) SendMessage(groupChatID, message string) error {
	return s.messageService.SendMessage(groupChatID, message)
}

// CreateTask 创建任务
func (s *LarkService) CreateTask(title string, dueTimestamp int64, isAllDay bool, assignees []map[string]interface{}) error {
	return s.taskService.CreateTask(title, dueTimestamp, isAllDay, assignees)
}