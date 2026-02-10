## 1. 项目初始化与基础设施

- [ ] 1.1 初始化 Go module (`go mod init`)，创建项目目录结构 (cmd/, internal/, web/)
- [ ] 1.2 安装 Go 依赖：Gin, GORM, SQLite driver, JWT, WebSocket, Viper, Zap, x/crypto
- [ ] 1.3 创建 config.yaml 默认配置文件，实现 Viper 配置加载 (internal/config/config.go)
- [ ] 1.4 实现 Zap 日志初始化 (开发/生产两种模式)
- [ ] 1.5 初始化 React + Vite + TypeScript 前端项目 (web/)，安装 Tailwind CSS v4、shadcn/ui、React Router、Zustand、Recharts、TanStack Table、Lucide React
- [ ] 1.6 配置 Vite 开发代理 (proxy /api/ 和 /ws/ 到后端)
- [ ] 1.7 创建 Makefile (dev-backend, dev-frontend, build 目标)

## 2. 数据库模型与迁移

- [ ] 2.1 实现 GORM 数据库连接初始化 (SQLite WAL mode, busy_timeout=5000)
- [ ] 2.2 定义 User model (internal/model/user.go) 含 GORM tags 和字段验证
- [ ] 2.3 定义 Server model (internal/model/server.go)
- [ ] 2.4 定义 Project model (internal/model/project.go)
- [ ] 2.5 定义 Environment model (internal/model/environment.go) 含唯一索引 (project_id, name)
- [ ] 2.6 定义 Build model (internal/model/build.go) 含唯一索引 (project_id, build_number)
- [ ] 2.7 定义 Notification model (internal/model/notification.go)
- [ ] 2.8 定义 AuditLog model (internal/model/audit_log.go)
- [ ] 2.9 实现 AutoMigrate 自动建表 + 默认 admin 账户初始化

## 3. 公共工具包

- [ ] 3.1 实现 AES-256-GCM 加密/解密工具 (internal/pkg/crypto.go)
- [ ] 3.2 实现 bcrypt 密码哈希与校验工具
- [ ] 3.3 实现统一 API 响应格式 (internal/pkg/response.go)：Success/Error/Paginated
- [ ] 3.4 实现请求参数校验工具 (internal/pkg/validator.go)

## 4. 认证系统

- [ ] 4.1 实现 JWT 签发 (Access Token + Refresh Token) 与验证 (internal/service/auth_service.go)
- [ ] 4.2 实现 Auth 中间件：从 Header 提取 Bearer Token 并验证 (internal/middleware/auth.go)
- [ ] 4.3 实现 RBAC 中间件：RequireRole() 和 RequireOwnerOrRole() (internal/middleware/rbac.go)
- [ ] 4.4 实现 CORS 中间件 (internal/middleware/cors.go)
- [ ] 4.5 实现 Auth handler：POST /login, POST /logout, POST /refresh, GET /me, PUT /profile (internal/handler/auth_handler.go)
- [ ] 4.6 实现 User repository CRUD (internal/repository/user_repo.go)
- [ ] 4.7 实现 User service 业务逻辑 (internal/service/user_service.go)

## 5. 用户管理

- [ ] 5.1 实现 User handler：GET/POST/PUT/DELETE /users (internal/handler/user_handler.go)
- [ ] 5.2 实现管理员创建用户（密码 bcrypt 哈希）
- [ ] 5.3 实现用户列表分页查询
- [ ] 5.4 实现禁止自删和最后一个管理员保护

## 6. 服务器管理

- [ ] 6.1 实现 Server repository CRUD (internal/repository/server_repo.go)
- [ ] 6.2 实现 Server service：创建/更新时加密密码和私钥，标签过滤 (internal/service/server_service.go)
- [ ] 6.3 实现 Server handler：CRUD + 按标签过滤 + 角色权限控制 (internal/handler/server_handler.go)
- [ ] 6.4 实现 SSH 连接测试功能 (POST /servers/:id/test)
- [ ] 6.5 实现删除服务器时检查环境引用（有引用则 409）

## 7. 项目管理

- [ ] 7.1 实现 Project repository CRUD (internal/repository/project_repo.go)
- [ ] 7.2 实现 Project service：创建时生成 webhook_secret，加密 repo_password，按角色过滤列表 (internal/service/project_service.go)
- [ ] 7.3 实现 Project handler：CRUD + owner 权限校验 (internal/handler/project_handler.go)
- [ ] 7.4 实现 Environment CRUD（project 子资源）含唯一性校验
- [ ] 7.5 实现项目删除级联（environments, builds, artifacts, workspace 目录）
- [ ] 7.6 实现项目导出 JSON (GET /projects/:id/export)
- [ ] 7.7 实现项目导入 JSON (POST /projects/import) 含名称冲突处理

## 8. WebSocket Hub

- [ ] 8.1 实现 WebSocket Hub (internal/ws/hub.go)：客户端注册/注销、按频道订阅、消息广播
- [ ] 8.2 实现 WebSocket handler：/ws/builds/:id/logs 和 /ws/notifications (internal/handler/ws_handler.go)
- [ ] 8.3 实现 WebSocket 连接的 JWT 认证 (query param token 验证)

## 9. 构建引擎

- [ ] 9.1 实现构建调度器 (internal/engine/scheduler.go)：buffered channel 信号量、goroutine 执行池、Submit/Run/Shutdown
- [ ] 9.2 实现 Git 操作 (internal/engine/git.go)：clone（首次）/ fetch+reset（后续），支持 username/password 和 token 认证，workspace cleanup（git clean 保留依赖目录）
- [ ] 9.3 实现构建流水线 (internal/engine/pipeline.go)：status 状态机转换、环境变量注入、exec 构建脚本、逐行日志捕获 (io.MultiWriter → 文件 + WebSocket)
- [ ] 9.4 实现构建产物收集：打包 build_output_dir 为 tar.gz，存储到 artifacts 目录
- [ ] 9.5 实现产物保留策略：构建后检查数量，超出 max_artifacts 删除最旧的
- [ ] 9.6 实现构建取消：通过 context.Cancel 终止进程、更新状态
- [ ] 9.7 实现 Build repository CRUD + build_number 自增 (internal/repository/build_repo.go)
- [ ] 9.8 实现 Build service 业务逻辑 (internal/service/build_service.go)
- [ ] 9.9 实现 Build handler：触发构建、构建历史、构建详情、取消、下载产物 (internal/handler/build_handler.go)

## 10. 部署系统

- [ ] 10.1 定义 Deployer 接口 (internal/deployer/deployer.go) + DeployOptions 结构体 + 工厂函数
- [ ] 10.2 实现 RsyncDeployer (internal/deployer/rsync.go)：通过 exec rsync 命令执行同步
- [ ] 10.3 实现 SFTPDeployer (internal/deployer/sftp.go)：使用 x/crypto/ssh + sftp 包上传文件
- [ ] 10.4 实现 SCPDeployer (internal/deployer/scp.go)：使用 x/crypto/ssh 执行 SCP 传输
- [ ] 10.5 实现部署后 SSH 远程脚本执行 (post_deploy_script)
- [ ] 10.6 将部署步骤集成到构建流水线 (build → deploy → post-deploy)
- [ ] 10.7 实现手动部署/重部署 (POST /builds/:id/deploy)
- [ ] 10.8 实现回滚 (POST /builds/:id/rollback)：重新部署历史产物

## 11. Webhook

- [ ] 11.1 实现 Webhook handler (internal/handler/webhook_handler.go)：解析 push 事件、验证 secret、匹配分支触发构建
- [ ] 11.2 支持 GitHub 和 GitLab webhook payload 格式解析

## 12. 通知系统

- [ ] 12.1 实现 Notification repository CRUD (internal/repository/notification_repo.go)
- [ ] 12.2 实现 Notification service：构建完成时为相关用户创建通知、通过 WebSocket Hub 广播 (internal/service/notification_service.go)
- [ ] 12.3 实现 Notification handler：列表、标记已读、全部已读 (internal/handler/notification_handler.go)

## 13. 审计日志

- [ ] 13.1 实现审计日志中间件 (internal/middleware/audit.go)：自动记录 state-changing 请求
- [ ] 13.2 实现 AuditLog repository：写入 + 分页查询 + 多条件过滤 (internal/repository/audit_repo.go)
- [ ] 13.3 实现 System handler：审计日志查询 API (GET /system/audit-logs)

## 14. 系统备份与恢复

- [ ] 14.1 实现系统备份导出 (POST /system/backup)：打包 SQLite + config.yaml 为 tar.gz
- [ ] 14.2 实现系统恢复 (POST /system/restore)：上传 tar.gz 解压覆盖、重新连接数据库

## 15. 路由注册与服务器入口

- [ ] 15.1 实现 Gin 路由注册：所有 API routes + 中间件挂载 + WebSocket endpoints
- [ ] 15.2 实现 main.go 入口：加载配置 → 初始化 DB → 初始化调度器 → 注册路由 → 启动服务器
- [ ] 15.3 实现 embed.FS 前端静态文件服务 + SPA fallback (非 /api/ 和 /ws/ 的 GET 请求返回 index.html)
- [ ] 15.4 实现优雅关机 (graceful shutdown)：等待构建完成、关闭调度器、关闭 DB

## 16. 前端基础

- [ ] 16.1 配置 shadcn/ui 组件系统 (安装所需组件：Button, Input, Table, Dialog, Form, Card, Tabs, Badge, DropdownMenu, Sheet, Toast, Tooltip 等)
- [ ] 16.2 实现 App Layout 组件：可折叠侧边栏 + 顶栏 (Logo, 通知铃铛, 用户菜单)
- [ ] 16.3 实现 React Router 路由配置 + 路由守卫 (未认证跳转 /login，按角色隐藏菜单)
- [ ] 16.4 实现 API 客户端 (web/src/lib/api.ts)：Fetch 封装、自动携带 Token、401 自动跳转 login、统一错误处理
- [ ] 16.5 实现 Auth Store (web/src/stores/auth-store.ts)：Token 管理、用户信息、登录/登出 actions
- [ ] 16.6 实现 Notification Store (web/src/stores/notification-store.ts)：通知列表、未读计数、WebSocket 连接管理

## 17. 前端页面实现

- [ ] 17.1 实现 LoginPage：用户名/密码表单、登录请求、错误提示
- [ ] 17.2 实现 DashboardPage：统计卡片 (4个)、活跃构建列表、7天构建趋势折线图 (Recharts)、最近构建表格 (TanStack Table)
- [ ] 17.3 实现 ProjectListPage：卡片/表格视图切换、搜索过滤、创建按钮
- [ ] 17.4 实现 ProjectFormPage (创建/编辑)：多步表单 (基础信息 → 环境配置 → 确认)，环境可动态添加/删除，服务器下拉选择
- [ ] 17.5 实现 ProjectDetailPage：项目信息卡片、环境 Tab 切换、每个环境下的构建历史表格、触发构建/部署/下载按钮、webhook URL 展示
- [ ] 17.6 实现 BuildDetailPage：构建状态 badge、commit 信息、实时日志终端组件 (等宽字体、自动滚动、ANSI 颜色支持)、取消/重新部署/下载操作
- [ ] 17.7 实现 ServerListPage：服务器表格、标签筛选、创建/编辑/删除/测试连接操作
- [ ] 17.8 实现 ServerFormPage：SSH 配置表单 (密码/密钥切换)、标签输入、连接测试
- [ ] 17.9 实现 UserListPage (admin)：用户表格、创建/编辑/删除 Dialog、角色选择、启用/禁用切换
- [ ] 17.10 实现 AuditLogPage：审计日志表格、多条件筛选 (操作类型、用户、日期范围)
- [ ] 17.11 实现 SettingsPage (admin)：系统备份导出/导入恢复、项目导入按钮
- [ ] 17.12 实现 NotificationBell 组件：未读计数 badge、下拉通知列表、标记已读、WebSocket 实时更新

## 18. 前端实时通信

- [ ] 18.1 实现 useWebSocket hook (web/src/hooks/use-websocket.ts)：连接管理、自动重连、认证
- [ ] 18.2 实现 BuildLogViewer 组件：WebSocket 连接 /ws/builds/:id/logs、实时渲染日志行、自动滚动到底部
- [ ] 18.3 集成通知 WebSocket：全局连接 /ws/notifications、收到新通知时更新 store 和弹出 toast

## 19. 整合与打包

- [ ] 19.1 实现 Go embed.FS 内嵌前端 build 产物到二进制
- [ ] 19.2 完善 Makefile：前端 build → Go embed → 交叉编译 (linux/windows/darwin)
- [ ] 19.3 端到端冒烟测试：启动服务 → 登录 → 创建项目/环境/服务器 → 触发构建 → 验证日志流 → 下载产物
- [ ] 19.4 编写 README.md：项目介绍、快速开始、配置说明、开发指南
