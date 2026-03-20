# 项目标签展示字典显示文本

> 状态: 已执行

## 目标

在 `web/src/pages/projects/` 下，项目标签在列表、详情等展示处使用数据字典 `project_tags` 的 **label（显示文本）**，而非仅展示逗号分隔的 **value（存储值）**；与表单页已存在的 `label || value` 行为一致。

## 内容

1. **`list.tsx`**：加载的 `dictTags` 已存在；为表格列、卡片 Badge、顶部筛选按钮增加 `value → label` 映射展示；筛选与存储仍以 value 为准。搜索关键词可同时匹配 label 与 value（字典已加载时）。
2. **`detail.tsx`**：请求 `/dictionaries/code/project_tags/items`，在标题行标签 Badge 上用字典 label 展示。
3. 运行 `cd web && bun run lint` 与 `bun run build` 验证前端。

## 影响范围

- `web/src/pages/projects/list.tsx`：标签展示与筛选按钮使用字典 label；搜索可匹配 label。
- `web/src/pages/projects/detail.tsx`：项目标题旁标签 Badge 使用字典 label。

## 历史补丁
