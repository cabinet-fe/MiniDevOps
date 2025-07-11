# MiniDevOps 开发指南

## 快速开始

### 1. 安装依赖

```bash
# 进入服务器目录
cd server

# 安装Go依赖
go mod tidy

# 安装Air热重载工具
make install-air
```

### 2. 开发模式

使用以下任一方式启动开发服务器：

#### 方式一：使用 Make 命令

```bash
make dev
```

#### 方式二：直接使用 Air

```bash
air
```

### 3. 其他命令

```bash
# 构建项目
make build

# 运行项目（不使用热重载）
make run

# 清理构建文件
make clean

# 格式化代码
make fmt

# 运行测试
make test

# 查看所有可用命令
make help
```

## Air 配置说明

Air 配置文件位于 `.air.toml`，主要配置项：

- **工作目录**: 项目根目录 (`root = "."`)
- **构建命令**: `go build -o ./tmp/main ./cmd`
- **监听目录**: `cmd`, `internal`, `configs`
- **监听文件**: `.go`, `.tpl`, `.tmpl`, `.html`, `.yaml`, `.yml`
- **排除目录**: `web`, `tmp`, `vendor`, `testdata`, `.git`, `.idea`, `.vscode`
- **排除文件**: 测试文件 (`*_test.go`)、数据库文件 (`*.db`)、日志文件 (`*.log`)

## 项目结构

```
server/
├── cmd/                    # 主程序入口
│   └── main.go
├── internal/               # 内部代码
│   ├── models/             # 数据模型
│   ├── service/            # 业务逻辑服务
│   ├── router/             # 路由配置
│   ├── db/                 # 数据库相关
│   └── utils/              # 工具函数
├── configs/                # 配置文件
├── scripts/                # 脚本文件
├── tmp/                    # Air临时文件（自动生成）
├── .air.toml               # Air配置文件
├── Makefile                # Make命令
└── README.md               # 项目说明
```

## 开发注意事项

1. **数据库文件**: SQLite 数据库文件 (`*.db`) 会自动生成，不需要手动创建
2. **热重载**: 修改 `internal/`, `cmd/`, `configs/` 目录下的 Go 文件会自动触发重新构建
3. **端口**: 默认服务端口为 8080，可通过环境变量 `PORT` 修改
4. **日志**: 构建错误日志会保存在 `build-errors.log` 文件中

## 环境变量

```bash
export PORT=8080                                    # 服务端口
export DB_PATH=minidevops.db                        # 数据库文件路径
export JWT_SECRET=minidevops-secret-key-2024        # JWT密钥
export MOUNT_PATH=~/dev-ops                         # 代码挂载路径
```

## 默认管理员账号

- 用户名: `admin`
- 密码: `admin123`

## API 测试

服务启动后，可以通过以下方式测试 API：

```bash
# 登录
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'

# 获取用户信息（需要先登录获取token）
curl -X GET http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 故障排除

1. **Air 无法启动**: 检查是否正确安装了 Air 工具
2. **构建失败**: 查看 `build-errors.log` 文件
3. **端口占用**: 修改 `.air.toml` 中的 `app_port` 配置
4. **数据库错误**: 删除 `*.db` 文件重新启动
