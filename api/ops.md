# 运维与仪表盘

仪表盘卡片，以及进程 / 开发环境相关接口。

通用约定（信封、分页、认证）见 [.agents/api.md](../.agents/api.md)。
业务语义与权限模型见 [docs/DESIGN.md](../docs/DESIGN.md)。

## 仪表盘

### GET /dashboard/layout — 获取当前用户仪表盘布局

权限：`dashboard:view`
响应 200：data = DashboardLayout
错误：403

### PUT /dashboard/layout — 保存当前用户仪表盘布局

权限：`dashboard:view`
请求：{ cards* }
响应 200：data = DashboardLayout
错误：403

### GET /dashboard/build-summary — 构建摘要卡片数据

权限：`cicd.build_runs:view`
响应 200：data = BuildSummary
错误：403

### GET /dashboard/system-info — 系统信息卡片数据

权限：`dashboard.system_info:view`
响应 200：data = SystemInfo
错误：403
说明：有权限的非超管也可见；不因此获得运维写权限。

### GET /dashboard/system-status — 系统状态卡片数据

权限：`dashboard.system_status:view`
响应 200：data = SystemStatus
错误：403
说明：`disk_*` 为关键数据目录所在分区的宿主机磁盘占用；`directories` 为各关键目录自身占用大小（非分区剩余空间）。

## 运维

### GET /ops/processes — 列出主机进程（仅超管）

权限：`ops.processes:view`
查询参数：keyword: string, pid: integer, port: integer, sort: string
响应 200：data = object
错误：403
说明：返回过滤后的完整进程列表，不分页。

### POST /ops/processes/{pid}/kill — 按 PID 结束进程

权限：`ops.processes:execute`
路径参数：pid*: integer
响应 200：Terminated
错误：403

### GET /ops/dev-environments — 列出开发环境

权限：`ops.dev_environments:view`
响应 200：data = object
错误：403

### POST /ops/dev-environments — 创建自定义开发环境

权限：`ops.dev_environments:create`
请求：{ name*, executable*, description, detect_script, install_script, upgrade_script, uninstall_script, versions_script, switch_script, default_version }
响应 201
错误：403
说明：自定义脚本以 Bedrock 进程 UID 运行，无沙箱。

### PUT /ops/dev-environments/{id} — 更新自定义开发环境

权限：`ops.dev_environments:update`
路径参数：id*: integer
请求：{ name*, executable*, description, detect_script, install_script, upgrade_script, uninstall_script, versions_script, switch_script, default_version }
响应 200：data = DevEnvironment
错误：403

### DELETE /ops/dev-environments/{id} — 删除自定义开发环境

权限：`ops.dev_environments:delete`
路径参数：id*: integer
响应 200
错误：403

### POST /ops/dev-environments/{id}/detect — 检测开发环境

权限：`ops.dev_environments:execute`
路径参数：id*: integer
响应 200：data = DevEnvironmentDetectResult
错误：403

### POST /ops/dev-environments/{id}/install — 安装开发环境（异步）

权限：`ops.dev_environments:execute`
路径参数：id*: integer
请求：{ version }
响应 202：data = DevEnvJob
错误：403

### POST /ops/dev-environments/{id}/upgrade — 升级开发环境（异步）

权限：`ops.dev_environments:execute`
路径参数：id*: integer
请求：{ version }
响应 202：data = DevEnvJob
错误：403

### POST /ops/dev-environments/{id}/uninstall — 卸载开发环境（异步）

权限：`ops.dev_environments:execute`
路径参数：id*: integer
请求：{ version }
响应 202：data = DevEnvJob
错误：403

### POST /ops/dev-environments/{id}/switch — 切换开发环境版本（异步）

权限：`ops.dev_environments:execute`
路径参数：id*: integer
请求：{ version }
响应 202：data = DevEnvJob
错误：403

### GET /ops/dev-environments/{id}/sources — 列出安装源

权限：`ops.dev_environments:view`
路径参数：id*: integer
响应 200：data = DevEnvInstallSourceList
错误：403

### POST /ops/dev-environments/{id}/sources — 添加安装源

权限：`ops.dev_environments:create`
路径参数：id*: integer
请求：{ name*, base_url*, priority*, enabled* }
响应 201：data = DevEnvInstallSource
错误：403

### PUT /ops/dev-environments/{id}/sources/{sourceId} — 更新安装源

权限：`ops.dev_environments:update`
路径参数：id*: integer, sourceId*: integer
请求：{ name*, base_url*, priority*, enabled* }
响应 200：data = DevEnvInstallSource
错误：403

### DELETE /ops/dev-environments/{id}/sources/{sourceId} — 删除安装源

权限：`ops.dev_environments:delete`
路径参数：id*: integer, sourceId*: integer
响应 200
错误：403

### POST /ops/dev-environments/{id}/sources/{sourceId}/ping — 探测安装源

权限：`ops.dev_environments:execute`
路径参数：id*: integer, sourceId*: integer
响应 200：data = DevEnvInstallSourcePingResult
错误：403

### GET /ops/dev-environments/{id}/jobs — 列出开发环境任务

权限：`ops.dev_environments:view`
路径参数：id*: integer
查询参数：page: integer, page_size: integer, status: string
响应 200：data = DevEnvJobPage
错误：403

### GET /ops/dev-environments/{id}/jobs/{jobId} — 获取开发环境任务

权限：`ops.dev_environments:view`
路径参数：id*: integer, jobId*: integer
响应 200：data = DevEnvJob
错误：403

### GET /ops/dev-environments/{id}/jobs/{jobId}/logs — 获取开发环境任务日志

权限：`ops.dev_environments:view`
路径参数：id*: integer, jobId*: integer
响应 200：data = text/plain
错误：403

### POST /ops/dev-environments/{id}/jobs/{jobId}/retry — 重试开发环境任务

权限：`ops.dev_environments:execute`
路径参数：id*: integer, jobId*: integer
响应 202：data = DevEnvJob
错误：403

## 对象形状

### BuildSummary

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `running` | `integer` |  |  |
| `queued` | `integer` |  |  |
| `success_rate` | `number` |  |  |
| `recent` | `DashboardRecentBuildRun[]` |  |  |

### DashboardCardLayout

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `'build_summary' \| 'system_info' \| 'system_status'` | 是 |  |
| `visible` | `boolean` | 是 |  |
| `order` | `integer` | 是 |  |

### DashboardLayout

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `cards` | `DashboardCardLayout[]` | 是 |  |

### DashboardRecentBuildRun

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `build_job_id` | `integer` |  |  |
| `build_number` | `integer` |  |  |
| `status` | `string` |  |  |
| `branch` | `string` |  |  |
| `created_at` | `string(date-time)` |  |  |

### DevEnvInstallSource

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `environment_id` | `integer` |  |  |
| `name` | `string` |  |  |
| `base_url` | `string` |  |  |
| `priority` | `integer` |  |  |
| `enabled` | `boolean` |  |  |
| `created_at` | `string(date-time)` |  |  |
| `updated_at` | `string(date-time)` |  |  |

### DevEnvInstallSourceInput

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `name` | `string` | 是 |  |
| `base_url` | `string` | 是 |  |
| `priority` | `integer` | 是 |  |
| `enabled` | `boolean` | 是 |  |

### DevEnvInstallSourceList

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `items` | `DevEnvInstallSource[]` |  |  |

### DevEnvInstallSourcePingResult

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `ok` | `boolean` |  |  |
| `detail` | `string` |  |  |

### DevEnvJob

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `environment_id` | `integer` |  |  |
| `operation` | `'install' \| 'upgrade' \| 'uninstall' \| 'switch'` |  |  |
| `requested_version` | `string` |  |  |
| `status` | `'queued' \| 'running' \| 'success' \| 'failed' \| 'interrupted'` |  |  |
| `source_id` | `integer` |  |  |
| `command_snapshot` | `string` |  |  |
| `error_message` | `string` |  |  |
| `created_by` | `integer` |  |  |
| `started_at` | `string(date-time)` |  |  |
| `finished_at` | `string(date-time)` |  |  |
| `created_at` | `string(date-time)` |  |  |
| `environment` | `DevEnvironment` |  |  |
| `source` | `DevEnvInstallSource` |  |  |

### DevEnvJobPage

组合：`Page` + `inline`

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `items` | `any[]` | 是 |  |
| `total` | `integer` | 是 |  |
| `page` | `integer` | 是 |  |
| `page_size` | `integer` | 是 |  |
| `total_pages` | `integer` | 是 |  |
| `items` | `DevEnvJob[]` |  |  |

### DevEnvironment

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` |  |  |
| `name` | `string` |  |  |
| `kind` | `'builtin' \| 'custom'` |  |  |
| `executable` | `string` |  |  |
| `description` | `string` |  |  |
| `detect_script` | `string` |  |  |
| `install_script` | `string` |  |  |
| `upgrade_script` | `string` |  |  |
| `uninstall_script` | `string` |  |  |
| `versions_script` | `string` |  |  |
| `switch_script` | `string` |  |  |
| `default_version` | `string` |  |  |
| `created_by` | `integer` |  |  |
| `created_at` | `string(date-time)` |  |  |
| `updated_at` | `string(date-time)` |  |  |
| `sources` | `DevEnvInstallSource[]` |  |  |

### DevEnvironmentDetectResult

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `detected` | `boolean` |  |  |
| `output` | `string` |  |  |

### DevEnvironmentInput

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `name` | `string` | 是 |  |
| `executable` | `string` | 是 |  |
| `description` | `string` |  |  |
| `detect_script` | `string` |  |  |
| `install_script` | `string` |  |  |
| `upgrade_script` | `string` |  |  |
| `uninstall_script` | `string` |  |  |
| `versions_script` | `string` |  |  |
| `switch_script` | `string` |  |  |
| `default_version` | `string` |  |  |

### DevEnvironmentJobInput

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `version` | `string` |  |  |

### DirectoryUsage

关键数据目录的占用大小（目录树合计），不是所在分区的剩余空间。

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `path` | `string` |  | 目录绝对/配置路径 |
| `used_bytes` | `integer` |  | 该目录内容占用字节数 |

### ProcessInfo

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `pid` | `integer` |  |  |
| `name` | `string` |  |  |
| `cpu_percent` | `number` |  |  |
| `memory_bytes` | `integer` |  |  |
| `username` | `string` |  |  |
| `num_threads` | `integer` |  |  |
| `status` | `string` |  | OS process status (e.g. R |
| `start_time` | `integer` |  | Process start time as Unix epoch milliseconds |
| `cmdline` | `string` |  |  |
| `ports` | `integer[]` |  |  |

### SystemInfo

Complete read-only system information; this does not grant operations write access.

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `version` | `string` |  |  |
| `os` | `string` |  |  |
| `arch` | `string` |  |  |
| `runtime` | `string` |  |  |
| `hostname` | `string` |  |  |
| `start_time` | `string(date-time)` |  |  |

### SystemStatus

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `cpu_usage_percent` | `number` |  |  |
| `memory_used_bytes` | `integer` |  |  |
| `memory_total_bytes` | `integer` |  |  |
| `memory_usage_percent` | `number` |  |  |
| `disk_used_bytes` | `integer` |  | 宿主机数据盘已用字节（关键目录所在分区） |
| `disk_total_bytes` | `integer` |  | 宿主机数据盘总容量 |
| `disk_free_bytes` | `integer` |  | 宿主机数据盘可用字节 |
| `disk_usage_percent` | `number` |  | 宿主机数据盘占用百分比 |
| `health` | `'ok' \| 'degraded'` |  |  |
| `directories` | `DirectoryUsage[]` |  | 关键目录各自占用大小 |
| `collected_at` | `string(date-time)` |  |  |
