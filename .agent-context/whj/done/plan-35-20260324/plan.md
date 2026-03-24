# 仪表盘资源展示调整

> 状态: 已执行

## 目标

- 移除易误导的「本应用使用内存」指标（基于 `runtime.MemStats.Alloc`，长期运行下会缓慢上升属 Go 运行时常态，非可靠「泄漏」检测）。
- 仪表盘系统内存与磁盘用量文案同时展示已用与总容量。

## 内容

1. 后端：从 `DashboardSystemResources` 与 `collectDashboardSystemResources` 中移除 `app_memory_used_bytes` 及相关 `runtime.ReadMemStats` 调用。
2. 前端：`dashboard.tsx` 删除本应用内存环形图项；系统内存与磁盘主文案格式为「已用 / 总计」或等价含总容量的展示。
3. 运行 `go test ./internal/service/...` 与 `cd web && bun run lint`、`bun run build` 验证。

## 影响范围

- `internal/service/dashboard_metrics.go`
- `internal/service/build_service.go`
- `web/src/pages/dashboard.tsx`

## 历史补丁
