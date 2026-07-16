# AGENTS.md

Bedrock 2.0 单体开发平台（Go Server + Vue 前端）。本文只保留**常用命令、目录导航与读写指引**；不写业务规则与详细设计。

| 文档 | 用途 |
| ---- | ---- |
| [docs/PRD.md](docs/PRD.md) | 产品需求 |
| [docs/DESIGN.md](docs/DESIGN.md) | 架构与领域设计（权威） |
| [docs/ROADMAP.md](docs/ROADMAP.md) | 路线图 |
| [api/openapi.yaml](api/openapi.yaml) | API 契约（OpenAPI 3.2，唯一手改） |
| [.agents/fe.md](.agents/fe.md) | **前端**约定与工作流 |
| [.agents/be.md](.agents/be.md) | **后端**约定与工作流 |

**按需阅读（省上下文）：**

- 改 `web/`、Vue / Veltra / cat-kit → 先读 [`.agents/fe.md`](.agents/fe.md)
- 改 `cmd/`、`internal/`、`api/`、migration、OpenAPI → 先读 [`.agents/be.md`](.agents/be.md)
- 领域行为、权限语义、流水线状态机等 → 查 [docs/DESIGN.md](docs/DESIGN.md)，勿在本文件或 fe/be 中复制产品设计

提交代码时遵循 [`.agents/skills/git-commit`](.agents/skills/git-commit)。

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

# 契约与检查
# OpenAPI 3.2 源：api/openapi.yaml（唯一手改）
# 生成 3.1 投影（禁止手改）：api/openapi.3.1.projection.yaml
make openapi-projection
make openapi-check
make ga-guardrails       # 禁止把 1.x 数据迁移当作支持路径

# 测试 / GA 冒烟
go test ./...
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
cd web && vp install
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
├── api/
│   ├── openapi.yaml            # OpenAPI 3.2（唯一手改）
│   └── openapi.3.1.projection.yaml
├── web/                        # Vue 3 前端 → .agents/fe.md
│   └── src/
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
│   └── skills/
├── config.yaml / config.example.yaml
├── Makefile
└── data/                       # gitignore（db、工作区、制品等）
```

## 跨切卫生（极简）

- **不要**把业务规则、权限/流水线/AI 等领域设计写进本文件；权威在 DESIGN。
- FE / BE 具体禁止项与编码约定分别见 fe.md / be.md。
- 契约：只改 `api/openapi.yaml`，投影用 `make openapi-projection` 生成。
- **不提供** 1.x → 2.0 数据迁移；已接受风险（HTTP + access Web Storage / refresh HttpOnly Cookie 不设 Secure、同 UID、自定义超管命令）见 DESIGN §1.4。
