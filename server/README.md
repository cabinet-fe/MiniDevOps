# MiniDevOps Server

这是 MiniDevOps 项目的后端服务，基于 Go 语言和 Fiber 框架开发。

## 技术栈

- **Go**: v1.25.0
- **Fiber**: v2 (Web 框架)
- **GORM**: ORM 数据库操作
- **SQLite3**: 数据库
- **JWT**: 用户认证

## 项目结构

```
server/
├── cmd/                    # 主程序入口
│   └── main.go
├── internal/               # 内部代码
│   ├── models/             # 数据模型
│   │   ├── user.go         # 用户模型
│   │   ├── role.go         # 角色模型
│   │   ├── permission.go   # 权限模型
│   │   ├── repository.go   # 仓库模型
│   │   ├── task.go         # 任务模型
│   │   ├── remote.go       # 远程服务器模型
│   │   └── config.go       # 系统配置模型
│   ├── service/            # 业务逻辑服务
│   │   └── auth/           # 认证服务
│   ├── router/             # 路由配置
│   │   ├── auth/           # 认证路由
│   │   ├── user/           # 用户管理路由
│   │   ├── role/           # 角色管理路由
│   │   ├── permission/     # 权限管理路由
│   │   ├── repository/     # 仓库管理路由
│   │   ├── task/           # 任务管理路由
│   │   ├── remote/         # 远程服务器路由
│   │   └── config/         # 系统配置路由
│   ├── db/                 # 数据库相关
│   │   ├── connection.go   # 数据库连接
│   │   ├── migration.go    # 数据库迁移
│   │   └── seed.go         # 种子数据
│   └── utils/              # 工具函数
│       ├── auth.go         # 认证工具
│       ├── middleware.go         # 中间件
│       └── response.go     # 响应工具
├── configs/                # 配置文件
│   └── config.go
├── go.mod                  # Go模块文件
└── README.md
```

## 功能模块

### 1. 用户管理

- 用户增删改查
- 用户角色授权

### 2. 角色管理

- 角色增删改查
- 角色权限设置

### 3. 权限管理

- 权限增删改查
- 菜单和按钮权限管理

### 4. 认证模块

- 用户登录/登出
- JWT 令牌管理

### 5. 代码仓库管理

- 仓库信息管理
- 支持 Git 仓库

### 6. 任务管理

- 构建任务管理
- 自动化部署
- 构建物下载

### 7. 远程服务器管理

- 服务器信息管理
- SSH 连接配置

### 8. 系统配置

- 系统参数配置
- 挂载路径设置

## 安装和运行

### 前置要求

- Go 1.25.0 或更高版本

### 安装依赖

```bash
go mod tidy
```

### 运行项目

```bash
go run cmd/main.go
```

服务器将在 `http://localhost:8080` 启动。

### 环境变量配置

可以通过环境变量配置以下参数：

- `PORT`: 服务端口 (默认: 8080)
- `DB_PATH`: 数据库文件路径 (默认: minidevops.db)
- `JWT_SECRET`: JWT 密钥 (默认: minidevops-secret-key-2024)
- `MOUNT_PATH`: 代码挂载路径 (默认: ~/dev-ops)

## API 接口

### 认证相关

- `POST /api/v1/login` - 用户登录
- `POST /api/v1/logout` - 用户登出
- `GET /api/v1/profile` - 获取用户信息

### 用户管理

- `GET /api/v1/users` - 获取用户列表
- `POST /api/v1/users` - 创建用户
- `GET /api/v1/users/:id` - 获取用户详情
- `PUT /api/v1/users/:id` - 更新用户
- `DELETE /api/v1/users/:id` - 删除用户

### 角色管理

- `GET /api/v1/roles` - 获取角色列表
- `POST /api/v1/roles` - 创建角色
- `GET /api/v1/roles/:id` - 获取角色详情
- `PUT /api/v1/roles/:id` - 更新角色
- `DELETE /api/v1/roles/:id` - 删除角色

### 权限管理

- `GET /api/v1/permissions` - 获取权限列表
- `POST /api/v1/permissions` - 创建权限
- `GET /api/v1/permissions/:id` - 获取权限详情
- `PUT /api/v1/permissions/:id` - 更新权限
- `DELETE /api/v1/permissions/:id` - 删除权限

### 仓库管理

- `GET /api/v1/repositories` - 获取仓库列表
- `POST /api/v1/repositories` - 创建仓库
- `GET /api/v1/repositories/:id` - 获取仓库详情
- `PUT /api/v1/repositories/:id` - 更新仓库
- `DELETE /api/v1/repositories/:id` - 删除仓库

### 任务管理

- `GET /api/v1/tasks` - 获取任务列表
- `POST /api/v1/tasks` - 创建任务
- `GET /api/v1/tasks/:id` - 获取任务详情
- `PUT /api/v1/tasks/:id` - 更新任务
- `DELETE /api/v1/tasks/:id` - 删除任务
- `POST /api/v1/tasks/:id/build` - 构建任务
- `POST /api/v1/tasks/:id/push` - 推送任务
- `GET /api/v1/tasks/:id/download` - 下载构建物

### 远程服务器管理

- `GET /api/v1/remotes` - 获取远程服务器列表
- `POST /api/v1/remotes` - 创建远程服务器
- `GET /api/v1/remotes/:id` - 获取远程服务器详情
- `PUT /api/v1/remotes/:id` - 更新远程服务器
- `DELETE /api/v1/remotes/:id` - 删除远程服务器

### 系统配置

- `GET /api/v1/configs` - 获取配置列表
- `POST /api/v1/configs` - 创建配置
- `GET /api/v1/configs/:key` - 获取配置详情
- `PUT /api/v1/configs/:key` - 更新配置
- `DELETE /api/v1/configs/:key` - 删除配置

## 默认管理员账号

系统初始化后会自动创建默认管理员账号：

- 用户名: `admin`
- 密码: `admin123`

请在首次登录后及时修改密码。

## 开发规范

- 使用有意义的变量名，避免缩写
- 显式错误处理，不要忽略错误
- 适当的代码注释
- 遵循 Go 语言最佳实践
- 使用 GORM 进行数据库操作
- 统一的 JSON 响应格式
