# BuildFlow

CI/CD 构建部署平台。Go 后端 + React 前端单体仓库，前端产物嵌入后端二进制发布。

## 常用命令

```bash
# 开发
make dev-backend          # 启动后端 (localhost:8080)
make dev-frontend         # 启动前端 dev server (Vite, 代理 → :8080)

# 构建
make build                # 完整构建：前端 → 嵌入 → Go 二进制
make build-frontend       # 仅构建前端
make build-backend        # 仅构建后端（CGO_ENABLED=1）
make build-linux          # 交叉编译 Linux amd64

# 前端
cd web && bun run lint    # oxlint 检查
cd web && bun run build   # TypeScript 编译 + Vite 构建

# 清理
make clean
```

## 技术栈

| 层级 | 技术 | 版本 |
|------|------|------|
| 语言 | Go | 1.25.6 |
| Web 框架 | Gin | 1.11 |
| ORM | GORM + SQLite | 1.31 |
| 认证 | JWT (golang-jwt/v5) | 5.3 |
| 日志 | zap | 1.27 |
| 配置 | Viper | 1.21 |
| 前端框架 | React | 19 |
| 构建工具 | Vite | 8.x |
| CSS | Tailwind CSS | 4.x |
| UI 组件 | shadcn/ui (Radix) | — |
| 状态管理 | Zustand | 5.x |
| 路由 | React Router | 7.x |
| 图表 | Recharts | 3.x |
| 类型检查 | TypeScript | 5.9 |
| 包管理器 | bun | 1.x |
| Lint | oxlint | — |

## 目录结构

```
.
├── cmd/server/              # 入口，embed 前端产物
│   ├── main.go              # 应用启动、路由注册、DI 组装
│   └── dist/                # 前端构建产物（git 忽略）
├── internal/
│   ├── config/              # Viper 配置加载
│   ├── model/               # GORM 模型 + DB 初始化
│   ├── repository/          # 数据访问层
│   ├── service/             # 业务逻辑层
│   ├── handler/             # HTTP handler（Gin）
│   ├── middleware/           # 认证、RBAC、CORS、审计
│   ├── engine/              # 构建引擎（Pipeline + Scheduler）
│   ├── deployer/            # 部署器（SSH/SCP/SFTP/Rsync）
│   ├── pkg/                 # 通用工具（加密、响应封装）
│   └── ws/                  # WebSocket Hub
├── web/                     # React 前端
│   └── src/
│       ├── components/      # 组件（ui/ 为 shadcn 组件，layout/ 为布局）
│       ├── pages/           # 页面组件
│       ├── hooks/           # 自定义 hooks（auth、websocket）
│       ├── stores/          # Zustand stores
│       ├── lib/             # API 客户端、工具函数、常量
│       ├── App.tsx          # 路由配置
│       └── main.tsx         # 入口
├── config.yaml              # 运行时配置
├── Makefile                 # 构建脚本
├── go.mod / go.sum
└── data/                    # 运行时数据（SQLite、工作空间、产物、日志）
```

## 架构约定

### 后端分层

`handler → service → repository → model`，单向依赖，禁止跨层调用。

- **model**：纯数据结构 + GORM tag，不含业务逻辑。
- **repository**：仅数据库 CRUD，方法签名以 `Find`/`Create`/`Update`/`Delete`/`List` 开头。
- **service**：业务编排，可组合多个 repository。
- **handler**：请求解析、参数校验、调用 service、统一响应。

### API 规范

- 基础路径：`/api/v1`
- 统一响应格式：`{ code: int, message: string, data?: T }`
- 分页：`{ items, total, page, page_size, total_pages }`
- 认证：Bearer JWT，access_token + refresh_token
- RBAC 角色：`admin`、`ops`、`dev`

### 前端规范

- 路径别名：`@` → `web/src/`
- UI 组件基于 shadcn/ui，放在 `components/ui/`
- 页面组件放在 `pages/` 下按功能分目录
- API 调用统一通过 `lib/api.ts` 的 `api` 对象
- 状态管理用 Zustand store，放在 `stores/`
- 开发时前端 Vite proxy 到后端 `:8080`

### 代码风格

- Go：标准 `gofmt` 格式化，包名小写单词
- TypeScript：oxlint 规则
- JSON 字段命名：`snake_case`
- 前端组件文件命名：`kebab-case.tsx`
