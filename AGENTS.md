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

- 改 `web-v2/`、Vue / Veltra / cat-kit → 先读 [`.agents/fe.md`](.agents/fe.md)
- 改 `cmd/`、`internal/`、`api/`、migration、OpenAPI → 先读 [`.agents/be.md`](.agents/be.md)
- 领域行为、权限语义、流水线状态机等 → 查 [docs/DESIGN.md](docs/DESIGN.md)，勿在本文件或 fe/be 中复制产品设计

提交代码时遵循 [`.agents/skills/git-commit`](.agents/skills/git-commit)。

## 常用命令

**启动开发服务前请先检查是否已在运行，避免重复启动。**

```bash
# 开发（FRONTEND_DIR 默认 web-v2）
make dev                 # 后端 :8080（-tags dev）+ 前端 Vite 代理
make dev-backend         # 仅后端
make dev-frontend        # 仅前端（web-v2）

# 构建
make build               # 前端 → cmd/server/dist → Go 二进制
make build-frontend      # 仅前端（FRONTEND_DIR）
make build-backend       # 仅后端
make build-linux         # Linux amd64
make build-linux-arm64   # Linux arm64
make build-agent-linux   # Deploy Agent

# 契约与检查
# OpenAPI 3.2 源：api/openapi.yaml
# 生成 3.1 投影（禁止手改）：api/openapi.3.1.projection.yaml
make openapi-projection
make openapi-check

# 测试
go test ./...
go test ./internal/cicd/...
go test -run TestXxx ./internal/...
# 三数据库合同测试（需本地或 CI 服务）
go test ./internal/platform/db/... -tags=contract

# 前端（web-v2；推荐 Vite+ 工作流）
cd web-v2 && vp install
cd web-v2 && vp dev
cd web-v2 && vp check    # format + lint + typecheck
cd web-v2 && vp build
# package.json scripts 亦映射到 vp（需全局安装 vp）

# 清理
make clean
```

> Makefile 目标以实现仓库为准逐步对齐；新增脚本时同步更新本文。

## 目录结构（目标态）

```text
.
├── cmd/
│   ├── server/                 # 入口、DI、embed dist
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
├── web-v2/                     # Vue 3 前端 → 约定见 .agents/fe.md
│   └── src/
│       ├── api/
│       ├── stores/
│       ├── router/
│       ├── composables/
│       ├── layouts/
│       ├── components/
│       ├── views/
│       └── lib/
├── docs/
│   ├── PRD.md
│   ├── DESIGN.md
│   └── ROADMAP.md
├── .agents/
│   ├── fe.md                   # 前端 agent 约定
│   ├── be.md                   # 后端 agent 约定
│   └── skills/
├── config.yaml
├── Makefile
└── data/                       # gitignore（db、工作区、制品等）
```

## 跨切卫生（极简）

- **不要**把业务规则、权限/流水线/AI 等领域设计写进本文件；权威在 DESIGN。
- FE / BE 具体禁止项与编码约定分别见 fe.md / be.md。
- 契约：只改 `api/openapi.yaml`，投影用 `make openapi-projection` 生成。
