## Context

BuildFlow 是一个从零开始的代码构建交付平台，面向小团队（几人规模）。当前没有现有代码或遗留系统，是全新的 greenfield 项目。

核心约束：
- 单二进制部署（Go embed.FS 内嵌前端）
- 宿主机直接执行构建（非 Docker 隔离）
- SQLite 单文件数据库（几人规模无需分布式数据库）
- 所有依赖使用最新稳定版本

利益方：小团队的开发和运维人员，管理员负责系统配置。

## Goals / Non-Goals

**Goals:**

- 单二进制文件部署，零外部数据库依赖
- 高并发构建调度，支持可配置的并发上限
- 实时构建日志流（WebSocket），毫秒级推送
- 可插拔的部署策略（rsync / SFTP / SCP），便于扩展
- 三级角色权限（Admin / Ops / Dev），细粒度 API 级控制
- 前端现代化 UI，基于 shadcn/ui + Tailwind CSS
- 敏感数据（密码、密钥、Token）AES-256-GCM 加密存储

**Non-Goals:**

- 不支持 Docker 容器化构建隔离（当前只在宿主机执行）
- 不支持分布式 Worker 节点（单机架构）
- 不支持多租户隔离
- 不支持外部通知（邮件、钉钉等），仅站内通知
- 不支持构建流水线编排（多阶段 Pipeline），只支持单脚本执行
- 不做前端 SSR，纯 SPA 模式

## Decisions

### D1: Web 框架选择 Gin

**选择**: Gin v1.11.0
**备选**: Fiber (fasthttp)、Echo、标准库 net/http

**理由**: Gin 社区最大、中间件生态最丰富、文档最多。构建平台的性能瓶颈在 IO（Git 操作、文件传输）而非 HTTP 框架本身，Gin 的性能绰绰有余。Fiber 基于 fasthttp 与标准库不兼容，长期维护有风险。

### D2: 数据库选择 SQLite + WAL

**选择**: SQLite (WAL mode, busy_timeout=5000)
**备选**: PostgreSQL、MySQL

**理由**: 几人团队规模，SQLite 完全满足并发需求。WAL 模式允许读写并发。单文件特性使备份/恢复极为简单（拷贝文件即可）。零运维成本。如果未来需要扩展到多实例，可以通过更换 GORM driver 迁移到 PostgreSQL，代码层面改动极小。

### D3: 构建调度使用 goroutine + buffered channel

**选择**: 自建调度器（buffered channel 做信号量）
**备选**: 第三方任务队列（Asynq、Machinery）、消息队列（Redis/RabbitMQ）

**理由**: 几人使用场景下，内置调度器足够。buffered channel 天然做信号量控制并发数，goroutine 做异步执行，无额外依赖。任务队列（Asynq 等）引入 Redis 依赖，违背"零外部依赖"的设计目标。

### D4: Git 操作使用 exec git CLI 而非 go-git

**选择**: 通过 `os/exec` 调用系统 git 命令
**备选**: go-git 纯 Go 实现

**理由**: 虽然 go-git 是纯 Go 实现无外部依赖，但在实际构建场景中，exec git 更可靠：支持所有 git 特性（submodule、LFS、shallow clone）、性能更好（大仓库 clone 速度明显快于 go-git）、错误信息更友好。构建平台的宿主机必然已安装 git，不存在依赖问题。

### D5: 前端状态管理使用 Zustand

**选择**: Zustand v5
**备选**: Redux Toolkit、Jotai、Context API

**理由**: Zustand 极简（无 Provider 包裹）、TypeScript 友好、体积极小。构建平台的全局状态不复杂（认证状态、通知列表），不需要 Redux 的重量级方案。

### D6: 前端内嵌方案

**选择**: Go 1.16+ embed.FS 静态内嵌
**方案**: 前端 Vite 构建产物输出到 `web/dist/`，Go 通过 `//go:embed web/dist` 指令嵌入，`gin.StaticFS` 提供服务。所有非 `/api/` 和 `/ws/` 的请求 fallback 到 `index.html`（SPA 路由支持）。

**开发模式**: 前端 Vite dev server (端口 5173) + 后端 Gin (端口 8080)，Vite 配置 proxy 代理 `/api/` 和 `/ws/` 到后端。

### D7: 敏感数据加密

**选择**: AES-256-GCM 对称加密
**密钥**: 从 config.yaml 的 `encryption.key` 读取（32 字节 hex）
**加密字段**: server.password、server.private_key、project.repo_password、environment.env_vars 中的敏感值

**理由**: 数据库文件可能被备份/传输，敏感信息不应明文存储。AES-256-GCM 提供加密 + 完整性校验，标准库支持好。

### D8: 构建日志双写

**选择**: 构建过程中日志同时写入文件和 WebSocket
**实现**: 构建脚本的 stdout/stderr 通过 `io.MultiWriter` 同时写入：(1) 本地 log 文件（持久化）；(2) WebSocket Hub 广播（实时推送）。客户端断开重连时，先读取已有 log 文件内容，再切换到 WebSocket 实时流。

### D9: 构建产物清理策略

**选择**: 每个项目可配置 `max_artifacts`（默认 5），构建成功后检查产物数量，超出上限时删除最旧的。
**存储**: 产物以 `tar.gz` 格式归档到 `data/artifacts/project-{id}/build-{number}.tar.gz`。
**回滚**: 回滚操作本质是重新部署某个历史产物，不触发新构建。

### D10: 多环境作为项目子配置

**选择**: `environments` 表关联 `projects` 表，每个环境独立的分支、构建脚本、部署目标、环境变量。
**理由**: 同一个项目的不同环境共享仓库配置和授权信息，避免重复维护。在 UI 上，项目详情页通过 Tab 切换环境。

## Risks / Trade-offs

**[宿主机执行不隔离]** → 构建脚本可以访问宿主机所有资源，有安全风险。缓解：限制构建用户权限、信任团队内部用户（小团队场景）。

**[SQLite 写锁]** → SQLite 写操作串行，高并发写入时可能有等待。缓解：WAL 模式 + busy_timeout 减轻影响；读操作不受影响；几人规模下写入压力极小。

**[单点故障]** → 单机部署无高可用。缓解：定期自动备份 SQLite 文件；几人小团队场景可接受。

**[构建脏状态]** → 宿主机构建可能残留上次构建的文件。缓解：每次构建前执行 `git clean -fdx && git reset --hard`，但保留 .gitignore 中的依赖目录（通过 `-e node_modules -e vendor` 排除）。

**[rsync Windows 不兼容]** → Windows 服务器不原生支持 rsync。缓解：提供 SFTP 作为 fallback 方案，用户按目标服务器系统选择部署方式。

## Open Questions

- 是否需要支持构建超时自动取消？（建议：是，可配置每个环境的超时时间）
- 首次启动是否自动创建默认 admin 账户？（建议：是，默认 admin/admin123，强制首次登录修改密码）
