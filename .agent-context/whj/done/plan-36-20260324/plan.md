# 部署仅发产物、Agent 鉴权与 Agent 配置文件

> 状态: 已执行

## 目标

1. 「重新部署」仅将已有构建产物发布到目标环境，不再执行克隆与构建。
2. 修复 Agent 部署时 `Authorization` 使用数据库中密文导致远端返回 `unauthorized` 的问题（与连接测试解密逻辑一致）。
3. `buildflow-agent` 支持同目录默认配置文件（与可执行文件同级），并与环境变量、命令行参数合理叠加。

## 内容

1. **BuildService**：新增 `TriggerDeployBuild(sourceBuildID, userID)`：校验源构建成功且存在产物路径；新建一条 `trigger_type = deploy` 的构建记录并分配新编号，复用分支/提交信息与产物路径。
2. **BuildHandler.Deploy**：改为调用上述方法并 `Submit` 新构建 ID；响应返回新构建。
3. **Pipeline.Execute**：若 `trigger_type == "deploy"`，走 `executeDeployOnly`：写日志、将产物归档解压到临时目录、解密 `AgentToken`（及现有密码/密钥逻辑保持一致）、调用既有部署与部署后脚本、成功/失败收尾。
4. **pipeline**：实现 `extractArtifactArchive`（zip / tar.gz）供仅部署路径使用；产物路径解析与下载接口一致。
5. **cmd/agent**：支持 `-config`；未指定时尝试 `<可执行文件目录>/buildflow-agent.yaml`；YAML 提供 `addr`、`token`、`tls_cert`、`tls_key`；优先级：配置文件 < 环境变量 < 命令行（Parse 后非默认值覆盖）。
6. **验证**：`go test ./...`，`cd web && bun run lint`（如有前端文案常量则顺带）。

## 影响范围

- `internal/service/build_service.go`（`TriggerDeployBuild`、日志阶段推断）
- `internal/service/build_service_test.go`
- `internal/handler/build_handler.go`
- `internal/engine/pipeline.go`（仅部署路径、`deployFromSource`、产物解压、AgentToken 解密）
- `cmd/agent/main.go`（默认 `buildflow-agent.yaml`、`-config`、与 env/flag 合并）
- `web/src/components/environment-builds-table.tsx`
- `web/src/pages/builds/detail.tsx`
- `go.mod` / `go.sum`（`gopkg.in/yaml.v3`）
- `README.md`（Agent 部署与对接说明）
- `web/src/pages/project-manual.tsx`（独立 Agent 章节）

## 历史补丁

- patch-1: 文档：部署 Agent 使用说明
