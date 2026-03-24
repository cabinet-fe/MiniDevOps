# 页面优化与进程清理修补

> 状态: 已执行

## 目标

- 在 dashboard 的系统资源卡片中增加本应用内存占用，并将所有资源使用信息展示为环形进度。
- 确保 build 构建任务通过 `os/exec` 执行后进程和内存被正确释放（包括子进程不遗留）。

## 内容

1. **调整系统状态信息接口**：在后端的系统状态或者相关 API 中，添加“本应用内存占用”数据（基于 `runtime.MemStats` 或 `gopsutil`）。
2. **重构前端仪表盘资源展示**：修改 `web/src/pages/dashboard.tsx`，在系统资源面板引入环形进度组件，展示 CPU、系统内存占用以及本应用内存占用。
3. **完善子进程资源清理逻辑**：在后端执行构建任务的代码处（如 `internal/engine` 相关文件使用 `exec.Command` 的地方），使用进程组（Process Group）或其他方式确保父进程和所有子进程均被彻底杀死，释放内存。
4. **编译与验证**：运行前端构建与检查并测试后端用例。

## 影响范围

- `internal/service/build_service.go`
- `internal/service/dashboard_metrics.go`
- `web/src/pages/dashboard.tsx`
- `internal/engine/pipeline.go`

## 历史补丁
