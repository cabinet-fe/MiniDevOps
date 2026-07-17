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
请求：{ name, description, enabled, cli_key, system_prompt, skill_ids, build_job_ids, output_dir, artifact_format, max_artifacts, timeout_sec }
响应 201
说明：创建后同步持久工作区 `{workspace}/agents/agent-{id}/`（技能解压到 `.agents/skills`，`job-{id}` 软链到构建任务工作区）。

### GET /ai/agents/{id} — 获取 Agent

权限：`ai.agents:view`
路径参数：id*: integer
响应 200

### PUT /ai/agents/{id} — 更新 Agent

权限：`ai.agents:update`
路径参数：id*: integer
请求：{ name, description, enabled, cli_key, system_prompt, skill_ids, build_job_ids, output_dir, artifact_format, max_artifacts, timeout_sec }
响应 200
说明：更新后重新同步持久工作区。

### DELETE /ai/agents/{id} — 删除 Agent

权限：`ai.agents:delete`
路径参数：id*: integer
响应 200
说明：删除记录并清理 `{workspace}/agents/agent-{id}/`。

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
说明：在 Agent 持久工作区执行；产出目录为 `runs/run-{id}/output`（环境变量 `BEDROCK_AGENT_OUTPUT`）；成功且非空时打包制品并写入 `artifact_path`。

### POST /ai/agents/{id}/api-runs — API 触发 Agent 运行（需 PAT scope）

权限：`ai.agents:execute`
路径参数：id*: integer
响应 202
错误：401 / 403
说明：JWT with `ai.agents:execute` or PAT with scope `agents:run`.

### GET /ai/runs — 列出 Agent 运行记录

权限：`ai.runs:view`
查询参数：page: integer, page_size: integer, agent_id: integer, status: string
响应 200

### GET /ai/runs/{id} — 获取 Agent 运行记录

权限：`ai.runs:view`
路径参数：id*: integer
响应 200

### GET /ai/runs/{id}/artifact — 下载 Agent 运行制品

权限：`ai.runs:view`
路径参数：id*: integer
响应 200：data = binary
错误：404
说明：仅当运行成功且存在 `artifact_path` 时可下载（zip 或 tar.gz，由 Agent 的 `artifact_format` 决定）。

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
| `skill_ids` | `integer[]` |  | 解压到工作区 `.agents/skills/{id}/` |
| `build_job_ids` | `integer[]` |  | 软链 `job-{id}` → `{workspace}/repo-{repoID}/job-{jobID}` |
| `output_dir` | `string` |  | 相对名提示，默认 `output`；实际每次 run 使用独立 `runs/run-{id}/output` |
| `artifact_format` | `'zip' \| 'gzip'` |  | 默认 `gzip` |
| `max_artifacts` | `integer` |  | 按 Agent 保留最近 N 个制品文件，默认 10 |
| `stream_output` | `boolean` |  | 启用后使用 CLI 默认可读流式输出；关闭时部分 CLI 仅输出最终摘要（如 Reasonix `-p`），默认 `false` |
| `timeout_sec` | `integer` |  |  |

### AgentRun

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `agent_id` | `integer` |  |  |
| `trigger_type` | `string` |  |  |
| `status` | `string` |  |  |
| `work_dir` | `string` |  | 本次使用的 Agent 根目录 |
| `artifact_path` | `string` |  | 打包后的制品路径；空表示无制品 |
| `build_run_id` | `integer` |  |  |
| `project_id` | `integer` |  |  |
| `doc_node_id` | `integer` |  |  |
| `error_message` | `string` |  |  |
| `output_text` | `string` |  |  |
| `created_at` | `string` |  |  |

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
