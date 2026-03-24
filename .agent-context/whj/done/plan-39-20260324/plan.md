# 项目删除与构建状态修复

> 状态: 已执行

## 目标

1. 在前端提供项目删除能力（后端已有 `DELETE /api/v1/projects/:id`），删除时确保与定时任务、数据一致性一致。
2. 删除项目或单个环境时，清除该范围下所有环境变量、变量组关联、构建记录及磁盘上的日志/产物/工作区（或环境级目录），满足「删环境即删相关构建信息」。
3. 服务进程异常退出（如构建中宕机）后重启，将仍处于 `pending`/`cloning`/`building`/`deploying` 的构建标记为失败，消除仪表盘「运行中」、最近构建、项目详情环境列表长期显示「构建中」的问题。

## 内容

1. **启动时回收未结束的构建**
   - 在 `cmd/server/main.go` 中，于数据库初始化之后、启动 `Scheduler` 之前，调用业务层方法（可放在 `BuildService` 或 `BuildRepository`）：将所有状态属于 `pending`、`cloning`、`building`、`deploying` 的 `Build` 批量更新为 `failed`。
   - 设置 `error_message` 为明确中文说明（如服务异常中断）；设置 `finished_at`；若有 `started_at` 则计算 `duration_ms`；将 `current_stage` 与终态一致或保留原阶段均可，但需与现有前端展示逻辑一致。
   - 打日志记录回收数量。

2. **项目删除：定时任务与权限**
   - 在 `ProjectHandler.Delete` 中，删除数据库记录之前先 `GetByID` 取得项目及其 `environments`，对每个环境 ID 调用 `cronNotifier.Remove(envID)`，避免已删环境仍被 cron 触发。
   - 将 `DELETE /projects/:id` 路由加上 `middleware.RequireRole("ops", "admin")`，与服务器删除等运维操作对齐（若与产品要求冲突可再 patch 调整）。

3. **删除单个环境：构建数据与磁盘**
   - 扩展 `BuildRepository`：`DeleteByEnvironmentID`；删除前可查询该环境下所有构建用于清理文件。
   - 在 `ProjectService.DeleteEnvironment` 中：在删除环境行之前，删除该环境全部构建；对每条构建尽力删除 `log_path`、`artifact_path` 对应文件；删除该环境工作区目录 `workspace/project-{pid}/env-{eid}` 与缓存目录 `cache/project-{pid}/env-{eid}`（路径与 `internal/engine/pipeline.go` 约定一致，需读 `config`）。
   - 在 `ProjectHandler.DeleteEnvironment` 成功删除后调用 `cronNotifier.Remove(envID)`。

4. **前端：项目删除入口**
   - 在项目列表页与/或项目详情页增加删除按钮（建议详情页主操作区 + 列表项危险操作），使用确认对话框（文案说明将删除所有环境、构建记录与相关文件）。
   - 使用现有 `api.delete("/projects/:id")`；成功后跳转列表或提示刷新；`ops`/`admin` 可见或按路由权限隐藏（与后端角色一致）。

5. **验证**
   - 运行 `go test ./...`。
   - 运行 `cd web && bun run lint` 与 `bun run build`。

## 影响范围

- `internal/repository/build_repo.go`：环境维度构建查询/删除、启动时 `MarkInterruptedBuilds`。
- `cmd/server/main.go`：启动时回收未结束构建；`DELETE /projects/:id` 限制 `ops`/`admin`。
- `internal/handler/project_handler.go`：删除项目前移除各环境 cron；删除环境后 `cronNotifier.Remove`。
- `internal/service/project_service.go`：`DeleteEnvironment` 删除构建记录并清理日志、产物、工作区与缓存目录。
- `web/src/pages/projects/detail.tsx`：运维/管理员删除项目入口与确认框。
- `web/src/pages/projects/list.tsx`：列表卡片/表格删除入口与确认框。

## 历史补丁
