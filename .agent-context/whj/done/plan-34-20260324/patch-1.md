# 环境列表分页 API 与构建弹窗样式

## 补丁内容

1. **构建弹窗样式**：环境页「查看构建」原先在 `DialogContent` 上叠加 `max-w-3xl` 与 `overflow-y-auto`，与默认的 `flex flex-col overflow-hidden` 布局冲突，且未使用 `DialogBody`/`DialogDescription`，与设置页、变量组等弹窗不一致。已改为 `sm:max-w-3xl` + `DialogHeader`（含 `DialogDescription`）+ `DialogBody` 承载 `EnvironmentBuildsTable`，与同仓库其它对话框一致。

2. **独立环境分页接口**：新增 `GET /api/v1/environments`，服务端关联 `projects` 表返回 `project_name`，支持 `page`、`page_size`、`project_id`、`name`（环境名模糊匹配），`dev` 角色仅能看到本人创建项目下的环境（与项目列表一致）。`internal/repository/environment_repo.go` 增加 `ListJoined`，`ProjectService.ListEnvironmentsGlobal` 与 `ProjectHandler.ListEnvironmentsGlobal` 暴露路由。

3. **环境列表页**：主表数据改为调用 `/environments` 分页；项目下拉仍通过分页拉取 `/projects`（仅用于筛选选项）。名称过滤 300ms debounce，并在名称变化时重置到第 1 页以避免竞态。主表底部分页与构建表格分页控件风格一致。

## 影响范围

- 新增文件: 无
- 修改文件: `internal/repository/environment_repo.go`, `internal/service/project_service.go`, `internal/handler/project_handler.go`, `cmd/server/main.go`, `web/src/pages/environments/list.tsx`
