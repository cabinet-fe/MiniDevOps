# 审计日志日期组件与备份导出修复

> 状态: 已执行

## 目标

1. 审计日志筛选中的开始/结束日期改用 shadcn Calendar（Popover）等 UI 组件，替代原生 `type="date"` 的 Input。
2. 修复系统设置「导出备份」下载的 tar.gz 无法打开的问题（避免响应被全局 gzip 中间件二次压缩）。

## 内容

1. 后端：`cmd/server/main.go` 中为 `gzip.Gzip` 增加 `WithExcludedPaths`，排除 `/api/v1/system/backup`，使备份流仅为 handler 内单层 gzip。
2. 前端：通过 shadcn 添加 `calendar` 组件（若缺依赖则安装）；在 `web/src/pages/audit-logs.tsx` 用 Popover + Calendar 选择日期，状态仍为 `YYYY-MM-DD` 字符串以兼容现有 API。
3. 运行 `go test ./...`、`cd web && bun run lint` 与 `bun run build` 验证。

## 影响范围

- `cmd/server/main.go`
- `web/package.json`、`web/bun.lock`（`date-fns`、`react-day-picker`）
- `web/src/components/ui/calendar.tsx`、`web/src/components/ui/button.tsx`
- `web/src/pages/audit-logs.tsx`

## 历史补丁
