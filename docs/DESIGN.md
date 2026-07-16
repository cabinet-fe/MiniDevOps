# Bedrock 2.0 详细设计

| 项 | 内容 |
| --- | --- |
| 文档版本 | 2.0.0 |
| 状态 | 已确认基线 |
| 输入 | [PRD.md](./PRD.md)、[ROADMAP.md](./ROADMAP.md)、PRD 审查 canvas、`refactor.md`、现有 1.x 代码基线 |
| 范围 | 架构、领域模型、鉴权、迁移、API、异步任务、存储、前端、测试与发布；**本文为 2.0 技术真源** |

产品需求语义以 PRD 为准；分期与 Gate 以 ROADMAP 为准；本文闭合 PRD/canvas 中未决的实现选择，并覆盖已接受风险的补偿控制。

---

## 1. 已确认决策基线

### 1.1 产品与范围

| # | 主题 | 决策 |
| --- | --- | --- |
| D1 | 交付切片 | 分阶段完成全量 2.0 GA：P0→P1→P2→P3→P4→P5（见 ROADMAP） |
| D2 | 用户角色 | 多角色；权限取**并集**；**不支持**显式 deny |
| D3 | 1.x 升级 | **仅全新安装**；不提供 1.x 数据迁移 |
| D4 | 对象 ACL | **仅产品项目**使用成员 ACL；CI/CD、AI、凭证等只依赖全局 RBAC |
| D5 | 全局项目权限 | 显式 `project.projects:view_all` / `manage_all`；普通 `:update` 不隐含全局越权 |
| D6 | AI 文档发布 | 同节点双态草稿；人工确认发布；`expected_version` 乐观锁 |
| D7 | AI CLI | Claude Code / OpenCode / Reasonix / Codex **并行**交付，均为 GA 条件 |
| D8 | 构建事件触发 Agent | 默认 `artifact_ready`；BuildJob 可覆盖为 `distribution_finished` |
| D9 | 父菜单可见性 | 至少一个可见后代 → 自动补齐父级分组；叶子仍需自身 `:view` |
| D10 | 系统信息 | 非超管可看与超管相同的**只读**系统信息；运维写操作仍仅超管 |
| D11 | 自定义命令 | 自定义开发环境/CLI 命令脚本：**仅超管**可维护与执行；快照 + 强提示 + 审计 |

### 1.2 运行与安全

| # | 主题 | 决策 |
| --- | --- | --- |
| D12 | 命令执行 | 构建与 AI CLI **同 Bedrock 进程 UID** 直接运行；**无 OS/容器沙箱**；产品不得声称沙箱安全 |
| D13 | Web 会话 | 允许 HTTP；`access_token` 存 **Web Storage** + Bearer；`refresh_token` 仅服务端 **Set-Cookie**（HttpOnly，不设 Secure） |
| D14 | 凭证授权时点 | **绑定/修改**时校验 `credential:use`；之后执行仅需任务 `execute` |
| D15 | Webhook | 优先平台签名头 + delivery ID 去重；保留 URL secret 兼容；日志脱敏 |
| D16 | Cron | 每任务 IANA 时区；禁止同任务重叠；停机错过的触发**跳过** |
| D17 | PAT scope | 固定白名单：`skills:read`、`agents:run`；哈希存储、一次性回显、可过期/吊销；**不替代 HTTPS/TLS**；供 Skill 安装器与 `agents:run` API 对接 |
| D18 | 重启恢复 | `queued` 恢复调度；`running` → `interrupted`（可人工重试）；不做断点续跑 |
| D19 | 平台支持 | 生产：Linux amd64/arm64；macOS 仅开发；部署目标继续支持 Linux/Windows |
| D20 | 非功能验收 | **仅功能 Gate**；不设容量/延迟 SLO |

### 1.3 数据与契约

| # | 主题 | 决策 |
| --- | --- | --- |
| D21 | Build/Deploy 状态 | 归档成功后 BuildRun `status=success`；当前 `distribution_summary`；redeploy **追加** `BuildDeployAttempt` |
| D22 | DeployTarget | BuildJob **1:N 私有**；复制任务时复制配置，不跨任务共享 |
| D23 | 菜单/资源 | `RbacResource` 唯一资源树 + 一对一 `MenuMetadata`；统一事务管理 |
| D24 | Schema | 应用内版本化 Go migration + `schema_migrations`；禁止仅靠 AutoMigrate |
| D25 | DB 切换 | 改 driver 只连接目标库，**不搬迁数据** |
| D26 | API | OpenAPI **3.2** 唯一源；自动生成不可手改的 **3.1 兼容投影** |
| D27 | 存储 | `StorageObject` 注册表 + `StorageService`；日志独立分段；保守默认限额 |
| D28 | 前端迁移 | 旁路 `web-v2/`；达标后一次切换 embed |
| D29 | Run 快照 | 创建运行时强制写入最小配置快照（只读复现） |
| D30 | 状态字段 | `status`（结果）与 `stage`（活动阶段）分离；流水线**无**内嵌 agent 阶段 |

### 1.4 已接受风险（必须对外声明）

1. **同 UID 执行**：获得脚本/Agent/自定义命令执行权的用户，可触及 Bedrock 进程可见的文件、环境变量与已注入凭证。RBAC/ACL **不能**替代 OS 隔离。
2. **HTTP + 浏览器会话存储**：`access_token`（Web Storage）与响应可能被窃听或被同机恶意脚本读取；`refresh_token` 为 HttpOnly Cookie（不设 Secure，HTTP 下仍可能被网络窃听）；`password_cipher` **不替代** TLS。

补偿控制：超管门控运维与自定义命令；脚本/开发环境编辑权限收紧；审计；生产强烈建议 HTTPS；登录页与文档说明风险边界。

---

## 2. 部署形态与支持矩阵

```text
┌─────────────────────────────────────────────┐
│  Bedrock Server（单体二进制）                 │
│  - Go API / WS / Scheduler / Cron             │
│  - web-v2 产物 embed                          │
│  - 本机构建 + 本机 AI CLI                     │
│  - 本地工作区 / 制品 / 日志 / 缓存 / 对象存储目录 │
└──────────────────────┬──────────────────────┘
                       │ rsync / sftp / scp / agent / local
                       ▼
┌─────────────────────────────────────────────┐
│  目标服务器 + 可选 Deploy Agent（独立二进制） │
└─────────────────────────────────────────────┘
```

| 组件 | 支持 |
| --- | --- |
| Server 生产 | Linux amd64、Linux arm64 |
| Server 开发 | macOS（不承诺生产特性如全部开发环境安装路径） |
| Deploy Agent / 远程目标 | Linux、Windows（沿用现有部署器能力） |
| 数据库 | sqlite（默认）、postgres/postgresql、mysql |
| 前端发布 | embed 进 Server；开发态 Vite 代理 |

---

## 3. 后端包结构与分层

### 3.1 目标目录

```text
cmd/
  server/                 # 瘦入口：配置、migration、DI、路由装配、embed
  agent/                  # Deploy Agent（独立）
internal/
  platform/               # config、db 工厂、migration runner、健康检查
  auth/                   # JWT、PAT、login cipher
  rbac/                   # 资源树、权限合并、中间件
  system/                 # User、Role、Dictionary、OperationLog、Menu 维护
  cicd/                   # Repository、BuildJob、BuildRun、Server、Credential
  engine/                 # Pipeline、Scheduler、Cron、Git（依赖 cicd 接口）
  deployer/               # rsync/sftp/scp/agent/local
  ops/                    # Process、DevEnvironment
  project/                # ProductProject、Requirement、ApiDoc
  ai/                     # CliRuntime、AiAgent、AgentRun、Skill
  dashboard/              # Layout + 卡片数据源
  storage/                # StorageObject + StorageService
  ws/                     # Hub + 频道
  pkg/                    # response、crypto、errors、id
api/
  openapi.yaml            # OpenAPI 3.2 源（唯一手改契约）
  openapi.3.1.projection.yaml  # 生成物，禁止手改
web-v2/                   # Vue 3 前端
docs/
  PRD.md / DESIGN.md / ROADMAP.md
```

### 3.2 分层规则

`handler → service → repository → model`，单向依赖，禁止跨层调用。

- **model**：结构 + GORM tag；无业务逻辑。
- **repository**：`Find/Create/Update/Delete/List/Count`；`New*Repository(db)`。
- **service**：编排；可组合多 repository / storage / engine 端口。
- **handler**：解析、校验、调用 service、统一响应；`RegisterRoutes(rg)`。
- **DI**：`cmd/server` 手动组装；按域注册路由，避免巨型 `main` 路由墙。

### 3.3 可复用与重写边界

| 复用（适配） | 重写 / 新建 |
| --- | --- |
| Pipeline 阶段语义、Deployer 五法、Git、Webhook 解析器、WS 日志模式、AES-GCM、Deploy Agent | 动态 RBAC、多库工厂、版本化 migration、Repository/BuildJob 拆分、AgentRun 异步、PM 域、Skills/PAT、StorageObject、web-v2 |

---

## 4. 身份、RBAC 与项目 ACL

### 4.1 身份模型

- **User**：可禁用；绑定 **多个 Role**；权限 = 各角色权限码并集。
- **Super Admin**：内置；不可删除；恒拥有全部权限；运维唯一准入。
- **自定义 Role**：绑定 `{path}:action` 集合。
- **PAT**：属于 User；scope ⊆ {`skills:read`,`agents:run`}；存储哈希；创建时明文回显一次。

### 4.2 权限码

格式：`{path}:action`，`path` 为 `.` 分层，整串仅一个 `:`。

常用 action：`view`、`create`、`update`、`delete`、`execute`、`download`、`cancel`、`retry`、`redeploy`、`install`、`test`、`use`、`view_all`、`manage_all`。

### 4.3 菜单唯一真源

```text
RbacResource (type=menu|page|action|card, path, parent, enabled, sort_key)
    1 ── 0..1 MenuMetadata (title, route, icon_base64, icon_mime)
```

- 登录 / `GET /auth/me` 返回裁剪后的菜单树（含一级图标）。
- 前端**只渲染**下发菜单，不硬编码全量再隐藏。
- 父级规则：存在可见后代则自动出现分组；叶子需自身 `:view`。
- 图标：原始体积 ≤ 32KB；超限 400。
- 运维相关 path：即使角色误勾选，服务端仍拒绝非超管。

### 4.4 项目 ACL（唯一对象级 ACL）

| 项目角色 | 能力 |
| --- | --- |
| Owner | 项目内全部管理；转让；归档/解散（仍受全局 RBAC） |
| Admin | 管理成员（除转让 Owner）、需求与文档 |
| Member | 按细则创建/编辑需求与文档、评论 |
| Readonly | 只读 |

鉴权公式：

```text
允许 = 全局功能权限(path:action)
     AND (
           超管
        OR 持有 project.projects:view_all/manage_all（按动作）
        OR 是项目成员且项目角色允许该动作
         )
```

- 无 `view_all`：列表仅返回自己加入的项目。
- `manage_all`：可管理全部项目成员与内容，**无需**加入项目。
- Owner 转让：仅当前 Owner 或 `manage_all`。
- **唯一对象级 ACL**：仅产品项目（`ProductProject` / 成员）使用对象级成员 ACL；CI/CD、运维、凭证、AI 等域仍为全局 RBAC only。

### 4.5 凭证

- 密文 AES-GCM；API 永不回显明文。
- 引用绑定（仓库认证、服务器认证、任务变量等）时校验操作者 `cicd.credentials:use`。
- Cron/Webhook 执行使用绑定快照；不要求「触发者」现场具备 `use`。
- 删除保护：仍被引用时拦截并提示。

### 4.6 认证流

| 机制 | 规则 |
| --- | --- |
| Web JWT | access 短 TTL（Web Storage + Bearer）；refresh 长 TTL（HttpOnly Cookie，不设 Secure）；401 → `/auth/refresh` → 重试 |
| 登录 | 仅接受 `password_cipher`（前端）；服务端亦可兼容明文 `password` 供调试，但 web-v2 **禁止**提交明文 |
| Refresh | `/auth/refresh` 读 Cookie（可选 body 兜底）；失败清会话并跳转登录 |
| PAT | `Authorization: Bearer <pat>`；与 JWT 分流校验；按 scope 映射端点 |
| Webhook | 无 Bearer；见 §8 |
| WS | query `token`（JWT）；生产建议 HTTPS 以降低日志泄露面 |

**不提供**：refresh rotation 作为首期必做（可后续增强）；但用户禁用后，后续请求必须失败（校验时查库 `disabled`）。

---

## 5. 核心领域模型

### 5.1 实体关系（概念）

```mermaid
flowchart TB
  User --> Role
  Role --> RolePermission
  RbacResource --> MenuMetadata
  User --> ProjectMember
  ProductProject --> ProjectMember
  ProductProject --> Requirement
  ProductProject --> ApiDocNode
  Repository --> BuildJob
  BuildJob --> DeployTarget
  BuildJob --> BuildRun
  BuildRun --> BuildDeployAttempt
  DeployTarget --> Server
  Repository --> Credential
  Server --> Credential
  AiAgent --> CliRuntime
  AiAgent --> SkillPackage
  AiAgent --> AgentTrigger
  AgentTrigger --> AgentRun
  BuildRun -->|artifact_ready_or_override| AgentRun
  ApiDocNode -->|generate| AgentRun
  StorageObject --> AttachmentRef
```

### 5.2 CI/CD 状态机

**BuildRun.status**（结果）：`queued` | `running` | `success` | `failed` | `cancelled` | `interrupted`

**BuildRun.stage**（活动）：`pending` | `cloning` | `building` | `archiving` | `distributing` | `idle`

规则：

1. 克隆→构建→归档成功后：`status=success`，制品可下载；`stage` 进入 `distributing` 或 `idle`。
2. **分发失败不将 `status` 改为 `failed`**；更新 `distribution_summary`：`none` | `running` | `all_success` | `partial` | `all_failed` | `cancelled`。
3. 构建阶段失败 → `failed`；用户取消构建中 → `cancelled`；若已 `success` 仅取消分发 → 保持 `success` + summary 反映取消。
4. **禁止**流水线内嵌同步 `agent` 阶段；构建事件异步创建 `AgentRun`。
5. `retry`：新建 BuildRun；`redeploy`：**同一** BuildRun，追加 `BuildDeployAttempt`，summary 指向最新一批结果。

**BuildDeployAttempt**：每次分发/重新分发对每个目标一行（或一批次 + 每目标行）；含目标配置快照、状态、日志引用、起止时间。

**最小快照（BuildRun.snapshot_json）** 至少含：trigger 载荷、resolved commit、脚本 SHA-256、环境变量**名**列表、DeployTarget 副本、制品格式、触发者/系统主体。

### 5.3 AgentRun / 安装任务

| 任务 | status | 重启 |
| --- | --- | --- |
| AgentRun | pending/queued/running/success/failed/cancelled/interrupted | queued 恢复；running→interrupted；重试建议**新 Run** |
| DevEnvJob / CliInstallJob | 同上 | running→interrupted/failed，保留日志；人工重试新任务 |

构建事件默认：`artifact_ready`（归档成功且制品路径有效）。BuildJob.`agent_trigger_event` 可覆盖为 `distribution_finished`（本轮分发流程结束，无论成功失败）或 `none`。可选 `agent_id` 绑定默认智能体；亦可在 AgentTrigger 中按 Job 过滤。事件**异步**创建独立 AgentRun；Agent 失败**不**修改 BuildRun.status。流水线**禁止**内嵌同步 agent 阶段。

### 5.4 文档节点双态

```text
ApiDocNode:
  published_content
  draft_content
  content_version      # 已发布版本号
  draft_base_version   # 生成/编辑草稿时所基于的已发布版本
  draft_updated_at
  draft_source_run_id  # 可选，关联 AgentRun
```

- 生成/导入更新草稿：只写 `draft_content`。
- `POST .../docs/{id}/publish`：body 含 `expected_version`；匹配则 draft→published，`content_version++`，清空 draft；冲突 409。
- 提供 diff 摘要 API（published vs draft）。
- `POST .../docs/generate` 契约归属项目管理域；接通 `AgentRun` 见 **P4**。P3 可返回明确的「依赖 AI 域」错误/501，**不得**绕过草稿模型直接写 published。

### 5.5 Skill 可见性

- `public`：具备 Skills 查看权限的用户可见/可用。
- `private`：仅创建者（及后续若扩展的显式授权——首期不做对象 ACL）可见。
- 更新覆盖当前包，不保留历史版本；Run 快照保存 package digest。

---

## 6. 数据库与 Migration

### 6.1 配置

```yaml
database:
  driver: sqlite # sqlite | postgres | mysql
  path: ./data/db.sqlite
  host: 127.0.0.1
  port: 5432
  name: bedrock
  user: bedrock
  password: "***"
  ssl_mode: disable
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 1h
```

### 6.2 Migration 机制

- 表 `schema_migrations(version TEXT PRIMARY KEY, applied_at TIMESTAMP)`。
- Go 注册表：`migrations.Register(version, up func(ctx, db, driver))`。
- 启动时事务顺序执行未应用版本；失败拒绝启动。
- 公共 GORM/SQL 操作 + **少量驱动分支**（如部分索引类型）。
- **禁止**把业务数据转换塞进日常启动；2.0 无 1.x 迁移任务。
- 合同测试：同一套 repository 用例在 sqlite/postgres/mysql 上跑。

### 6.3 切换语义

修改 `driver` 并重启 = 连接**另一空库或已有 2.0 库**；不会从 SQLite「搬家」到 Postgres。产品文案与启动日志必须写明。

---

## 7. API 契约

### 7.1 约定

| 项 | 规则 |
| --- | --- |
| 前缀 | `/api/v1` |
| 成功 | `{ "code": 0, "message": "success", "data": ... }` |
| 错误 | `{ "code": <http或业务码>, "message": "...", "request_id": "..." }` |
| 分页 | `{ items, total, page, page_size, total_pages }`；稳定排序默认 `id desc` 或文档约定字段 |
| 异步创建 | `202` + `{ id, status }`；随后查详情/WS |
| 幂等 | 写接口支持 `Idempotency-Key`（Webhook delivery、手动触发等） |
| 并发 | 文档发布等使用 `expected_version` / `If-Match` |
| 鉴权 | `Authorization: Bearer <jwt\|pat>` |
| 契约源 | `api/openapi.yaml`（3.2）；CI 生成 3.1 投影；前端类型从投影或 3.2 工具生成；Gin 实现合同测试对照源 |

### 7.2 错误码基线

| HTTP | 场景 |
| --- | --- |
| 400 | 参数/JSON/登录 cipher 无效 |
| 401 | 未认证、密钥错误、PAT 无效 |
| 403 | RBAC/ACL/超管门控/PAT scope |
| 404 | 资源不存在 |
| 409 | 版本冲突、状态不允许、引用冲突 |
| 422 | 语义校验失败（如缺 SKILL.md） |
| 429 | 限流（可选，首期可占位） |
| 500/503 | 内部错误/依赖不可用 |

### 7.3 路由域（与 PRD 对齐，实现以 OpenAPI 为准）

- Auth / Users / Roles / RBAC resources / Menus / Dictionaries / Operation logs / Tokens
- Dashboard layout + card data
- Ops processes / dev-environments / per-environment sources
- Repositories / webhook / build-jobs / build-runs / servers / credentials
- Projects / members / requirements / docs（含 generate、publish、diff）
- AI CLIs / agents / triggers / runs / skills

Webhook 路径（2.0）：`POST /api/v1/webhook/repos/:repository_id/:secret`（可短期双路由兼容旧 path，GA 前文档化弃用）。

---

## 8. Webhook 与 Cron

### 8.1 Webhook

校验顺序：

1. 解析 repository + URL secret（恒需匹配，作为第一道门或与签名并用）。
2. 若存在平台签名头（GitHub/GitLab/Gitea/Gitee/Bitbucket 等），**校验签名**；失败 401。
3. 使用 delivery ID（或平台等价唯一键）做幂等；重复投递返回成功且不重复触发。
4. 无签名的 generic/手动调用：允许仅 URL secret；必须审计并限流。
5. 日志与错误信息**脱敏** secret。

分支匹配：与 BuildJob 分支规则一致；可选 `build_job_id` 查询参数收窄范围。

### 8.2 Cron

- 表达式 + **每任务** `timezone`（IANA）。
- 同 Job 若上一次 Run 仍非终态 → **跳过**本次触发并记审计/指标。
- 服务停机期间错过的触发：**不补跑**。
- 与全局 `max_concurrent` 队列协同：触发成功仅表示入队。

---

## 9. 异步任务与调度

```text
API/Cron/Webhook/Event
        │
        ▼
   持久化 Run 行 (queued)
        │
        ▼
   内存 Scheduler（可配置并发）
        │
        ▼
   Executor (engine / ai / ops install)
```

| 策略 | 行为 |
| --- | --- |
| 入队 | DB 先写 queued，再投递 channel |
| 重启 | 扫描：queued 重新 Submit；running → interrupted |
| 取消 | cancel map + context；终态写库 |
| 并发 | `build.max_concurrent`；Agent/安装可共用或分池（配置项） |
| 通知 | 终态站内通知 + WS |

---

## 10. 存储设计

### 10.1 StorageObject

| 字段（概念） | 说明 |
| --- | --- |
| id | 主键 |
| kind | attachment / doc_import / skill_zip / artifact / other |
| sha256 | 内容摘要 |
| size | 字节 |
| content_type | MIME |
| path | 相对存储根路径 |
| ref_count / refs | 引用保护 |
| created_by | 上传者 |
| deleted_at | 软删 |
| purge_after | GC 时间 |

日志与构建实时日志：**不**强制进对象表，按 `{log_dir}/...` 分段文件 + 保留期清理。

### 10.2 默认限额（均可配置）

| 类型 | 默认 |
| --- | --- |
| 附件 | 20MB |
| 文档导入包 | 100MB |
| Skill ZIP | 50MB |
| 制品单件 | 5GB |
| 日志保留 | 30 天 |
| 工作区保留 | 7 天（终态后） |
| 软删后 GC | 24 小时 |

### 10.3 安全

- 上传：大小、MIME/扩展名白名单（按 kind）、ZIP 条目数/解压比限制、防 Zip Slip。
- Markdown 渲染：XSS 消毒（前端强制；后端存储原样但 API 可标记）。
- 路径：一律经 StorageService，禁止业务直接拼接用户输入路径。

---

## 11. 前端设计（web-v2）

### 11.1 技术栈

| 层 | 技术 |
| --- | --- |
| 框架 | Vue 3.5+（Composition API + SFC） |
| 语言 | TypeScript |
| 状态 | Pinia |
| 路由 | Vue Router |
| 构建 | Vite+（`vite-plus` / `vp`） |
| UI | `@veltra/desktop` + `@veltra/styles` + `@veltra/icons` + `@veltra/utils` + `@veltra/directives` + `@veltra/compositions` |
| 工具 | `@cat-kit/core`、`@cat-kit/fe`、`@cat-kit/http`、`@cat-kit/tsconfig` |
| 包管理 | bun（由 Vite+ 工作流包装） |

### 11.2 目录

```text
web-v2/src/
  main.ts
  App.vue
  router/
  stores/          # auth, notification, ...
  api/             # @cat-kit/http 客户端 + 生成类型
  composables/     # usePermission, useWebSocket, ...
  layouts/
  components/
  views/           # 按域：cicd, project, ai, ops, system, dashboard
  lib/             # constants, login-crypto, ...
```

### 11.3 关键规则

1. UI **优先** Veltra；无合适组件再自建。
2. HTTP **只**经统一 client；401 refresh 语义对齐 DESIGN。
3. 菜单、权限码来自 `/auth/me`；路由守卫 + 按钮级 `hasPermission`。
4. 登录仅 `password_cipher`；密钥：`window.__BEDROCK_ENCRYPTION_KEY__` > env。
5. 避免巨型 SFC（参考 vue-best-practices）；重型编辑器/终端可局部引入。
6. 开发端口代理 `/api`、`/ws` → `:8080`；embed 仍输出到构建产物并由 Go 注入加密密钥。

### 11.4 切换

- Makefile `FRONTEND_DIR ?= web-v2`（CI/Release 默认相同）。
- Go embed **只认** `cmd/server/dist`，不关心来源。
- Release：构建 `web-v2/dist` → 拷贝至 `cmd/server/dist` → `go build` embed。
- 旧 `web/`（或上一版产物）保留**至少一个发布周期**；回滚改 `FRONTEND_DIR` 或替换 `cmd/server/dist` 后重打包。
- 切换 Gate 证据：[roadmap/P5-switch-gate.md](./roadmap/P5-switch-gate.md)；Gate 条文见 ROADMAP P5。

---

## 12. 可观测、审计与通知

- **OperationLog**：写操作与关键安全事件（谁、何时、IP、action、资源、结果、详情）。
- **构建/Agent 日志**：文件 + WS 频道 `build_run:{id}` / `agent_run:{id}`。
- **通知**：终态推送；WS `notifications:{userId}`。
- **request_id**：中间件注入，错误响应回传。

---

## 13. 测试与验收策略

| 层级 | 内容 |
| --- | --- |
| 单元 | migration、权限合并、状态机、签名校验、存储路径安全 |
| 合同 | 三数据库 repository；OpenAPI 响应形状 |
| 集成 | Pipeline 分发失败、redeploy attempt、重启恢复、Webhook 幂等 |
| E2E | Playwright 冒烟：登录→菜单→构建→日志→（GA）Agent/文档发布 |
| 平台 | Linux amd64/arm64 发布包冒烟；macOS 开发路径 |
| 明确不做 | 容量压测、延迟 SLO |

功能验收对齐 PRD 模块清单，但行为以本文决策为准（例如 view_all、草稿发布、attempt 历史）。

---

## 14. 发布、备份与回滚

1. **发布物**：`bedrock` Server 单二进制（embed 前端）+ `bedrock-agent`；Linux amd64/arm64 命名 `bedrock-linux-amd64` / `bedrock-linux-arm64` 与对应 `bedrock-agent-*`；附带 SHA256。
2. **全新安装**：空数据目录 + 配置 + 启动（migration + 种子超管）。见 [ops-handbook.md](./ops-handbook.md)。
3. **备份**：SQLite 可用文件复制/专用备份命令；Postgres/MySQL 使用各自工具——平台可提供「备份指引」，**不假装统一物理备份**。
4. **前端回滚**：保留上一版 `web/` 或旧 `web-v2` 产物 tag；改 `FRONTEND_DIR` / CI 拷贝源重新打包。见 [release-checklist.md](./release-checklist.md)。
5. **无** 1.x 升级通道；文档与登录页显著位置声明。
6. **检查单**：[release-checklist.md](./release-checklist.md)；冒烟：`make smoke*`。

---

## 15. 与 1.x 概念映射（无数据迁移）

| 1.x | 2.0 |
| --- | --- |
| Project | Repository（+ 可选关联 ProductProject） |
| Environment | BuildJob + DeployTarget[] |
| Build | BuildRun |
| BuildDistribution | BuildDeployAttempt（可多轮历史） |
| 固定 admin/ops/dev | Super Admin + 自定义 Role |
| 内嵌 pipeline agent | 异步 AgentRun + 构建事件 |
| AgentProxy | CliRuntime |
| React web/ | Vue web-v2/ |

---

## 16. 文档关系

| 文档 | 职责 |
| --- | --- |
| PRD.md | 产品需求与验收意图 |
| ROADMAP.md | 分期、依赖、Gate |
| DESIGN.md | 技术真源（本文） |
| AGENTS.md | 命令、目录与读写指引；FE/BE 约定见 `.agents/fe.md` / `.agents/be.md` |
| ops-handbook.md | 安装、多库、备份、风险、回滚 |
| release-checklist.md | 发版检查与 checksum |
| known-issues.md | 非阻塞已知问题 |
| api/openapi.yaml | API 真源 |

冲突时：实现与 OpenAPI/DESIGN 对齐；需求争议回退 PRD，并开变更同步三份文档。

---

**文档结束。**
