# 认证与令牌

登录、刷新、登出、当前用户，以及个人访问令牌（PAT）。

通用约定（信封、分页、认证）见 [.agents/api.md](../.agents/api.md)。
业务语义与权限模型见 [docs/DESIGN.md](../docs/DESIGN.md)。

## 认证

### POST /auth/login — 登录

认证：不需要
请求：{ username*, password_cipher, password }
响应头：写入 HttpOnly `refresh_token`（Max-Age 取 jwt.refresh_ttl，不设 Secure；Path=/api/v1/auth）
响应 200：data = LoginResponse
错误：400 / 401

### POST /auth/refresh — 刷新访问令牌

认证：不需要
请求：{ refresh_token }
响应头：轮换 HttpOnly `refresh_token`（不设 Secure；Path=/api/v1/auth）
响应 200：data = AccessTokenResponse
错误：401
说明：从 Cookie 读取 `refresh_token`（登录/刷新时 Set-Cookie）。非浏览器客户端也可在 JSON body 里传 `refresh_token`。会轮换 Cookie；响应体只返回新的 `access_token`。

### POST /auth/logout — 登出

响应头：清除 `refresh_token` Cookie
响应 200
错误：401

### GET /auth/me — 当前用户（含权限与菜单）

响应 200：data = MeResponse
错误：401 / 404

## 个人访问令牌（PAT）

### GET /tokens — 列出个人访问令牌（仅元数据）

响应 200

### POST /tokens — 创建个人访问令牌

请求：{ name*, scopes*, expires_at }
响应 201：data = PATCreateResponse
说明：明文 token 仅在创建响应中返回一次，之后不可再读。服务端只存哈希。scopes 限于 `skills:read`、`agents:run`。不能替代 HTTPS/TLS。

### DELETE /tokens/{id} — 删除个人访问令牌

路径参数：id*: integer
响应 200

## 对象形状

### AccessTokenResponse

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `access_token` | `string` | 是 |  |

### LoginRequest

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `username` | `string` | 是 |  |
| `password_cipher` | `string` |  | hex(IV ‖ AES-256-CBC PKCS#7 密文)；Web 优先 |
| `password` | `string` |  | 明文密码，仅调试；Web 不得发送 |

### LoginResponse

`refresh_token` 只通过 Set-Cookie（HttpOnly）下发，不在 JSON 里返回。

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `access_token` | `string` | 是 | JWT access token（Bearer），客户端自行保存 |
| `user` | `User` | 是 |  |
| `permissions` | `string[]` |  |  |
| `menus` | `MenuNode[]` |  |  |

### MeResponse

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `user` | `User` |  |  |
| `permissions` | `string[]` |  |  |
| `menus` | `MenuNode[]` |  |  |

### MenuNode

侧边栏精简菜单；title/route/icon/children 映射为 NavItem（path ← route）。

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `path` | `string` |  | 资源 path（如 system.users） |
| `title` | `string` |  |  |
| `route` | `string` |  | 前端路由，用作 NavItem.path |
| `icon` | `string` |  | 一级图标（data URL 或 Base64） |
| `sort` | `integer` |  |  |
| `children` | `MenuNode[]` |  |  |

### PATCreateResponse

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `token` | `string` |  | 明文仅展示一次；不落日志、不可再读 |
| `metadata` | `PersonalAccessToken` |  |  |

### PersonalAccessToken

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `user_id` | `integer` |  |  |
| `name` | `string` |  |  |
| `token_prefix` | `string` |  |  |
| `scopes` | `('skills:read' \| 'agents:run')[]` |  |  |
| `expires_at` | `string(date-time)` |  |  |
| `revoked_at` | `string(date-time)` |  |  |
| `last_used_at` | `string(date-time)` |  |  |
| `created_at` | `string(date-time)` |  |  |
