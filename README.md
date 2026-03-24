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

从 [Releases](../../releases) 下载对应平台的二进制文件。系统分为 **主控 (Server)** 与 **执行端 (Agent)** 两部分。

#### 1. 发布文件说明

| 文件类型 | 文件名示例 | 平台 | 说明 |
|----------|------------|------|------|
| **Server** | `buildflow-linux-amd64` | Linux x86_64 | 包含 UI 与核心调度逻辑，通常部署在管理机。 |
| **Server** | `buildflow-windows-amd64.exe` | Windows x64 | — |
| **Agent** | `buildflow-agent-linux-amd64` | Linux x86_64 | 部署在目标生产服务器，接收并部署产物。 |
| **Agent** | `buildflow-agent-windows-amd64.exe` | Windows x64 | — |

#### 2. 主控启动（以 Linux 为例）

```bash
# 1. 赋予执行权限
chmod +x buildflow-linux-amd64

# 2. 创建配置文件 config.yaml
cat > config.yaml << 'EOF'
server:
  port: 8080
  host: "0.0.0.0"

database:
  path: "./data/db.sqlite"

jwt:
  secret: "your-secret-key-change-this"

build:
  workspace_dir: "./data/workspaces"
  artifact_dir: "./data/artifacts"
  log_dir: "./data/logs"
  cache_dir: "./data/caches"

encryption:
  key: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

admin:
  username: "admin"
  password: "admin123"
EOF

# 3. 运行主控
./buildflow-linux-amd64 --config config.yaml
```

启动后访问 `http://localhost:8080`，使用配置文件中的管理员账号登录。

#### 3. 目录与持久化说明

应用启动后会在同级或配置指定的路径下生成 `data/` 目录，其结构及用途如下：

| 目录/文件 | 说明 |
|-----------|------|
| `data/db.sqlite` | SQLite 数据库文件，存储所有项目、环境、用户及审计数据。**务必定期备份**。 |
| `data/workspaces/` | Git 工作区。每个环境对应一个目录，用于存放克隆的代码及执行构建。 |
| `data/artifacts/` | 产物归档区。存储历史构建生成的压缩包（Zip/Gzip），支持按需回滚部署。 |
| `data/logs/` | 实时日志存储。记录每次构建的完整输出，支持通过 WebSocket 实时查看。 |
| `data/caches/` | 构建缓存。通过配置环境的缓存路径（如 `node_modules`），在清理后仍可保留。 |

#### 4. 执行端启动 (Agent)

若环境部署方式选择 **HTTP Agent**，需在目标机器上运行 `buildflow-agent`。主控通过 HTTP 将构建产物推送到 Agent，由 Agent 解压到指定目录并可执行部署后脚本；Agent 与主控之间使用 **Bearer Token** 鉴权（请求头 `Authorization: Bearer <token>`），请确保主控「服务器管理」里填写的 Token 与 Agent 侧配置**完全一致**（明文一致即可，数据库存储为加密字段，部署时会自动解密后使用）。

**配置方式（优先级由低到高：YAML 文件 → 环境变量 → 命令行参数）**

| 来源 | 说明 |
|------|------|
| 默认配置文件 | 与可执行文件同目录下的 `buildflow-agent.yaml`（也可用 `-config` 指定路径）。不存在则忽略，不报错。 |
| 环境变量 | `BUILDFLOW_AGENT_ADDR`、`BUILDFLOW_AGENT_TOKEN`、`BUILDFLOW_AGENT_TLS_CERT`、`BUILDFLOW_AGENT_TLS_KEY` |
| 命令行 | `-addr`、`-token`、`-tls-cert`、`-tls-key`、`-config` |

YAML 示例（与二进制放在同一目录时无需 `-config`）：

```yaml
addr: ":9091"
token: "YOUR_SECRET_TOKEN"
# 可选：启用 HTTPS
# tls_cert: "/path/to/cert.pem"
# tls_key: "/path/to/key.pem"
```

最小启动示例（仅用命令行）：

```bash
chmod +x buildflow-agent-linux-amd64
./buildflow-agent-linux-amd64 -addr :9091 -token YOUR_SECRET_TOKEN
```

使用同目录配置文件时：

```bash
./buildflow-agent-linux-amd64
# 等价于读取 ./buildflow-agent.yaml 中的 addr / token 等
```

**与主控对接**：在「服务器管理」中新增服务器，部署方式选 **Agent**，填写 Agent 监听地址（如 `http://192.168.1.10:9091`；若 Agent 使用 HTTPS 则填 `https://...`）、以及上述 Token。保存后可用连接测试确认可达。

**自检**：Agent 启动后访问 `GET /healthz`（需携带相同 Bearer Token）应返回健康信息。

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
| `encryption.key` | AES-GCM 密钥（64 位 hex，**生产环境务必修改**）。用于加密凭据及敏感变量。 | — |
| `admin.username` | 初始管理员用户名 | `admin` |
| `admin.password` | 初始管理员密码（**首次启动后请修改**） | — |

所有配置项均可通过环境变量覆盖，前缀为 `BUILDFLOW_`，例如 `BUILDFLOW_SERVER_PORT=9090`。

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

开发模式下后端使用 `-tags dev` 编译，不嵌入前端资源。访问 `http://localhost:8070` 使用前端（Vite 自动代理 API 请求到后端）。

### 构建

```bash
# 构建 Linux amd64 二进制
make build-linux

# 构建 Windows amd64 二进制
make build-win

# 构建 Agent 二进制
make build-agent-linux
make build-agent-win

# 清理构建产物
make clean
```

### 项目结构

```
├── cmd/
│   ├── server/          # 主控入口，embed 前端产物
│   └── agent/           # 执行端入口，极简部署工具
├── internal/
│   ├── config/          # 配置加载
│   ├── model/           # GORM 模型 + DB 初始化
│   ├── repository/      # 数据访问层
│   ├── service/         # 业务逻辑层
│   ├── handler/         # HTTP 处理器 (Gin)
│   ├── middleware/      # 认证、RBAC、CORS、审计
│   ├── engine/          # 构建引擎 (Pipeline、Scheduler、Cron)
│   ├── deployer/        # 部署器 (Rsync/SFTP/SCP/Agent)
│   ├── pkg/             # 通用工具（加密、响应封装）
│   └── ws/              # WebSocket Hub
├── web/                 # React 前端 (Vite + shadcn/ui)
├── config.yaml          # 运行时配置示例
└── Makefile
```

## 技术栈

**后端**：Go · Gin · GORM · SQLite · JWT · WebSocket · Cron

**前端**：React 19 · TypeScript · Vite · Tailwind CSS · shadcn/ui · Zustand · Recharts

## 许可证

MIT
