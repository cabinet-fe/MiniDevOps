# 环境列表列顺序、构建分页与查看构建弹窗

> 状态: 已执行

## 目标

1. 项目详情「环境与构建」中每个环境的构建记录支持分页（与 API 一致，展示分页器）。
2. 全局「环境」页表格将「环境」列置于第一列；增加「查看构建」操作，弹窗内展示与项目详情该环境 Tab 下一致的构建表格与分页。
3. 抽取可复用的构建表格片段，避免两处逻辑分叉。

## 内容

1. 新增 `web/src/components/environment-builds-table.tsx`：从 `detail.tsx` 抽出构建表格行、状态徽章、操作列及底部分页（样式与审计日志分页一致）；对外接收 `projectId`、环境信息、`builds`、分页元数据、`onPageChange`、`onBuildAction` 等。
2. 调整 `web/src/pages/projects/detail.tsx`：拆分「仅拉取项目」与「按环境分页拉取构建」；用 `useEffect` 在切换 Tab 或翻页时请求对应页；轮询时刷新当前 Tab 当前页构建；将原 `BuildRow` 替换为上述组件。
3. 调整 `web/src/pages/environments/list.tsx`：表头与列顺序为「环境、项目、分支、操作」；操作区增加「查看构建」打开 `Dialog`，内嵌同一构建表格组件并绑定分页与 API。

## 影响范围

- `web/src/components/environment-builds-table.tsx`（新建）
- `web/src/pages/projects/detail.tsx`
- `web/src/pages/environments/list.tsx`
- `internal/repository/environment_repo.go`
- `internal/service/project_service.go`
- `internal/handler/project_handler.go`
- `cmd/server/main.go`

## 历史补丁

- patch-1: 环境列表分页 API 与构建弹窗样式
