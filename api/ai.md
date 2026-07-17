# AI

AI CLI、Agents、运行记录、Skills。

通用约定（信封、分页、认证）见 [.agents/api.md](../.agents/api.md)。
业务语义与权限模型见 [docs/DESIGN.md](../docs/DESIGN.md)。

## AI CLI

### GET /ai/clis — 列出 AI CLI

权限：`ai.clis:view`
响应 200：data = object
说明：四套并行 CLI（Claude Code、OpenCode、Reasonix、Codex）。与 Bedrock 同 UID 执行，无 OS/容器沙箱。

### POST /ai/clis/{key}/detect — 检测 AI CLI

权限：`ai.clis:execute`
路径参数：key*: string
响应 200：data = CliDetectResult

### POST /ai/clis/{key}/install — 安装 AI CLI

权限：`ai.clis:execute`
路径参数：key*: string
请求：{ version }
响应 200：data = CliExecuteResult

### POST /ai/clis/{key}/upgrade — 升级 AI CLI

权限：`ai.clis:execute`
路径参数：key*: string
请求：{ version }
响应 200：data = CliExecuteResult

### POST /ai/clis/{key}/uninstall — 卸载 AI CLI

权限：`ai.clis:execute`
路径参数：key*: string
响应 200：data = CliExecuteResult

### GET /ai/cli-sources — 列出 CLI 安装源

权限：`ai.clis:view`
查询参数：cli_key: string
响应 200

### POST /ai/cli-sources — 创建 CLI 安装源

权限：`ai.clis:create`
请求：{ cli_key*, name*, base_url*, priority, enabled }
响应 201

### PUT /ai/cli-sources/{id} — 更新 CLI 安装源

权限：`ai.clis:update`
路径参数：id*: integer
请求：{ cli_key*, name*, base_url*, priority, enabled }
响应 200

### DELETE /ai/cli-sources/{id} — 删除 CLI 安装源

权限：`ai.clis:delete`
路径参数：id*: integer
响应 200

## Agents

### GET /ai/agents — 列出 Agents

权限：`ai.agents:view`
查询参数：page: integer, page_size: integer
响应 200

### POST /ai/agents — 创建 Agent

权限：`ai.agents:create`
请求：{ name, description, enabled, cli_key, system_prompt, skill_ids, repository_id, timeout_sec }
响应 201

### GET /ai/agents/{id} — 获取 Agent

权限：`ai.agents:view`
路径参数：id*: integer
响应 200

### PUT /ai/agents/{id} — 更新 Agent

权限：`ai.agents:update`
路径参数：id*: integer
请求：{ name, description, enabled, cli_key, system_prompt, skill_ids, repository_id, timeout_sec }
响应 200

### DELETE /ai/agents/{id} — 删除 Agent

权限：`ai.agents:delete`
路径参数：id*: integer
响应 200

### GET /ai/agents/{id}/triggers — 列出触发器

权限：`ai.agents:view`
路径参数：id*: integer
响应 200

### POST /ai/agents/{id}/triggers — 创建触发器

权限：`ai.agents:update`
路径参数：id*: integer
请求：{ type*, enabled, cron_expression, cron_timezone, build_job_id, build_event }
响应 201
说明：类型包括 manual、api、cron（IANA 时区；不重叠、不补跑错过的任务）、build_event。

### PUT /ai/agents/{id}/triggers/{tid} — 更新触发器

权限：`ai.agents:update`
路径参数：id*: integer, tid*: integer
请求：{ type*, enabled, cron_expression, cron_timezone, build_job_id, build_event }
响应 200

### DELETE /ai/agents/{id}/triggers/{tid} — 删除触发器

权限：`ai.agents:update`
路径参数：id*: integer, tid*: integer
响应 200

### POST /ai/agents/{id}/runs — 手动触发 Agent 运行

权限：`ai.agents:execute`
路径参数：id*: integer
响应 202
说明：需要 `ai.agents:execute`。上下文仅为 system_prompt + 所选仓库。

### POST /ai/agents/{id}/api-runs — API 触发 Agent 运行（需 PAT scope）

权限：`ai.agents:execute`
路径参数：id*: integer
响应 202
错误：401 / 403
说明：JWT with `ai.agents:execute` or PAT with scope `agents:run`.

### GET /ai/runs — 列出 Agent 运行记录

权限：`ai.agents:view`
查询参数：page: integer, page_size: integer, agent_id: integer, status: string
响应 200

### GET /ai/runs/{id} — 获取 Agent 运行记录

权限：`ai.agents:view`
路径参数：id*: integer
响应 200

### POST /ai/runs/{id}/cancel — 取消 Agent 运行

权限：`ai.agents:execute`
路径参数：id*: integer
响应 200：Cancelled

## Skills

### GET /skills — 列出 Skills

权限：`ai.skills:view`
查询参数：page: integer, page_size: integer
响应 200
说明：公开 Skill 需 view 权限可见；私有仅创建者可见。

### POST /skills — 创建 Skill

权限：`ai.skills:create`
请求：multipart: { name, description, visibility, file* }
响应 201
错误：422
说明：需要 SKILL.md；防 Zip Slip / zip bomb；默认上限 50MB（经 StorageService）。

### GET /skills/{id} — 获取 Skill

权限：`ai.skills:view`
路径参数：id*: integer
响应 200

### PUT /skills/{id} — 覆盖更新 Skill

权限：`ai.skills:update`
路径参数：id*: integer
请求：multipart: { name, description, visibility, file* }
响应 200：Updated

### DELETE /skills/{id} — 删除 Skill

权限：`ai.skills:delete`
路径参数：id*: integer
响应 200

### GET /skills/{id}/package — 下载 Skill 包

权限：`ai.skills:download`
路径参数：id*: integer
响应 200：data = binary
错误：401 / 403
说明：JWT 需 `ai.skills:download`，或 PAT scope `skills:read`。

## 对象形状

### AgentTriggerInput

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `type` | `'manual' \| 'api' \| 'cron' \| 'build_event'` | 是 |  |
| `enabled` | `boolean` |  |  |
| `cron_expression` | `string` |  |  |
| `cron_timezone` | `string` |  | IANA timezone |
| `build_job_id` | `integer` |  |  |
| `build_event` | `'artifact_ready' \| 'distribution_finished'` |  |  |

### AiAgentInput

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `name` | `string` |  |  |
| `description` | `string` |  |  |
| `enabled` | `boolean` |  |  |
| `cli_key` | `string` |  |  |
| `system_prompt` | `string` |  |  |
| `skill_ids` | `integer[]` |  |  |
| `repository_id` | `integer` |  |  |
| `timeout_sec` | `integer` |  |  |

### CliDetectResult

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `detected` | `boolean` |  |  |
| `output` | `string` |  |  |
| `path` | `string` |  |  |
| `version` | `string` |  |  |
| `healthy` | `boolean` |  |  |
| `risk_notice` | `string` |  |  |

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
