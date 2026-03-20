# 服务器密钥可选与仪表盘环境列

> 状态: 已执行

## 目标

1. 服务器 SSH 私钥字段允许留空：目标机已配置 `authorized_keys` 时，依赖本机默认密钥或 SSH Agent 即可连接。
2. 仪表盘「构建列表」增加「环境名称」列，便于区分同一项目下不同环境的构建。

## 内容

1. **后端**：放宽服务器创建/更新时对 SSH 私钥的校验；在 `internal/deployer/ssh.go`（或实际建立 SSH 的代码路径）中，当私钥为空时跳过密钥解析，仅使用密码、SSH Agent 等已有逻辑。
2. **前端**：服务器表单将私钥输入标为可选，并调整占位/校验提示。
3. **仪表盘**：若构建列表 API 未返回环境名，在后端列表/详情 DTO 中补充 `environment_name`（或关联查询）；前端 `dashboard` 构建表格增加环境名称列。

## 影响范围

- `internal/deployer/ssh.go`：新增 `SSHAuthMethods`、空私钥时通过 `SSH_AUTH_SOCK` 使用 SSH Agent。
- `internal/service/server_service.go`：`key` 认证不再强制私钥；连接测试复用 `deployer.SSHAuthMethods`。
- `web/src/pages/servers/form.tsx`：私钥可选与文案。
- `web/src/pages/dashboard.tsx`：最近构建表与运行中卡片展示环境名（API 已有 `environment_name`）。

## 历史补丁
