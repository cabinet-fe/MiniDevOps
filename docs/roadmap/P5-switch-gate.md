# web-v2 切换 Gate 证据（P5 §3.1）

对照 [ROADMAP.md §7.3](../ROADMAP.md#73-web-v2-切换-gate必须全部满足) / [P5.md §3.1](./P5.md#31-web-v2-切换-gate必须全部满足)。

| # | Gate 项 | 证据 / 验证方式 | 状态 |
| --- | --- | --- | --- |
| 1 | 路由深链与书签 | `web-v2/src/router/index.ts` 含 `cicd/build-runs/:id` 等；`scripts/smoke/fresh-install.sh` 请求深链路径返回 SPA 200；Playwright `e2e/smoke.spec.ts` | 通过 |
| 2 | API 信封、401 refresh、登出 | `@cat-kit/http` + `web-v2/src/api/http.ts` Envelope/Token 插件；`stores/auth.ts` logout 清 token 并跳转登录 | 通过 |
| 3 | `password_cipher` 字节兼容；`__BEDROCK_ENCRYPTION_KEY__` | `internal/pkg/crypto_golden_test.go`；`cmd/server/embed_prod.go` 注入；smoke 登录用 golden cipher；fresh-install 检查 index 注入 | 通过 |
| 4 | WS：构建日志与通知；Agent 日志 | **构建日志**：`GET /ws/build-runs/:id/logs`（cicd WS handler）；web-v2 构建详情订阅。**通知**：`GET /ws/notifications` → 频道 `notifications:{userId}`（system WS handler）；REST `GET /api/v1/notifications`、`PUT .../read`、`PUT .../read-all`；Pipeline/Agent 终态经 `TerminalNotifier` 落库并推送；web-v2 顶栏铃铛 REST+WS。**Agent 日志**：`GET /ws/ai/runs/:id/logs`。冒烟：`scripts/smoke/api-e2e.sh` 断言 REST、无 token→401、Bun WebSocket 升级成功、触发构建后出现 `build_run_*` 通知 | 通过 |
| 5 | 制品下载、备份/恢复或 FormData | OpenAPI `build-runs/{id}/artifact`；项目附件/文档上传；系统备份相关 API（若启用） | 通过（API/UI 路径存在） |
| 6 | 菜单服务端下发 | `app-layout.vue` 仅渲染 `/auth/me` menus；无本地全量菜单真源；smoke 断言 menus 非空 | 通过 |
| 7 | `vp check && vp build`；`go build` embed | CI `frontend` + `backend` jobs；本地 `make build` | 通过 |

**默认 FRONTEND_DIR**：`Makefile` `FRONTEND_DIR ?= web-v2`；Release workflow 构建 `web-v2/dist` → `cmd/server/dist`。

**回滚**：见 [../release-checklist.md](../release-checklist.md#前端-embed-回滚)；旧 `web/` 保留一个发布周期。

复验命令：

```bash
make openapi-check
cd web-v2 && vp check && vp build && cd ..
make build-backend
make smoke-fresh-install
make smoke-api-e2e
# 可选 UI：cd web-v2 && bunx playwright test
```
