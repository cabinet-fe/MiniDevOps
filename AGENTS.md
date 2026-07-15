# Bedrock

磐石（Bedrock）2.0：项目开发基石平台。覆盖宿主机运维、CI/CD、产品协作（需求与接口文档）、AI 智能体与开放 Agent Skills。Go 后端 + Vue 3 前端单体仓库，前端产物通过 `embed` 嵌入后端二进制发布。

**权威文档：** [docs/PRD.md](docs/PRD.md)（需求）· [docs/DESIGN.md](docs/DESIGN.md)（技术真源）· [docs/ROADMAP.md](docs/ROADMAP.md)（分期与 Gate）· `api/openapi.yaml`（API 真源，OpenAPI 3.2）

本文描述 **2.0 目标架构与开发约定**。实现过程中若目录尚未完全落地，仍按本文与 DESIGN 推进，禁止回退到 1.x 口径（React、固定三角色、仅 SQLite AutoMigrate、流水线内嵌 Agent 等）。

## 常用命令

**启动开发服务前请先检查是否已在运行，避免重复启动。**

```bash
# 开发（FRONTEND_DIR 默认 web-v2）
make dev                 # 后端 :8080（-tags dev）+ 前端 Vite 代理
make dev-backend         # 仅后端
make dev-frontend        # 仅前端（web-v2）

# 构建
make build               # 前端 → cmd/server/dist → Go 二进制
make build-frontend      # 仅前端（FRONTEND_DIR）
make build-backend       # 仅后端
make build-linux         # Linux amd64（生产目标之一）
make build-linux-arm64   # Linux arm64
make build-agent-linux   # Deploy Agent

# 契约与检查
# OpenAPI 3.2 源：api/openapi.yaml
# 生成 3.1 投影（禁止手改）：api/openapi.3.1.projection.yaml
make openapi-projection
make openapi-check

# 测试
go test ./...
go test ./internal/cicd/...
go test -run TestXxx ./internal/...
# 三数据库合同测试（需本地或 CI 服务）
go test ./internal/platform/db/... -tags=contract

# 前端（web-v2；推荐 Vite+ 工作流）
cd web-v2 && vp install
cd web-v2 && vp dev
cd web-v2 && vp check    # format + lint + typecheck
cd web-v2 && vp build
# 或：bun run lint / bun run build

# 清理
make clean
```

> Makefile 目标以实现仓库为准逐步对齐；新增脚本时同步更新本文。

## 技术栈

| 层级 | 技术 | 说明 |
| -------- | ------------------- | ------ |
| 语言 | Go 1.25+ | 单体 Server + 独立 Deploy Agent |
| Web 框架 | Gin | `/api/v1`、`/ws` |
| ORM | GORM | sqlite / postgres / mysql |
| Schema | 版本化 Go migration + `schema_migrations` | **禁止**仅靠 AutoMigrate |
| 认证 | JWT + PAT | Bearer；PAT scope：`skills:read`、`agents:run` |
| 日志 | zap | 请求 `request_id` |
| 配置 | Viper | `BEDROCK_` 环境变量前缀 |
| WebSocket | gorilla/websocket | token 查询参数 |
| 定时任务 | robfig/cron/v3 | 每任务 IANA 时区 |
| 部署 | rsync / sftp / scp / agent / local | Deploy Agent 独立二进制 |
| 加密 | AES-GCM（存储）+ AES-256-CBC（`password_cipher`）+ bcrypt | — |
| 前端 | Vue 3.5+ / TypeScript / Pinia / Vue Router | 目录 `web-v2/` |
| 构建 | Vite+（`vite-plus` / `vp`） | — |
| UI | `@veltra/desktop` + styles/icons/utils/directives/compositions | 优先检索技能文档 |
| 工具库 | `@cat-kit/core`、`@cat-kit/fe`、`@cat-kit/http`、`@cat-kit/tsconfig` | 渐进式查阅 |
| 包管理 | bun | 可由 `vp install` 包装 |
| API 契约 | OpenAPI 3.2 源 + 3.1 投影 | 投影禁止手改 |

## 目录结构（目标态）

```text
.
├── cmd/
│   ├── server/                 # 入口、DI、embed dist、加密密钥注入
│   └── agent/                  # Deploy Agent
├── internal/
│   ├── platform/               # config、db 工厂、migration、健康检查
│   ├── auth/                   # JWT、PAT、登录解密
│   ├── rbac/                   # 资源树、权限并集、中间件
│   ├── system/                 # User、Role、Dictionary、OperationLog、Menu
│   ├── cicd/                   # Repository、BuildJob、BuildRun、Server、Credential
│   ├── engine/                 # Pipeline、Scheduler、Cron、Git
│   ├── deployer/               # 五种部署器 + SSH
│   ├── ops/                    # 进程、工具链
│   ├── project/                # 产品项目、需求、文档树
│   ├── ai/                     # CLI、Agent、Skill、AgentRun
│   ├── dashboard/              # 布局与卡片数据
│   ├── storage/                # StorageObject + StorageService
│   ├── ws/                     # Hub
│   └── pkg/                    # response、crypto、errors
├── api/
│   ├── openapi.yaml            # OpenAPI 3.2（唯一手改）
│   └── openapi.3.1.projection.yaml
├── web-v2/                     # Vue 3 前端
│   └── src/
│       ├── api/
│       ├── stores/
│       ├── router/
│       ├── composables/
│       ├── layouts/
│       ├── components/
│       ├── views/
│       └── lib/
├── docs/
│   ├── PRD.md
│   ├── DESIGN.md
│   └── ROADMAP.md
├── config.yaml
├── Makefile
└── data/                       # db、工作区、制品、日志、对象存储（gitignore）
```

## 架构约定

### 后端分层

`handler → service → repository → model`，单向依赖，禁止跨层与逆向依赖。

- **model**：纯数据结构 + GORM tag。
- **repository**：仅 CRUD；方法名以 `Find`/`Create`/`Update`/`Delete`/`List`/`Count` 开头。
- **service**：业务编排；构造函数 `New*Service(...)`。
- **handler**：校验与响应；`RegisterRoutes(rg *gin.RouterGroup)` 按域注册。
- **DI**：在 `cmd/server` 手动组装，不引入 DI 框架。

领域包边界见 [docs/DESIGN.md](docs/DESIGN.md) §3。引擎通过接口依赖 CI/CD，不直接耦合旧「Project/Environment」概念。

### 数据库

- 驱动：`sqlite`（默认）、`postgres`/`postgresql`、`mysql`。
- 启动：连通性失败则**拒绝启动**；执行未应用 migration。
- **改 driver ≠ 搬迁数据**；2.0 **只支持全新安装**，不迁移 1.x。
- 敏感字段 AES-GCM；用户密码 bcrypt。
- 种子：内置超级管理员（不可删除）。

### 权限与菜单

- 多角色权限**并集**，无 deny。
- 权限码：`{path}:action`。
- `RbacResource` + 一对一 `MenuMetadata` 为菜单唯一真源；登录/`/auth/me` 下发裁剪菜单树。
- 前端**禁止**硬编码全量菜单后再隐藏。
- 父级：有可见后代则自动补齐；叶子需自身 `:view`。
- 运维 path：**仅超级管理员**，角色误勾无效。
- 对象级 ACL：**仅产品项目成员**；`project.projects:view_all` / `manage_all` 显式绕过；普通 `:update` 不隐含全局越权。
- CI/CD、凭证、AI、Skills 等：**全局 RBAC only**（Skill 的 public/private 为可见性字段，不是通用 ACL）。

### 构建引擎与部署

流水线：`克隆 → 构建 → 归档 →（标记构建成功，可下载）→ 按 DeployTarget 顺序分发`。

- **status** 与 **stage** 分字段；归档成功后 `status=success`。
- 分发失败**不得**将 BuildRun 改为 `failed`；更新 `distribution_summary`。
- 重新分发：同一 BuildRun，**追加** `BuildDeployAttempt`，summary 反映最新一批。
- **禁止**在流水线内同步执行 AI Agent；构建事件异步创建 `AgentRun`（默认 `artifact_ready`，可覆盖 `distribution_finished`）。
- Scheduler：全局并发上限；DB 持久化 queued；重启时 queued 恢复，running → `interrupted`。
- Cron：每任务 IANA 时区；禁止同任务重叠；停机错过的触发跳过。
- DeployTarget：归属 BuildJob 私有 1:N。
- 凭证：绑定/修改时校验 `credential:use`；执行仅需任务 `execute`。
- Run 创建时写入最小配置快照（commit、脚本摘要、变量名、目标副本等）。

### AI / Skills

- CLI：Claude Code、OpenCode、Reasonix、Codex（均为 2.0 GA 条件）。
- Agent 上下文首期：系统提示词 + 代码仓库。
- Skills：开放 Agent Skills ZIP；缺 `SKILL.md` 拒绝；更新覆盖；PAT 下载。
- 自定义 CLI/工具链命令模板：**仅超管**；快照 + 审计。
- 文档生成：只写 draft；人工 publish + `expected_version`。

### 存储

- 统一 `StorageObject` + `StorageService`；业务禁止随意拼路径写盘。
- 默认限额见 DESIGN（附件 20MB、导入 100MB、Skill 50MB、制品 5GB 等）。
- 防 Zip Slip / 解压炸弹；Markdown XSS 在渲染侧消毒。
- 日志文件独立保留策略，不强制进对象表。

### WebSocket

- Hub：频道广播与按用户推送。
- 构建日志、Agent 日志、通知；query `token`。
- 单进程内存 Hub（2.0 不承诺多实例粘滞）。

### API

- 前缀 `/api/v1`；信封 `{ code, message, data? }`；分页字段 `items/total/page/page_size/total_pages`。
- 错误带 `request_id`；异步创建可用 202；写操作支持 `Idempotency-Key`（关键路径）。
- **唯一手改契约**：`api/openapi.yaml`（3.2）。变更后必须再生 3.1 投影；**禁止**直接编辑投影。
- Webhook：`POST /api/v1/webhook/repos/:repository_id/:secret`；签名优先 + URL secret fallback；delivery 去重；日志脱敏。
- 登录：web-v2 只提交 `password_cipher`；密钥优先 `window.__BEDROCK_ENCRYPTION_KEY__`（Go 注入），否则环境变量；安全上下文 Web Crypto，非安全上下文兼容库；失败不回退明文密码。

### 前端规范（web-v2）

- 路径别名 `@` → `web-v2/src/`。
- **组件与工具必须优先 `@veltra/*` 与 `@cat-kit/*`**（二者已含大量可复用能力）。写 UI 前先检索 `.agents/skills/veltra-ui`（尤其 `packages/desktop/`）；写工具/HTTP 前检索 `.agents/skills/cat-kit`。**渐进检索**，勿整包预加载。仅确认库内无合适能力后才自研或引入第三方。
- 字段与交互形态须对齐 `@veltra/desktop` 组件契约（查 `components/<name>/types.d.ts` 与 `api.md`）。例如侧栏菜单应对齐 `UNav` / `UDualNav` 的 `NavItem`（`title`/`path`/`icon`/`children` 等），表格列用 `defineTableColumns`，分页用 `UPaginator`，表单用 `UForm`/`UFormItem`。API/DTO 命名可与后端 `snake_case` 并存，但映射到组件 props 时保持类型兼容。
- Vue 开发遵循 `.agents/skills/vue-best-practices`（避免巨型组件、可预测状态、Composition API）。
- HTTP 只走 `@cat-kit/http` 封装客户端（含 refresh）；禁止页面内散落 `fetch`（除非 DESIGN 标明的特例并抽 helper）。
- 状态：Pinia；权限辅助：composables。
- Token：`localStorage` + Bearer（已接受风险；勿改称「安全 Cookie 方案」除非产品变更决策）。
- 枚举与后端 `snake_case` JSON 字段保持一致。

### 列表与分页（前后端统一封装）

- **后端**：列表接口统一分页信封 `{ items, total, page, page_size, total_pages }`（见 DESIGN）；在 `internal/pkg`（或 platform）提供可复用的分页解析/响应 helper（query：`page`/`page_size`，默认与上限约定一致）。无分页的列表仍可返回 `items`（或文档约定的数组），但禁止各 handler 手写不一致的分页字段名。
- **前端**：封装通用列表组件（如 `ResourceList` / `QueryList`），调用方传入：
  - **API**：请求函数（接收 filters + 分页参数，返回分页或纯列表数据）；
  - **filters 插槽**：过滤表单/条件区（内部收集后触发请求）；
  - **列/行展示**：基于 `UTable` / `UList` 等 Veltra 组件；
  - 组件内部负责请求、加载态、空态（`UEmpty`）、以及有分页时渲染 `UPaginator`。
- 是否展示分页由**接口返回形态**决定（有 `page`/`total` 等则分页；纯列表则不分页），页面勿各自复制请求+表格+分页样板代码。

### 安全边界（必须遵守表述）

1. 构建与 AI CLI **同 Bedrock UID** 执行 → **无** OS/容器隔离；不得在注释/文档中声称「沙箱安全」。
   - **用户可见说明：** 构建脚本与部署后脚本以运行 Bedrock 的同一操作系统用户执行，可访问该用户可读的文件与网络；请仅授予可信人员脚本编辑权限，并对写操作保留审计。
2. 允许 HTTP + localStorage → **无**传输层防窃听保证；不得声称 `password_cipher` 替代 TLS。
3. 生产强烈建议 HTTPS；运维与自定义命令仅超管。

### 代码风格

- Go：`gofmt`；测试 `*_test.go` 同包。
- TypeScript / Vue：`vp check` 或项目 lint；SFC/`kebab-case` 文件名。
- JSON：`snake_case`。
- 前端常量：`UPPER_SNAKE_CASE`（对象键 `snake_case`）。

### 测试门禁

- 涉及 schema：三驱动合同测试或明确标注驱动差异。
- 涉及 API：对照 OpenAPI。
- 涉及状态机/权限/Webhook/存储路径：必须有单测或集成测。
- 前端关键路径：优先 Playwright 冒烟（登录、菜单、构建日志）。
- **不做**容量/延迟 SLO 验收（见 ROADMAP）。

## 禁止事项

1. 跨层调用或在 handler 中直接操作 DB。
2. 前端硬编码全量菜单/权限表作为真源。
3. 绕过统一 HTTP 客户端与 OpenAPI 契约私自加字段。
4. 手改 `openapi.3.1.projection.yaml`。
5. 用 GORM AutoMigrate **替代**版本化 migration。
6. 将分发失败写为 BuildRun `failed`。
7. 在构建流水线内同步跑 AI Agent。
8. 实现 1.x → 2.0 静默数据迁移（产品已否决）。
9. 为非项目域引入对象 ACL，或用 `:update` 冒充 `manage_all`。
10. 把 HTTP/同 UID 模式描述为具备隔离或传输安全。

## 提交与技能

- 提交代码时遵循 `.agents/skills/git-commit`。
- Vue / Veltra / CatKit 相关改动必须先读对应 skill，再写代码。
