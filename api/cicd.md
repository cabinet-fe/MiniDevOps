# CI/CD

代码仓库、凭证、构建任务、构建运行、服务器、Webhook。

通用约定（信封、分页、认证）见 [.agents/api.md](../.agents/api.md)。
业务语义与权限模型见 [docs/DESIGN.md](../docs/DESIGN.md)。

## 代码仓库

### GET /repositories — 列出代码仓库

权限：`cicd.repositories:view`
查询参数：page: integer, page_size: integer, keyword: string
响应 200：data = RepositoryPage
错误：403

### POST /repositories — 创建代码仓库

权限：`cicd.repositories:create`
请求：{ name*, description, tags, repo_url*, auth_type, credential_id }
响应 201：data = Repository
错误：403

### GET /repositories/{id} — 获取代码仓库

权限：`cicd.repositories:view`
路径参数：id*: integer
响应 200：data = Repository
错误：404

### PUT /repositories/{id} — 更新代码仓库

权限：`cicd.repositories:update`
路径参数：id*: integer
请求：{ name, description, tags, repo_url, auth_type, credential_id, clear_credential }
响应 200：data = Repository
错误：403

### DELETE /repositories/{id} — 删除代码仓库

权限：`cicd.repositories:delete`
路径参数：id*: integer
响应 200
错误：409

### GET /repositories/{id}/branches — 列出远程分支

权限：`cicd.repositories:view`
路径参数：id*: integer
响应 200：data = object

### POST /repositories/{id}/test — 测试拉取 / 列分支

权限：`cicd.repositories:view`
路径参数：id*: integer
响应 200

## 凭证

### GET /credentials — 列出凭证（仅元数据，不含明文）

权限：`cicd.credentials:view`
查询参数：page: integer, page_size: integer, keyword: string
响应 200：data = CredentialPage

### POST /credentials — 创建凭证

权限：`cicd.credentials:create`
请求：{ name*, type*, username, secret*, passphrase, description }
响应 201：data = Credential

### GET /credentials/{id} — 获取凭证元数据

权限：`cicd.credentials:view`
路径参数：id*: integer
响应 200：data = Credential

### PUT /credentials/{id} — 更新凭证（secret 为空则保留原值）

权限：`cicd.credentials:update`
路径参数：id*: integer
请求：{ name, type, username, secret, passphrase, description }
响应 200：data = Credential

### DELETE /credentials/{id} — 删除凭证

权限：`cicd.credentials:delete`
路径参数：id*: integer
响应 200
错误：409

## 构建任务

### GET /build-jobs — 列出构建任务

权限：`cicd.build_jobs:view`
查询参数：page: integer, page_size: integer, repository_id: integer, keyword: string
响应 200：data = BuildJobPage

### POST /build-jobs — 创建构建任务

权限：`cicd.build_jobs:create`
请求：{ repository_id*, name*, description, enabled, branch, shallow_clone, build_script_type, build_script, work_dir, output_dir, cache_paths, env_var_names, trigger_manual, trigger_webhook, trigger_cron, webhook_secret, webhook_type, webhook_ref_path, webhook_commit_path, webhook_message_path, cron_expression, cron_timezone, max_artifacts, artifact_format, agent_trigger_event, agent_id, deploy_targets }
响应 201：data = BuildJob

### GET /build-jobs/{id} — 获取构建任务（含部署目标）

权限：`cicd.build_jobs:view`
路径参数：id*: integer
响应 200：data = BuildJob

### PUT /build-jobs/{id} — 更新构建任务 / 替换部署目标

权限：`cicd.build_jobs:update`
路径参数：id*: integer
请求：{ name, description, enabled, branch, shallow_clone, build_script_type, build_script, work_dir, output_dir, cache_paths, env_var_names, trigger_manual, trigger_webhook, trigger_cron, webhook_secret, webhook_type, webhook_ref_path, webhook_commit_path, webhook_message_path, cron_expression, cron_timezone, max_artifacts, artifact_format, agent_trigger_event, agent_id, deploy_targets }
响应 200：data = BuildJob

### DELETE /build-jobs/{id} — 删除构建任务

权限：`cicd.build_jobs:delete`
路径参数：id*: integer
响应 200

### GET /build-jobs/{id}/webhook-secret — 查看 Webhook 密钥与 URL

权限：`cicd.build_jobs:view`
路径参数：id*: integer
响应 200

### POST /build-jobs/{id}/webhook-secret/rotate — 轮换 Webhook 密钥

权限：`cicd.build_jobs:update`
路径参数：id*: integer
响应 200

### POST /build-jobs/{id}/runs — 入队构建运行

权限：`cicd.build_jobs:execute`
路径参数：id*: integer
请求：{ branch, trigger_type }
响应 202：data = BuildRun
说明：触发时只需 `cicd.build_jobs:execute`；不要求凭证 `:use`（执行时使用已绑定凭证快照）。

## 构建运行

### GET /build-runs — 列出构建运行

权限：`cicd.build_runs:view`
查询参数：page: integer, page_size: integer, build_job_id: integer, status: string
响应 200：data = BuildRunPage

### GET /build-runs/{id} — 获取构建运行详情（含部署尝试）

权限：`cicd.build_runs:view`
路径参数：id*: integer
响应 200：data = BuildRun

### POST /build-runs/{id}/cancel — 取消构建运行

权限：`cicd.build_jobs:execute`
路径参数：id*: integer
响应 200：data = BuildRun

### POST /build-runs/{id}/retry — 重试（新建一次构建运行）

权限：`cicd.build_jobs:execute`
路径参数：id*: integer
响应 202：data = BuildRun

### POST /build-runs/{id}/redeploy — 在同一构建运行上重新部署

权限：`cicd.build_jobs:execute`
路径参数：id*: integer
请求：{ target_ids }
响应 202：data = BuildRun

### GET /build-runs/{id}/artifact — 下载构建制品

权限：`cicd.build_runs:view`
路径参数：id*: integer
响应 200：data = binary

### GET /build-runs/{id}/log — 获取构建日志文本

权限：`cicd.build_runs:view`
路径参数：id*: integer
响应 200：data = text/plain

### GET /ws/build-runs/{id}/logs — 构建日志 WebSocket（实时）

路径前缀为 `/ws`（非 `/api/v1`）。查询参数 `token` 携带 JWT（与其它 WebSocket 一致）。

权限：`cicd.build_runs:view`
路径参数：id*: integer
查询参数：token*: string

连接成功后：

1. 服务端先按行回放已有日志文件（若有）。
2. 后续推送两类文本帧：
   - 日志行：追加到终端输出。
   - 控制帧 `__REFRESH__`：元数据（status / stage / distribution_summary / deploy_attempts 等）已变更；客户端应重新请求 `GET /build-runs/{id}`，勿写入日志视图。

`__REFRESH__` 仅经 WebSocket 广播，不写入日志文件。

## 服务器

### GET /servers — 列出服务器

权限：`cicd.servers:view`
查询参数：page: integer, page_size: integer, keyword: string, tag: string
响应 200：data = ServerPage

### POST /servers — 创建服务器

权限：`cicd.servers:create`
请求：{ name*, host, port, os_type, username, auth_type, credential_id, agent_url, agent_credential_id, description, tags }
响应 201：data = Server

### GET /servers/{id} — 获取服务器

权限：`cicd.servers:view`
路径参数：id*: integer
响应 200：data = Server

### PUT /servers/{id} — 更新服务器

权限：`cicd.servers:update`
路径参数：id*: integer
请求：{ name, host, port, os_type, username, auth_type, credential_id, clear_credential, agent_url, agent_credential_id, clear_agent_credential, description, tags }
响应 200：data = Server

### DELETE /servers/{id} — 删除服务器

权限：`cicd.servers:delete`
路径参数：id*: integer
响应 200
错误：409

### POST /servers/{id}/test — 测试 SSH / Agent 连通性

权限：`cicd.servers:view`
路径参数：id*: integer
响应 200

## Webhook

### POST /webhook/jobs/{build_job_id}/{secret} — 接收构建任务 Webhook

认证：不需要
路径参数：build_job_id*: integer, secret*: string
响应 202（可能为重复投递，`triggered=0`）
错误：401
说明：优先校验签名；也可用 URL 中的 secret。按 delivery 去重。

### POST /webhook/repos/{repository_id}/{secret} — 已废弃的仓库 Webhook（返回 410）

认证：不需要
状态：已废弃
路径参数：repository_id*: integer, secret*: string
错误：410

## 对象形状

### BuildDeployAttempt

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `build_run_id` | `integer` |  |  |
| `batch_no` | `integer` |  |  |
| `deploy_target_id` | `integer` |  |  |
| `target_snapshot_json` | `string` |  |  |
| `status` | `string` |  |  |
| `log_path` | `string` |  |  |
| `error_message` | `string` |  |  |
| `started_at` | `string(date-time)` |  |  |
| `finished_at` | `string(date-time)` |  |  |
| `created_at` | `string(date-time)` |  |  |

### BuildJob

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `repository_id` | `integer` |  |  |
| `name` | `string` |  |  |
| `description` | `string` |  |  |
| `enabled` | `boolean` |  |  |
| `branch` | `string` |  |  |
| `shallow_clone` | `boolean` |  |  |
| `build_script_type` | `string` |  |  |
| `build_script` | `string` |  |  |
| `work_dir` | `string` |  |  |
| `output_dir` | `string` |  |  |
| `cache_paths` | `string` |  |  |
| `env_var_names` | `string[]` |  |  |
| `trigger_manual` | `boolean` |  |  |
| `trigger_webhook` | `boolean` |  |  |
| `trigger_cron` | `boolean` |  |  |
| `webhook_secret` | `string` |  | Only present on secret view/rotate |
| `webhook_type` | `string` |  |  |
| `webhook_ref_path` | `string` |  |  |
| `webhook_commit_path` | `string` |  |  |
| `webhook_message_path` | `string` |  |  |
| `cron_expression` | `string` |  |  |
| `cron_timezone` | `string` |  |  |
| `max_artifacts` | `integer` |  |  |
| `artifact_format` | `string` |  |  |
| `agent_trigger_event` | `'artifact_ready' \| 'distribution_finished' \| 'none'` |  | Default artifact_ready; override distribution_finished or none |
| `agent_id` | `integer` |  | Optional agent bound for build-event trigger |
| `deploy_targets` | `DeployTarget[]` |  |  |
| `created_by` | `integer` |  |  |
| `created_at` | `string(date-time)` |  |  |
| `updated_at` | `string(date-time)` |  |  |

### BuildJobCreateRequest

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `repository_id` | `integer` | 是 |  |
| `name` | `string` | 是 |  |
| `description` | `string` |  |  |
| `enabled` | `boolean` |  |  |
| `branch` | `string` |  |  |
| `shallow_clone` | `boolean` |  |  |
| `build_script_type` | `string` |  |  |
| `build_script` | `string` |  |  |
| `work_dir` | `string` |  |  |
| `output_dir` | `string` |  |  |
| `cache_paths` | `string` |  |  |
| `env_var_names` | `string[]` |  |  |
| `trigger_manual` | `boolean` |  |  |
| `trigger_webhook` | `boolean` |  |  |
| `trigger_cron` | `boolean` |  |  |
| `webhook_secret` | `string` |  | Only present on secret view/rotate |
| `webhook_type` | `string` |  |  |
| `webhook_ref_path` | `string` |  |  |
| `webhook_commit_path` | `string` |  |  |
| `webhook_message_path` | `string` |  |  |
| `cron_expression` | `string` |  |  |
| `cron_timezone` | `string` |  |  |
| `max_artifacts` | `integer` |  |  |
| `artifact_format` | `string` |  |  |
| `agent_trigger_event` | `'artifact_ready' \| 'distribution_finished' \| 'none'` |  |  |
| `agent_id` | `integer` |  |  |
| `deploy_targets` | `DeployTarget[]` |  |  |

### BuildJobPage

组合：`Page` + `inline`

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `items` | `any[]` | 是 |  |
| `total` | `integer` | 是 |  |
| `page` | `integer` | 是 |  |
| `page_size` | `integer` | 是 |  |
| `total_pages` | `integer` | 是 |  |
| `items` | `BuildJob[]` |  |  |

### BuildJobUpdateRequest

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `name` | `string` |  |  |
| `description` | `string` |  |  |
| `enabled` | `boolean` |  |  |
| `branch` | `string` |  |  |
| `shallow_clone` | `boolean` |  |  |
| `build_script_type` | `string` |  |  |
| `build_script` | `string` |  |  |
| `work_dir` | `string` |  |  |
| `output_dir` | `string` |  |  |
| `cache_paths` | `string` |  |  |
| `env_var_names` | `string[]` |  |  |
| `trigger_manual` | `boolean` |  |  |
| `trigger_webhook` | `boolean` |  |  |
| `trigger_cron` | `boolean` |  |  |
| `webhook_secret` | `string` |  | Only present on secret view/rotate |
| `webhook_type` | `string` |  |  |
| `webhook_ref_path` | `string` |  |  |
| `webhook_commit_path` | `string` |  |  |
| `webhook_message_path` | `string` |  |  |
| `cron_expression` | `string` |  |  |
| `cron_timezone` | `string` |  |  |
| `max_artifacts` | `integer` |  |  |
| `artifact_format` | `string` |  |  |
| `agent_trigger_event` | `'artifact_ready' \| 'distribution_finished' \| 'none'` |  |  |
| `agent_id` | `integer` |  |  |
| `deploy_targets` | `DeployTarget[]` |  |  |

### BuildRun

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `build_job_id` | `integer` |  |  |
| `build_number` | `integer` |  |  |
| `status` | `'queued' \| 'running' \| 'success' \| 'failed' \| 'cancelled' \| 'interrupted'` |  |  |
| `stage` | `'pending' \| 'cloning' \| 'building' \| 'archiving' \| 'distributing' \| 'idle'` |  |  |
| `trigger_type` | `string` |  |  |
| `triggered_by` | `integer` |  |  |
| `branch` | `string` |  |  |
| `commit_hash` | `string` |  |  |
| `commit_message` | `string` |  |  |
| `log_path` | `string` |  |  |
| `artifact_path` | `string` |  |  |
| `duration_ms` | `integer` |  |  |
| `error_message` | `string` |  |  |
| `distribution_summary` | `'none' \| 'running' \| 'all_success' \| 'partial' \| 'all_failed' \| 'cancelled'` |  |  |
| `snapshot_json` | `string` |  |  |
| `started_at` | `string(date-time)` |  |  |
| `finished_at` | `string(date-time)` |  |  |
| `created_at` | `string(date-time)` |  |  |
| `deploy_attempts` | `BuildDeployAttempt[]` |  |  |

### BuildRunPage

组合：`Page` + `inline`

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `items` | `any[]` | 是 |  |
| `total` | `integer` | 是 |  |
| `page` | `integer` | 是 |  |
| `page_size` | `integer` | 是 |  |
| `total_pages` | `integer` | 是 |  |
| `items` | `BuildRun[]` |  |  |

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

### DeployTarget

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `build_job_id` | `integer` |  |  |
| `server_id` | `integer` |  |  |
| `remote_path` | `string` |  |  |
| `method` | `'rsync' \| 'sftp' \| 'scp' \| 'agent' \| 'local'` |  |  |
| `post_deploy_script` | `string` |  |  |
| `sort_order` | `integer` |  |  |

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
