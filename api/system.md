# 系统管理

用户、角色、RBAC 资源、菜单、字典、操作日志、通知。

通用约定（信封、分页、认证）见 [.agents/api.md](../.agents/api.md)。
业务语义与权限模型见 [docs/DESIGN.md](../docs/DESIGN.md)。

## 用户

### GET /users — 列出用户

权限：`system.users:view`
查询参数：page: integer, page_size: integer
响应 200：data = UserPage
错误：401 / 403

### POST /users — 创建用户

权限：`system.users:create`
请求：{ username*, password*, display_name, email, is_active, role_ids }
响应 201：data = User
错误：400 / 403

### GET /users/{id} — 获取用户

权限：`system.users:view`
路径参数：id*: integer
响应 200：data = User
错误：403 / 404

### PUT /users/{id} — 更新用户

权限：`system.users:update`
路径参数：id*: integer
请求：{ display_name, email, password, is_active, role_ids }
响应 200：data = User
错误：400 / 403

### DELETE /users/{id} — 删除用户

权限：`system.users:delete`
路径参数：id*: integer
响应 200
错误：400 / 403

## 角色

### GET /roles — 列出角色

权限：`system.roles:view`
查询参数：page: integer, page_size: integer
响应 200：data = RolePage
错误：403

### POST /roles — 创建角色

权限：`system.roles:create`
请求：{ name*, code*, description, permissions }
响应 201：data = Role
错误：400 / 403

### GET /roles/{id} — 获取角色

权限：`system.roles:view`
路径参数：id*: integer
响应 200：data = Role
错误：404

### PUT /roles/{id} — 更新角色

权限：`system.roles:update`
路径参数：id*: integer
请求：{ name, description }
响应 200：data = Role
错误：400

### DELETE /roles/{id} — 删除角色

权限：`system.roles:delete`
路径参数：id*: integer
响应 200

### PUT /roles/{id}/permissions — 替换角色权限码

权限：`system.roles:update`
路径参数：id*: integer
请求：{ permissions* }
响应 200：data = Role
错误：400 / 403

## RBAC 资源

### GET /rbac/resources — 列出 RBAC 资源树

权限：`system.resources:view`
查询参数：keyword: string, type: string(menu|page|action|card), enabled: boolean
响应 200：data = object
错误：403
说明：返回树形 `items`。有筛选时匹配 path / 菜单标题 / 路由，并保留匹配节点的祖先以维持树结构。

### POST /rbac/resources — 创建 RBAC 资源

权限：`system.resources:create`
请求：{ path*, type*, parent_id, enabled, sort_key, title, route }
响应 201：data = RbacResource
错误：400 / 403

### GET /rbac/resources/{id} — 获取 RBAC 资源

权限：`system.resources:view`
路径参数：id*: integer
响应 200：data = RbacResource
错误：404

### PUT /rbac/resources/{id} — 更新 RBAC 资源

权限：`system.resources:update`
路径参数：id*: integer
请求：{ enabled, sort_key, title, route }
响应 200：data = RbacResource
错误：400

### DELETE /rbac/resources/{id} — 删除 RBAC 资源

权限：`system.resources:delete`
路径参数：id*: integer
响应 200
错误：400

### PUT /rbac/resources/{id}/icon — 更新资源图标

权限：`system.resources:update`
路径参数：id*: integer
请求：{ icon_base64*, icon_mime }
响应 200：data = RbacResource
错误：400 / 403
说明：仅顶级菜单类型资源可设置图标。需要 `system.resources:update`。

## 菜单

### GET /menus — 列出菜单资源树

权限：`system.roles:update`
响应 200：data = object
错误：403
说明：仅菜单类型节点，供角色权限编辑器使用。需要 `system.roles:update`。

## 字典

### GET /dictionaries — 列出字典

权限：`system.dictionaries:view`
查询参数：page: integer, page_size: integer
响应 200：data = DictionaryPage
错误：403

### POST /dictionaries — 创建字典

权限：`system.dictionaries:create`
请求：{ name*, code*, description, items }
响应 201：data = Dictionary
错误：400

### GET /dictionaries/{id} — 获取字典

权限：`system.dictionaries:view`
路径参数：id*: integer
响应 200：data = Dictionary
错误：404

### PUT /dictionaries/{id} — 更新字典

权限：`system.dictionaries:update`
路径参数：id*: integer
请求：{ name, description, items }
响应 200：data = Dictionary
错误：400

### DELETE /dictionaries/{id} — 删除字典

权限：`system.dictionaries:delete`
路径参数：id*: integer
响应 200

## 操作日志

### GET /operation-logs — 列出操作日志

权限：`system.operation_logs:view`
查询参数：page: integer, page_size: integer, user_id: integer, action: string, resource_type: string, from: string(date), to: string(date)
响应 200：data = OperationLogPage
错误：403

## 通知

### GET /notifications — 列出当前用户通知

查询参数：page: integer, page_size: integer
响应 200：data = NotificationPage
错误：401

### PUT /notifications/read-all — 全部标为已读

响应 200
错误：401

### PUT /notifications/{id}/read — 单条标为已读

路径参数：id*: integer
响应 200
错误：401

## 对象形状

### DictItem

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `dictionary_id` | `integer` |  |  |
| `label` | `string` |  |  |
| `value` | `string` |  |  |
| `sort_order` | `integer` |  |  |
| `enabled` | `boolean` |  |  |

### Dictionary

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `name` | `string` |  |  |
| `code` | `string` |  |  |
| `description` | `string` |  |  |
| `items` | `DictItem[]` |  |  |
| `created_at` | `string(date-time)` |  |  |
| `updated_at` | `string(date-time)` |  |  |

### DictionaryCreateRequest

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `name` | `string` | 是 |  |
| `code` | `string` | 是 |  |
| `description` | `string` |  |  |
| `items` | `DictItem[]` |  |  |

### DictionaryPage

组合：`Page` + `inline`

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `items` | `any[]` | 是 |  |
| `total` | `integer` | 是 |  |
| `page` | `integer` | 是 |  |
| `page_size` | `integer` | 是 |  |
| `total_pages` | `integer` | 是 |  |
| `items` | `Dictionary[]` |  |  |

### DictionaryUpdateRequest

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `name` | `string` |  |  |
| `description` | `string` |  |  |
| `items` | `DictItem[]` |  |  |

### MenuMetadata

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `resource_id` | `integer` |  |  |
| `title` | `string` |  |  |
| `route` | `string` |  |  |
| `icon_base64` | `string` |  |  |
| `icon_mime` | `string` |  |  |

### Notification

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `user_id` | `integer` |  |  |
| `type` | `string` |  | e.g. build_run_success, build_run_failed, agent_run_success |
| `title` | `string` |  |  |
| `message` | `string` |  |  |
| `build_run_id` | `integer` |  |  |
| `agent_run_id` | `integer` |  |  |
| `is_read` | `boolean` |  |  |
| `created_at` | `string(date-time)` |  |  |

### NotificationPage

组合：`Page` + `inline`

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `items` | `any[]` | 是 |  |
| `total` | `integer` | 是 |  |
| `page` | `integer` | 是 |  |
| `page_size` | `integer` | 是 |  |
| `total_pages` | `integer` | 是 |  |
| `items` | `Notification[]` |  |  |

### OperationLog

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `user_id` | `integer` |  |  |
| `username` | `string` |  |  |
| `action` | `string` |  |  |
| `resource_type` | `string` |  |  |
| `resource_id` | `string` |  |  |
| `details` | `string` |  |  |
| `ip_address` | `string` |  |  |
| `created_at` | `string(date-time)` |  |  |

### OperationLogPage

组合：`Page` + `inline`

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `items` | `any[]` | 是 |  |
| `total` | `integer` | 是 |  |
| `page` | `integer` | 是 |  |
| `page_size` | `integer` | 是 |  |
| `total_pages` | `integer` | 是 |  |
| `items` | `OperationLog[]` |  |  |

### Page

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `items` | `any[]` | 是 |  |
| `total` | `integer` | 是 |  |
| `page` | `integer` | 是 |  |
| `page_size` | `integer` | 是 |  |
| `total_pages` | `integer` | 是 |  |

### RbacResource

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `path` | `string` |  |  |
| `type` | `'menu' \| 'page' \| 'action' \| 'card'` |  |  |
| `parent_id` | `integer` |  |  |
| `enabled` | `boolean` |  |  |
| `sort_key` | `integer` |  |  |
| `menu_metadata` | `MenuMetadata` |  |  |
| `children` | `RbacResource[]` |  |  |
| `created_at` | `string(date-time)` |  |  |
| `updated_at` | `string(date-time)` |  |  |

### RbacResourceCreateRequest

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `path` | `string` | 是 |  |
| `type` | `'menu' \| 'page' \| 'action' \| 'card'` | 是 |  |
| `parent_id` | `integer` |  |  |
| `enabled` | `boolean` |  |  |
| `sort_key` | `integer` |  |  |
| `title` | `string` |  |  |
| `route` | `string` |  |  |

### RbacResourceUpdateRequest

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `enabled` | `boolean` |  |  |
| `sort_key` | `integer` |  |  |
| `title` | `string` |  |  |
| `route` | `string` |  |  |

### Role

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `name` | `string` |  |  |
| `code` | `string` |  |  |
| `description` | `string` |  |  |
| `permissions` | `RolePermission[]` |  |  |
| `created_at` | `string(date-time)` |  |  |
| `updated_at` | `string(date-time)` |  |  |

### RoleCreateRequest

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `name` | `string` | 是 |  |
| `code` | `string` | 是 |  |
| `description` | `string` |  |  |
| `permissions` | `string[]` |  |  |

### RolePage

组合：`Page` + `inline`

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `items` | `any[]` | 是 |  |
| `total` | `integer` | 是 |  |
| `page` | `integer` | 是 |  |
| `page_size` | `integer` | 是 |  |
| `total_pages` | `integer` | 是 |  |
| `items` | `Role[]` |  |  |

### RolePermission

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `role_id` | `integer` |  |  |
| `permission` | `string` |  |  |

### User

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `username` | `string` |  |  |
| `display_name` | `string` |  |  |
| `email` | `string` |  |  |
| `avatar` | `string` |  |  |
| `is_active` | `boolean` |  |  |
| `is_super_admin` | `boolean` |  |  |
| `role_ids` | `integer[]` |  |  |
| `created_at` | `string(date-time)` |  |  |
| `updated_at` | `string(date-time)` |  |  |

### UserCreateRequest

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `username` | `string` | 是 |  |
| `password` | `string` | 是 |  |
| `display_name` | `string` |  |  |
| `email` | `string` |  |  |
| `is_active` | `boolean` |  |  |
| `role_ids` | `integer[]` |  |  |

### UserPage

组合：`Page` + `inline`

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `items` | `any[]` | 是 |  |
| `total` | `integer` | 是 |  |
| `page` | `integer` | 是 |  |
| `page_size` | `integer` | 是 |  |
| `total_pages` | `integer` | 是 |  |
| `items` | `User[]` |  |  |

### UserUpdateRequest

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `display_name` | `string` |  |  |
| `email` | `string` |  |  |
| `password` | `string` |  |  |
| `is_active` | `boolean` |  |  |
| `role_ids` | `integer[]` |  |  |
