# Windows 部署支持

> 状态: 已执行

## 目标

扩展部署能力，支持将构建产物推送到 Windows 远程服务器。采用双方案架构：优先使用 SSH（Windows OpenSSH），同时提供轻量级 Agent 方案作为补充。

## 内容

### 步骤 1：Server 模型扩展

- `Server` 模型新增 `OSType string` 字段（`linux`/`windows`，默认 `linux`）。
- 前端服务器表单增加"操作系统"选择（Linux / Windows）。
- 根据操作系统类型在部署时选择合适的路径分隔符和命令。

### 步骤 2：SSH for Windows 部署器

- 修改 `deployer/ssh.go`，检测目标 OS 类型：
  - Linux：现有 `sh -c` 方式不变。
  - Windows：远程命令使用 `cmd /c` 或 `powershell -Command`，路径使用 `\` 分隔符。
- 修改 `deployer/sftp.go`，Windows 路径兼容处理。
- 修改 `deployer/scp.go`，同上。
- `rsync` 在 Windows 上通常不可用，当 OS 为 windows 且 method 为 rsync 时，自动降级为 sftp 并记录日志。

### 步骤 3：轻量级分发 Agent 方案

为不方便安装 OpenSSH 的 Windows 服务器（或需要更灵活管控的场景），构建一个轻量级 Agent：

- 新建 `cmd/agent/main.go`：
  - 监听 HTTP(S) 端口接收文件推送请求。
  - 接口：`POST /upload` 接收 tar.gz 文件，解压到目标目录。
  - 接口：`POST /exec` 执行部署后脚本。
  - 使用预共享密钥认证（Bearer Token）。
- `Server` 模型新增 `AgentURL string` 和 `AgentToken string` 字段。
- 新建 `deployer/agent.go`，实现 `Deployer` 接口：
  - 将产物 tar.gz 通过 HTTP POST 发送给 Agent。
  - 调用 Agent 的 `/exec` 接口执行后续脚本。
- 前端：服务器的认证方式增加 `agent` 选项，配置 Agent URL 和 Token。
- `DEPLOY_METHODS` 常量增加 `agent` 选项。

### 步骤 4：连接测试适配

- `TestConnection` 接口根据 OS 类型和认证方式区分测试逻辑：
  - SSH (Linux/Windows)：现有 SSH 测试。
  - Agent：HTTP GET Agent 健康检查端点。

### 步骤 5：验证

- Linux 服务器部署不受影响。
- Windows 服务器通过 SSH 或 Agent 方式均可成功部署。
- 部署后脚本在 Windows 上正确执行（PowerShell/cmd）。

## 影响范围

- `cmd/agent/main.go`
- `internal/model/server.go`
- `internal/service/server_service.go`
- `internal/handler/server_handler.go`
- `internal/deployer/deployer.go`
- `internal/deployer/path.go`
- `internal/deployer/ssh.go`
- `internal/deployer/sftp.go`
- `internal/deployer/scp.go`
- `internal/deployer/rsync.go`
- `internal/deployer/agent.go`
- `internal/engine/pipeline.go`
- `web/src/lib/constants.ts`
- `web/src/pages/servers/form.tsx`
- `web/src/pages/projects/environment-form.tsx`

## 历史补丁
