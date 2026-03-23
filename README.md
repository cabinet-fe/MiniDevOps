# BuildFlow

轻量级 CI/CD 构建部署平台。Go 后端 + React 前端单体仓库，前端产物通过 `embed` 嵌入后端二进制，单文件即可部署。

## 功能特性

- **项目管理** — 多项目、多环境（开发/测试/生产），支持项目导入导出
- **构建引擎** — Git 克隆 → 脚本构建 → 产物归档（gzip/zip），支持并发限制
- **多种部署方式** — Rsync / SFTP / SCP / HTTP Agent 推送
- **定时构建** — 基于 Cron 表达式的定时触发
- **Webhook** — 支持外部系统（如 Gitea/GitHub）触发构建
- **环境变量** — 变量组共享，支持加密存储敏感数据
- **实时日志** — WebSocket 实时推送构建日志
- **通知系统** — 构建结果实时推送，支持站内通知
- **RBAC 权限** — admin / ops / dev 三级角色控制
- **审计日志** — 自动记录所有状态变更操作
- **数据字典** — 可配置的项目标签等枚举数据

## 快速开始

### 二进制部署

从 [Releases](../../releases) 下载对应平台的二进制文件：

| 文件 | 平台 |
|------|------|
| `buildflow-linux-amd64` | Linux x86_64 |
| `buildflow-linux-arm64` | Linux ARM64 |
| `buildflow-windows-amd64.exe` | Windows x64 |

```bash
# 下载（以 Linux x86_64 为例）
chmod +x buildflow-linux-amd64

# 创建配置文件
cat > config.yaml << 'EOF'
server:
  port: 8080
  host: "0.0.0.0"

database:
  path: "./data/db.sqlite"

jwt:
  secret: "your-secret-key-change-this"
  access_ttl: "2h"
  refresh_ttl: "168h"

build:
  max_concurrent: 3
  workspace_dir: "./data/workspaces"
  artifact_dir: "./data/artifacts"
  log_dir: "./data/logs"
  cache_dir: "./data/caches"

encryption:
  key: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

admin:
  username: "admin"
  password: "admin123"
  display_name: "Administrator"
EOF

# 启动
./buildflow-linux-amd64 --config config.yaml
```

启动后访问 `http://localhost:8080`，使用配置文件中的管理员账号登录。

### 配置说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `server.port` | 监听端口 | `8080` |
| `server.host` | 监听地址 | `0.0.0.0` |
| `database.path` | SQLite 数据库路径 | `./data/db.sqlite` |
| `jwt.secret` | JWT 签名密钥（**生产环境务必修改**） | — |
| `jwt.access_ttl` | Access Token 有效期 | `2h` |
| `jwt.refresh_ttl` | Refresh Token 有效期 | `168h` |
| `build.max_concurrent` | 最大并发构建数 | `3` |
| `build.workspace_dir` | Git 工作区目录 | `./data/workspaces` |
| `build.artifact_dir` | 构建产物目录 | `./data/artifacts` |
| `build.log_dir` | 构建日志目录 | `./data/logs` |
| `build.cache_dir` | 构建缓存目录 | `./data/caches` |
| `encryption.key` | AES-GCM（敏感字段）与登录可选 AES-CBC 共用密钥（64 位 hex，**生产环境务必修改**）。嵌入二进制会在响应 `index.html` 时注入该密钥到 `window.__BUILDFLOW_ENCRYPTION_KEY__`，与运行时配置一致；本地开发或非 Go 托管时可在 `web/.env` 设 `VITE_BUILDFLOW_ENCRYPTION_KEY` 对齐 | — |
| `admin.username` | 初始管理员用户名 | `admin` |
| `admin.password` | 初始管理员密码（**首次启动后请修改**） | — |

所有配置项均可通过环境变量覆盖，前缀为 `BUILDFLOW_`，例如 `BUILDFLOW_SERVER_PORT=9090`。

### 数据库迁移（仓库凭证）

从旧版本（项目内直接保存 `repo_username/repo_password`）升级到新版本（独立凭证表）时，无需手动执行 SQL：

1. 应用启动时会通过 GORM `AutoMigrate` 自动创建 `credentials` 表，并为 `projects` 表新增 `credential_id` 字段。
2. 启动后会自动扫描历史项目：`repo_auth_type != none` 且存在 `repo_password` 的记录会被迁移为独立凭证，并自动回填 `projects.credential_id`。
3. 迁移逻辑是幂等的，重复启动不会重复迁移同一项目。

建议线上升级步骤：

1. 停止旧版本服务。
2. 备份数据库文件（默认 `data/db.sqlite`）。
3. 部署并启动新版本，等待自动迁移完成。
4. 验证项目凭证是否已迁移成功（进入「凭证」页面检查）。

回滚方案：停止新版本，恢复备份数据库文件，并切回旧版本二进制。

## 开发指南

### 环境要求

- Go 1.25+
- Bun 1.x
- Node.js 22+（Bun 依赖）

### 本地开发

```bash
# 同时启动后端（:8080）和前端 Vite 开发服务器（:8070）
make dev
```

开发模式下后端使用 `-tags dev` 编译，不嵌入前端资源。访问 `http://localhost:8070` 使用前端（Vite 自动代理 API 请求到后端）。密文登录：开发依赖 `web/.env` 中的 `VITE_BUILDFLOW_ENCRYPTION_KEY` 与后端一致；**生产嵌入二进制**由服务端在返回的 `index.html` 中注入运行时 `encryption.key`，修改 `config.yaml` 后重启即可，无需为改密钥而重编前端（CI 中仍可注入 `VITE_*` 作为构建校验或静态托管场景的备用）。

### 构建

```bash
# 构建 Linux amd64 二进制
make build-linux

# 构建 Windows amd64 二进制
make build-win

# 清理构建产物
make clean
```

### 项目结构

```
├── cmd/server/          # 入口，embed 前端产物
├── internal/
│   ├── config/          # 配置加载
│   ├── model/           # GORM 模型 + DB 初始化
│   ├── repository/      # 数据访问层
│   ├── service/         # 业务逻辑层
│   ├── handler/         # HTTP 处理器（Gin）
│   ├── middleware/       # 认证、RBAC、CORS、审计
│   ├── engine/          # 构建引擎（Pipeline、Scheduler、Cron）
│   ├── deployer/        # 部署器（Rsync/SFTP/SCP/Agent）
│   ├── pkg/             # 通用工具（加密、响应封装）
│   └── ws/              # WebSocket Hub
├── web/                 # React 前端（Vite + shadcn/ui）
├── config.yaml          # 运行时配置
└── Makefile
```

## 技术栈

**后端**：Go · Gin · GORM · SQLite · JWT · WebSocket · Cron

**前端**：React 19 · TypeScript · Vite · Tailwind CSS · shadcn/ui · Zustand · Recharts

## 许可证

MIT
