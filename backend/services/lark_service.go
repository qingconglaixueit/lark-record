package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"lark-record/models"
	"net/http"
	"sync"
	"time"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

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

type LarkService struct {
	appID      string
	appSecret  string
	client     *lark.Client
	httpClient *http.Client
	// Wiki转换结果缓存
	wikiConvertCache sync.Map
}

func NewLarkService(appID, appSecret string) *LarkService {
	return &LarkService{
		appID:      appID,
		appSecret:  appSecret,
		httpClient: &http.Client{Timeout: 30 * time.Second},
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
func (s *LarkService) GetBitableTables(appToken string) ([]models.TableInfo, error) {
	s.initClient()

	ctx := context.Background()

	// 尝试直接使用bitable API
	req := larkbitable.NewListAppTableReqBuilder().
		AppToken(appToken).
		Build()

	resp, err := s.client.Bitable.AppTable.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("获取数据表列表失败: %v", err)
	}

	// 如果返回91402错误，可能是wiki token，尝试处理
	if !resp.Success() {
		if resp.Code == 91402 || resp.Code == 99991663 {
			fmt.Println("🔍 检测到Wiki Token，尝试处理...")

			// 尝试通过HTTP API直接获取数据表
			tables, err := s.fetchWikiTablesDirectly(appToken)
			if err != nil {
				return []models.TableInfo{}, fmt.Errorf("Wiki链接处理失败: %v", err)
			}

			if len(tables) > 0 {
				fmt.Printf("✅ 成功获取到Wiki中的数据表: %d 个\n", len(tables))
				return tables, nil
			}

			return []models.TableInfo{}, nil
		} else {
			return nil, fmt.Errorf("获取数据表列表失败: %s (Code: %d)", resp.Msg, resp.Code)
		}
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

	return tables, nil
}

// fetchWikiTablesDirectly 直接通过HTTP API获取Wiki节点关联的数据表
func (s *LarkService) fetchWikiTablesDirectly(wikiToken string) ([]models.TableInfo, error) {
	token, err := s.getTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("获取访问令牌失败: %w", err)
	}

	// 第一步：通过 wiki token 获取节点信息（获取 obj_token）
	// 使用正确的接口: https://open.feishu.cn/open-apis/wiki/v2/spaces/get_node
	getNodeURL := fmt.Sprintf("https://open.feishu.cn/open-apis/wiki/v2/spaces/get_node?user_id_type=user_id&token=%s", wikiToken)
	nodeReq, err := http.NewRequest("GET", getNodeURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建节点请求失败: %w", err)
	}
	nodeReq.Header.Set("Authorization", "Bearer "+token)

	nodeResp, err := s.httpClient.Do(nodeReq)
	if err != nil {
		return nil, fmt.Errorf("获取节点信息失败: %w", err)
	}
	defer nodeResp.Body.Close()

	nodeBody, err := io.ReadAll(nodeResp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取节点响应失败: %w", err)
	}

	fmt.Printf("📋 获取到Wiki节点响应: %s\n", string(nodeBody))

	type GetNodeResponse struct {
		Code int    `json:"code"`
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
func (s *LarkService) searchAllBitablesInWiki(accessToken string) ([]models.TableInfo, error) {
	// 这个函数被调用时，我们不知道具体的wikiNodeToken，所以直接返回空
	// 因为fetchWikiTablesDirectly已经尝试过获取节点信息，如果失败，我们直接返回空
	fmt.Println("🔍 无法获取特定Wiki节点，返回空列表")
	return []models.TableInfo{}, nil
}

// fetchWikiSpaceTables 从Wiki空间获取所有bitable的数据表
func (s *LarkService) fetchWikiSpaceTables(wikiToken, accessToken string) ([]models.TableInfo, error) {
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
func (s *LarkService) searchNodeForTables(nodeToken, nodeTitle, accessToken string, isRoot bool, targetWikiToken string) ([]models.TableInfo, error) {
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
func (s *LarkService) searchChildrenForTables(nodeToken, nodeTitle, accessToken, targetWikiToken, spaceID string) ([]models.TableInfo, error) {
	var allTables []models.TableInfo

	// 获取子节点列表（使用正确的API，根据飞书开放文档）
	childrenURL := fmt.Sprintf("https://open.feishu.cn/open-apis/wiki/v2/spaces/%s/nodes?page_size=50&parent_node_token=%s", spaceID, nodeToken)
	childrenReq, err := http.NewRequest("GET", childrenURL, nil)
	if err != nil {
		return allTables, fmt.Errorf("创建子节点列表请求失败: %w", err)
	}
	childrenReq.Header.Set("Authorization", "Bearer "+accessToken)

	childrenResp, err := s.httpClient.Do(childrenReq)
	if err != nil {
		fmt.Printf("    ⚠️  获取子节点列表失败: %v\n", err)
		return allTables, nil
	}
	defer childrenResp.Body.Close()

	childrenBody, err := io.ReadAll(childrenResp.Body)
	if err != nil {
		fmt.Printf("    ⚠️  读取子节点响应失败: %v\n", err)
		return allTables, nil
	}

	// 打印原始响应以调试
	fmt.Printf("    📋 子节点原始响应(%d字节): %s\n", len(childrenBody), string(childrenBody))

	type ChildrenResponse struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Items []struct {
				NodeToken string `json:"node_token"`
				ObjToken  string `json:"obj_token"`
				ObjType   string `json:"obj_type"`
				Title     string `json:"title"`
				HasChild  bool   `json:"has_child"`
			} `json:"items"`
			HasMore   bool   `json:"has_more"`
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

// GetTableFields 获取数据表的所有字段
func (s *LarkService) GetTableFields(appToken, tableID string) ([]models.Field, error) {
	s.initClient()

	ctx := context.Background()

	// 首先检查 appToken 是否是 wiki token，如果是需要先获取 obj_token
	realAppToken := appToken
	
	// 尝试使用 SDK 获取字段，如果失败则可能需要处理 wiki token
	req := larkbitable.NewListAppTableFieldReqBuilder().
		AppToken(realAppToken).
		TableId(tableID).
		Build()

	resp, err := s.client.Bitable.AppTableField.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("获取字段列表失败: %v", err)
	}

	// 如果获取失败，可能是 wiki token，尝试HTTP API直接获取
	if !resp.Success() {
		fmt.Println("🔍 SDK获取失败，可能是 Wiki Token，尝试处理...")
		
		token, err := s.getTenantAccessToken()
		if err != nil {
			return nil, fmt.Errorf("获取访问令牌失败: %w", err)
		}

		// 尝试判断是否为 wiki token：如果以 "BEsNwa" 等开头，很可能是 wiki token
		// 或者尝试调用 wiki API 检查
		getNodeURL := fmt.Sprintf("https://open.feishu.cn/open-apis/wiki/v2/spaces/get_node?user_id_type=user_id&token=%s", appToken)
		nodeReq, err := http.NewRequest("GET", getNodeURL, nil)
		if err == nil {
			nodeReq.Header.Set("Authorization", "Bearer "+token)
			nodeResp, nodeErr := s.httpClient.Do(nodeReq)
			if nodeErr == nil {
				defer nodeResp.Body.Close()
				nodeBody, _ := io.ReadAll(nodeResp.Body)
				
				type GetNodeResponse struct {
					Code int    `json:"code"`
					Data struct {
						Node struct {
							ObjToken  string `json:"obj_token"`
							ObjType   string `json:"obj_type"`
							Title     string `json:"title"`
						} `json:"node"`
					} `json:"data"`
				}
				var nodeResult GetNodeResponse
				if json.Unmarshal(nodeBody, &nodeResult) == nil && nodeResult.Code == 0 {
					if nodeResult.Data.Node.ObjType == "bitable" && nodeResult.Data.Node.ObjToken != "" {
						fmt.Printf("✅ 检测到 Wiki Token，获取到 ObjToken: %s\n", nodeResult.Data.Node.ObjToken)
						realAppToken = nodeResult.Data.Node.ObjToken
					}
				}
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
			fields = append(fields, models.Field{
				FieldName: field.FieldName,
				FieldType: fmt.Sprintf("%d", field.Type),
				FieldID:   field.FieldId,
			})
		}
		fmt.Printf("✅ 成功获取到字段: %d 个\n", len(fields))
		return fields, nil
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

	return "", fmt.Errorf("新增记录失败: 未获取到记录ID")
}

// CheckFieldsCompleted 检查记录中的指定字段是否已完成
func (s *LarkService) CheckFieldsCompleted(appToken, tableID, recordID string, checkFields []string) (bool, error) {
	s.initClient()

	ctx := context.Background()

	req := larkbitable.NewGetAppTableRecordReqBuilder().
		AppToken(appToken).
		TableId(tableID).
		RecordId(recordID).
		Build()

	resp, err := s.client.Bitable.AppTableRecord.Get(ctx, req)
	if err != nil {
		return false, fmt.Errorf("获取记录失败: %v", err)
	}

	if !resp.Success() {
		return false, fmt.Errorf("获取记录失败: %s", resp.Msg)
	}

	if resp.Data == nil || resp.Data.Record == nil {
		return false, fmt.Errorf("记录数据为空")
	}

	// 检查字段是否都已完成
	record := resp.Data.Record
	for _, fieldName := range checkFields {
		value := record.Fields[fieldName]
		if value == nil || value == "" {
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

// getTenantAccessToken 获取租户访问令牌
func (s *LarkService) getTenantAccessToken() (string, error) {
	reqBody := map[string]string{
		"app_id":     s.appID,
		"app_secret": s.appSecret,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("请求体序列化失败: %w", err)
	}

	req, err := http.NewRequest(
		"POST",
		"https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal",
		bytes.NewReader(jsonData),
	)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	type TenantAccessTokenResponse struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		Expire            int    `json:"expire"`
		TenantAccessToken string `json:"tenant_access_token"`
	}

	var result TenantAccessTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 0 {
		return "", fmt.Errorf("获取令牌失败: %s (code: %d)", result.Msg, result.Code)
	}

	return result.TenantAccessToken, nil
}
