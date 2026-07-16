# 后端约定

改 `cmd/`、`internal/`、`api/`、migration、契约时阅读本文。领域产品规则（权限语义、流水线状态机、AI/Skills、存储限额等）见 [docs/DESIGN.md](../docs/DESIGN.md)，**勿在此复制业务设计**。仓库入口与命令见 [AGENTS.md](../AGENTS.md)。

## 技术栈

| 项 | 技术 | 说明 |
| -- | ---- | ---- |
| 语言 | Go 1.25+ | 单体 Server + 独立 Deploy Agent |
| Web | Gin | `/api/v1`、`/ws` |
| ORM | GORM | sqlite / postgres / mysql |
| Schema | 版本化 Go migration + `schema_migrations` | **禁止**仅靠 AutoMigrate |
| 认证 | JWT + PAT | Bearer |
| 日志 | zap | 请求 `request_id` |
| 配置 | Viper | `BEDROCK_` 前缀 |
| 加密 | AES-GCM（存储）+ AES-256-CBC（`password_cipher`）+ bcrypt | — |
| 契约 | OpenAPI 3.2 源 + 3.1 投影 | 投影禁止手改 |

## 分层

`handler → service → repository → model`，单向依赖，禁止跨层与逆向依赖。

- **model**：纯数据结构 + GORM tag
- **repository**：仅 CRUD；方法名以 `Find` / `Create` / `Update` / `Delete` / `List` / `Count` 开头
- **service**：业务编排；构造函数 `New*Service(...)`
- **handler**：校验与响应；`RegisterRoutes(rg *gin.RouterGroup)` 按域注册
- **DI**：在 `cmd/server` 手动组装，不引入 DI 框架

领域包边界见 [docs/DESIGN.md](../docs/DESIGN.md) §3。

## 数据库

- 驱动：`sqlite`（默认）、`postgres`/`postgresql`、`mysql`
- 启动：连通性失败则**拒绝启动**；执行未应用 migration
- **改 driver ≠ 搬迁数据**；2.0 只支持全新安装（不迁移 1.x）——产品决策见 DESIGN / ROADMAP
- 敏感字段 AES-GCM；用户密码 bcrypt
- Schema 变更必须走版本化 migration，不得用 AutoMigrate 替代

## API 约定

- 前缀 `/api/v1`；信封 `{ code, message, data? }`
- 分页字段：`items` / `total` / `page` / `page_size` / `total_pages`
- query：`page` / `page_size`（默认与上限与 DESIGN 一致）；在 `internal/pkg`（或 platform）复用分页 helper，禁止各 handler 手写不一致的字段名
- 无分页列表仍可返回 `items`（或文档约定的数组）
- 错误带 `request_id`；异步创建可用 202；关键写操作支持 `Idempotency-Key`
- JSON：`snake_case`

前端列表组件约定见 [.agents/fe.md](fe.md)。

## OpenAPI 工作流

1. **唯一手改**：`api/openapi.yaml`（3.2）
2. 变更后执行 `make openapi-projection` 生成 `api/openapi.3.1.projection.yaml`
3. **禁止**直接编辑投影文件
4. 可用 `make openapi-check` 校验

## 代码风格

- Go：`gofmt`；测试 `*_test.go` 同包

## 测试门禁

- 涉及 schema：三驱动合同测试（`go test ./internal/platform/db/... -tags=contract`）或明确标注驱动差异
- 涉及 API：对照 OpenAPI
- 涉及状态机 / 权限 / Webhook / 存储路径：必须有单测或集成测
- GA 冒烟：`make smoke` / `scripts/smoke/*`；投影：`make openapi-check`（禁止手改投影）
- 不做容量/延迟 SLO 验收（见 ROADMAP）

## 安全表述（工程卫生）

注释与用户可见文案不得夸大隔离或传输安全；具体安全边界与用户说明措辞见 [docs/DESIGN.md](../docs/DESIGN.md)。

## 禁止事项（后端工程）

1. 跨层调用，或在 handler 中直接操作 DB。
2. 用 GORM AutoMigrate **替代**版本化 migration。
3. **禁止**手改 `openapi.3.1.projection.yaml`（CI：`make openapi-check`）。
4. 绕过 OpenAPI 契约私自加字段（先改 `openapi.yaml` 再投影）。
5. 把 1.x→2.0 数据迁移脚本当作正式支持路径（CI：`make ga-guardrails`）。

领域级禁止项（如分发失败不得标 BuildRun `failed`、流水线内不同步跑 AI、非项目域对象 ACL 等）以 [docs/DESIGN.md](../docs/DESIGN.md) 为准，不在此展开。
