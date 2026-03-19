# 环境配置改进：分支选择、脚本类型、工作空间隔离

## 补丁内容

### 1. 分支做成可选的，自动获取分支列表

- 后端新增 `GitListBranches` 函数（`internal/engine/git.go`），通过 `git ls-remote --heads` 获取远程仓库分支列表
- 后端新增 `GET /api/v1/projects/:id/branches` API（`internal/handler/project_handler.go`）
- 前端环境表单中分支字段从输入框改为 Combobox（下拉选择 + 自由输入），打开表单时自动加载分支列表
- 前端触发构建对话框的分支字段同样改为 Combobox
- 新增 shadcn Command 组件（`web/src/components/ui/command.tsx`），基于 cmdk 库
- 分支字段改为可选，不填则使用空字符串（后端默认 main）

### 2. 构建脚本支持选择类型

- Environment 模型新增 `BuildScriptType` 字段（`internal/model/environment.go`），支持 bash/node/python
- Pipeline 执行时根据脚本类型选择对应解释器：bash → `sh -c`，node → `node -e`，python → `python3 -c`
- 前端环境表单新增脚本类型下拉选择器，根据类型动态切换占位符提示
- 项目详情页环境信息卡片展示脚本类型 Badge
- 导入导出功能同步支持 `build_script_type` 字段

### 3. 环境代码仓库隔离

- 工作目录从 `project-{pid}` 改为 `project-{pid}/env-{eid}`
- 不同环境各自有独立的代码仓库副本，避免分支切换冲突

## 影响范围

- 修改文件: `internal/model/environment.go` — Environment 新增 BuildScriptType 字段
- 修改文件: `internal/engine/git.go` — 新增 GitListBranches 函数
- 修改文件: `internal/engine/pipeline.go` — 工作目录按环境隔离 + 按脚本类型选择解释器
- 修改文件: `internal/handler/project_handler.go` — 环境 CRUD 支持 build_script_type + 新增 ListBranches
- 修改文件: `internal/service/project_service.go` — EnvironmentExport 增加 build_script_type
- 修改文件: `cmd/server/main.go` — 注册 branches 路由
- 新增文件: `web/src/components/ui/command.tsx` — shadcn Command 组件
- 修改文件: `web/src/lib/constants.ts` — 新增 BUILD_SCRIPT_TYPES 常量
- 修改文件: `web/src/pages/projects/environment-form.tsx` — 分支 Combobox + 脚本类型选择
- 修改文件: `web/src/pages/projects/detail.tsx` — 触发构建分支 Combobox + 脚本类型展示
- 修改文件: `web/package.json` — 新增 cmdk 依赖
