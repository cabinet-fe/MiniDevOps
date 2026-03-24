# 备份导出精简（排除审计与构建）

> 状态: 已执行

## 目标

系统「备份下载」导出的 SQLite 不再包含审计日志、构建记录及与构建强相关的站内通知，以减小备份体积并避免导出运行期流水数据；配置 `config.yaml` 行为不变。

## 内容

1. 在 `internal/pkg` 增加从现有库路径生成「精简库」临时文件的逻辑：对源库只读连接后使用 `VACUUM INTO` 得到一致快照，再在副本中清空 `audit_logs`、`builds`、`notifications` 表并 `VACUUM`，返回临时文件路径与清理回调。
2. 修改 `SystemHandler.Backup`：在存在数据库文件时写入精简副本到 tar；精简失败则返回错误（见补丁 patch-1，已取消整库回退）。
3. 运行 `go test ./...` 确认通过；如有必要为精简逻辑补充单测（可用内存库或临时文件）。

## 影响范围

- `internal/pkg/backup_sqlite.go`（新建：精简备份 SQLite 快照）
- `internal/pkg/backup_sqlite_test.go`（新建：单测）
- `internal/handler/system_handler.go`（`Backup` 使用精简库；失败时返回错误，见 patch-1）
- `internal/handler/build_handler.go`（见 patch-1）
- `internal/engine/scheduler.go`（见 patch-1）
- `internal/engine/scheduler_test.go`（见 patch-1）
- `cmd/server/main.go`（见 patch-1）
- `web/src/pages/builds/detail.tsx`（见 patch-1）
- `web/src/pages/settings.tsx`（备份卡片说明与行为一致）

## 历史补丁

- patch-1: 移除备份整库回退与构建回滚，修复队列与日志 404
