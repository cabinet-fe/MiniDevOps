# 代码库安全与死代码清理

> 状态: 已执行

## 目标

扫描 BuildFlow 代码库，识别未使用代码与潜在安全问题，在可验证范围内移除死代码、修复或缓解漏洞，并保持测试与 Lint 通过。

## 内容

1. 运行 Go 静态检查：`go vet ./...`、如环境存在则运行 `staticcheck ./...`（或 `go run honnef.co/go/tools/cmd/staticcheck@latest`）关注 `U1000` 等未使用符号。
2. 运行前端检查：`cd web && bun run lint` 与 `bun run build`，处理可安全修复的告警。
3. 人工审查高风险路径（handler 上传/路径、认证、命令执行、加密），对确认的漏洞做最小修复。
4. 移除或收敛确认未使用的导出/函数（仅限无测试/无引用），避免破坏 `main` 或 `embed` 等隐式引用。
5. 全量 `go test ./...` 与上述 Lint/构建复验，直至通过。

## 影响范围

- `cmd/server/main.go`：`-version` 标志、启动日志写入 `version`
- `internal/engine/pipeline.go`：构建脚本解释器分支消除 SA4006（无未使用的 `interpreterArgs` 初值）
- `internal/handler/webhook_handler.go`：Bitbucket webhook 错误文案 ST1005
- `internal/service/server_service.go`：Agent 校验与连接相关错误文案 ST1005

## 历史补丁
