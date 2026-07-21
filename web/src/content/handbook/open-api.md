# 开放接口

面向脚本与外部系统对接的 HTTP / WebSocket 接口参考。接口契约真源为仓库 `api/` 目录（按域拆分），本页与其保持一致；业务语义与权限模型见 `docs/DESIGN.md`。

**Base URL**：`{host}/api/v1`（下文路径均省略该前缀）
**WebSocket**：`{host}/ws`（路径前缀不是 `/api/v1`，见「10. WebSocket」）
**数据格式**：JSON，字段统一 `snake_case`

---

## 1. 通用约定

### 1.1 响应信封

所有 JSON 响应使用统一信封：

```json
{ "code": 0, "message": "success", "data": {} }
```

- 成功：`code` 为 `0`；`message` 为 `success`（200）或 `created`（201）。
- 失败：`code` 等于 HTTP 状态码，并附带 `request_id` 便于排查：

```json
{ "code": 403, "message": "forbidden", "request_id": "req_01J..." }
```

- 异步创建（入队类操作）返回 `202`，`data` 中含资源 ID 与当前状态。

### 1.2 错误码

| HTTP | 场景 |
| --- | --- |
| 400 | 参数 / JSON / 登录 cipher 无效 |
| 401 | 未认证、密钥错误、PAT 无效 |
| 403 | RBAC / ACL / 超管门控 / PAT scope 不足 |
| 404 | 资源不存在 |
| 409 | 版本冲突、状态不允许、引用冲突（如删除被引用的仓库） |
| 413 | 上传超限（附件 / 导入包） |
| 422 | 语义校验失败（如 Skill 包缺 SKILL.md） |
| 429 | 限流 |
| 500 / 503 | 内部错误 / 依赖不可用 |

### 1.3 分页

- 请求参数：`page`、`page_size`。
- 响应 `data`：`items`、`total`、`page`、`page_size`、`total_pages`。
- 部分列表不分页（如菜单分组、RBAC 资源树），响应 `data.items` 或直接返回数组，以各域说明为准。

### 1.4 认证方式

| 方式 | 用法 |
| --- | --- |
| 登录 JWT | `Authorization: Bearer <access_token>`，有效期取 `jwt.access_ttl`（默认 2h） |
| 个人访问令牌（PAT） | `Authorization: Bearer br_...`，适合脚本 / 外部系统；明文仅创建时返回一次 |
| 刷新令牌 | HttpOnly Cookie `refresh_token`（Path=/api/v1/auth，不设 Secure）；非浏览器客户端可在 `POST /auth/refresh` 的 body 中传 `refresh_token` |

刷新流程：收到 `401` → 调用 `POST /auth/refresh`（Cookie 自动携带）→ 获得新 `access_token` 并轮换 Cookie → 重试原请求。

### 1.5 幂等

需要防止重复提交的写接口支持 `Idempotency-Key` 请求头；Webhook 按 delivery 去重（重复投递返回 `202` 且 `triggered=0`）。

---

## 2. 快速上手

```bash
HOST=http://127.0.0.1:8080

# 1. 登录（脚本调试可用明文 password；Web 端只发 password_cipher）
curl -fsS -X POST "$HOST/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"your-password"}' \
  -c cookies.txt
# 从响应 data.access_token 取出令牌
ACCESS_TOKEN=<粘贴 access_token>

# 2. 携带令牌调用任意接口
curl -fsS "$HOST/api/v1/auth/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 3. 创建个人访问令牌（PAT），明文仅在响应 data.token 出现一次
curl -fsS -X POST "$HOST/api/v1/resource/tokens" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"ci","scopes":["agents:run"],"expires_in_days":90}'

# 4. 用 PAT 触发 Agent 运行（202 异步）
curl -fsS -X POST "$HOST/api/v1/ai/agents/1/api-runs" \
  -H "Authorization: Bearer br_..."

# 5. 手动触发构建（202 异步）
curl -fsS -X POST "$HOST/api/v1/build-jobs/1/runs" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"trigger_type":"manual"}'

# 6. 实时构建日志（WebSocket，token 走查询参数）
# wscat -c "ws://127.0.0.1:8080/ws/build-runs/123/logs?token=$ACCESS_TOKEN"

# 7. 健康检查（无需认证）
curl -fsS "$HOST/api/v1/health"
# → data: { "status": "ok", "version": "...", "driver": "sqlite" }
```

---

## 3. 认证（Auth）

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| POST | `/auth/login` | 无 | 登录，写入 refresh Cookie |
| POST | `/auth/refresh` | 无 | 刷新访问令牌并轮换 Cookie |
| POST | `/auth/logout` | 登录 | 登出，清除 refresh Cookie |
| GET | `/auth/me` | 登录 | 当前用户 + 权限 + 菜单 |

**关键字段**：

- 登录请求：`username*`、`password_cipher`（hex(IV ‖ AES-256-CBC PKCS#7)，Web 优先）、`password`（明文，仅调试）。
- 登录响应 `data`：`access_token`、`user`、`permissions`（功能 `full_code[]`）、`menus`（两层导航）。`refresh_token` 只经 Set-Cookie 下发，不在 JSON 返回。
- `/auth/me` 响应 `data`：`user`、`permissions`、`menus`；菜单已按 `hidden` / 启用状态 / `super_admin_only` / `{menuCode}:view` 过滤，空分组不返回。
- 错误：登录 400 / 401；refresh 401；me 401 / 404。

---

## 4. 资源管理（Resource）

#### 代码仓库

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/resource/repositories` | `resource_repositories:view` | 分页列出，支持 `page`/`page_size`/`keyword` |
| POST | `/resource/repositories` | `resource_repositories:create` | 创建，201 |
| GET | `/resource/repositories/{id}` | `resource_repositories:view` | 获取单个 |
| PUT | `/resource/repositories/{id}` | `resource_repositories:update` | 更新 |
| DELETE | `/resource/repositories/{id}` | `resource_repositories:delete` | 删除；被引用返回 409 |
| GET | `/resource/repositories/{id}/branches` | `resource_repositories:view` | 读分支缓存（仅本地；未同步过 `items=[]`、`synced_at=null`） |
| POST | `/resource/repositories/{id}/sync-branches` | `resource_repositories:update` | 同步单仓分支缓存 |
| POST | `/resource/repositories/sync-branches` | `resource_repositories:update` | 批量同步，请求 `{ ids?: integer[] }`；`data.items` 含 `{ id, ok, branch_count?, error?, synced_at? }` |
| POST | `/resource/repositories/{id}/test` | `resource_repositories:view` | 测试拉取/列分支；成功时一并刷新分支缓存 |

**关键字段**：创建 `name*`、`repo_url*`，`auth_type` ∈ `none|credential`，`credential_id`；更新额外支持 `clear_credential`（解绑凭证）。响应含 `branches`、`branches_synced_at` 缓存字段。

#### 凭证

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/resource/credentials` | `resource_credentials:view` | 分页列出（`page`/`page_size`/`keyword`），仅元数据 |
| POST | `/resource/credentials` | `resource_credentials:create` | 创建，201 |
| GET | `/resource/credentials/{id}` | `resource_credentials:view` | 获取元数据 |
| PUT | `/resource/credentials/{id}` | `resource_credentials:update` | 更新；`secret`/`passphrase` 为空则保留原值 |
| DELETE | `/resource/credentials/{id}` | `resource_credentials:delete` | 删除；被引用返回 409 |

**关键字段**：`name*`、`type*` ∈ `password|token|ssh_key|api_key`，`secret*`（创建必填）、`username`、`passphrase`。敏感字段永不回显：响应仅含 `has_secret`/`has_passphrase` 布尔位。

#### 服务器

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/resource/servers` | `resource_servers:view` | 分页列出，支持 `page`/`page_size`/`keyword`/`tag` |
| POST | `/resource/servers` | `resource_servers:create` | 创建，201 |
| GET | `/resource/servers/{id}` | `resource_servers:view` | 获取单个 |
| PUT | `/resource/servers/{id}` | `resource_servers:update` | 更新 |
| DELETE | `/resource/servers/{id}` | `resource_servers:delete` | 删除；被引用返回 409 |
| POST | `/resource/servers/{id}/test` | `resource_servers:view` | 测试 SSH / Agent 连通性 |

**关键字段**：`name*`；连接 `host`/`port`/`os_type`/`username`/`auth_type`/`credential_id`；Agent `agent_url`/`agent_credential_id`。更新支持 `clear_credential`、`clear_agent_credential` 解绑。响应含 `status`。

#### AI CLI（含安装源）

四套并行 CLI：`key*` ∈ `claude_code|opencode|reasonix|codex`，与 Bedrock 同 UID 执行，无 OS/容器沙箱。权限统一为 `ops_dev_environments:*`（仅超管），路径挂在 `/resource` 下。

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/resource/clis` | `ops_dev_environments:view` | 列出四套 CLI 运行时（`install_status`/`installed_version`/`healthy` 等） |
| POST | `/resource/clis/{key}/detect` | `ops_dev_environments:execute` | 检测安装情况（`detected`/`path`/`version`/`healthy`/`risk_notice`） |
| POST | `/resource/clis/{key}/check-update` | `ops_dev_environments:execute` | 经 `npm view` 查最新版，返回 `current`/`latest`/`update_available` |
| POST | `/resource/clis/{key}/install` | `ops_dev_environments:execute` | 安装，可传 `{ version }` |
| POST | `/resource/clis/{key}/upgrade` | `ops_dev_environments:execute` | 升级，可传 `{ version }` |
| POST | `/resource/clis/{key}/uninstall` | `ops_dev_environments:execute` | 卸载 |
| GET | `/resource/cli-sources` | `ops_dev_environments:view` | 列出安装源，可按 `cli_key` 过滤 |
| POST | `/resource/cli-sources` | `ops_dev_environments:create` | 创建安装源，201 |
| PUT | `/resource/cli-sources/{id}` | `ops_dev_environments:update` | 更新安装源 |
| DELETE | `/resource/cli-sources/{id}` | `ops_dev_environments:delete` | 删除安装源 |

**关键字段**：安装源 `cli_key*`、`name*`、`base_url*`（npm Registry 地址，安装/升级时拼为 `npm --registry`；未配置启用源则用 npm 默认 Registry）、`priority`、`enabled`。install/upgrade/uninstall 返回 `{ success, output, error }`。

#### 个人访问令牌（PAT）

按 `user_id` 隔离：仅能操作本人令牌。

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/resource/tokens` | `resource_tokens:view` | 分页列出，仅元数据 |
| POST | `/resource/tokens` | `resource_tokens:create` | 创建，201；响应含明文 `token`（仅一次） |
| DELETE | `/resource/tokens/{id}` | `resource_tokens:delete` | 删除 |

**关键字段**：创建 `name*`、`scopes*` ∈ `skills:read|agents:run|docs:write|docs:publish`（仅这四个，可多选）。过期三选一：都不传 = 永不过期；`expires_in_days` 仅允许 `30|90|180|365`；`expires_at` 为自定义 UTC 绝对时间（必须晚于当前，与 `expires_in_days` 互斥）。明文 token（`br_` 前缀 + hex）仅在创建响应的 `data.token` 返回一次，服务端只存哈希、不落日志、不可再读；元数据见 `data.metadata`（`token_prefix`/`scopes`/`expires_at`/`last_used_at` 等）。

---

## 5. CI/CD

#### 构建任务

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/build-jobs` | `cicd_build_jobs:view` | 分页列出；参数 `page`、`page_size`、`repository_id`、`keyword` |
| POST | `/build-jobs` | `cicd_build_jobs:create` | 创建构建任务，201 |
| GET | `/build-jobs/{id}` | `cicd_build_jobs:view` | 详情，含 `deploy_targets` |
| PUT | `/build-jobs/{id}` | `cicd_build_jobs:update` | 更新任务；`deploy_targets` 整体替换 |
| DELETE | `/build-jobs/{id}` | `cicd_build_jobs:delete` | 删除构建任务 |
| GET | `/build-jobs/{id}/webhook-secret` | `cicd_build_jobs:view` | 查看 Webhook 密钥与 URL |
| POST | `/build-jobs/{id}/webhook-secret/rotate` | `cicd_build_jobs:update` | 轮换 Webhook 密钥 |
| POST | `/build-jobs/{id}/runs` | `cicd_build_jobs:execute` | 入队构建运行，202 异步 |

**关键字段**（创建/更新）：`repository_id*`、`name*`、`description`、`enabled`、`branch`、`shallow_clone`、`build_script_type`、`build_script`、`work_dir`、`output_dir`、`cache_paths`、`env_var_names[]`；触发开关 `trigger_manual`/`trigger_webhook`/`trigger_cron`；Webhook 解析 `webhook_type`、`webhook_ref_path`、`webhook_commit_path`、`webhook_message_path`；定时 `cron_expression`、`cron_timezone`；制品 `max_artifacts`、`artifact_format`；构建事件联动 `agent_trigger_event`（`artifact_ready|distribution_finished|none`）、`agent_id`；`deploy_targets[]` 含 `{ server_id, remote_path, method(rsync|sftp|scp|agent|local), post_deploy_script, sort_order }`。

**密钥**：`webhook_secret` 仅在「查看/轮换密钥」两个接口的响应 `data` 中出现；轮换后旧密钥立即失效，需同步更新 Git 平台侧配置。

**入队**：`POST /build-jobs/{id}/runs` 请求体 `{ branch, trigger_type }`，响应 202、`data` = BuildRun；触发只需 `cicd_build_jobs:execute`（执行时使用已绑定的凭证快照）。

#### 构建运行

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/build-runs` | `cicd_build_runs:view` | 分页列出；参数 `page`、`page_size`、`build_job_id`、`status` |
| GET | `/build-runs/{id}` | `cicd_build_runs:view` | 详情，含 `deploy_attempts` |
| POST | `/build-runs/{id}/cancel` | `cicd_build_jobs:execute` | 取消运行 |
| POST | `/build-runs/{id}/retry` | `cicd_build_jobs:execute` | 重试：新建一次运行，202 |
| POST | `/build-runs/{id}/redeploy` | `cicd_build_jobs:execute` | 复用当前运行制品重新部署，202 |
| GET | `/build-runs/{id}/artifact` | `cicd_build_runs:view` | 下载构建制品（二进制流，非 JSON 信封） |
| GET | `/build-runs/{id}/log` | `cicd_build_runs:view` | 构建日志全文（text/plain） |

**关键字段**（BuildRun）：`build_number`、`status`（`queued|running|success|failed|cancelled|interrupted`）、`stage`（`pending|cloning|building|archiving|distributing|idle`）、`trigger_type`、`triggered_by`、`branch`、`commit_hash`、`commit_message`、`duration_ms`、`error_message`、`distribution_summary`（`none|running|all_success|partial|all_failed|cancelled`）、`deploy_attempts[]`（`batch_no`、`deploy_target_id`、`status`、`error_message`）。

**行为**：`retry` 创建全新运行（新 `build_number`）；`redeploy` 请求体 `{ target_ids }` 指定目标；`cancel` 终止排队/执行中的运行。

#### Webhook（Git 平台回调）

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| POST | `/webhook/jobs/:build_job_id/:secret` | 无 | 接收构建任务 Webhook，202 |
| POST | `/webhook/repos/:repository_id/:secret` | 无 | 已废弃的仓库级路径，固定返回 410 Gone |

**说明**：优先校验请求签名，也可用 URL 中的 `secret`；按 delivery 去重，重复投递返回 202 且 `triggered=0`；校验失败返回 401。

---

## 6. AI

#### Agents

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/ai/agents` | `ai_agents:view` | 分页列出 |
| POST | `/ai/agents` | `ai_agents:create` | 创建，201 立即返回；工作区后台异步初始化 |
| GET | `/ai/agents/{id}` | `ai_agents:view` | 详情，含工作区状态 |
| PUT | `/ai/agents/{id}` | `ai_agents:update` | 更新；`workspace_status` 重置为 `pending` 并后台重建工作区 |
| DELETE | `/ai/agents/{id}` | `ai_agents:delete` | 删除记录并清理 `{workspace}/agents/agent-{id}/` |
| POST | `/ai/agents/{id}/runs` | `ai_agents:execute` | 手动触发运行，202；未启用或工作区非 ready 返回 400 |
| POST | `/ai/agents/{id}/api-runs` | `ai_agents:execute` | API 触发运行，202；JWT 或 PAT，PAT 需 scope `agents:run` |

**关键字段**：创建/更新 `{ name, description, enabled, cli_key, system_prompt, skill_ids, repo_bindings, output_dir, stream_output, timeout_sec }`；`repo_bindings` 为 `{ repository_id*, branch }[]`（同 Agent 内 `repository_id` 唯一，`branch` 缺省 `main`，不校验远程分支存在）；`skill_ids` 解压到工作区 `.agents/skills/{name}/`；`output_dir` 为相对产出目录名（默认 `output`，跨 Run 固定复用、不清空）；`stream_output` 默认 false（关闭时部分 CLI 仅输出最终摘要）。

**工作区**：响应含 `workspace_status`（`pending|ready|failed`）与 `workspace_error`；仅 `ready` 可创建 Run，失败不回滚删除 Agent。运行环境注入 `BEDROCK_AGENT_WORKDIR`（持久根工作区，跨 Run 复用）与 `BEDROCK_AGENT_OUTPUT`（固定产出目录）；Run 无专属目录、无文件制品归档与下载。

#### Agent 触发器

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/ai/agents/{id}/triggers` | `ai_agents:view` | 列出触发器 |
| POST | `/ai/agents/{id}/triggers` | `ai_agents:update` | 创建，201 |
| PUT | `/ai/agents/{id}/triggers/{tid}` | `ai_agents:update` | 更新 |
| DELETE | `/ai/agents/{id}/triggers/{tid}` | `ai_agents:update` | 删除 |

**关键字段**：`type*`（`manual|api|cron|build_event`）、`enabled`、`cron_expression`、`cron_timezone`（IANA 时区；cron 不重叠、不补跑错过的任务）、`build_job_id`、`build_event`（`artifact_ready|distribution_finished`）。

#### Runs

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/ai/runs` | `ai_runs:view` | 分页列出，可加 `agent_id`、`status` 过滤 |
| GET | `/ai/runs/{id}` | `ai_runs:view` | 详情；无制品字段 / 下载端点 |
| POST | `/ai/runs/{id}/cancel` | `ai_agents:execute` | 取消运行 |

**关键字段**：`id`、`agent_id`、`trigger_type`、`status`、`work_dir`（同 Agent 各 Run 复用同一路径）、`build_run_id`、`project_id`、`doc_node_id`、`error_message`、`output_text`、`duration_ms`（未结束为 0）、`started_at`/`finished_at`。

#### Skills

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/skills` | `ai_skills:view` | 分页列出；公开 Skill 凭权限可见，私有仅创建者可见 |
| POST | `/skills` | `ai_skills:create` | 创建，201；校验失败 422 |
| GET | `/skills/{id}` | `ai_skills:view` | 详情 |
| PUT | `/skills/{id}` | `ai_skills:update` | 覆盖更新 |
| DELETE | `/skills/{id}` | `ai_skills:delete` | 删除 |
| GET | `/skills/{id}/package` | `ai_skills:download` | 下载技能包（二进制）；或 PAT scope `skills:read` |

**关键字段**：创建/更新为 multipart 表单 `{ name, description, visibility, file* }`（ZIP 上传）；包内必须含 `SKILL.md`，服务端防 Zip Slip / zip bomb，默认上限 50MB；ZIP 内含 SKILL.md 的包装目录与 `__MACOSX` 会被剥离。

---

## 7. 项目（Project）

#### 项目

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/projects` | `project_projects:view` | 分页列出 |
| POST | `/projects` | `project_projects:create` | 创建项目，创建者自动成为 Owner |
| GET | `/projects/{id}` | `project_projects:view` | 项目详情 |
| PUT | `/projects/{id}` | `project_projects:update` | 更新项目 |
| DELETE | `/projects/{id}` | `project_projects:delete` | 解散项目 |
| POST | `/projects/{id}/archive` | `project_projects:update` | 归档项目 |

**关键字段**：创建 `{ name*, slug*, description, repository_id, tags }`；更新额外支持 `status`（`active|archived`）与 `clear_repository`（解绑仓库）。列表参数 `keyword`、`status`、`page`/`page_size`。详情 `data` 含 `my_role`（`owner|admin|member|readonly`）与 `permissions`（update/archive/delete/manage_members/transfer_owner 能力布尔位）。

#### 项目成员

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/projects/{id}/members` | `project_projects:view` | 列出成员 |
| POST | `/projects/{id}/members` | `project_projects:update` | 添加非 Owner 成员 |
| PUT | `/projects/{id}/members/{userID}` | `project_projects:update` | 修改非 Owner 成员角色 |
| DELETE | `/projects/{id}/members/{userID}` | `project_projects:update` | 移除非 Owner 成员（操作 Owner 返回 409） |
| POST | `/projects/{id}/members/transfer-owner` | `project_projects:update` | 转让所有者 `{ user_id* }` |

**关键字段**：添加/改角色 `{ user_id*, role* }`，`role` ∈ `admin|member|readonly`（Owner 角色不可通过成员接口设置）。

#### 需求（含评论与附件）

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/projects/meta/requirement-statuses` | `project_requirements:view` | 需求状态选项（`data.items[]`） |
| GET | `/projects/{id}/requirements` | `project_requirements:view` | 分页列出需求 |
| POST | `/projects/{id}/requirements` | `project_requirements:create` | 创建需求 |
| GET | `/projects/{id}/requirements/{rid}` | `project_requirements:view` | 获取需求 |
| PUT | `/projects/{id}/requirements/{rid}` | `project_requirements:update` | 更新需求 |
| DELETE | `/projects/{id}/requirements/{rid}` | `project_requirements:delete` | 删除需求 |
| GET | `/projects/{id}/requirements/{rid}/comments` | `project_requirements:view` | 列出评论 |
| POST | `/projects/{id}/requirements/{rid}/comments` | `project_requirements:create` | 添加评论 `{ content* }` |
| PUT | `/projects/{id}/requirements/{rid}/comments/{cid}` | `project_requirements:update` | 编辑评论 `{ content* }` |
| DELETE | `/projects/{id}/requirements/{rid}/comments/{cid}` | `project_requirements:delete` | 删除评论 |
| GET | `/projects/{id}/requirements/{rid}/attachments` | `project_requirements:view` | 列出附件 |
| POST | `/projects/{id}/requirements/{rid}/attachments` | `project_requirements:update` | 上传附件（multipart `{ file* }`，默认 20MB，超限 413） |
| DELETE | `/projects/{id}/requirements/{rid}/attachments/{aid}` | `project_requirements:update` | 删除附件 |
| GET | `/projects/{id}/requirements/{rid}/attachments/{aid}/download` | `project_requirements:view` | 下载附件（二进制流，非 JSON 信封） |

**关键字段**：需求 `{ title*, description, status, priority, assignee_id, repository_id, tags }`；`priority` ∈ `low|normal|high|urgent`；`status` 取值须为 `requirement_status` 字典中 enabled 的项。列表支持 `keyword`、`status`、`priority`、`assignee_id`、`sort`。

#### 项目文档

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/projects/{id}/docs` | `project_docs:view` | 文档树（含已发布与草稿内容） |
| POST | `/projects/{id}/docs` | `project_docs:create` | 创建目录或文档节点 |
| POST | `/projects/{id}/docs/upload` | `project_docs:create` | 上传单个 Markdown 为草稿（multipart `{ parent_id, file* }`） |
| POST | `/projects/{id}/docs/import-zip` | `project_docs:create` | 导入 Markdown zip 为草稿（默认 100MB，含 Zip Slip/条目数/压缩比防护） |
| POST | `/projects/{id}/docs/push` | `project_docs:create` 或 PAT `docs:write` | 按路径 upsert 草稿（外部 API） |
| POST | `/projects/{id}/docs/publish-path` | `project_docs:update` 或 PAT `docs:publish` | 按路径发布草稿（外部 API） |
| POST | `/projects/{id}/docs/generate` | `project_docs:execute` | AI 生成文档，异步 202 |
| GET | `/projects/{id}/docs/{nid}` | `project_docs:view` | 获取节点 |
| PUT | `/projects/{id}/docs/{nid}` | `project_docs:update` | 重命名节点或写入草稿 |
| DELETE | `/projects/{id}/docs/{nid}` | `project_docs:delete` | 删除节点及其全部子节点（级联，不可恢复） |
| POST | `/projects/{id}/docs/{nid}/move` | `project_docs:update` | 移动节点 `{ parent_id, sort_order }` |
| POST | `/projects/{id}/docs/{nid}/publish` | `project_docs:update` | 发布草稿 |
| GET | `/projects/{id}/docs/{nid}/diff` | `project_docs:view` | 比较草稿与已发布版本 |

**关键字段与行为**：

- 创建节点 `{ parent_id, kind*, name*, sort_order, repository_id, draft_content }`，`kind` ∈ `dir|doc`；更新 `{ name, repository_id, draft_content }`。
- 草稿/发布状态机：内容先写入 `draft_content`，发布后才生成对外可见版本；`publish` 需带 `expected_version*`（当前 `content_version`）做乐观并发控制，冲突返回 409。
- `push`/`publish-path`：`api_dir` 为空表示根目录，`/` 分隔，拒绝 `..`、绝对路径与空段；目录不存在自动创建；`api_doc_name` 缺 `.md` 后缀时服务端补齐。push 只写草稿不自动发布（新建 201 / 更新 200）；publish-path 无草稿 400、路径不存在 404、版本冲突 409。两者均要求项目 ACL，可用 JWT 或对应 scope 的 PAT 鉴权。
- `generate` 请求 `{ agent_id*, node_id }`，返回 202 创建异步 AgentRun；结果仅写入 `draft_content`（可选 `draft_source_run_id`），不自动发布。
- `diff` 返回 `{ node_id, content_version, has_draft, published_lines, draft_lines, added_lines, removed_lines }`。

---

## 8. 系统管理（System）

#### 用户

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/system/users` | `system_users:view` | 列出用户（分页） |
| POST | `/system/users` | `system_users:create` | 创建用户，201 |
| GET | `/system/users/{id}` | `system_users:view` | 获取用户 |
| PUT | `/system/users/{id}` | `system_users:update` | 更新用户 |
| DELETE | `/system/users/{id}` | `system_users:delete` | 删除用户 |

**关键字段**：创建 `username*`、`password*`，可选 `display_name`、`email`、`is_active`、`role_ids[]`；更新字段相同、全部可选。`role_ids` 不可包含内置 `super_admin` 角色；超管用户的内置角色绑定由服务端维持。响应 `User`：`id`、`username`、`display_name`、`email`、`is_active`、`is_super_admin`、`role_ids`；`password` 任何响应均不回显。

#### 角色

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/system/roles` | `system_roles:view` | 列出角色（分页） |
| POST | `/system/roles` | `system_roles:create` | 创建角色，201 |
| GET | `/system/roles/{id}` | `system_roles:view` | 获取角色 |
| PUT | `/system/roles/{id}` | `system_roles:update` | 更新角色（仅 name/description） |
| DELETE | `/system/roles/{id}` | `system_roles:delete` | 删除角色（内置角色不可删） |
| PUT | `/system/roles/{id}/permissions` | `system_roles:update` | 整体替换角色权限码 |
| GET | `/system/roles/permission-catalog` | `system_roles:update` | 绑权目录（三层：分组→菜单→功能） |

**关键字段**：创建 `name*`、`code*`，`permissions` 为功能 `full_code[]`；替换权限 `permissions*`（拒绝内置角色，且拒绝写入不存在或 `super_admin_only` 的功能）。绑权取功能 `full_code`（格式 `{menu.code}:{code}`，如 `system_users:view`）。

#### 菜单分组

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/system/menu-groups` | `system_resources:view` | 列出菜单分组（不分页，`data.items`） |
| POST | `/system/menu-groups` | `system_resources:create` | 创建，201 |
| GET | `/system/menu-groups/{id}` | `system_resources:view` | 获取 |
| PUT | `/system/menu-groups/{id}` | `system_resources:update` | 更新 |
| DELETE | `/system/menu-groups/{id}` | `system_resources:delete` | 删除；分组下仍有菜单时拒绝（400） |

**关键字段**：`name*`、`code*`（全局唯一、不含 `.`），可选 `route_prefix`、`sort_key`、`enabled`。

#### RBAC 资源

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/system/rbac/resources` | `system_resources:view` | 资源树（菜单→功能，不分页） |
| POST | `/system/rbac/resources` | `system_resources:create` | 创建资源，201 |
| GET | `/system/rbac/resources/{id}` | `system_resources:view` | 获取资源 |
| PUT | `/system/rbac/resources/{id}` | `system_resources:update` | 更新资源 |
| DELETE | `/system/rbac/resources/{id}` | `system_resources:delete` | 删除资源 |
| PUT | `/system/rbac/resources/{id}/icon` | `system_resources:update` | 更新资源图标（仅菜单） |

**关键字段**：创建 `code*`（不含 `.`）、`type*`（`menu|action|card`）；菜单必须带 `group_id` 且 `parent_id` 为空，功能必须挂菜单 `parent_id`。`full_code` 规则：菜单 = `code`，功能 = `{menu.code}:{code}`。`super_admin_only` 仅超管可设/改。改菜单 `code` 会级联重算子功能 `full_code` 并清理失效的 `role_permissions`。列表查询参数 `keyword`、`type`、`enabled`、`group_id`（筛选时保留匹配节点祖先以维持树结构）。图标请求 `icon_base64*`、`icon_mime`。

#### 字典

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/system/dictionaries` | `system_dictionaries:view` | 列出字典（分页） |
| POST | `/system/dictionaries` | `system_dictionaries:create` | 创建，201 |
| GET | `/system/dictionaries/{id}` | `system_dictionaries:view` | 获取 |
| PUT | `/system/dictionaries/{id}` | `system_dictionaries:update` | 更新 |
| DELETE | `/system/dictionaries/{id}` | `system_dictionaries:delete` | 删除 |

**关键字段**：`name*`、`code*`；`items` 为 `{ label, value, sort_order, enabled }[]`。

#### 操作日志

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/system/operation-logs` | `system_operation_logs:view` | 列出操作日志（分页） |

**关键字段**：查询参数 `user_id`、`action`、`resource_type`、`from`/`to`（date）。响应项：`username`、`action`、`resource_type`、`resource_id`、`details`、`ip_address`、`created_at`。

#### 通知

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/system/notifications` | 登录即可 | 列出当前用户通知（分页） |
| PUT | `/system/notifications/read-all` | 登录即可 | 全部标为已读 |
| PUT | `/system/notifications/{id}/read` | 登录即可 | 单条标为已读 |

**关键字段**：`type`（如 `build_run_success`、`build_run_failed`、`agent_run_success`）、`title`、`message`、`is_read`、`build_run_id`/`agent_run_id`、`created_at`。

---

## 9. 仪表盘与运维（Dashboard & Ops）

#### 仪表盘

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/dashboard/layout` | `dashboard:view` | 获取当前用户仪表盘布局 |
| PUT | `/dashboard/layout` | `dashboard:view` | 保存当前用户仪表盘布局 |
| GET | `/dashboard/build-summary` | `cicd_build_runs:view` | 构建摘要卡片数据 |
| GET | `/dashboard/agent-run-summary` | `ai_runs:view` | 智能体运行摘要卡片数据 |
| GET | `/dashboard/system-info` | `dashboard:system_info` | 系统信息卡片（version/os/arch/runtime/hostname/start_time） |
| GET | `/dashboard/system-status` | `dashboard:system_status` | 系统状态卡片（CPU/内存/磁盘/健康度） |

**关键字段**：PUT layout 请求 `cards*`：每张卡片 `id*`（`build_summary|agent_run_summary|system_info|system_status`）、`visible*`、`x*`/`y*`（0-based 行列起点）、`w*`（2–12）、`h*`（≥2）；`order` 由服务端按 `y*12+x` 归一。摘要卡片响应：`running`、`queued`、`success_rate`、`recent[]`。system-status 响应中 `disk_*` 是关键目录所在分区的宿主机磁盘占用，`directories[]` 是各目录自身大小。

#### 进程（仅超管）

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/ops/processes` | `ops_processes:view` | 列出主机进程（过滤后完整列表，不分页） |
| POST | `/ops/processes/{pid}/kill` | `ops_processes:execute` | 按 PID 结束进程（危险操作，需二次确认） |

**关键字段**：列表参数 `keyword`、`pid`、`port`、`sort`；每项含 `pid`、`name`、`cpu_percent`、`memory_bytes`、`username`、`num_threads`、`status`、`start_time`（Unix 毫秒）、`cmdline`、`ports[]`。

#### 开发环境（仅超管）

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/ops/dev-environments` | `ops_dev_environments:view` | 列出开发环境（含内置与自定义） |
| POST | `/ops/dev-environments` | `ops_dev_environments:create` | 创建自定义环境，201 |
| PUT | `/ops/dev-environments/{id}` | `ops_dev_environments:update` | 更新自定义环境 |
| DELETE | `/ops/dev-environments/{id}` | `ops_dev_environments:delete` | 删除自定义环境 |
| POST | `/ops/dev-environments/{id}/detect` | `ops_dev_environments:execute` | 运行检测脚本 |
| POST | `/ops/dev-environments/{id}/install` | `ops_dev_environments:execute` | 安装，异步 202 |
| POST | `/ops/dev-environments/{id}/upgrade` | `ops_dev_environments:execute` | 升级，异步 202 |
| POST | `/ops/dev-environments/{id}/uninstall` | `ops_dev_environments:execute` | 卸载，异步 202 |
| POST | `/ops/dev-environments/{id}/switch` | `ops_dev_environments:execute` | 切换版本，异步 202 |

**关键字段**：`name*`、`executable*`，可选 `description` 与 `detect/install/upgrade/uninstall/versions/switch_script`、`default_version`；install 等操作可传 `{ version }`。**任意命令执行风险**：自定义脚本以 Bedrock 进程 UID 直接在宿主机运行，无沙箱，务必仅允许可信管理员配置。

#### 开发环境安装源

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/ops/dev-environments/{id}/sources` | `ops_dev_environments:view` | 列出安装源 |
| POST | `/ops/dev-environments/{id}/sources` | `ops_dev_environments:create` | 添加，201 |
| PUT | `/ops/dev-environments/{id}/sources/{sid}` | `ops_dev_environments:update` | 更新 |
| DELETE | `/ops/dev-environments/{id}/sources/{sid}` | `ops_dev_environments:delete` | 删除 |
| POST | `/ops/dev-environments/{id}/sources/{sid}/ping` | `ops_dev_environments:execute` | 探测可用性（`ok`、`detail`） |

**关键字段**：`name*`、`base_url*`、`priority*`（int）、`enabled*`（bool）。

#### 开发环境任务

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/ops/dev-environments/{id}/jobs` | `ops_dev_environments:view` | 分页列出任务（参数 `status`） |
| GET | `/ops/dev-environments/{id}/jobs/{jid}` | `ops_dev_environments:view` | 任务详情 |
| GET | `/ops/dev-environments/{id}/jobs/{jid}/logs` | `ops_dev_environments:view` | 任务日志（text/plain） |
| POST | `/ops/dev-environments/{id}/jobs/{jid}/retry` | `ops_dev_environments:execute` | 重试任务，202 返回新任务 |

**关键字段**：任务对象 `operation`（`install|upgrade|uninstall|switch`）、`status`（`queued|running|success|failed|interrupted`）、`requested_version`、`source_id`、`command_snapshot`、`error_message`、`started_at`/`finished_at`。

---

## 10. WebSocket

统一挂载在 `{host}/ws`（不是 `/api/v1`），通过查询参数 `token` 携带 `access_token` 鉴权，并校验对应权限码。

| 路径 | 权限 | 说明 |
| --- | --- | --- |
| `/ws/build-runs/{id}/logs` | `cicd_build_runs:view` | 构建运行实时日志 |
| `/ws/ai/runs/{id}/logs` | `ai_runs:view` | Agent 运行实时日志 |
| `/ws/notifications` | 登录即可 | 当前用户通知推送 |

**构建日志协议**：连接后先按行回放已有日志，随后推送两类文本帧——日志行（追加到终端）与控制帧 `__REFRESH__`（表示 `status`/`stage`/`distribution_summary`/`deploy_attempts` 已变更，应重新 `GET /build-runs/{id}`，勿写入日志视图；该帧仅经 WS 广播，不落入日志文件）。

```text
ws://127.0.0.1:8080/ws/build-runs/123/logs?token=<access_token>
```

---

## 11. 完整契约索引

本页为面向使用的完整参考；字段级对象形状（DTO 表格）以仓库契约为准：

| 域 | 仓库文件 | 内容 |
| --- | --- | --- |
| 认证 | `api/auth.md` | login / refresh / logout / me |
| 资源 | `api/resource.md` | 仓库、服务器、凭证、AI CLI、PAT |
| CI/CD | `api/cicd.md` | 构建任务、构建运行、Webhook |
| AI | `api/ai.md` | Agents、触发器、Runs、Skills |
| 项目 | `api/project.md` | 项目、成员、需求、评论、附件、文档发布 |
| 系统 | `api/system.md` | 用户、角色、菜单分组、RBAC 资源、字典、操作日志、通知 |
| 运维 | `api/ops.md` | 仪表盘卡片、进程、开发环境 |
