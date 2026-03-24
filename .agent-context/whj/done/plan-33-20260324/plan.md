# 环境列表与移除高级触发

> 状态: 已执行

## 目标

从项目详情页移除「高级触发」功能；新增跨项目环境列表页，支持按项目筛选、按环境名过滤，并提供快捷触发构建（默认分支），不提供新增环境入口。

## 内容

1. 修改 `web/src/pages/projects/detail.tsx`：删除「高级触发」按钮、`TriggerBuildDialog` 组件及相关 state、仅被该对话框使用的 import。
2. 新增 `web/src/pages/environments/list.tsx`：分页拉取全部项目（含 environments），扁平化为行；项目选择器（含「全部项目」）、环境名称输入过滤；表格展示项目名（链到详情）、环境名、分支、触发构建按钮；无新建按钮。
3. 在 `web/src/App.tsx` 注册路由 `/environments`；在 `web/src/components/layout/sidebar.tsx` 与 `header.tsx` 增加导航与面包屑元数据。

## 影响范围

- `web/src/pages/projects/detail.tsx` — 移除高级触发按钮与 `TriggerBuildDialog`
- `web/src/pages/environments/list.tsx` — 新增跨项目环境列表页
- `web/src/App.tsx` — 路由 `/environments`
- `web/src/components/layout/sidebar.tsx` — 侧栏「环境」入口
- `web/src/components/layout/header.tsx` — 面包屑元数据

## 历史补丁
