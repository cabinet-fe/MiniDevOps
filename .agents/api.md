# API 契约

本文件记录前后端共同遵守的 HTTP / JSON 约定。具体接口写在 [`api/`](../api/README.md) 下按域拆分的 Markdown 里。业务含义、权限和状态流转以 [docs/DESIGN.md](../docs/DESIGN.md) 为准。

## 路由与数据格式

- HTTP API 统一使用 `/api/v1` 前缀，WebSocket 使用 `/ws`。
- JSON 字段统一使用 `snake_case`，前后端 DTO 和枚举值保持一致。
- 所有响应使用 `{ code, message, data? }` 信封；错误响应还应包含 `request_id`。
- 异步创建返回 `202`，并在 `data` 中提供资源 ID 和当前状态。
- 需要防止重复提交的写接口应支持 `Idempotency-Key`。

### 信封里的 `code`

实现见 `internal/pkg/response.go`：

- 成功：`code` 为 `0`；`message` 为 `success`（200）或 `created`（201）。
- 失败：`code` 等于对应的 HTTP 状态码（如 400、401、403），并带上 `request_id`。
- 常见错误码含义见 [DESIGN §7.2](../docs/DESIGN.md)。

## 列表与分页

- 分页请求使用 `page` 和 `page_size`，后端统一通过 `internal/pkg` 中的分页工具解析。
- 分页结果放在响应的 `data` 中，字段为 `items`、`total`、`page`、`page_size` 和 `total_pages`。
- 不分页的列表可以返回 `{ items }`，也可以直接返回契约文档里明确约定的数组。
- 不要在 handler 或页面中另行定义分页参数、响应结构或默认值。

## HTTP 认证

- 请求通过 `Authorization: Bearer <access_token>` 携带访问令牌。
- `refresh_token` 由服务端写入 HttpOnly Cookie；有效期取 `jwt.refresh_ttl`，未配置时为 7 天。为兼容 HTTP 部署，该 Cookie 不设置 `Secure`。
- 收到 `401` 后，客户端调用 `POST /auth/refresh`，刷新成功后重试原请求。
- 前端统一使用基于 `@cat-kit/http` 的客户端，并启用 `credentials: true`。页面中不要直接调用 `fetch`；确有特殊需求时，应先在 DESIGN 中说明，并封装为共享 helper。

## 修改契约文档

接口真源是 [`api/`](../api/README.md) 下按域拆分的 Markdown（例如用户相关看 `api/system.md`，构建相关看 `api/cicd.md`）。

改接口时按这个顺序：

1. 先改对应域的 `api/<域>.md`（路径、字段、响应、错误码）。
2. 再同步后端 handler / service 与前端调用，以及相关测试。

任何新增或变更的接口字段都应先写进契约文档，前后端不得自行扩展一套未记录的结构。
