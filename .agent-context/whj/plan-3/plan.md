# 开发体验与安全改进

> 状态: 未执行

## 目标

添加一键启动前后端开发服务的命令，并将数据库初始化中硬编码的管理员账号密码改为从配置文件读取，消除凭据泄露风险。

## 内容

### 步骤 1：Makefile 添加 `make dev` 一键启动

- 文件：`Makefile`
- 添加 `dev` 目标，并行启动前端 Vite dev server 和后端 Go 服务
- 使用后台进程 + `trap` 信号捕获实现：两个子进程同时启动，任一退出或收到中断信号时同时终止
- 终端输出混合显示前后端日志，方便开发调试

### 步骤 2：配置文件增加初始管理员配置项

- 文件：`config.yaml`、`internal/config/config.go`
- 在 `config.yaml` 中新增 `admin` 配置块：
  ```yaml
  admin:
    username: "admin"
    password: "admin123"
    display_name: "Administrator"
  ```
- 在 `config.go` 的配置结构体中添加对应字段，设定合理默认值
- 配置文件中的密码仅作为首次初始化种子，日志中不得明文输出

### 步骤 3：改造数据库初始化逻辑

- 文件：`internal/model/database.go`
- 将硬编码的 `"admin"` / `"admin123"` 替换为从 `config.C.Admin` 读取
- 保留"仅在用户表为空时创建"的逻辑不变
- 确保配置项为空时给出明确错误提示而非静默创建

### 步骤 4：验证

- `make dev` 可同时启动前后端，Ctrl+C 可正常终止两个进程
- 修改 `config.yaml` 中 admin 密码后，删除 `data/` 目录重新启动，可用新密码登录
- 后端编译通过：`go build ./...`

## 影响范围

## 历史补丁
