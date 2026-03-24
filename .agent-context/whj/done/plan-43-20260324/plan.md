# Windows 仪表盘 CPU 占用修正

> 状态: 已执行

## 目标

修正 Windows 下仪表盘系统 CPU 百分比与任务管理器等指标明显不一致的问题：按 `GetSystemTimes` 文档，内核时间已包含空闲时间，当前实现将 `idle+kernel+user` 当作总时间会导致空闲时仍显示较高占用。

## 内容

1. 调整 `internal/service/dashboard_metrics_proc_windows.go`：用 `kernel+user` 作为总时间增量、`kernel-idle+user` 作为忙碌时间增量计算 CPU%。
2. 在可跨平台编译的代码中抽出纯公式函数（或等价实现），在 `internal/service/dashboard_metrics_test.go` 增加单测覆盖典型增量（含近似全空闲场景）。
3. 运行 `go test ./internal/service/...`（或 `go test ./...`）确认通过。

## 影响范围

- `internal/service/dashboard_metrics.go`
- `internal/service/dashboard_metrics_proc_windows.go`
- `internal/service/dashboard_metrics_test.go`

## 历史补丁
