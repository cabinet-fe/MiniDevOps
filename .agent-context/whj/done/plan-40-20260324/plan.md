# 环境部署方式：本机复制（local）

> 状态: 已执行

## 目标

为环境新增部署方式 **本机部署**：在 BuildFlow 服务所在机器上，将构建阶段的产物目录（与现有流水线中 `deployFromSource` 的 `sourceDir` 一致，即工作区内 `BuildOutputDir` 对应目录；仅部署流程则为解压后的临时目录内容）**递归复制**到配置的本机目标路径 `deploy_path`，不经过 SSH/Agent。用户选择该方式时无需绑定远程服务器。

## 内容

1. **后端 Deployer**
   - 在 `internal/deployer/` 新增本机部署实现（如 `local.go`），`Deploy` 将 `DeployOptions.SourceDir` 下文件递归复制到 `DeployOptions.RemotePath`（此处复用字段表示本机绝对路径）。
   - 复制策略：创建目标目录（若不存在）；对源树内每个文件复制到目标相对路径，**同名覆盖**；源树中不存在的目标残留文件是否删除——实现时二选一并在代码注释中写明（建议默认**不删除**残留，避免误删；若与 rsync 默认行为差异大，可在计划实施时改为「先清空目标再复制」并记入影响说明）。
   - `NewDeployer` 增加 `case "local"` 返回该实现；**禁止**未知方法静默回退 rsync（可顺带将 `default` 改为明确错误或保留现状并在协议中说明——若改动面大则本计划仅加 `local` 分支）。

2. **流水线与校验**（`internal/engine/pipeline.go`）
   - 扩展 `deployFromSource`：当 `env.DeployMethod == "local"` 时，**不要求** `DeployServerID`；仅要求 `DeployPath` 非空且为**本机可用**路径（建议校验 `filepath.IsAbs`，非法则失败并日志说明）。
   - 非 local 时保持现有「必须绑定服务器 + 路径」逻辑。
   - **部署后脚本**：`local` 模式下若配置了 `PostDeployScript`，应在**本机**、以 `DeployPath` 为工作目录执行（shell），与 SSH 远程执行语义对齐；需新增本地执行辅助函数或集中在 deployer 包，避免复用 `ExecuteRemoteScriptInDir` 误连远程。

3. **API / 模型**
   - `Environment` 已有 `deploy_method` 字符串字段，无需迁移；在 handler 或服务层校验：`local` 时允许 `deploy_server_id` 为空；非 `local` 仍要求服务器与路径。
   - 检查创建/更新环境的校验逻辑（`project_handler` / `project_service`），保证前后端一致。

4. **前端**
   - `web/src/lib/constants.ts` 的 `DEPLOY_METHODS` 增加一项，如 `{ value: "local", label: "本机部署" }`。
   - `environment-form.tsx`：选择 `local` 时隐藏或禁用部署服务器选择，并校验仅填写部署路径；编辑回填时兼容旧数据（服务器 ID 可空）。
   - 项目详情等展示部署方式的页面若有硬编码映射，一并补充 `local` 的展示文案。

5. **文档与验证**
   - 更新仓库根目录 `AGENTS.md` 中部署器表格，增加 `local` 行说明。
   - `go test ./...`；对 `LocalDeployer` 可用临时目录编写简短单测（复制、覆盖、目标不存在等）。
   - 前端 `cd web && bun run lint`（如有 touched 文件）。

## 影响范围

- `internal/deployer/local.go`（新建）、`internal/deployer/local_test.go`（新建）、`internal/deployer/deployer.go`
- `internal/engine/pipeline.go`
- `internal/service/project_service.go`
- `web/src/lib/constants.ts`、`web/src/pages/projects/environment-form.tsx`、`web/src/pages/projects/detail.tsx`
- `AGENTS.md`

## 历史补丁
