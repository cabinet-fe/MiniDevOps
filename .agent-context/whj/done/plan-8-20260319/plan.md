# 构建缓存与工作空间优化

> 状态: 已执行

## 目标

引入构建缓存机制，避免每次构建重复下载依赖（如 node_modules、Go modules 等），显著缩短构建时间。同时优化工作空间管理，支持查看和清理工作空间。

## 内容

### 步骤 1：构建缓存机制设计与实现

- 新增配置 `build.cache_dir` 指定缓存根目录。
- `Environment` 模型新增 `CachePaths string` 字段（JSON 数组格式，如 `["node_modules", ".npm"]`）。
- Pipeline 构建前：将缓存目录中该环境对应的缓存恢复到工作目录。
- Pipeline 构建后：将指定路径的目录同步回缓存目录。
- 缓存按 `project-{pid}/env-{eid}/` 路径隔离。

### 步骤 2：前端环境配置增加缓存路径

- 环境表单新增"构建缓存路径"配置项（多行文本，每行一个路径）。
- 项目详情页显示环境缓存配置状态。

### 步骤 3：工作空间管理

- 新增 `GET /api/v1/system/workspaces` 接口，返回各项目工作空间的磁盘占用。
- 新增 `DELETE /api/v1/system/workspaces/:projectId` 接口，清理指定项目的工作空间。
- 新增 `DELETE /api/v1/system/caches/:projectId` 接口，清理指定项目的构建缓存。
- 系统设置页增加"存储管理"区域，展示工作空间和缓存的磁盘使用情况，支持清理操作。

### 步骤 4：验证

- 首次构建正常（无缓存），二次构建依赖不重复安装。
- 清理缓存/工作空间后再次构建正常。

## 影响范围

- `config.yaml` — 新增 `build.cache_dir` 配置项
- `internal/config/config.go` — BuildConfig 新增 CacheDir 字段及路径解析
- `internal/model/environment.go` — Environment 新增 CachePaths 字段
- `internal/engine/pipeline.go` — 新增缓存恢复/保存逻辑 + parseCachePaths/copyDir/copyFile 辅助函数
- `internal/handler/project_handler.go` — 环境创建/更新请求结构新增 CachePaths 字段
- `internal/handler/system_handler.go` — 新增 ListWorkspaces/CleanWorkspace/CleanCache 接口
- `internal/service/project_service.go` — EnvironmentExport 新增 CachePaths 字段
- `cmd/server/main.go` — Pipeline 构造传递 cacheDir、创建缓存目录、注册新路由
- `web/src/pages/projects/environment-form.tsx` — 新增「构建缓存路径」配置项
- `web/src/pages/settings.tsx` — 新增「存储管理」区域

## 历史补丁
