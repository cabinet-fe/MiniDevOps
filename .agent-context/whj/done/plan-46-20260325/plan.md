# 环境表单分发方式与项目详情切换

> 状态: 已执行

## 目标

1. 编辑环境时，选择分发方式后清空已选服务器，且服务器下拉仅展示支持该分发方式的服务器。
2. 项目详情页支持在不返回列表的情况下快速切换到其他项目。

## 内容

1. **environment-form.tsx**：梳理分发方式（部署方式）与服务器字段的联动；在分发方式变更时重置服务器相关 state（含多分发目标时的各条）；根据当前选中的分发方式过滤 `servers` 列表（与后端 `Server` 支持的部署方式字段对齐）。
2. **detail.tsx**：在页面头部或侧栏附近增加项目切换入口（如 Combobox/Select 拉取项目列表并 `navigate` 到对应 `/projects/:id`），保持与现有 UI 风格一致。
3. 运行 `cd web && bun run lint` 与必要的类型检查，修复新增问题。

## 影响范围

- `web/src/pages/projects/environment-form.tsx`：分发方式变更时清空该条 `server_id`；按 `auth_type`/`agent_url` 过滤可选服务器；加载服务器列表后自动清除与当前分发方式不兼容的已选服务器。
- `web/src/pages/projects/detail.tsx`：项目标题旁增加可搜索的 Popover 项目列表，用于快速 `navigate` 到其它项目详情。

## 历史补丁
