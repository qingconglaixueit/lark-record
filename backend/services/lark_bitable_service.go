package services

import (
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
	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
)

// LarkBitableService 处理飞书多维表格相关操作
type LarkBitableService struct {
	BaseService
	bitablesCache     sync.Map
	bitablesCacheTime sync.Map
	tablesCache       sync.Map
	tablesCacheTime   sync.Map
}

// NewLarkBitableService 创建新的LarkBitableService实例
func NewLarkBitableService(appID, appSecret string) *LarkBitableService {
	baseService := NewBaseService(appID, appSecret)
	return &LarkBitableService{
		BaseService: baseService,
	}
}

// initClient 初始化飞书客户端（如果未初始化）
func (s *LarkBitableService) initClient() {
	if s.client == nil {
		s.client = lark.NewClient(s.appID, s.appSecret)
	}
}

// getTenantAccessToken 获取租户访问令牌，使用BaseService中的统一实现
func (s *LarkBitableService) getTenantAccessToken() (string, error) {
	return s.GetTenantAccessToken()
}

// GetBitables 获取用户有权限访问的所有多维表格（带缓存）
func (s *LarkBitableService) GetBitables() ([]models.Bitable, error) {
	// 检查缓存
	if cachedBitables, ok := s.bitablesCache.Load("all"); ok {
		if cachedTime, ok := s.bitablesCacheTime.Load("all"); ok {
			// 缓存有效期
			if time.Since(cachedTime.(time.Time)) < BitablesCacheExpireTime {
				fmt.Println("使用缓存的多维表格列表")
				return cachedBitables.([]models.Bitable), nil
			}
		}
	}
	
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

	// 缓存结果
	s.bitablesCache.Store("all", bitables)
	s.bitablesCacheTime.Store("all", time.Now())
	
	return bitables, nil
}

// GetBitableTables 获取多维表格中的所有数据表（带缓存）
func (s *LarkBitableService) GetBitableTables(appToken string, isWiki bool) ([]models.TableInfo, error) {
	// 缓存键
	cacheKey := fmt.Sprintf("%s:%t", appToken, isWiki)
	
	// 检查缓存
	if cachedTables, ok := s.tablesCache.Load(cacheKey); ok {
		if cachedTime, ok := s.tablesCacheTime.Load(cacheKey); ok {
			// 缓存有效期
			if time.Since(cachedTime.(time.Time)) < TablesCacheExpireTime {
				fmt.Println("使用缓存的数据表列表")
				return cachedTables.([]models.TableInfo), nil
			}
		}
	}
	
	s.initClient()

	// 如果URL中包含"wiki"字符串，直接处理为Wiki表格
	isWikiToken := isWiki || strings.Contains(appToken, "wiki") || strings.Contains(appToken, "Wiki")
	
	if isWikiToken {
		fmt.Println("🔍 检测到Wiki链接，直接处理...")
		// 尝试通过HTTP API直接获取数据表
		tables, err := s.fetchWikiTablesDirectly(appToken)
		if err != nil {
			return []models.TableInfo{}, fmt.Errorf("Wiki链接处理失败: %v", err)
		}

		if len(tables) > 0 {
			fmt.Printf("✅ 成功获取到Wiki中的数据表: %d 个\n", len(tables))
			// 缓存结果
			s.tablesCache.Store(cacheKey, tables)
			s.tablesCacheTime.Store(cacheKey, time.Now())
			return tables, nil
		}

		return []models.TableInfo{}, nil
	}

	// 否则，尝试直接使用bitable API
	ctx := context.Background()
	req := larkbitable.NewListAppTableReqBuilder().
		AppToken(appToken).
		Build()

	resp, err := s.client.Bitable.AppTable.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("获取数据表列表失败: %v", err)
	}

	if !resp.Success() {
		return nil, fmt.Errorf("获取数据表列表失败: %s (Code: %d)", resp.Msg, resp.Code)
	}

	var tables []models.TableInfo
	if resp.Data != nil && resp.Data.Items != nil {
		for _, table := range resp.Data.Items {
			tableID := ""
			tableName := ""

			if table.TableId != nil {
				tableID = *table.TableId
			}
			if table.Name != nil {
				tableName = *table.Name
			}

			if tableID != "" {
				tables = append(tables, models.TableInfo{
					TableID: tableID,
					Name:    tableName,
				})
			}
		}
	} else {
		fmt.Println("⚠️  数据响应为空或items为空")
	}

	// 确保总是返回空数组而不是nil
	if tables == nil {
		tables = []models.TableInfo{}
	}

	// 缓存结果
	s.tablesCache.Store(cacheKey, tables)
	s.tablesCacheTime.Store(cacheKey, time.Now())
	
	return tables, nil
}

// fetchWikiTablesDirectly 直接通过HTTP API获取Wiki节点关联的数据表
func (s *LarkBitableService) fetchWikiTablesDirectly(wikiToken string) ([]models.TableInfo, error) {
	token, err := s.getTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("获取访问令牌失败: %w", err)
	}

	// 第一步：通过 wiki token 获取节点信息（获取 obj_token）
	// 使用正确的接口: https://open.feishu.cn/open-apis/wiki/v2/spaces/get_node
	getNodeURL := fmt.Sprintf("https://open.feishu.cn/open-apis/wiki/v2/spaces/get_node?user_id_type=user_id&token=%s", wikiToken)
	_, nodeBody, err := s.handleHTTPRequest("GET", getNodeURL, token, nil)
	if err != nil {
		return nil, fmt.Errorf("获取节点信息失败: %w", err)
	}

	fmt.Printf("📋 获取到Wiki节点响应: %s\n", string(nodeBody))

	type GetNodeResponse struct {
		Code int `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Node struct {
				NodeToken string `json:"node_token"`
				ObjToken  string `json:"obj_token"`
				ObjType   string `json:"obj_type"`
				Title     string `json:"title"`
				SpaceID   string `json:"space_id"`
			} `json:"node"`
		} `json:"data"`
	}

	var nodeResult GetNodeResponse

	if err := json.Unmarshal(nodeBody, &nodeResult); err != nil {
		return nil, fmt.Errorf("解析节点信息失败: %w", err)
	}

	if nodeResult.Code != 0 {
		return nil, fmt.Errorf("获取节点信息失败: %s (Code: %d)", nodeResult.Msg, nodeResult.Code)
	}

	node := nodeResult.Data.Node
	fmt.Printf("✅ 获取到节点: 标题=%s, ObjType=%s, ObjToken=%s, SpaceID=%s\n", node.Title, node.ObjType, node.ObjToken, node.SpaceID)

	// 第二步：如果节点本身是bitable，使用 obj_token 作为 app_token 获取数据表
	if node.ObjType == "bitable" && node.ObjToken != "" {
		return s.fetchBitableTables(node.ObjToken, node.Title, token)
	}

	// 如果节点不是bitable，尝试搜索子节点
	fmt.Printf("🔍 节点类型为 %s，尝试搜索子节点...\n", node.ObjType)
	tables, err := s.searchChildrenForTables(node.NodeToken, node.Title, token, wikiToken, node.SpaceID)
	if err != nil {
		fmt.Printf("⚠️ 搜索子节点失败: %v\n", err)
		return []models.TableInfo{}, nil
	}
	if len(tables) > 0 {
		return tables, nil
	}

	return []models.TableInfo{}, fmt.Errorf("未找到多维表格数据表")
}

// searchAllBitablesInWiki 从Wiki空间搜索所有bitable节点
func (s *LarkBitableService) searchAllBitablesInWiki(accessToken string) ([]models.TableInfo, error) {
	// 这个函数被调用时，我们不知道具体的wikiNodeToken，所以直接返回空
	// 因为fetchWikiTablesDirectly已经尝试过获取节点信息，如果失败，我们直接返回空
	fmt.Println("🔍 无法获取特定Wiki节点，返回空列表")
	return []models.TableInfo{}, nil
}

// fetchWikiSpaceTables 从Wiki空间获取所有bitable的数据表
func (s *LarkBitableService) fetchWikiSpaceTables(wikiToken, accessToken string) ([]models.TableInfo, error) {
	// 使用传入的wikiToken作为space_id
	wikiSpaceID := wikiToken

	// 获取Wiki空间的节点列表（根节点）
	nodesURL := fmt.Sprintf("https://open.feishu.cn/open-apis/wiki/v2/spaces/%s/nodes?page_size=50&user_id_type=user_id", wikiSpaceID)
	req, err := http.NewRequest("GET", nodesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建节点列表请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取节点列表失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取节点列表响应失败: %w", err)
	}

	fmt.Printf("📋 获取到Wiki空间节点列表: %s\n", string(body))

	var nodesResult WikiNodesResponse
	if err := json.Unmarshal(body, &nodesResult); err != nil {
		return nil, fmt.Errorf("解析节点列表失败: %w", err)
	}

	if nodesResult.Code != 0 {
		return nil, fmt.Errorf("获取节点列表失败: %s (Code: %d)", nodesResult.Msg, nodesResult.Code)
	}

	// 查找所有bitable节点并尝试获取数据表
	fmt.Printf("🔍 在 %d 个根节点中查找bitable节点并递归搜索子节点...\n", len(nodesResult.Data.Items))

	var allTables []models.TableInfo

	// 直接遍历所有根节点，对于有子节点的节点，递归搜索所有bitable
	for _, node := range nodesResult.Data.Items {
		if node.HasChild {
			fmt.Printf("🔍 根节点 '%s' 有子节点，开始递归搜索...\n", node.Title)
			tables, err := s.searchChildrenForTables(node.NodeToken, node.Title, accessToken, wikiToken, wikiSpaceID)
			if err != nil {
				fmt.Printf("⚠️  搜索节点 '%s' 失败: %v\n", node.Title, err)
				continue
			}
			allTables = append(allTables, tables...)
		}

		// 如果这个节点本身就是bitable，获取其数据表
		if node.ObjType == "bitable" && node.ObjToken != "" {
			fmt.Printf("✅ 根节点本身是Bitable: %s\n", node.Title)
			tables, err := s.fetchBitableTables(node.ObjToken, node.Title, accessToken)
			if err != nil {
				fmt.Printf("⚠️  获取 '%s' 的数据表失败: %v\n", node.Title, err)
			} else {
				allTables = append(allTables, tables...)
			}
		}
	}

	if len(allTables) > 0 {
		fmt.Printf("✅ 成功获取到Wiki空间中的所有数据表: %d 个\n", len(allTables))
		return allTables, nil
	}

	return nil, fmt.Errorf("在Wiki空间的 %d 个节点中，未找到包含数据表的Bitable节点。\n\n建议：\n1. 检查飞书应用的权限设置\n2. 或使用第二个链接（直接多维表格链接）", len(nodesResult.Data.Items))
}

// searchNodeForTables 递归搜索节点及其子节点中的bitable数据表
func (s *LarkBitableService) searchNodeForTables(nodeToken, nodeTitle, accessToken string, isRoot bool, targetWikiToken string) ([]models.TableInfo, error) {
	var allTables []models.TableInfo

	if isRoot {
		fmt.Printf("  ↳ 处理根节点: %s (%s)\n", nodeTitle, nodeToken)
	}

	// 如果匹配目标Wiki Token，优先处理
	if nodeToken == targetWikiToken {
		// 获取子节点（使用children API）
		// 由于这里无法获取到正确的 space_id，返回空列表
		fmt.Printf("⚠️ searchNodeForTables 中无法获取 space_id，跳过处理")
		return []models.TableInfo{}, nil
	}

	// 直接从 WikiNodesResponse 结构中访问节点信息，避免重复API调用
	// 这里我们不再单独获取节点信息，而是使用已有的数据
	// 如果需要遍历子节点，使用 children API

	return allTables, nil
}

// searchChildrenForTables 获取节点的子节点并搜索其中的bitable数据表
func (s *LarkBitableService) searchChildrenForTables(nodeToken, nodeTitle, accessToken, targetWikiToken, spaceID string) ([]models.TableInfo, error) {
	var allTables []models.TableInfo

	// 获取子节点列表（使用正确的API，根据飞书开放文档）
	childrenURL := fmt.Sprintf("https://open.feishu.cn/open-apis/wiki/v2/spaces/%s/nodes?page_size=50&parent_node_token=%s", spaceID, nodeToken)
	_, childrenBody, err := s.handleHTTPRequest("GET", childrenURL, accessToken, nil)
	if err != nil {
		fmt.Printf("    ⚠️  获取子节点列表失败: %v\n", err)
		return allTables, nil
	}

	// 打印原始响应以调试
	fmt.Printf("    📋 子节点原始响应(%d字节): %s\n", len(childrenBody), string(childrenBody))

	type ChildrenResponse struct {
		Code int `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Items []struct {
				NodeToken string `json:"node_token"`
				ObjToken  string `json:"obj_token"`
				ObjType   string `json:"obj_type"`
				Title     string `json:"title"`
				HasChild  bool `json:"has_child"`
			} `json:"items"`
			HasMore   bool `json:"has_more"`
			PageToken string `json:"page_token"`
		} `json:"data"`
	}

	var childrenResult ChildrenResponse
	if err := json.Unmarshal(childrenBody, &childrenResult); err != nil {
		fmt.Printf("    ⚠️  解析子节点列表失败: %v\n", err)
		return allTables, nil
	}

	if childrenResult.Code == 0 && len(childrenResult.Data.Items) > 0 {
		fmt.Printf("    🔍 找到 %d 个子节点\n", len(childrenResult.Data.Items))
		for _, child := range childrenResult.Data.Items {
			// 如果子节点是bitable，直接获取其数据表
			if child.ObjType == "bitable" && child.ObjToken != "" {
				fmt.Printf("      ✅ 找到Bitable子节点: %s\n", child.Title)
				tables, err := s.fetchBitableTables(child.ObjToken, child.Title, accessToken)
				if err != nil {
					fmt.Printf("        ⚠️  获取 '%s' 的数据表失败: %v\n", child.Title, err)
					continue
				}
				allTables = append(allTables, tables...)
			}
			// 如果子节点还有子节点，递归搜索（这里暂时不递归，避免深度过大）
		}
	} else if childrenResult.Code != 0 {
		fmt.Printf("    ⚠️  获取子节点列表失败: %s (Code: %d)\n", childrenResult.Msg, childrenResult.Code)
	}

	return allTables, nil
}

// fetchBitableTables 获取指定bitable的所有数据表
func (s *LarkBitableService) fetchBitableTables(bitableToken, bitableName, accessToken string) ([]models.TableInfo, error) {
	fmt.Printf("✅ 找到Bitable节点: 标题=%s, ObjToken=%s\n", bitableName, bitableToken)

	// 尝试获取这个bitable的数据表列表
	tablesURL := fmt.Sprintf("https://open.feishu.cn/open-apis/bitable/v1/apps/%s/tables", bitableToken)
	_, tablesBody, err := s.handleHTTPRequest("GET", tablesURL, accessToken, nil)
	if err != nil {
		return nil, fmt.Errorf("获取数据表列表失败: %w", err)
	}

	type TablesResponse struct {
		Code int `json:"code"`
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