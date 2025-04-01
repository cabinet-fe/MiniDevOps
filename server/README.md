# MiniDevOps 迷你构建工具

迷你构建工具是一个轻量级 CI/CD 工具，支持用户登录、管理项目代码仓库（从码云自动拉取代码）、执行自动化构建任务，并提供构建历史追溯功能。

## 技术栈

- **后端框架**：Go + Go Fiber
- **数据库**：SQLite（轻量嵌入式）
- **ORM**：Ent
- **认证**：JWT
- **API 文档**：Swagger

## 项目结构

```
server/
├── internal/               # 内部包
│   ├── auth/               # 鉴权模块
│   ├── model/              # 数据模型 (Ent)
│   ├── gitee/              # 码云API集成
│   ├── builder/            # 构建任务引擎
│   ├── task/               # 任务调度管理
│   ├── handler/            # API处理器
│   ├── middleware/         # 中间件
│   ├── config/             # 配置
│   └── router/             # 路由定义
├── docs/                   # API文档
├── data/                   # 数据存储目录
│   ├── projects/           # 项目代码目录
│   ├── output/             # 构建输出目录
│   └── logs/               # 构建日志目录
├── main.go                 # 主入口文件
├── go.mod                  # Go模块定义
└── README.md               # 项目说明
```

## 功能特性

- **用户认证**：用户注册、登录、JWT 认证
- **项目管理**：创建、查询、更新和删除项目
- **码云集成**：自动拉取码云代码仓库
- **构建任务**：执行自定义构建命令
- **日志记录**：记录构建过程日志
- **WebSocket**：实时日志推送
- **API 文档**：集成 Swagger 文档

## 环境变量

| 环境变量名   | 描述                  | 默认值                                        |
| ------------ | --------------------- | --------------------------------------------- |
| SERVER_ADDR  | 服务监听地址          | :8080                                         |
| DATABASE_URL | SQLite 数据库连接 URL | file:./data/minidevops.db?cache=shared&\_fk=1 |
| JWT_SECRET   | JWT 签名密钥          | minidevops_secret_key                         |

## 本地开发

### 前置条件

- Go 1.20 或更高版本
- Git

### 安装依赖

```bash
cd server
go mod download
```

### 生成 Ent 模型代码

```bash
go run -mod=mod entgo.io/ent/cmd/ent generate ./internal/model/schema
```

### 启动服务

```bash
go run main.go
```

访问 http://localhost:8080/swagger/ 查看 API 文档。

## API 接口

### 认证

- POST `/api/register` - 注册新用户
- POST `/api/login` - 用户登录

### 项目管理

- GET `/api/projects` - 获取项目列表
- POST `/api/projects` - 创建新项目
- GET `/api/projects/{id}` - 获取项目详情
- PUT `/api/projects/{id}` - 更新项目
- DELETE `/api/projects/{id}` - 删除项目

### 构建任务

- POST `/api/projects/{id}/build` - 启动构建任务
- GET `/api/projects/{id}/builds` - 获取项目构建历史
- GET `/api/builds/{id}` - 获取构建任务详情
- GET `/api/builds/{id}/logs` - 获取构建日志
- GET `/api/ws/builds/{id}/logs` - WebSocket 实时日志
