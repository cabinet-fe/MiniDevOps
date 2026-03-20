# Release 与 Makefile 增加 Agent 二进制

> 状态: 已执行

## 目标

在 GitHub Release 工作流中与 server 一并发布 `cmd/agent` 交叉编译产物；在 Makefile 中增加仅编译 agent 的目标（不构建前端），命名与现有 `build-linux` / `build-win` 风格一致。

## 内容

1. 在 `cmd/agent/main.go` 增加 `var version = "dev"`，以便与 server 一样通过 `-ldflags "-X main.version=..."` 注入版本号。
2. 修改 `.github/workflows/release.yml`：在现有 matrix 构建步骤中于 server 之后追加 `go build` 输出 `buildflow-agent-${suffix}`，与 server 共用同一 artifact 上传（单矩阵行包含两个二进制）。
3. 修改 `Makefile`：新增 `build-agent-linux`、`build-agent-win`（仅 `go build ./cmd/agent`，使用相同 `LDFLAGS`）；更新 `.PHONY` 与 `clean` 规则以包含 `buildflow-agent*`。

## 影响范围

- `cmd/agent/main.go`：版本变量与 healthz 展示
- `.github/workflows/release.yml`：同矩阵构建并上传 agent 二进制
- `Makefile`：`build-agent-linux`、`build-agent-win`、`.PHONY`、`clean`（沿用 `buildflow*`）

## 历史补丁
