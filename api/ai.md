# AI

Agents、运行记录、Skills。

通用约定（信封、分页、认证）见 [.agents/api.md](../.agents/api.md)。
业务语义与权限模型见 [docs/DESIGN.md](../docs/DESIGN.md)。
AI CLI 运行时管理（列表/检测/安装/升级/卸载/安装源）已迁入资源管理域，见 [resource.md](resource.md)。

## Agents

工作区与制品语义：

- 每个 Agent 唯一对应持久根工作区 `{workspace}/agents/agent-{id}/`；所有 Run 直接在该根目录执行，跨 Run 复用，启动新 Run 时不清空根目录已有文件。
- 每个 Agent 另有一个固定产出目录 `{agentRoot}/{output_dir}`（`output_dir` 默认为相对名 `output`）。CLI 注入 `BEDROCK_AGENT_WORKDIR`（根）与 `BEDROCK_AGENT_OUTPUT`（固定产出目录）。不创建 `runs/run-{id}/output` 或任何 per-run 输出子目录；后续 Run 可覆盖同一产出目录内容（运行前可清空该产出子目录，但不删除 Agent 根下其他文件）。
- AgentRun 只保存状态、日志、文本输出和 `work_dir` 等运行记录，不绑定、归档或提供文件制品下载（无 `artifact_path` / `GET /ai/runs/:id/artifact`）。此约束不影响 CI/CD BuildRun 的制品归档与下载。

### GET /ai/agents — 列出 Agents

权限：`ai.agents:view`
查询参数：page: integer, page_size: integer
响应 200

### POST /ai/agents — 创建 Agent

权限：`ai.agents:create`
请求：{ name, description, enabled, cli_key, system_prompt, skill_ids, build_job_ids, output_dir, stream_output, timeout_sec }
响应 201
说明：创建后同步持久根工作区 `{workspace}/agents/agent-{id}/`（技能解压到 `.agents/skills`，`job-{id}` 软链到构建任务工作区）。`output_dir` 为相对产出目录名，默认 `output`。

### GET /ai/agents/{id} — 获取 Agent

权限：`ai.agents:view`
路径参数：id*: integer
响应 200

### PUT /ai/agents/{id} — 更新 Agent

权限：`ai.agents:update`
路径参数：id*: integer
请求：{ name, description, enabled, cli_key, system_prompt, skill_ids, build_job_ids, output_dir, stream_output, timeout_sec }
响应 200
说明：更新后重新同步持久根工作区，但不清空其中已有文件。

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
说明：直接在 Agent 持久根工作区执行；环境提供 `BEDROCK_AGENT_WORKDIR` 与 `BEDROCK_AGENT_OUTPUT`（固定产出目录）。不创建 Run 专属目录，不归档文件制品。

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
说明：返回状态、日志/文本输出与 `work_dir` 等记录；AgentRun 无文件制品字段或下载端点。

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
| `output_dir` | `string` |  | 相对产出目录名；默认 `output`；路径为 `{agentRoot}/{output_dir}`，跨 Run 固定复用 |
| `stream_output` | `boolean` |  | 启用后使用 CLI 默认可读流式输出；关闭时部分 CLI 仅输出最终摘要（如 Reasonix `-p`），默认 `false` |
| `timeout_sec` | `integer` |  |  |

### AgentRun

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `agent_id` | `integer` |  |  |
| `trigger_type` | `string` |  |  |
| `status` | `string` |  |  |
| `work_dir` | `string` |  | Agent 持久根工作区；同一 Agent 的 Run 复用相同路径 |
| `build_run_id` | `integer` |  |  |
| `project_id` | `integer` |  |  |
| `doc_node_id` | `integer` |  |  |
| `error_message` | `string` |  |  |
| `output_text` | `string` |  |  |
| `created_at` | `string` |  |  |

CLI 相关对象形状（CliDetectResult、CliCheckUpdateResult、CliExecuteResult、CliInstallSourceInput、CliRuntimeDefinition）见 [resource.md](resource.md)。
