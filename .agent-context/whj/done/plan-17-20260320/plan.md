# 构建引擎与应用健壮性修复

> 状态: 已执行

## 目标

修复构建引擎中的并发安全问题、资源泄漏、panic 恢复缺失等健壮性问题，确保后端进程不会因构建任务异常而崩溃，同时改善优雅关闭流程和 Git 操作的可靠性。

## 内容

### 步骤 1：构建引擎 panic 恢复

为所有后台 goroutine 添加 panic 恢复机制，防止单个任务异常导致整个进程崩溃。

- **Scheduler worker goroutine**（`internal/engine/scheduler.go`）：在 `run()` 中启动的每个 worker goroutine 入口添加 `defer recover()`，panic 时记录日志并标记构建为 failed。
- **Pipeline.Execute**（`internal/engine/pipeline.go`）：在 `Execute` 方法开头添加 `defer recover()`，捕获到 panic 时调用 `failBuild` 标记失败。
- **Cron 回调**（`internal/engine/cron.go`）：在 `addEntry` 的回调函数中添加 `defer recover()`。
- **Hub.run**（`internal/ws/hub.go`）：在 `run()` 方法中添加 `defer recover()`，并在 recover 后重启 run loop。

### 步骤 2：日志并发写竞争修复

修复 Pipeline 中 stdout/stderr 双 goroutine 并发写日志文件的数据竞争。

- 为 `writeLine` 函数（`internal/engine/pipeline.go`）添加 `sync.Mutex` 保护，确保对 `logFile` 的写入和 WebSocket 广播串行化。
- 使用 `sync.WaitGroup` 等待 stdout 扫描 goroutine 完成后再返回 `Execute`，避免 `logFile` 在 goroutine 仍在写入时被关闭。

### 步骤 3：构建取消功能修复

当前 `BuildHandler.Cancel` 只更新数据库状态但不取消正在运行的构建进程。

- 在 `BuildHandler.Cancel`（`internal/handler/build_handler.go`）中增加对 `scheduler.Cancel(buildID)` 的调用，使 context 取消传播到正在执行的 Pipeline。
- 需要让 handler 层能够访问 scheduler 实例（通过依赖注入或在 BuildHandler 中持有引用）。

### 步骤 4：资源泄漏修复

- **SSH 临时密钥文件**（`internal/deployer/ssh.go`）：在 `buildSSHOptionsSlice` 创建的临时密钥文件，需要在使用完毕后清理。修改 `runAndLog` 或调用方，在部署命令执行完成后 `os.Remove` 临时文件。可考虑返回一个 cleanup 函数由调用方在 defer 中执行。

### 步骤 5：Shutdown 流程优化

调整优雅关闭的顺序和完整性，防止关闭期间 panic。

- **调整关闭顺序**（`cmd/server/main.go`）：
  1. `srv.Shutdown(ctx)` — 先停止接受新 HTTP 请求
  2. `cronScheduler.Stop()` — 停止定时任务
  3. `hub.Shutdown()` — 关闭 WebSocket 连接（需新增）
  4. `scheduler.Shutdown()` — 等待所有构建任务完成
  5. `sqlDB.Close()` — 关闭数据库连接
  6. `logger.Sync()` — 刷新日志
- **Scheduler Submit 保护**（`internal/engine/scheduler.go`）：添加 `closed` 标志位（atomic 或 sync.Once），`Submit` 在已关闭时返回 error 而非向已关闭 channel 发送。
- **Hub Shutdown**（`internal/ws/hub.go`）：添加 `Shutdown()` 方法，关闭所有连接并退出 run loop。
- **数据库连接关闭**（`cmd/server/main.go`）：在 shutdown 中获取 `*sql.DB` 并调用 `Close()`。

### 步骤 6：Git 操作健壮性增强

改善构建时 Git 操作的可靠性，防止因残留锁文件或依赖冲突导致拉取失败。

- **Git lock 文件清理**（`internal/engine/git.go`）：在 `GitCloneOrPull` 执行 fetch 前，检测并删除 `.git/index.lock` 等残留锁文件，避免上次构建异常中断后 git 操作被锁住。
- **git clean 策略优化**：当前 `git clean -fd` 排除了 `node_modules`、`vendor` 等依赖目录（使用 `-e` 标志），这是合理的缓存策略。确认此行为在日志中有提示信息，帮助排查因缓存导致的偶发问题。
- **git 操作超时保护**：确保 `runGit` 使用的 context 有合理超时，防止 fetch/clone 因网络问题永久阻塞（当前依赖上层 context，需确认上层 context 有 timeout 或 cancel）。

## 影响范围

- `internal/engine/scheduler.go` — panic 恢复、Submit 返回 error、closed 标志位、Shutdown 设置 closed
- `internal/engine/pipeline.go` — Execute 方法 panic 恢复、writeLine 加 Mutex、stdout scanner WaitGroup
- `internal/engine/cron.go` — cron 回调 panic 恢复
- `internal/engine/git.go` — 新增 cleanGitLockFiles 函数、fetch 前清理残留锁文件、日志优化
- `internal/ws/hub.go` — run() panic 恢复并自动重启、新增 quit channel 和 Shutdown() 方法
- `internal/handler/build_handler.go` — BuildScheduler 接口增加 Cancel 方法和 Submit 返回 error、Cancel handler 调用 scheduler.Cancel
- `internal/deployer/ssh.go` — buildSSHOptionsSlice/buildSSHOptions 返回 cleanup 函数、临时密钥文件 chmod 0600
- `internal/deployer/scp.go` — Deploy 方法增加 defer cleanup()
- `internal/deployer/rsync.go` — Deploy 方法增加 defer cleanup()
- `cmd/server/main.go` — 优雅关闭顺序调整为 HTTP→Cron→Hub→Scheduler→DB，新增 hub.Shutdown() 和 sqlDB.Close()

## 历史补丁
