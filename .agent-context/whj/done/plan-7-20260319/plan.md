# 定时构建（Cron）与构建触发增强

> 状态: 已执行

## 目标

为环境增加 Cron 定时构建支持，用户可为每个环境配置 Cron 表达式实现定时自动构建。同时增强构建触发的灵活性，支持指定分支/Commit 构建。

## 内容

### 步骤 1：后端 Cron 模型与调度器

- `Environment` 模型新增 `CronExpression string` 和 `CronEnabled bool` 字段。
- 新建 `internal/engine/cron.go`，基于 `robfig/cron/v3` 实现定时调度器：
  - 启动时从数据库加载所有启用 Cron 的环境。
  - 每个环境注册 Cron 任务，触发时创建 Build 记录（trigger_type="cron"）并提交 scheduler。
  - 提供 Add/Remove/Update 方法动态管理。
- 在 `main.go` 中初始化并启动 CronScheduler。

### 步骤 2：环境 API 支持 Cron 字段

- 修改环境 CRUD 的请求/响应体，新增 `cron_expression` 和 `cron_enabled` 字段。
- 创建/更新环境时验证 Cron 表达式合法性。
- 更新环境后通知 CronScheduler 动态刷新。

### 步骤 3：前端环境配置增加 Cron UI

- 环境表单新增"定时构建"区域：Cron 开关、Cron 表达式输入、表达式可读说明（如 "每天 02:00"）。
- 项目详情页环境信息卡片展示下次执行时间。

### 步骤 4：构建触发增强 — 指定分支/Commit

- `POST /projects/:id/builds` 请求体增加可选字段 `branch` 和 `commit_hash`。
- Pipeline 执行时如指定了 branch 则覆盖环境默认分支，如指定了 commit 则 checkout 到指定 commit。
- 前端触发构建对话框增加可选的分支和 commit 输入。

### 步骤 5：验证

- Cron 表达式合法性校验。
- Cron 触发的构建可正常执行和查看。
- 指定分支/Commit 构建结果正确。

## 影响范围

- `go.mod` / `go.sum` — 新增 `github.com/robfig/cron/v3` 依赖
- `internal/model/environment.go` — Environment 模型新增 CronExpression、CronEnabled、BuildScriptType 字段
- `internal/model/build.go` — Build 模型新增 Branch 字段
- `internal/engine/cron.go` — **新建**，CronScheduler 定时构建调度器
- `internal/engine/git.go` — 新增 GitListBranches 函数
- `internal/engine/pipeline.go` — 支持 build-level branch 覆盖和 commit checkout、工作目录按环境隔离、按脚本类型选择解释器
- `internal/repository/environment_repo.go` — 新增 ListCronEnabled 方法
- `internal/service/build_service.go` — TriggerBuild 增加 branch 参数
- `internal/service/project_service.go` — EnvironmentExport 增加 cron 和 build_script_type 字段
- `internal/handler/project_handler.go` — 环境 CRUD 支持 cron、build_script_type 字段、CronNotifier 接口、cron 表达式校验、新增 ListBranches
- `internal/handler/build_handler.go` — TriggerBuild/Retry 支持 branch/commit_hash
- `internal/handler/webhook_handler.go` — 适配 TriggerBuild 新签名
- `cmd/server/main.go` — 初始化 CronScheduler、传递给 ProjectHandler、注册 branches 路由、优雅关闭
- `web/src/components/ui/command.tsx` — **新建**，shadcn Command 组件
- `web/src/lib/constants.ts` — 新增 BUILD_SCRIPT_TYPES 常量
- `web/src/pages/projects/detail.tsx` — 环境信息卡片展示 Cron/脚本类型、高级触发对话框分支 Combobox、环境编辑按钮
- `web/src/pages/projects/environment-form.tsx` — **新建**，环境表单对话框含 Cron UI、分支 Combobox、脚本类型选择、宽弹框适配和 CodeMirror 代码高亮
- `web/src/pages/builds/detail.tsx` — 构建详情展示分支字段
- `web/package.json` — 新增 cmdk, @uiw/react-codemirror 等 UI 及代码高亮依赖
## 历史补丁

- patch-1: 环境配置改进：分支选择、脚本类型、工作空间隔离
- patch-2: 优化环境弹框宽度与代码输入高亮
