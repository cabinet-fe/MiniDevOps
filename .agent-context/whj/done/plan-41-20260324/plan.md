# Webhook 环境参数与复制修复

> 状态: 已执行

## 目标

修复项目详情页仓库与 Webhook 复制按钮在部分浏览器/非 HTTPS 场景下不生效的问题；Webhook 支持可选查询参数 `environment_id`，仅对指定环境在分支匹配时触发构建；前端展示项目级与各环境级 Webhook 地址。

## 内容

1. 前端：在 `web/src/lib/utils.ts` 增加带降级策略的 `copyTextToClipboard`；`detail.tsx` 的 `UrlRow` 使用该方法并 `type="button"`；在项目信息区保留全局 Webhook URL，在每个环境卡片内增加该环境专用的 Webhook URL（带 `environment_id`）及复制。
2. 后端：`internal/handler/webhook_handler.go` 读取 `environment_id` 查询参数，若存在则仅尝试触发该 ID 对应环境（仍要求分支与 payload 一致）；响应中可附带 `environment_id` 便于调试。
3. 验证：`go test ./internal/handler/...`（或全量 `go test ./...`）、`cd web && bun run build`。

## 影响范围

- 修改文件: `web/src/lib/utils.ts`
- 修改文件: `web/src/pages/projects/detail.tsx`
- 修改文件: `internal/handler/webhook_handler.go`
- 修改文件: `AGENTS.md`

## 历史补丁
