## Why

团队需要一个轻量级的代码构建交付平台，将"拉取代码 → 构建 → 打包 → 推送到服务器"这条核心链路自动化。现有方式依赖手动操作（本地构建、SCP 上传、SSH 重启），效率低且容易出错。市面上的 CI/CD 工具（Jenkins、GitLab CI）过于重量级，部署和维护成本高，不适合小团队。需要一个单二进制部署、开箱即用的解决方案。

## What Changes

- 创建全新的 Go + React 全栈应用 "BuildFlow"
- 后端：Go (Gin) + SQLite + GORM，单二进制部署
- 前端：React + shadcn/ui + Tailwind CSS，通过 Go embed.FS 内嵌
- 实现用户认证与三级角色权限系统（Admin / Ops / Dev）
- 实现项目管理，支持多环境配置（dev / staging / prod）
- 实现远程服务器管理，支持 SSH 密码和密钥两种认证
- 实现构建引擎，支持并发调度、实时日志流、构建产物归档与保留策略
- 实现三种部署方式（rsync / SFTP / SCP）+ 部署后远程脚本执行
- 实现 Git Webhook 触发构建
- 实现站内实时通知（WebSocket）
- 实现审计日志记录所有关键操作
- 实现系统备份/恢复与项目导入/导出

## Capabilities

### New Capabilities

- `auth`: JWT 认证体系，包括登录/登出、Token 刷新、RBAC 中间件（admin/ops/dev 三级权限）
- `user-management`: 用户 CRUD，管理员管理所有用户，普通用户管理个人资料
- `project-management`: 项目 CRUD，仓库配置（URL + 用户名/密码/Token 授权），多环境子配置（分支、构建脚本、输出目录、环境变量），项目导入/导出
- `server-management`: 远程服务器 CRUD，SSH 连接配置（密码/密钥），连接测试，标签分组
- `build-engine`: 构建调度器（goroutine + semaphore 并发控制），构建流水线（clone → build → artifact），构建日志逐行捕获，产物归档与按项目配置保留最近 N 个，构建取消，构建历史与回滚
- `deployment`: 可插拔部署策略（rsync / SFTP / SCP），部署后通过 SSH 执行远程脚本，部署目标从服务器池选择或自定义
- `realtime`: WebSocket 实时通信，构建日志流推送，站内通知推送
- `webhook`: Git push 事件触发自动构建，基于 secret 的请求验证
- `audit-log`: 审计日志中间件，记录用户操作（who/what/when/where），支持查询和过滤
- `backup-restore`: 系统备份（SQLite + 配置打包），系统恢复（上传覆盖），项目级导入/导出（JSON 序列化）
- `dashboard`: 仪表盘总览，统计卡片（项目数/构建数/成功率），活跃构建实时状态，最近 7 天构建趋势图，最近构建列表

### Modified Capabilities

（无，全新项目）

## Impact

- **代码**：从零创建完整的 Go 后端 + React 前端项目
- **API**：全新 RESTful API（/api/v1/*）+ WebSocket 端点（/ws/*）
- **依赖**：Go 模块（Gin、GORM、JWT、WebSocket、go-git、SSH 等）；npm 包（React、Vite、Tailwind、shadcn/ui、Zustand、Recharts 等）
- **数据存储**：SQLite 数据库（users、projects、environments、servers、builds、notifications、audit_logs），本地文件系统（工作目录、构建产物、日志）
- **外部系统交互**：Git 远程仓库（clone/pull）、远程服务器（SSH/rsync/SFTP/SCP）、Git 平台 Webhook
- **部署产物**：单个 Go 二进制文件 + config.yaml 配置文件
