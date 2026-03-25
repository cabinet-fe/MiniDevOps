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
│   │   ├── environment.go        # Environment（构建脚本、Cron 等）
│   │   ├── distribution.go       # Distribution（环境多分发目标）
│   │   ├── build_distribution.go # BuildDistribution（构建-分发执行记录）
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
│   │   ├── pipeline.go           # Pipeline：克隆 → 构建 → 归档 → 分发（多目标）
│   │   ├── pipeline_distribute.go # 分发阶段与单目标部署封装
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

Pipeline 执行流程：`克隆/拉取 → 执行构建脚本 → 归档产物（gzip/zip）→ 标记构建阶段成功（可下载）→ 按环境多条 Distribution 顺序分发`。分发失败**不会**将 `Build.Status` 置为 `failed`；汇总见 `Build.distribution_summary` 与 `BuildDistribution` 各行状态。`TriggerType == redistribute` 或仅重新分发时在同一 `Build` 上只跑分发阶段。

- **Scheduler**：通过 `config.build.max_concurrent` 限制并发构建数，`Submit(buildID)` 入队，`Cancel(buildID)` 取消。
- **CronScheduler**：基于 `robfig/cron/v3`，从 `Environment.CronExpression` 加载定时任务。
- **Git**：支持 URL 内嵌 token、HTTP Basic、明文密码三种认证方式。

### 部署器（分发）

`Deployer` 接口统一 `Deploy(ctx, DeployOptions) error`，通过 `NewDeployer(method)` 工厂方法创建；每条 `Distribution` 调用一次：

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
- Webhook（公开，无需 JWT；`curl`/CI/代码托管平台均可调用）：
  - **方法与路径**：`POST /api/v1/webhook/{project_id}/{webhook_secret}`。`{project_id}` 为正整数项目 ID；`{webhook_secret}` 为路径段，须与数据库中该项目的 `webhook_secret` **完全一致**（与「项目」里展示的 Webhook 密钥一致，需正确 URL 编码；含 `/`、`+` 等特殊字符时路径必须编码）。**无** `Authorization` 头。
  - **查询参数（可选）**：`environment_id` — 无符号十进制整数。若提供：仅处理该 ID 对应环境（且须属于该项目）；仍要求**推送分支**与环境的 **Branch** 字段一致才触发。不提供：遍历该项目下所有环境，对每个「分支名与解析出的分支一致」的环境各触发一次构建。
  - **请求头 `Content-Type`**：建议 `application/json`；请求体为**原始 JSON 字节**（不能为空或非法 JSON）。
  - **平台识别（决定如何解析请求体）**：按以下顺序匹配解析器；**手动调用**时至少满足其一即可。
    1. `X-Gitea-Event` 非空 → **Gitea**（需 `push` 事件，否则 400）。
    2. `X-Gitee-Event` 非空 → **码云 Gitee**（需 `Push Hook` 或 `Tag Push Hook`，否则 400）。
    3. `X-GitHub-Event` 非空 → **GitHub**（需 `push`，否则 400）。
    4. `X-Gitlab-Event` 非空 → **GitLab**（需 `Push Hook`，否则 400）。
    5. `X-Event-Key` 非空 → **Bitbucket**（需 `repo:push`，否则 400）。
    6. 以上均无：若项目 `webhook_type` 已设为 `github` / `gitlab` / `gitee` / `gitea` / `bitbucket` / `generic`（且非 `auto`），则**固定**使用该解析器。
    7. 若项目配置了 `webhook_ref_path`（JSONPath），则走 **generic**。
    8. 否则 400「无法识别 webhook 平台」。**仅**依赖 `auto` 且无上述头、且未配 `webhook_ref_path` 时，手动请求容易失败，应在项目中指定 `webhook_type` 或携带对应平台请求头。
  - **请求体字段（解析后用于触发）**：解析得到 `ref`（如 `refs/heads/main`）、可选 `commit` 哈希与说明；分支名为 `ref` 去掉 `refs/heads/` 前缀后的字符串（与环境的 **Branch** 做**字符串相等**比较）。
    - **GitHub / Gitea / Gitee（Push / Tag Push）**：JSON 至少含 `ref`；提交信息常用 `head_commit.id`、`head_commit.message`，缺省时哈希可用 `after`。码云官方示例见 [Gitee WebHook 推送数据格式说明](https://help.gitee.com/webhook/gitee-webhook-push-data-format)。
    - **GitLab（Push Hook）**：至少含 `ref`；`checkout_sha` 为提交哈希；`commit.message` 可选。
    - **Bitbucket（repo:push）**：按服务端结构解析 `push.changes[0].new` 等（见 `webhook_handler.go`）。
    - **generic**：在项目配置 `webhook_ref_path`（必填）、可选 `webhook_commit_path`、`webhook_message_path`；JSONPath 为点分字段，支持 `field[0]` 数组下标，见 `extractJSONValue`。
  - **成功响应**（HTTP 200）：`{ "code": 0, "message": "success", "data": { "triggered": <int>, "branch": "<分支名>", "environment_id": <uint> } }`，仅当查询带 `environment_id` 时 `data` 含 `environment_id`。
  - **失败响应**：HTTP 4xx/5xx，`{ "code": <http码>, "message": "<原因>" }`（如项目不存在 404、密钥错误 401、解析失败 400）。
  - **手动调用示例**  
    - GitHub：`curl -sS -X POST 'https://{host}/api/v1/webhook/{project_id}/{secret}' -H 'Content-Type: application/json' -H 'X-GitHub-Event: push' -d '{"ref":"refs/heads/main","after":"abc123","head_commit":{"id":"abc123","message":"manual"}}'`  
    - 码云：请求体与上类似，请求头使用 `X-Gitee-Event: Push Hook`（标签推送为 `Tag Push Hook`）；仓库 WebHook 还会带 `User-Agent: git-oschina-hook`、`X-Gitee-Token` / `X-Gitee-Timestamp` 等，**无需**与 BuildFlow URL 密钥相同，鉴权仍以路径中的 `webhook_secret` 为准。  
    可选查询：`?environment_id=3`。
- 文件下载：直接返回二进制流（产物下载）
- 重新分发：`POST /api/v1/builds/:id/deploy`（body 可选 `distribution_ids`），对**已成功且有产物**的同一构建记录触发仅分发阶段

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
