# 飞书记录助手

一个功能强大的Chrome浏览器扩展，帮助用户快速将数据记录到飞书多维表格，并在指定字段完成时自动发送消息通知。

## 功能特性

- 🔧 **飞书应用配置**：支持配置飞书应用ID和密钥，验证权限
- 📊 **多维表格管理**：支持访问知识库、个人云文档和用户创建的多维表格
- ✍️ **数据快速录入**：提供简洁的用户界面，快速录入数据到指定字段
- 🔍 **字段检测**：自动检测记录中其他字段是否已完成
- 📢 **消息通知**：当指定字段全部有值时，自动发送消息到飞书群聊

## 技术架构

### 前端技术栈
- **Chrome Extension API**：Chrome扩展开发框架
- **Vanilla JavaScript**：原生JavaScript，无需依赖
- **CSS3**：现代CSS，响应式设计

### 后端技术栈
- **Go 1.21+**：后端服务语言
- **Gin框架**：高性能HTTP框架
- **飞书开放平台SDK**：官方Go SDK (github.com/larksuite/oapi-sdk-go/v3)
- **CORS支持**：支持跨域请求

## 项目结构

```
lark-record/
├── manifest.json              # Chrome扩展配置文件
├── README.md                  # 项目说明文档
├── backend/                   # Go后端服务
│   ├── main.go               # 后端入口文件
│   ├── go.mod                # Go模块依赖
│   ├── handlers/             # HTTP请求处理器
│   │   └── lark_handler.go  # 飞书相关接口处理
│   ├── services/             # 业务逻辑层
│   │   └── lark_service.go  # 飞书API封装
│   └── models/               # 数据模型
│       └── config.go        # 配置数据结构
├── options/                  # 配置页面
│   ├── options.html         # 配置页面HTML
│   └── options.js           # 配置页面逻辑
├── popup/                   # 弹窗页面
│   ├── popup.html          # 弹窗HTML
│   └── popup.js            # 弹窗逻辑
├── background/              # 后台服务
│   └── background.js       # 扩展后台脚本
├── styles/                  # 样式文件
│   ├── options.css         # 配置页面样式
│   └── popup.css           # 弹窗样式
└── icons/                   # 图标文件
```



## 安装步骤

### 前置要求

1. **Chrome浏览器**：版本88或更高
2. **Go环境**：版本1.21或更高
3. **（可选）飞书开发者账号**：如果需要使用自己的应用，可创建飞书应用

### 安装后端服务

1. 克隆或下载项目到本地
2. 进入后端目录并安装依赖：
```bash
cd backend
go mod tidy
```

3. 启动后端服务：
```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动。

### 安装Chrome扩展

1. 打开Chrome浏览器，访问 `chrome://extensions/`
2. 启用"开发者模式"（右上角开关）
3. 点击"加载已解压的扩展程序"
4. 选择项目根目录（包含manifest.json的目录）
5. 扩展安装成功后，会在浏览器工具栏看到扩展图标

## 使用指南

### 1. 配置飞书应用

#### 创建飞书应用

1. 访问[飞书开放平台](https://open.feishu.cn/)
2. 创建应用，获取 `App ID` 和 `App Secret`
3. 在应用权限中添加以下权限：
   - `bitable:app`：读取多维表格
   - `bitable:app:readonly`：只读多维表格
   - `drive:drive`：访问云文档
   - `im:message`：发送消息

#### 配置扩展

1. 点击扩展图标，选择“选项”或右键点击扩展选择“选项”
2. 手动填写飞书应用的 `App ID` 和 `App Secret`
3. 点击“测试配置”验证配置是否正确
4. 配置通过后，添加你的多维表格URL（支持直接表格和知识库表格）

### 2. 选择多维表格和字段

1. 从列表中选择要使用的多维表格
2. 从下拉菜单中选择数据表
3. **待写入字段**：勾选用户在Popup中需要输入的字段（至少一个）
4. **需检测的字段**：勾选需要检测是否完成的字段
5. **群聊ID**：填写要接收通知的飞书群聊ID（可选）
6. 点击"保存配置"

### 3. 记录数据

1. 点击扩展图标打开Popup
2. 如果已配置，会显示多维表格列表
3. 选择要使用的表格
4. 在输入框中填写待写入字段的值
5. 点击"记录数据"提交
6. 成功后，数据会被添加到飞书多维表格

### 4. 消息通知

当记录的其他字段（需检测的字段）全部有值时：
- 后端会自动检测到字段完成
- 向配置的群聊发送完成通知消息
- 通知内容包含记录ID等信息

## 详细使用指南

### 一、后端服务使用

#### 启动后端服务

1. 确保已安装Go 1.21+环境：
   - [Go下载地址](https://go.dev/dl/)
   - 验证安装：`go version`

2. 进入后端目录：
```bash
cd backend
```

3. 安装依赖：
```bash
go mod tidy
```

4. 启动服务：
```bash
go run main.go
```

5. 验证服务是否运行：
   - 访问 `http://localhost:8080`
   - 应看到 "飞书记录助手后端服务运行中" 的提示

#### 后端配置

- 配置文件：`backend/config.json`
- 默认端口：`8080`，可在 `main.go` 中修改

#### 后端日志

- 日志文件：`backend/server.log`
- 包含API请求、飞书交互和错误信息

### 二、前端扩展使用

#### 配置页面

1. 打开扩展选项页面：
   - 点击浏览器工具栏中的扩展图标
   - 选择"选项"
   - 或右键点击扩展图标，选择"选项"

2. 飞书应用配置：
   - 使用内置配置或手动输入App ID和App Secret
   - 点击"测试配置"验证

3. 多维表格配置：
   - 添加多维表格URL
   - 选择表格和数据表
   - 配置字段和群聊

#### 弹窗页面

1. 打开弹窗：
   - 点击浏览器工具栏中的扩展图标

2. 记录数据：
   - 选择多维表格
   - 填写字段值
   - 点击"记录数据"提交

### 三、飞书应用权限设置

#### 1. 进入飞书开放平台

- 访问：[飞书开放平台](https://open.feishu.cn/)
- 登录飞书账号

#### 2. 创建或选择应用

- 创建应用：[创建飞书应用指南](https://open.feishu.cn/document/home/introduction-to-application-creation/create-an-application)
- 或选择已创建的应用

#### 3. 配置应用权限

1. 进入应用的"权限管理"页面
2. 添加以下权限：

| 权限名称 | 权限ID | 权限范围 | 操作步骤 |
|---------|--------|---------|--------|
| 读取多维表格 | `bitable:app` | 应用 | [添加权限指南](https://open.feishu.cn/document/uAjLw4CM/ugTN1YjL4UTN24CO1UjN/trouble-shooting/how-to-apply-for-permissions#359b9f0a) |
| 只读多维表格 | `bitable:app:readonly` | 应用 | [添加权限指南](https://open.feishu.cn/document/uAjLw4CM/ugTN1YjL4UTN24CO1UjN/trouble-shooting/how-to-apply-for-permissions#359b9f0a) |
| 访问云文档 | `drive:drive` | 应用 | [添加权限指南](https://open.feishu.cn/document/uAjLw4CM/ugTN1YjL4UTN24CO1UjN/trouble-shooting/how-to-apply-for-permissions#359b9f0a) |
| 发送消息 | `im:message` | 应用 | [添加权限指南](https://open.feishu.cn/document/uAjLw4CM/ugTN1YjL4UTN24CO1UjN/trouble-shooting/how-to-apply-for-permissions#359b9f0a) |

3. 点击"申请权限"按钮
4. 等待权限审核通过（通常即时通过）

#### 4. 配置安全设置

1. 进入"安全设置"页面
2. 配置"重定向URL"（如需要）
3. 配置"IP白名单"（如需要）

### 四、获取飞书用户ID

#### 方法一：从飞书客户端获取

1. 打开飞书客户端
2. 点击左上角头像
3. 选择"设置" -> "账号与安全"
4. 在"飞书ID"处查看

#### 方法二：通过飞书API获取

- 调用用户信息API：[获取用户信息API文档](https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/contact-v3/user/get)
- 需要 `contact:user.base:readonly` 权限

### 五、创建飞书群聊

#### 创建群聊

1. 打开飞书客户端
2. 点击左侧"消息"面板
3. 点击右上角"+"号
4. 选择"创建群聊"
5. 选择群成员
6. 设置群名称
7. 点击"创建"

#### 获取群聊ID

1. 打开飞书客户端
2. 进入目标群聊
3. 右键点击群聊名称
4. 选择"复制群ID"
5. 将复制的群ID粘贴到配置页面

更多详情：[飞书群聊管理指南](https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/chat/overview)

### 六、将应用添加为多维表格协作者

#### 1. 打开多维表格

- 在飞书客户端或网页端打开目标多维表格

#### 2. 打开协作设置

1. 点击右上角"分享"按钮
2. 或点击"设置" -> "协作设置"

#### 3. 添加应用为协作者

1. 在协作设置中，点击"添加协作者"
2. 选择"按成员/应用添加"
3. 在搜索框中输入应用名称
4. 选择应用
5. 设置协作权限（建议设置为"编辑者"或"查看者"）
6. 点击"添加"

更多详情：[多维表格协作指南](https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/bitable-v1/app-collaborator/overview)

## 配置说明

### 飞书应用权限

确保飞书应用具有以下权限：

| 权限名称     | 权限ID               | 说明                 |
| ------------ | -------------------- | -------------------- |
| 获取多维表格 | bitable:app          | 读取和管理多维表格   |
| 只读多维表格 | bitable:app:readonly | 读取多维表格数据     |
| 云文档访问   | drive:drive          | 访问个人和知识库文档 |
| 发送消息     | im:message           | 向群聊发送消息       |

### 获取群聊ID

1. 打开飞书客户端
2. 进入目标群聊
3. 右键点击群聊名称，选择"复制群ID"
4. 将复制的群ID粘贴到配置页面

## 开发指南

### 启动开发环境

1. **启动后端服务**：
```bash
cd backend
go run main.go
```

2. **加载扩展**：
   - 打开Chrome扩展管理页面
   - 加载已解压的扩展程序
   - 选择项目目录

3. **调试扩展**：
   - Popup页面：右键点击扩展图标 -> 检查弹出内容
   - 后台脚本：在扩展管理页面点击"Service Worker"查看日志
   - 配置页面：打开配置页面后按F12

### 修改后端API端口

如需修改后端服务端口，编辑 [`backend/main.go`](backend/main.go:40)：

```go
log.Println("服务器启动在 :8080 端口")
if err := r.Run(":8080"); err != nil {
    log.Fatalf("服务器启动失败: %v", err)
}
```

同时需要修改前端API调用地址：
- [`options/options.js`](options/options.js:21) 和 [`popup/popup.js`](popup/popup.js:63) 中的 `http://localhost:8080`

### 添加新功能

1. **后端**：在 [`backend/handlers/`](backend/handlers/) 添加新的处理器
2. **前端**：在 [`options/options.js`](options/options.js) 或 [`popup/popup.js`](popup/popup.js) 添加新的UI和逻辑
3. **API路由**：在 [`backend/main.go`](backend/main.go) 注册新的路由

## 常见问题

### Q1: 配置测试失败怎么办？

**A:** 请检查以下几点：
- 确认飞书应用ID和密钥是否正确
- 确认飞书应用已启用相关权限
- 确认后端服务已启动（http://localhost:8080）
- 检查浏览器控制台是否有错误信息

### Q2: 找不到多维表格？

**A:** 请确认：
- 飞书应用是否已授予相关权限
- 用户是否有权限访问目标多维表格
- 多维表格是否存在于知识库、个人云文档或用户创建的文档中

### Q3: 数据记录成功但没有收到消息通知？

**A:** 请检查：
- 是否配置了"需检测的字段"
- 这些字段是否全部有值
- 是否正确配置了群聊ID
- 飞书应用是否有发送消息权限

### Q4: 扩展显示"未配置"？

**A:** 请确认：
- 是否已完成飞书应用配置
- 是否已保存配置
- 配置中的必填项是否都已填写

### Q5: 如何获取飞书App ID和App Secret？

**A:** 
1. 访问 [飞书开放平台](https://open.feishu.cn/)
2. 登录并进入"应用管理"
3. 创建应用或选择已有应用
4. 在"凭证与基础信息"页面查看App ID和App Secret

## 注意事项

1. **后端服务**：使用扩展前必须先启动后端服务
2. **权限管理**：确保飞书应用具有足够的权限
3. **数据安全**：App Secret等敏感信息请妥善保管
4. **网络环境**：确保能访问飞书API（open.feishu.cn）
5. **字段类型**：确保字段类型与输入数据类型匹配

## 技术支持

如有问题或建议，请通过以下方式联系：
- 提交Issue到项目仓库
- 发送邮件至开发者邮箱

## 许可证

本项目采用 MIT 许可证。

## 更新日志

### v1.0.0 (2024-01-16)
- ✨ 初始版本发布
- ✅ 实现飞书API对接
- ✅ 实现多维表格管理
- ✅ 实现数据快速录入
- ✅ 实现字段检测和消息通知

## 贡献指南

欢迎提交Pull Request来改进本项目！

## 致谢

- [飞书开放平台](https://open.feishu.cn/) - 提供API支持
- [Gin框架](https://gin-gonic.com/) - 高性能HTTP框架
- [Chrome扩展文档](https://developer.chrome.com/docs/extensions/) - 扩展开发指南