# 构建与分发阶段分离（多分发目标）

> 状态: 已执行

## 目标

将流水线拆为两个独立阶段：**构建**（含拉取、执行构建脚本、打包产物）与**分发**（原「部署」）。构建脚本与产物归档成功后，该构建实例即视为**构建成功**，`Build` 状态为成功且可下载产物；分发失败**不再**将整条构建记为失败。一个环境支持配置**多条**分发（多服务器/多路径/多方式等），无需兼容旧版环境上的单路 `deploy_*` 字段。完整主流程为：等待 → 拉取代码 → 执行构建（构建结束即可视为可复用产物，不必因分发失败而重跑构建脚本）→ 若配置了分发则执行分发（可多条）。支持在构建详情中对已成功构建**重新分发**（不重新跑构建）。

## 内容

### 1. 领域模型与存储（破坏性变更，不做旧数据迁移）

- 从 `Environment` 移除单路部署字段：`deploy_server_id`、`deploy_path`、`deploy_method`、`post_deploy_script`（以实际代码为准，名称若有出入以仓库为准）。
- 新增 **分发配置** 表（名称待定，如 `distributions`）：归属 `environment_id`，字段至少包含：关联服务器、远程路径、部署方式、部署后脚本（可选）、排序等；GORM `AutoMigrate`，在 `database.go` 注册。
- 新增 **构建-分发执行记录** 表（如 `build_distributions`）：`build_id`、指向分发配置 `distribution_id`（或快照字段，若希望配置删除后历史仍可读则存快照）、`status`（等待/进行中/成功/失败/跳过/取消等）、`error_message`、`started_at`、`finished_at`，必要时支持**同一构建对同一条分发的重试**（多行按 `attempt` 或只保留最新一次，实现时选一种并写清语义）。
- 更新 `Build`：明确「构建成功」与「分发汇总」的展示字段（若用 `current_stage` 表达分发汇总，需与 `BUILD_STATUSES` 前端枚举对齐；或增加独立字段/关联查询，避免 `success` 与分发失败在 UI 上矛盾）。

### 2. 流水线 `internal/engine/pipeline.go`（及拆出的分发逻辑）

- **主路径（非仅分发触发）**：在产物归档成功并写入 `artifact_path` 后，将 `Build` 更新为**构建阶段成功**（`status=success`，`duration`/`finished_at` 语义以「构建结束」为准，或与现有一致但文档写清）；**不得**因后续分发失败而调用 `failBuild`。
- **分发阶段**：在构建已成功的前提下，按环境的多条分发配置顺序或并行执行（实现时定序；日志写入同一 build 日志或分子段，需可读）。单条分发失败：只更新对应 `build_distributions` 行，可选更新构建的「分发汇总」展示字段；**不**改变 `Build.Status` 为 `failed`。
- **取消**：取消构建时，构建阶段与进行中的分发均应能中止；语义写清（如分发中取消记为 cancelled）。
- **仅分发 / 重新分发**：沿用或调整现有「仅部署」入口（当前为 `TriggerType == "deploy"` + `executeDeployOnly`）：应对**已有成功构建 + 已有产物**触发，仅执行分发（不写失败到 `Build`）。支持对**同一次成功构建**多次「重新分发」（API 与 scheduler 行为一致）。
- 复用现有 `deployer` 包；将原 `deployFromSource` 从「单环境单目标」改为「对单条分发配置调用」。

### 3. Service / Repository / Handler / API

- 环境 CRUD：增删改查「分发配置」列表（嵌套在环境保存接口或独立子资源，二选一并保持 REST 清晰）。
- 构建列表/详情：返回每条分发的状态与错误信息；下载产物接口保持以「构建成功」为准。
- 新增或调整：**触发重新分发**（参数：构建 ID，可选分发 ID 列表；未指定则全部重试）。

### 4. 定时任务与 Webhook

- 检查 `cron` / webhook 触发的构建是否仍按环境拉取配置；改为读取多分发配置，无分发则构建成功后跳过分发阶段。

### 5. 前端 `web/`

- `constants.ts`：构建阶段/状态与分发状态枚举、文案、颜色。
- 环境表单：多行「分发」编辑（增删排序），对接新 API。
- 构建详情：展示构建成功与**各分发目标**状态；**重新部署**按钮（调用重新分发 API）。
- 列表与仪表盘：若依赖「整单成功/失败」，改为区分「构建成功」与「分发是否全部成功」（按产品设计：可仅图标提示，不把整单标红为失败）。

### 6. 测试与文档

- `go test ./internal/engine/...` 与相关 service 测试：覆盖「构建成功 + 分发失败 → Build 仍为 success」、多分发、仅分发触发。
- 实施后更新 `AGENTS.md` 中流水线与部署器章节（若本次变更改变了对外行为描述）。

## 影响范围

- `internal/model/`：`environment.go`、`build.go`、新增 `distribution.go`、`build_distribution.go`；`database.go` 迁移与丢弃旧 `deploy_*` 列
- `internal/repository/`：`distribution_repo.go`、`build_distribution_repo.go`，`build_repo.go`、`environment_repo.go`、`server_repo.go` 调整
- `internal/engine/`：`pipeline.go`、`pipeline_distribute.go`、`scheduler.go`
- `internal/service/`：`project_service.go`、`build_service.go`、`server_service.go` 及测试
- `internal/handler/`：`project_handler.go`、`build_handler.go`
- `cmd/server/main.go`：DI
- `web/src/`：`lib/constants.ts`、`components/build-log-viewer.tsx`、`environment-builds-table.tsx`、`pages/projects/environment-form.tsx`、`pages/projects/detail.tsx`、`pages/builds/detail.tsx`、`pages/dashboard.tsx`
- `AGENTS.md`：流水线与部署器说明

## 历史补丁

