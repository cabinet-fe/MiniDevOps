# 后端约定

## 技术栈

| 项     | 技术                                                      | 说明                            |
| ------ | --------------------------------------------------------- | ------------------------------- |
| 语言   | Go 1.26+                                                  | 单体 Server + 独立 Deploy Agent |
| Web    | Gin                                                       | `/api/v1`、`/ws`                |
| ORM    | GORM                                                      | sqlite / postgres / mysql       |
| Schema | 版本化 Go migration + `schema_migrations`                 | **禁止**仅靠 AutoMigrate        |
| 认证   | JWT + PAT                                                 | Bearer                          |
| 日志   | zap                                                       | 请求 `request_id`               |
| 配置   | Viper                                                     | `BEDROCK_` 前缀                 |
| 加密   | AES-GCM（存储）+ AES-256-CBC（`password_cipher`）+ bcrypt | —                               |
| 契约   | `api/*.md`（按域拆分）                                    | 见 [.agents/api.md](api.md)     |

## 分层

`handler → service → repository → model`，单向依赖，禁止跨层与逆向依赖。

- **model**：纯数据结构 + GORM tag
- **repository**：仅 CRUD；方法名以 `Find` / `Create` / `Update` / `Delete` / `List` / `Count` 开头
- **service**：业务编排；构造函数 `New*Service(...)`
- **handler**：校验与响应；`RegisterRoutes(rg *gin.RouterGroup)` 按域注册
- **DI**：在 `cmd/server` 手动组装，不引入 DI 框架

## 数据库

- 驱动：`sqlite`（默认）、`postgres`/`postgresql`、`mysql`
- 启动：连通性失败则**拒绝启动**；执行未应用 migration
- **改 driver ≠ 搬迁数据**；只支持全新安装——产品决策见 DESIGN / ROADMAP
- 敏感字段 AES-GCM；用户密码 bcrypt
- Schema 变更必须走版本化 migration，不得用 AutoMigrate 替代

## 代码风格

- Go：`gofmt`；测试 `*_test.go` 同包

## 测试门禁

- 涉及 schema：三驱动合同测试（`go test ./internal/platform/db/... -tags=contract`）或明确标注驱动差异
- 涉及 API：对照 [`api/`](../api/README.md) 契约文档（工作流见 [.agents/api.md](api.md)）
- 涉及状态机 / 权限 / Webhook / 存储路径：必须有单测或集成测
- GA 冒烟：`make smoke` / `scripts/smoke/*`
- 不做容量/延迟 SLO 验收（见 ROADMAP）

## 安全表述（工程卫生）

注释与用户可见文案不得夸大隔离或传输安全；具体安全边界与用户说明措辞见 [docs/DESIGN.md](../docs/DESIGN.md)。

## 禁止事项（后端工程）

1. 跨层调用，或在 handler 中直接操作 DB。
2. 用 GORM AutoMigrate **替代**版本化 migration。
3. 信封 / 分页字段约定见 [.agents/api.md](api.md)，勿在本文件重复或另起一套。
