# 资源管理

代码仓库、服务器、凭证（与 CI/CD 构建域解耦的共享资源）。

通用约定（信封、分页、认证）见 [.agents/api.md](../.agents/api.md)。
业务语义与权限模型见 [docs/DESIGN.md](../docs/DESIGN.md)。

## 代码仓库

### GET /resource/repositories — 列出代码仓库

权限：`resource.repositories:view`
查询参数：page: integer, page_size: integer, keyword: string
响应 200：data = RepositoryPage
错误：403

### POST /resource/repositories — 创建代码仓库

权限：`resource.repositories:create`
请求：{ name*, description, tags, repo_url*, auth_type, credential_id }
响应 201：data = Repository
错误：403

### GET /resource/repositories/{id} — 获取代码仓库

权限：`resource.repositories:view`
路径参数：id*: integer
响应 200：data = Repository
错误：404

### PUT /resource/repositories/{id} — 更新代码仓库

权限：`resource.repositories:update`
路径参数：id*: integer
请求：{ name, description, tags, repo_url, auth_type, credential_id, clear_credential }
响应 200：data = Repository
错误：403

### DELETE /resource/repositories/{id} — 删除代码仓库

权限：`resource.repositories:delete`
路径参数：id*: integer
响应 200
错误：409

### GET /resource/repositories/{id}/branches — 列出远程分支

权限：`resource.repositories:view`
路径参数：id*: integer
响应 200：data = object

### POST /resource/repositories/{id}/test — 测试拉取 / 列分支

权限：`resource.repositories:view`
路径参数：id*: integer
响应 200

## 凭证

### GET /resource/credentials — 列出凭证（仅元数据，不含明文）

权限：`resource.credentials:view`
查询参数：page: integer, page_size: integer, keyword: string
响应 200：data = CredentialPage

### POST /resource/credentials — 创建凭证

权限：`resource.credentials:create`
请求：{ name*, type*, username, secret*, passphrase, description }
响应 201：data = Credential

### GET /resource/credentials/{id} — 获取凭证元数据

权限：`resource.credentials:view`
路径参数：id*: integer
响应 200：data = Credential

### PUT /resource/credentials/{id} — 更新凭证（secret 为空则保留原值）

权限：`resource.credentials:update`
路径参数：id*: integer
请求：{ name, type, username, secret, passphrase, description }
响应 200：data = Credential

### DELETE /resource/credentials/{id} — 删除凭证

权限：`resource.credentials:delete`
路径参数：id*: integer
响应 200
错误：409

## 服务器

### GET /resource/servers — 列出服务器

权限：`resource.servers:view`
查询参数：page: integer, page_size: integer, keyword: string, tag: string
响应 200：data = ServerPage

### POST /resource/servers — 创建服务器

权限：`resource.servers:create`
请求：{ name*, host, port, os_type, username, auth_type, credential_id, agent_url, agent_credential_id, description, tags }
响应 201：data = Server

### GET /resource/servers/{id} — 获取服务器

权限：`resource.servers:view`
路径参数：id*: integer
响应 200：data = Server

### PUT /resource/servers/{id} — 更新服务器

权限：`resource.servers:update`
路径参数：id*: integer
请求：{ name, host, port, os_type, username, auth_type, credential_id, clear_credential, agent_url, agent_credential_id, clear_agent_credential, description, tags }
响应 200：data = Server

### DELETE /resource/servers/{id} — 删除服务器

权限：`resource.servers:delete`
路径参数：id*: integer
响应 200
错误：409

### POST /resource/servers/{id}/test — 测试 SSH / Agent 连通性

权限：`resource.servers:view`
路径参数：id*: integer
响应 200

## 对象形状

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
