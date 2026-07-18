# 资源管理

代码仓库、服务器、凭证、AI CLI 运行时、个人访问令牌（与 CI/CD / AI 业务域解耦的共享资源）。

通用约定（信封、分页、认证）见 [.agents/api.md](../.agents/api.md)。
业务语义与权限模型见 [docs/DESIGN.md](../docs/DESIGN.md)。

## 代码仓库

### GET /resource/repositories — 列出代码仓库

权限：`resource_repositories:view`
查询参数：page: integer, page_size: integer, keyword: string
响应 200：data = RepositoryPage
错误：403

### POST /resource/repositories — 创建代码仓库

权限：`resource_repositories:create`
请求：{ name*, description, tags, repo_url*, auth_type, credential_id }
响应 201：data = Repository
错误：403

### GET /resource/repositories/{id} — 获取代码仓库

权限：`resource_repositories:view`
路径参数：id*: integer
响应 200：data = Repository
错误：404

### PUT /resource/repositories/{id} — 更新代码仓库

权限：`resource_repositories:update`
路径参数：id*: integer
请求：{ name, description, tags, repo_url, auth_type, credential_id, clear_credential }
响应 200：data = Repository
错误：403

### DELETE /resource/repositories/{id} — 删除代码仓库

权限：`resource_repositories:delete`
路径参数：id*: integer
响应 200
错误：409

### GET /resource/repositories/{id}/branches — 列出远程分支

权限：`resource_repositories:view`
路径参数：id*: integer
响应 200：data = object

### POST /resource/repositories/{id}/test — 测试拉取 / 列分支

权限：`resource_repositories:view`
路径参数：id*: integer
响应 200

## 凭证

### GET /resource/credentials — 列出凭证（仅元数据，不含明文）

权限：`resource_credentials:view`
查询参数：page: integer, page_size: integer, keyword: string
响应 200：data = CredentialPage

### POST /resource/credentials — 创建凭证

权限：`resource_credentials:create`
请求：{ name*, type*, username, secret*, passphrase, description }
响应 201：data = Credential

### GET /resource/credentials/{id} — 获取凭证元数据

权限：`resource_credentials:view`
路径参数：id*: integer
响应 200：data = Credential

### PUT /resource/credentials/{id} — 更新凭证（secret 为空则保留原值）

权限：`resource_credentials:update`
路径参数：id*: integer
请求：{ name, type, username, secret, passphrase, description }
响应 200：data = Credential

### DELETE /resource/credentials/{id} — 删除凭证

权限：`resource_credentials:delete`
路径参数：id*: integer
响应 200
错误：409

## 服务器

### GET /resource/servers — 列出服务器

权限：`resource_servers:view`
查询参数：page: integer, page_size: integer, keyword: string, tag: string
响应 200：data = ServerPage

### POST /resource/servers — 创建服务器

权限：`resource_servers:create`
请求：{ name*, host, port, os_type, username, auth_type, credential_id, agent_url, agent_credential_id, description, tags }
响应 201：data = Server

### GET /resource/servers/{id} — 获取服务器

权限：`resource_servers:view`
路径参数：id*: integer
响应 200：data = Server

### PUT /resource/servers/{id} — 更新服务器

权限：`resource_servers:update`
路径参数：id*: integer
请求：{ name, host, port, os_type, username, auth_type, credential_id, clear_credential, agent_url, agent_credential_id, clear_agent_credential, description, tags }
响应 200：data = Server

### DELETE /resource/servers/{id} — 删除服务器

权限：`resource_servers:delete`
路径参数：id*: integer
响应 200
错误：409

### POST /resource/servers/{id}/test — 测试 SSH / Agent 连通性

权限：`resource_servers:view`
路径参数：id*: integer
响应 200

## AI CLI

运行时管理入口在运维「开发环境」页的「智能体 CLI」区块；无独立菜单。路径仍挂在 `/resource`（模型归属资源域），权限统一为 `ops_dev_environments:*`（仅超管）。

### GET /resource/clis — 列出 AI CLI

权限：`ops_dev_environments:view`
响应 200：data = object
说明：四套并行 CLI（Claude Code、OpenCode、Reasonix、Codex）。与 Bedrock 同 UID 执行，无 OS/容器沙箱。

### POST /resource/clis/{key}/detect — 检测 AI CLI

权限：`ops_dev_environments:execute`
路径参数：key*: string
响应 200：data = CliDetectResult

### POST /resource/clis/{key}/check-update — 检查 AI CLI 更新

权限：`ops_dev_environments:execute`
路径参数：key*: string
响应 200：data = CliCheckUpdateResult
说明：通过 `npm view <package> version` 查询最新版本（按启用安装源优先级尝试 `--registry`，无源则用默认 Registry）。与已安装版本比较后返回是否可更新。未安装时 `update_available` 为 false。

### POST /resource/clis/{key}/install — 安装 AI CLI

权限：`ops_dev_environments:execute`
路径参数：key*: string
请求：{ version }
响应 200：data = CliExecuteResult

### POST /resource/clis/{key}/upgrade — 升级 AI CLI

权限：`ops_dev_environments:execute`
路径参数：key*: string
请求：{ version }
响应 200：data = CliExecuteResult

### POST /resource/clis/{key}/uninstall — 卸载 AI CLI

权限：`ops_dev_environments:execute`
路径参数：key*: string
响应 200：data = CliExecuteResult

### GET /resource/cli-sources — 列出 CLI 安装源

权限：`ops_dev_environments:view`
查询参数：cli_key: string
响应 200
说明：安装源为可选 npm Registry。安装/升级时将 `base_url` 拼为 `npm --registry`；未配置启用源时使用 npm 默认 Registry。

### POST /resource/cli-sources — 创建 CLI 安装源

权限：`ops_dev_environments:create`
请求：{ cli_key*, name*, base_url*, priority, enabled }
响应 201
说明：`base_url` 为 npm Registry 地址（如 `https://registry.npmjs.org`）。

### PUT /resource/cli-sources/{id} — 更新 CLI 安装源

权限：`ops_dev_environments:update`
路径参数：id*: integer
请求：{ cli_key*, name*, base_url*, priority, enabled }
响应 200

### DELETE /resource/cli-sources/{id} — 删除 CLI 安装源

权限：`ops_dev_environments:delete`
路径参数：id*: integer
响应 200

## 个人访问令牌（PAT）

PAT 按 `user_id` 隔离：仅能列出/创建/删除本人令牌。Bearer PAT 的鉴权消费方式见 [auth.md](auth.md)。

### GET /resource/tokens — 列出个人访问令牌（仅元数据）

权限：`resource_tokens:view`
查询：`page`、`page_size`（标准分页）
响应 200：data = 分页信封（`items`、`total`、`page`、`page_size`、`total_pages`）

### POST /resource/tokens — 创建个人访问令牌

权限：`resource_tokens:create`
请求：{ name*, scopes*, expires_at }
响应 201：data = PATCreateResponse
说明：明文 token 仅在创建响应中返回一次，之后不可再读。服务端只存哈希。scopes 限于 `skills:read`、`agents:run`。不能替代 HTTPS/TLS。

### DELETE /resource/tokens/{id} — 删除个人访问令牌

权限：`resource_tokens:delete`
路径参数：id*: integer
响应 200

## 对象形状

### CliDetectResult

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `detected` | `boolean` |  |  |
| `output` | `string` |  |  |
| `path` | `string` |  |  |
| `version` | `string` |  |  |
| `healthy` | `boolean` |  |  |
| `risk_notice` | `string` |  |  |

### CliCheckUpdateResult

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `current_version` | `string` |  | 已安装版本；未安装为空 |
| `latest_version` | `string` |  | Registry 上的最新版本 |
| `update_available` | `boolean` |  | `latest_version` 高于 `current_version` 时为 true |
| `package` | `string` |  | npm 包名 |
| `registry` | `string` |  | 成功查询所用的 Registry；默认源时为空 |
| `output` | `string` |  | 查询过程日志 |
| `error` | `string` |  | 查询失败时的错误信息 |

### CliExecuteResult

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `success` | `boolean` |  |  |
| `output` | `string` |  |  |
| `error` | `string` |  |  |

### CliInstallSourceInput

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `cli_key` | `string` | 是 |  |
| `name` | `string` | 是 |  |
| `base_url` | `string` | 是 |  |
| `priority` | `integer` |  |  |
| `enabled` | `boolean` |  |  |

### CliRuntimeDefinition

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `key` | `'claude_code' \| 'opencode' \| 'reasonix' \| 'codex'` |  |  |
| `name` | `string` |  |  |
| `binary_name` | `string` |  |  |
| `description` | `string` |  |  |
| `install_status` | `string` |  |  |
| `installed_path` | `string` |  |  |
| `installed_version` | `string` |  |  |
| `healthy` | `boolean` |  |  |
| `risk_notice` | `string` |  |  |
| `api_base_env` | `string` |  |  |
| `default_args` | `string` |  |  |

### Credential

Metadata only; secret never returned

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `name` | `string` |  |  |
| `type` | `'password' \| 'token' \| 'ssh_key' \| 'api_key'` |  |  |
| `username` | `string` |  |  |
| `description` | `string` |  |  |
| `has_secret` | `boolean` |  |  |
| `has_passphrase` | `boolean` |  |  |
| `created_by` | `integer` |  |  |
| `created_at` | `string(date-time)` |  |  |
| `updated_at` | `string(date-time)` |  |  |

### CredentialCreateRequest

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `name` | `string` | 是 |  |
| `type` | `string` | 是 |  |
| `username` | `string` |  |  |
| `secret` | `string` | 是 |  |
| `passphrase` | `string` |  |  |
| `description` | `string` |  |  |

### CredentialPage

组合：`Page` + `inline`

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `items` | `any[]` | 是 |  |
| `total` | `integer` | 是 |  |
| `page` | `integer` | 是 |  |
| `page_size` | `integer` | 是 |  |
| `total_pages` | `integer` | 是 |  |
| `items` | `Credential[]` |  |  |

### CredentialUpdateRequest

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `name` | `string` |  |  |
| `type` | `string` |  |  |
| `username` | `string` |  |  |
| `secret` | `string` |  | Empty keeps existing |
| `passphrase` | `string` |  | Empty keeps existing |
| `description` | `string` |  |  |

### Repository

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `name` | `string` |  |  |
| `description` | `string` |  |  |
| `tags` | `string` |  |  |
| `repo_url` | `string` |  |  |
| `auth_type` | `'none' \| 'credential'` |  |  |
| `credential_id` | `integer` |  |  |
| `created_by` | `integer` |  |  |
| `created_at` | `string(date-time)` |  |  |
| `updated_at` | `string(date-time)` |  |  |

### RepositoryCreateRequest

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `name` | `string` | 是 |  |
| `description` | `string` |  |  |
| `tags` | `string` |  |  |
| `repo_url` | `string` | 是 |  |
| `auth_type` | `string` |  |  |
| `credential_id` | `integer` |  |  |

### RepositoryPage

组合：`Page` + `inline`

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `items` | `any[]` | 是 |  |
| `total` | `integer` | 是 |  |
| `page` | `integer` | 是 |  |
| `page_size` | `integer` | 是 |  |
| `total_pages` | `integer` | 是 |  |
| `items` | `Repository[]` |  |  |

### RepositoryUpdateRequest

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `name` | `string` |  |  |
| `description` | `string` |  |  |
| `tags` | `string` |  |  |
| `repo_url` | `string` |  |  |
| `auth_type` | `string` |  |  |
| `credential_id` | `integer` |  |  |
| `clear_credential` | `boolean` |  |  |

### Server

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `name` | `string` |  |  |
| `host` | `string` |  |  |
| `port` | `integer` |  |  |
| `os_type` | `string` |  |  |
| `username` | `string` |  |  |
| `auth_type` | `string` |  |  |
| `credential_id` | `integer` |  |  |
| `agent_url` | `string` |  |  |
| `agent_credential_id` | `integer` |  |  |
| `description` | `string` |  |  |
| `tags` | `string` |  |  |
| `status` | `string` |  |  |
| `created_by` | `integer` |  |  |
| `created_at` | `string(date-time)` |  |  |
| `updated_at` | `string(date-time)` |  |  |

### ServerCreateRequest

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `name` | `string` | 是 |  |
| `host` | `string` |  |  |
| `port` | `integer` |  |  |
| `os_type` | `string` |  |  |
| `username` | `string` |  |  |
| `auth_type` | `string` |  |  |
| `credential_id` | `integer` |  |  |
| `agent_url` | `string` |  |  |
| `agent_credential_id` | `integer` |  |  |
| `description` | `string` |  |  |
| `tags` | `string` |  |  |

### ServerPage

组合：`Page` + `inline`

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `items` | `any[]` | 是 |  |
| `total` | `integer` | 是 |  |
| `page` | `integer` | 是 |  |
| `page_size` | `integer` | 是 |  |
| `total_pages` | `integer` | 是 |  |
| `items` | `Server[]` |  |  |

### ServerUpdateRequest

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `name` | `string` |  |  |
| `host` | `string` |  |  |
| `port` | `integer` |  |  |
| `os_type` | `string` |  |  |
| `username` | `string` |  |  |
| `auth_type` | `string` |  |  |
| `credential_id` | `integer` |  |  |
| `clear_credential` | `boolean` |  |  |
| `agent_url` | `string` |  |  |
| `agent_credential_id` | `integer` |  |  |
| `clear_agent_credential` | `boolean` |  |  |
| `description` | `string` |  |  |
| `tags` | `string` |  |  |

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
