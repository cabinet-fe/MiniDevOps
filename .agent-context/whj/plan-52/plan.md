# 智能体与 CLI 代理管理（含构建接入）

> 状态: 已执行

## 目标

1. 侧边栏新增「智能体」分组：智能体管理、代理管理。
2. 智能体可配置自定义提示词、绑定一种 CLI 代理、勾选多个项目范围。
3. 代理管理覆盖 opencode / claude（Claude Code）/ reasonix：检测是否安装、版本、安装与更新。
4. 环境级挂载智能体；构建脚本成功后按顺序执行，日志写入构建输出；失败不回滚构建成功（对齐分发失败语义）。

## 内容

### 1. 数据模型

- 新建 `internal/model/agent.go`：
  - `Agent`：`name`、`prompt`（text）、`proxy_key`（`opencode`|`claude`|`reasonix`）、`enabled`
  - `AgentProject`：`agent_id` + `project_id` 联合唯一
  - `EnvironmentAgent`：`environment_id` + `agent_id` + `sort_order`
- `Environment` 增加 `AgentIDs []uint` json `agent_ids` gorm:"-"
- `database.go` AutoMigrate 注册上述表

### 2. Agent CRUD API

- repository / service / handler：`agent_repo.go`、`agent_service.go`、`agent_handler.go`
- 路由：`GET/POST /api/v1/agents`，`GET/PUT/DELETE /api/v1/agents/:id`
- Body：`{ name, prompt, proxy_key, enabled, project_ids: number[] }`
- 写操作 `ops`/`admin`，登录可读
- 审计：`agents` → `agent`

### 3. AgentProxy 服务（无 DB）

- `agent_proxy_service.go` + `agent_proxy_handler.go`
- 内置目录 opencode / claude / reasonix：LookPath、`--version`、install、upgrade
- `GET /api/v1/agent-proxies`，`POST .../:key/install`，`POST .../:key/upgrade`
- 审计：`agent-proxies` → `agent_proxy`

### 4. 环境挂载 agent_ids

- 扩展环境 Create/Update/Get，同步 `EnvironmentAgent`（模式同 `var_group_ids`）
- 前端 `environment-form`：构建后智能体多选（仅展示已勾选当前项目的智能体）

### 5. Pipeline 接入

- 新建 `pipeline_agent.go`：`markBuildArtifactSuccess` 之后、分发之前执行挂载智能体
- 失败写 ERROR 日志并继续，不改 Build.Status；`redistribute` 跳过
- `NewPipeline` 注入 agent repository

### 6. 前端

- 侧边栏「智能体」分组：`/agents`、`/agent-proxies`
- App 路由 + header 面包屑
- `pages/agents/list.tsx`、`pages/agent-proxies/list.tsx`
- `constants.ts`：`AGENT_PROXIES`

### 7. 测试

- Agent 校验/范围与 proxy 探测单测；运行相关 `go test`

## 影响范围

- internal/model/agent.go（新建）
- internal/model/environment.go
- internal/model/database.go
- internal/repository/agent_repo.go（新建）
- internal/service/agent_service.go（新建）
- internal/service/agent_service_test.go（新建）
- internal/service/agent_proxy_service.go（新建）
- internal/service/agent_proxy_service_test.go（新建）
- internal/service/project_service.go
- internal/handler/agent_handler.go（新建）
- internal/handler/agent_proxy_handler.go（新建）
- internal/handler/project_handler.go
- internal/engine/pipeline.go
- internal/engine/pipeline_agent.go（新建）
- internal/middleware/audit.go
- cmd/server/main.go
- web/src/lib/constants.ts
- web/src/components/layout/sidebar.tsx
- web/src/components/layout/header.tsx
- web/src/App.tsx
- web/src/pages/agents/list.tsx（新建）
- web/src/pages/agent-proxies/list.tsx（新建）
- web/src/pages/projects/environment-form.tsx
- web/src/pages/projects/detail.tsx

## 历史补丁
