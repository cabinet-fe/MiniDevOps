# BuildFlow

CI/CD 构建部署平台。Go 后端 + React 前端单体仓库，前端产物通过 `embed` 嵌入后端二进制发布。

## 常用命令

**启动开发服务时请先查看是否已经启动，避免重复启动！**

```bash
# 开发
make dev                 # 同时启动后端和前端
make dev-backend         # 仅后端 (localhost:8080)
make dev-frontend        # 仅前端 dev server (Vite :8070, 代理 → :8080)

# 构建
make build               # 完整构建：前端 → 嵌入 → Go 二进制
make build-frontend      # 仅构建前端（输出到 cmd/server/dist/）
make build-backend       # 仅构建后端（CGO_ENABLED=1）
make build-linux         # 交叉编译 Linux amd64

# 测试
go test ./...                          # 全量测试
go test ./internal/engine/...          # 单包测试
go test -run TestXxx ./internal/...    # 单用例

# 前端
cd web && bun run lint   # oxlint 检查
cd web && bun run build  # TypeScript 编译 + Vite 构建

# 清理
make clean               # 删除 buildflow 二进制、dist、data 目录
```

## 技术栈

| 层级 | 技术 | 版本 |
| -------- | ------------------- | ------ |
| 语言 | Go | 1.25.6 |
| Web 框架 | Gin | 1.11 |
| ORM | GORM + SQLite | 1.31 |
| 认证 | JWT (golang-jwt/v5) | 5.3 |
| 日志 | zap | 1.27 |
| 配置 | Viper | 1.21 |
| WebSocket | gorilla/websocket | 1.5 |
| 定时任务 | robfig/cron/v3 | 3.0 |
| SFTP | pkg/sftp | — |
| 加密 | AES-GCM（敏感字段存储）+ 登录可选 AES-256-CBC（`password_cipher`）+ bcrypt | — |
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
├── cmd/server/                   # 入口，embed 前端产物
│   ├── main.go                   # 应用启动、路由注册、DI 组装
│   └── dist/                     # 前端构建产物（git 忽略）
├── internal/
│   ├── config/                   # Viper 配置加载（支持 BUILDFLOW_ 环境变量前缀）
│   ├── model/                    # GORM 模型 + DB 初始化 + 种子数据
│   │   ├── database.go           # InitDB()、AutoMigrate、默认 admin
│   │   ├── build.go              # Build
│   │   ├── project.go            # Project
│   │   ├── environment.go        # Environment（含构建脚本、部署配置、Cron）
│   │   ├── server.go             # Server
│   │   ├── user.go               # User
│   │   ├── variable.go           # EnvVar, VarGroup, VarGroupItem
│   │   ├── dictionary.go         # Dictionary, DictItem
│   │   ├── audit_log.go          # AuditLog
│   │   └── notification.go       # Notification
│   ├── repository/               # 数据访问层
│   ├── service/                  # 业务逻辑层
│   ├── handler/                  # HTTP handler（Gin）
│   ├── middleware/               # 认证、RBAC、CORS、审计
│   │   ├── auth.go               # JWT 校验，注入 ctxUserID/ctxUsername/ctxRole
│   │   ├── rbac.go               # RequireRole, RequireOwnerOrRole
│   │   ├── cors.go               # CORS 配置
│   │   └── audit.go              # POST/PUT/PATCH/DELETE 自动审计
│   ├── engine/                   # 构建引擎
│   │   ├── pipeline.go           # Pipeline：克隆 → 构建 → 归档 → 部署
│   │   ├── scheduler.go          # Scheduler：并发构建限制、Submit/Cancel
│   │   ├── cron.go               # CronScheduler：定时构建 Add/Remove/Update
│   │   └── git.go                # GitCloneOrPull, GitListBranches
│   ├── deployer/                 # 部署器
│   │   ├── deployer.go           # Deployer 接口 + NewDeployer(method) 工厂
│   │   ├── rsync.go / sftp.go / scp.go / agent.go / local.go
│   │   ├── ssh.go                # SSH 连接、远程脚本执行
│   │   └── path.go               # 跨平台路径处理
│   ├── pkg/                      # 通用工具
│   │   ├── response.go           # Success/Error/Paginated 响应封装
│   │   └── crypto.go             # AES-GCM 加解密、bcrypt 哈希
│   └── ws/                       # WebSocket Hub
│       └── hub.go                # Hub/Client：频道广播、用户推送
├── web/                          # React 前端
│   └── src/
│       ├── components/
│       │   ├── ui/               # shadcn/ui 组件
│       │   ├── layout/           # AppLayout, Sidebar, Header
│       │   ├── build-log-viewer.tsx
│       │   ├── notification-bell.tsx
│       │   └── dashboard/        # 仪表盘图表组件
│       ├── pages/
│       │   ├── dashboard.tsx
│       │   ├── login.tsx
│       │   ├── projects/         # list, detail, form, environment-form
│       │   ├── builds/           # detail
│       │   ├── servers/          # list, form
│       │   ├── users/            # list
│       │   ├── dictionaries/     # list
│       │   ├── audit-logs.tsx
│       │   └── settings.tsx
│       ├── hooks/                # useAuth, useWebSocket
│       ├── stores/               # auth-store, notification-store
│       ├── lib/
│       │   ├── api.ts            # 统一 HTTP 客户端（401 自动 refresh 重试）
│       │   ├── constants.ts      # 枚举常量（角色、状态、部署方式等）
│       │   └── utils.ts          # cn()（clsx + tailwind-merge）
│       ├── App.tsx               # BrowserRouter 路由配置 + ProtectedRoute
│       └── main.tsx              # 入口
├── config.yaml                   # 运行时配置
├── Makefile                      # 构建脚本
├── go.mod / go.sum
└── data/                         # 运行时数据（SQLite、工作空间、产物、日志）
```

## 架构约定

### 后端分层

`handler → service → repository → model`，单向依赖，禁止跨层调用。

- **model**：纯数据结构 + GORM tag，不含业务逻辑。`database.go` 负责 `InitDB()`、`AutoMigrate` 和种子数据。
- **repository**：仅数据库 CRUD，方法签名以 `Find`/`Create`/`Update`/`Delete`/`List`/`Count` 开头。通过 `New*Repository(db *gorm.DB)` 构造。
- **service**：业务编排，可组合多个 repository。通过 `New*Service(repo, ...)` 构造。
- **handler**：请求解析、参数校验、调用 service、统一响应。通过 `New*Handler(service)` 构造后调用 `RegisterRoutes(rg *gin.RouterGroup)` 注册路由。

DI 在 `cmd/server/main.go` 中手动组装，不使用框架。

### 构建引擎

Pipeline 执行流程：`克隆/拉取 → 执行构建脚本 → 归档产物（gzip/zip） → 部署`。

- **Scheduler**：通过 `config.build.max_concurrent` 限制并发构建数，`Submit(buildID)` 入队，`Cancel(buildID)` 取消。
- **CronScheduler**：基于 `robfig/cron/v3`，从 `Environment.CronExpression` 加载定时任务。
- **Git**：支持 URL 内嵌 token、HTTP Basic、明文密码三种认证方式。

### 部署器

`Deployer` 接口统一 `Deploy(ctx, DeployOptions) error`，通过 `NewDeployer(method)` 工厂方法创建：

| 方法 | 实现 | 说明 |
|------|------|------|
| `rsync` | RsyncDeployer | rsync 增量同步 |
| `sftp` | SFTPDeployer | SFTP 传输 |
| `scp` | SCPDeployer | SCP 传输 |
| `agent` | AgentDeployer | HTTP Agent 推送（zip/gzip） |
| `local` | LocalDeployer | 本机目录复制（产物目录递归复制到本机绝对路径，不删除目标多余文件） |

SSH 连接支持密码、密钥、SSH Agent 三种认证。`path.go` 处理 Windows/Linux 远程路径差异。

### WebSocket

- `Hub` 管理所有连接，支持频道广播 (`BroadcastToChannel`) 和用户推送 (`BroadcastToUser`)。
- 构建日志：`/ws/builds/:id/logs`，实时流式推送。
- 通知推送：`/ws/notifications`，按用户 ID 定向推送。
- 客户端连接自动携带 JWT token 作为查询参数。

### API 规范

- 基础路径：`/api/v1`
- 统一响应：`{ code: int, message: string, data?: T }`
- 分页响应：`{ items, total, page, page_size, total_pages }`
- 认证：Bearer JWT，access_token（2h） + refresh_token（168h）
- 登录：`POST /api/v1/auth/login` 支持 `password`（明文，可选）与 `password_cipher`（hex，可选）。若 `password_cipher` 非空则仅解密该字段（AES-256-CBC，格式为 `hex(IV(16 字节) || PKCS#7 密文)`）；否则使用 `password`。解密失败返回 400「登录参数无效」，与凭据错误 401 区分。前端仅提交 `password_cipher`（`web/src/lib/login-crypto.ts`）：密钥来源 **优先** `window.__BUILDFLOW_ENCRYPTION_KEY__`（嵌入二进制由 Go 在返回的 `index.html` 中注入，与**运行时** `encryption.key` 一致，改配置重启即可，无需重编前端）；否则使用 `VITE_BUILDFLOW_ENCRYPTION_KEY`（dev、`vite preview`、非 Go 托管静态资源等，见 `web/.env`）。**安全上下文**（HTTPS、`localhost` 等，即存在 `crypto.subtle`）下用 **Web Crypto** 加密；**非安全上下文**（如纯 HTTP 内网 IP）下用 **`crypto-es`**（AES-256-CBC）；无有效密钥或加密失败时抛错，不再回退为明文 `password`。
- RBAC 角色：`admin`（全权限）、`ops`（运维操作）、`dev`（只读 + 触发构建）
- WebSocket：`/ws/` 前缀，token 通过查询参数传递
- Webhook：`POST /api/v1/webhook/:projectId/:secret`（公开，无需认证）；可选查询参数 `environment_id` 时仅在推送分支与该环境分支一致时触发该环境构建
- 文件下载：直接返回二进制流（产物下载）

### 前端规范

- 路径别名：`@` → `web/src/`
- Vite 开发端口 `:8070`，代理 `/api` 和 `/ws` 到后端 `:8080`
- UI 组件基于 shadcn/ui，放在 `components/ui/`
- 页面组件放在 `pages/` 下按功能分目录
- API 调用统一通过 `lib/api.ts` 的 `api` 对象，401 自动 refresh 重试
- 状态管理用 Zustand store，放在 `stores/`
- 枚举常量集中定义在 `lib/constants.ts`，前后端保持一致
- WebSocket 使用 `hooks/use-websocket.ts`，内置自动重连
- 路由保护通过 `ProtectedRoute` 组件，未登录跳转 `/login`

### 代码风格

- Go：标准 `gofmt` 格式化，包名小写单词
- Go 测试：`*_test.go` 同包放置，使用标准 `testing` 包
- TypeScript：oxlint 规则
- JSON 字段命名：`snake_case`
- 前端组件文件命名：`kebab-case.tsx`
- 前端常量命名：`UPPER_SNAKE_CASE`（对象键用 `snake_case`）

### 数据库约定

- SQLite 单文件存储：`data/db.sqlite`
- Schema 变更通过 GORM `AutoMigrate` 自动迁移，无手动 migration 文件
- 敏感字段（密钥、密码）使用 `pkg/crypto.go` 的 AES-GCM 加密存储
- 用户密码使用 bcrypt 哈希
- 默认管理员账号在 `InitDB()` 中创建，配置来自 `config.yaml`
