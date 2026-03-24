# Windows 仪表盘 CPU/内存采集

> 状态: 已执行

## 目标

在 Windows 上运行 BuildFlow 服务端时，仪表盘「系统资源」能正确显示 CPU 使用率与系统内存占用（与 Linux 下通过 /proc 采集的行为一致）；磁盘采集保持现有实现。

## 内容

1. 将 `collectDashboardSystemResources` 中的 CPU、内存采集从硬编码 `/proc/stat`、`/proc/meminfo` 拆为按平台实现：`!windows` 保留现有逻辑；`windows` 使用 `kernel32.GetSystemTimes` 双采样计算 CPU 占比、`GlobalMemoryStatusEx` 读取物理内存总量与可用量并推导已用与百分比。
2. 共享 `roundSingleDecimal` 与采样间隔常量，Windows 侧复用与 Linux 相近的节流逻辑以降低接口延迟。
3. 运行 `go test ./internal/service/...` 与 `GOOS=windows GOARCH=amd64 go build -o /dev/null ./...` 验证。

## 影响范围

- 修改文件: `internal/service/dashboard_metrics.go`
- 新增文件: `internal/service/dashboard_metrics_proc_linux.go`
- 新增文件: `internal/service/dashboard_metrics_proc_windows.go`

## 历史补丁
