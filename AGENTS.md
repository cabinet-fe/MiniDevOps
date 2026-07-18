# AGENTS.md

**按需阅读（省上下文）：**

- 改 `web/` 先读 [`.agents/fe.md`](.agents/fe.md)
- 改 `cmd/`、`internal/`、migration → 先读 [`.agents/be.md`](.agents/be.md)
- 改 HTTP / JSON 信封 / 分页 / API 契约 → 先读 [`.agents/api.md`](.agents/api.md)
- 领域行为、权限语义、流水线状态机等 → 查 [docs/DESIGN.md](docs/DESIGN.md)，勿在本文件或 fe/be/api 中复制产品设计

## 常用命令

**启动开发服务前请先检查是否已在运行，避免重复启动。**

```bash
# 开发（FRONTEND_DIR 默认 web）
make dev                 # 后端 :8080（-tags dev）+ 前端 Vite 代理
make dev-backend         # 仅后端
make dev-frontend        # 仅前端（web）

# 构建（web → cmd/server/dist → go build embed）
make build               # 前端 → cmd/server/dist → Go 二进制
make build-frontend      # 仅前端（FRONTEND_DIR）
make build-backend       # 仅后端
make build-linux         # Linux amd64 Server
make build-linux-arm64   # Linux arm64 Server
make build-agent-linux   # Deploy Agent amd64
make build-agent-linux-arm64
make checksums           # 对已构建 linux 产物输出 SHA256

# 检查
make ga-guardrails       # 禁止把 1.x 数据迁移当作支持路径

# 测试 / GA 冒烟
go test ./...
go test ./internal/resource/...
go test ./internal/cicd/...
go test -run TestXxx ./internal/...
go test ./internal/platform/db/... -tags=contract   # 三库合同（需 DSN）
make smoke               # fresh-install + api-e2e + recovery + 3db + linux 包
make smoke-fresh-install
make smoke-api-e2e
make smoke-three-db
make smoke-linux-package
make smoke-restart-recovery

# 前端（web；推荐 Vite+ 工作流）
cd web && bun install
cd web && vp dev
cd web && vp check    # format + lint + typecheck
cd web && vp build
# Playwright 冒烟（需后端已启动）：cd web && bunx playwright test

# 清理
make clean
```

> Makefile 目标以实现仓库为准。发布检查单：[docs/release-checklist.md](docs/release-checklist.md)；操作手册：[docs/ops-handbook.md](docs/ops-handbook.md)。

## 目录结构（GA）

```text
.
├── cmd/
│   ├── server/                 # 入口、DI、embed dist（web 产物）
│   └── agent/                  # Deploy Agent
├── internal/
│   ├── platform/               # config、db、migration、健康检查
│   ├── auth/
│   ├── rbac/
│   ├── system/
│   ├── resource/               # 资源管理：仓库、服务器、凭证、AI CLI、访问令牌
│   ├── cicd/
│   ├── engine/
│   ├── deployer/
│   ├── ops/
│   ├── project/
│   ├── ai/
│   ├── dashboard/
│   ├── storage/
│   ├── ws/
│   └── pkg/
├── api/                        # HTTP 契约（Markdown，按域拆分）
│   ├── README.md
│   ├── auth.md
│   ├── system.md
│   ├── resource.md
│   ├── cicd.md
│   ├── ops.md
│   ├── project.md
│   └── ai.md
├── web/                        # Vue 3 前端 → .agents/fe.md
│   └── src/
│       └── views/system/resources/  # 权限资源页：分组面板 + 菜单/功能面板
├── scripts/
│   ├── check-ga-guardrails.sh
│   └── smoke/                  # fresh-install / api-e2e / 3db / linux-package
├── docs/
│   ├── PRD.md
│   ├── DESIGN.md
│   ├── ROADMAP.md
│   ├── ops-handbook.md
│   ├── release-checklist.md
│   ├── known-issues.md
│   └── roadmap/
├── .agents/
│   ├── fe.md
│   ├── be.md
│   ├── api.md
│   └── skills/
├── config.yaml / config.example.yaml
├── Makefile
└── data/                       # gitignore（db、工作区、制品等）
```

**以上目录结构在文件更改时需要同步更新!**
