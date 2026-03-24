# 移除备份整库回退与构建回滚，修复队列与日志 404

## 补丁内容

- **备份**：`Backup` 仅在精简快照成功时打包数据库；去掉「精简失败则打包整库」的回退，避免与「不含构建记录」的约定不一致，失败时返回明确错误（设置响应头与流之前校验）。
- **构建回滚**：删除 `POST /builds/:id/rollback` 及详情页「回滚」按钮；该操作仅 `Submit` 旧构建 ID，易与预期不符且易引入异常状态。
- **构建一直等待**：`max_concurrent` 为 0 时调度器在 `semaphore` 上死锁，首个任务永不执行；`NewScheduler` 将并发下限钳为 1（例如恢复的配置里误写 0）。
- **日志 404**：排队/运行中阶段日志文件尚未创建时，`GET /builds/:id/log` 返回 200 空正文，避免浏览器控制台对 404 的噪声；终态仍按原逻辑返回 404。
- **设置页**：备份说明与上述行为一致。

## 影响范围

- 新增文件: `internal/engine/scheduler_test.go`
- 修改文件: `internal/handler/system_handler.go`
- 修改文件: `internal/handler/build_handler.go`
- 修改文件: `internal/engine/scheduler.go`
- 修改文件: `cmd/server/main.go`
- 修改文件: `web/src/pages/builds/detail.tsx`
- 修改文件: `web/src/pages/settings.tsx`
