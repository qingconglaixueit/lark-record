# 飞书记录助手 - 项目结构说明

本文档详细说明了飞书记录助手项目的文件结构和各文件的用途。

## 📁 项目根目录

```
lark-record/
├── manifest.json              # Chrome扩展配置文件（必需）
├── README.md                  # 项目完整文档
├── QUICKSTART.md              # 快速入门指南
├── TESTING.md                 # 测试指南
├── PROJECT_STRUCTURE.md       # 项目结构说明（本文件）
├── .gitignore                 # Git忽略文件配置
├── start.bat                  # Windows启动脚本
└── start.sh                   # Linux/Mac启动脚本
```

### manifest.json
- **用途**：Chrome扩展的核心配置文件
- **功能**：
  - 定义扩展名称、版本、描述
  - 配置权限（storage, activeTab）
  - 设置host_permissions（飞书API域名）
  - 配置popup页面、选项页面、后台脚本
  - 设置图标文件

### README.md
- **用途**：项目的完整说明文档
- **功能**：
  - 功能特性介绍
  - 技术架构说明
  - 安装和使用指南
  - 常见问题解答

### QUICKSTART.md
- **用途**：快速入门指南
- **功能**：
  - 5步快速安装流程
  - 简化的配置说明
  - 基本使用方法

### TESTING.md
- **用途**：测试指南
- **功能**：
  - 完整的测试流程
  - 功能测试用例
  - 问题排查指南
  - 测试报告模板

### .gitignore
- **用途**：Git版本控制忽略配置
- **功能**：
  - 忽略编译产物
  - 忽略临时文件
  - 忽略操作系统文件

### start.bat / start.sh
- **用途**：后端服务启动脚本
- **功能**：
  - 检查Go环境
  - 安装依赖
  - 启动后端服务

## 📁 backend/ - Go后端服务

```
backend/
├── main.go                   # 后端入口文件
├── go.mod                     # Go模块依赖定义
├── go.sum                     # Go模块依赖锁定
├── handlers/                  # HTTP请求处理器
│   └── lark_handler.go       # 飞书相关接口处理
├── services/                  # 业务逻辑层
│   └── lark_service.go       # 飞书API封装
└── models/                    # 数据模型
    └── config.go             # 配置数据结构
```

### main.go
- **用途**：后端服务的入口文件
- **功能**：
  - 初始化Gin路由
  - 配置CORS
  - 注册API路由
  - 启动HTTP服务（端口8080）
- **关键路由**：
  - POST /api/config - 保存配置
  - GET /api/config - 获取配置
  - GET /api/bitables - 获取多维表格
  - GET /api/bitables/tables - 获取数据表
  - GET /api/bitables/fields - 获取字段
  - POST /api/records - 新增记录
  - GET /api/records/check - 检查记录状态

### go.mod
- **用途**：Go模块依赖定义
- **依赖包**：
  - github.com/gin-gonic/gin - Web框架
  - github.com/gin-contrib/cors - CORS中间件
  - github.com/larksuite/oapi-sdk-go/v3 - 飞书SDK

### handlers/lark_handler.go
- **用途**：处理飞书相关的HTTP请求
- **主要函数**：
  - `SaveConfig()` - 保存飞书配置
  - `GetConfig()` - 获取飞书配置
  - `GetBitables()` - 获取多维表格列表
  - `GetBitableTables()` - 获取数据表列表
  - `GetTableFields()` - 获取字段列表
  - `AddRecord()` - 新增记录
  - `CheckRecordStatus()` - 检查记录状态
- **全局变量**：
  - `configCache` - 配置缓存
  - `cacheMutex` - 缓存锁

### services/lark_service.go
- **用途**：封装飞书API调用
- **主要函数**：
  - `GetAccessToken()` - 获取访问令牌
  - `GetBitables()` - 获取多维表格
  - `GetBitableTables()` - 获取数据表
  - `GetTableFields()` - 获取字段
  - `AddRecord()` - 新增记录
  - `GetRecord()` - 获取记录
  - `CheckFieldsCompleted()` - 检查字段是否完成
  - `SendMessage()` - 发送消息到群聊

### models/config.go
- **用途**：定义数据结构
- **主要结构体**：
  - `Config` - 飞书配置
  - `Bitable` - 多维表格
  - `Field` - 表格字段
  - `Record` - 记录
  - `AddRecordRequest` - 新增记录请求
  - `SendMessageRequest` - 发送消息请求

## 📁 options/ - 配置页面

```
options/
├── options.html              # 配置页面HTML
└── options.js                # 配置页面逻辑
```

### options.html
- **用途**：配置页面的HTML结构
- **主要区域**：
  1. 飞书应用配置（App ID、Secret）
  2. 多维表格选择
  3. 数据表选择
  4. 字段选择（待写入字段、需检测字段）
  5. 消息通知配置（群聊ID）
  6. 保存配置
  7. 当前配置显示

### options.js
- **用途**：配置页面的交互逻辑
- **主要功能**：
  - 测试飞书应用配置
  - 加载多维表格列表
  - 选择多维表格
  - 加载数据表
  - 加载字段列表
  - 保存配置到Chrome存储和后端
  - 显示当前配置
  - 加载已保存的配置

## 📁 popup/ - 弹窗页面

```
popup/
├── popup.html                # 弹窗页面HTML
└── popup.js                  # 弹窗页面逻辑
```

### popup.html
- **用途**：弹窗页面的HTML结构
- **主要状态**：
  1. 未配置状态
  2. 表格选择状态
  3. 数据输入状态
  4. 加载中状态

### popup.js
- **用途**：弹窗页面的交互逻辑
- **主要功能**：
  - 检查配置状态
  - 加载多维表格列表
  - 显示多维表格列表
  - 选择多维表格
  - 加载字段信息
  - 显示输入字段
  - 提交记录数据
  - 切换表格
  - 导航到配置页面

## 📁 background/ - 后台服务

```
background/
└── background.js             # Chrome扩展后台脚本
```

### background.js
- **用途**：Chrome扩展的后台服务脚本
- **主要功能**：
  - 处理扩展安装事件
  - 处理扩展更新事件
  - 响应来自popup的消息
  - 检查配置状态
  - 获取配置信息
  - 监听图标点击事件

## 📁 styles/ - 样式文件

```
styles/
├── options.css               # 配置页面样式
└── popup.css                 # 弹窗页面样式
```

### options.css
- **用途**：配置页面的样式
- **主要特性**：
  - 响应式设计
  - 渐变背景
  - 现代UI风格
  - 动画效果
  - 滚动条美化

### popup.css
- **用途**：弹窗页面的样式
- **主要特性**：
  - 固定宽度（400px）
  - 最小高度（500px）
  - 渐变背景
  - 状态卡片设计
  - 动画过渡效果

## 📁 utils/ - 工具函数

```
utils/
├── api.js                    # API调用工具
└── storage.js                # Chrome存储工具
```

### api.js
- **用途**：封装后端API调用
- **主要函数**：
  - `saveConfig()` - 保存配置
  - `getConfig()` - 获取配置
  - `getBitables()` - 获取多维表格
  - `getBitableTables()` - 获取数据表
  - `getTableFields()` - 获取字段
  - `addRecord()` - 新增记录
  - `checkRecordStatus()` - 检查记录状态

### storage.js
- **用途**：封装Chrome存储API
- **主要函数**：
  - `saveConfig()` - 保存配置
  - `getConfig()` - 获取配置
  - `isConfigured()` - 检查是否已配置
  - `clearConfig()` - 清除配置
  - `saveData()` - 保存数据
  - `getData()` - 获取数据
  - `removeData()` - 删除数据
  - `onStorageChanged()` - 监听存储变化

## 📁 icons/ - 图标文件

```
icons/
├── create-icons.html         # 图标生成工具
├── icon16.png                # 16x16图标
├── icon48.png                # 48x48图标
└── icon128.png               # 128x128图标
```

### create-icons.html
- **用途**：生成Chrome扩展图标
- **功能**：
  - 生成三个尺寸的图标（16x16、48x48、128x128）
  - 使用Canvas绘制图标
  - 提供下载功能
- **图标设计**：
  - 渐变背景（紫色系）
  - 表格图标（白色）
  - 圆角矩形设计

## 🔄 数据流程

### 配置流程
```
用户输入 → options.html → options.js → 
api.js → 后端API → handlers/lark_handler.go → 
services/lark_service.go → 飞书API
```

### 数据录入流程
```
用户操作 → popup.html → popup.js → 
api.js → 后端API → handlers/lark_handler.go → 
services/lark_service.go → 飞书API → 
数据存储 → 字段检测 → 消息发送
```

### 配置存储流程
```
保存配置 → options.js → 
chrome.storage.local → 
后端API → handlers/lark_handler.go → 
configCache
```

## 🔑 关键技术点

### 前端技术
- **Chrome Extension API**：扩展开发基础
- **Vanilla JavaScript**：无框架依赖
- **CSS3**：现代样式和动画
- **LocalStorage**：数据持久化

### 后端技术
- **Go 1.21+**：后端语言
- **Gin框架**：Web框架
- **飞书SDK**：API封装
- **CORS**：跨域支持

### 集成技术
- **RESTful API**：前后端通信
- **JSON**：数据交换格式
- **HTTP**：通信协议
- **WebSocket**：实时通信（可选）

## 📝 开发建议

### 添加新功能
1. 在[`backend/handlers/`](backend/handlers/)添加新的处理器
2. 在[`backend/services/`](backend/services/)添加业务逻辑
3. 在[`backend/main.go`](backend/main.go)注册新路由
4. 在[`utils/api.js`](utils/api.js)添加前端API调用
5. 在[`options/options.js`](options/options.js)或[`popup/popup.js`](popup/popup.js)添加UI逻辑

### 修改样式
1. 编辑[`styles/options.css`](styles/options.css)或[`styles/popup.css`](styles/popup.css)
2. 保持设计一致性
3. 考虑响应式设计

### 调试技巧
1. 使用Chrome开发者工具
2. 查看后端控制台日志
3. 检查网络请求
4. 验证Chrome存储数据

## 🎯 项目特色

1. **前后端分离**：清晰的架构设计
2. **RESTful API**：标准化的接口设计
3. **配置灵活**：支持多种配置场景
4. **用户体验**：简洁直观的界面
5. **文档完善**：详细的使用和开发文档

---

本文档帮助您快速了解项目结构，便于后续开发和维护。