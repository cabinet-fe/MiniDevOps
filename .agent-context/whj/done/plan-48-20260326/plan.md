# 构建访问控制与代码健壮性修补

> 状态: 已执行

## 目标

与项目列表规则对齐：`dev` 仅能访问其创建项目下的构建（HTTP、WebSocket、仪表盘统计）；`admin`/`ops` 保持全量。收紧 WebSocket `CheckOrigin` 与 HTTP CORS 一致。修正 JSON 绑定、解密失败、字典项更新死代码、用户列表 Count、审计写入等可观测性问题。

## 内容

1. 新增 `middleware` 项目访问判断（`dev` 对比 `project.created_by`，`admin`/`ops` 放行），WebSocket 使用 JWT 中的 `role` 做同样判断。
2. `BuildHandler` 注入 `ProjectRepository`：对单构建相关接口及 `TriggerBuild`、`ListByProject` 做项目级校验；`ListAll` 与仪表盘相关接口经 `BuildService` 按角色过滤。
3. `ProjectRepository.ListIDsByCreatedBy`、`Count(createdBy *uint)`；`BuildRepository` 对列表与统计增加可选 `project_id IN (...)` 过滤（`nil` 表示不限）。
4. `BuildService` 内解析 `dev` 的可访问项目 ID 列表，贯通 `ListAll`、`GetDashboardStats`、`GetActiveBuildsList`、`GetRecentBuilds`、`GetBuildTrend`。
5. `WSHandler`：`CheckOrigin` 与 `CORSConfig` 一致；升级前校验构建所属项目访问权。
6. `decryptServerSecrets` 解密失败返回错误并在分发路径失败；`build_handler` 校验 `ShouldBindJSON`；`dict_handler` 去掉无意义的 `_ = item`；`UserRepository.List` 检查 `Count` 错误；审计 `Create` 失败时 `log.Printf` 记录。

## 影响范围

- `internal/middleware/project_access.go`（新建）
- `internal/middleware/cors.go`
- `internal/middleware/audit.go`
- `internal/repository/project_repo.go`
- `internal/repository/build_repo.go`
- `internal/repository/user_repo.go`
- `internal/service/build_service.go`
- `internal/handler/build_handler.go`
- `internal/handler/ws_handler.go`
- `internal/handler/dict_handler.go`
- `internal/engine/pipeline.go`
- `internal/engine/pipeline_distribute.go`
- `cmd/server/main.go`

## 历史补丁
