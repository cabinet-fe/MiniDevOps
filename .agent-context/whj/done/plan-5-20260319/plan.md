# 前端 UI 重构：表单弹框化与布局优化

> 状态: 已执行

## 目标

将项目表单和服务器表单从独立页面迁移为弹框（Dialog）形式，统一交互体验；优化前端整体布局，对标优秀 shadcn 开源项目的设计水准。

## 内容

### 步骤 1：项目表单迁移到弹框

- 文件：`web/src/pages/projects/form.tsx` → 重构为 Dialog 组件
- 在项目列表页（`web/src/pages/projects/list.tsx`）中集成新建/编辑弹框
- 新建按钮和表格行编辑按钮触发弹框，提交后刷新列表
- 移除路由中 `/projects/new` 和 `/projects/:id/edit` 的独立页面路由

### 步骤 2：服务器表单迁移到弹框

- 文件：`web/src/pages/servers/form.tsx` → 重构为 Dialog 组件
- 在服务器列表页（`web/src/pages/servers/list.tsx`）中集成新建/编辑弹框
- 移除路由中 `/servers/new` 和 `/servers/:id/edit` 的独立页面路由

### 步骤 3：清理冗余路由

- 文件：`web/src/App.tsx`
- 移除已弃用的表单页面路由条目
- 确保导航链接和面包屑更新一致

### 步骤 4：布局优化

- 参照 shadcn/ui 官方示例和优秀开源项目（如 taxonomy、shadcn-admin）的布局模式
- 优化 Sidebar：改进折叠交互、导航分组、活跃状态样式
- 优化 Header：精简高度、改进面包屑、响应式适配
- 优化内容区：统一页面容器间距、标题区域布局、空状态展示
- 整体色调和间距微调，提升视觉一致性

### 步骤 5：验证

- TypeScript 编译通过：`cd web && bun run build`
- Lint 通过：`cd web && bun run lint`
- 各页面功能正常：列表页新建/编辑弹框流程完整、布局在桌面和移动端均表现正常

## 影响范围

- `web/src/App.tsx` — 移除 4 条独立表单页面路由，移除对应导入
- `web/src/pages/projects/form.tsx` — 从 `ProjectFormPage` 重构为 `ProjectFormDialog` 弹框组件
- `web/src/pages/projects/list.tsx` — 集成 `ProjectFormDialog`，新建/编辑按钮改为打开弹框
- `web/src/pages/projects/detail.tsx` — 集成 `ProjectFormDialog`，编辑按钮改为打开弹框
- `web/src/pages/servers/form.tsx` — 从 `ServerFormPage` 重构为 `ServerFormDialog` 弹框组件
- `web/src/pages/servers/list.tsx` — 集成 `ServerFormDialog`，编辑按钮改为打开弹框
- `web/src/components/layout/sidebar.tsx` — 导航分组、Tooltip、渐变 Logo、活跃态蓝色图标
- `web/src/components/layout/header.tsx` — 精简高度、ChevronRight 面包屑、毛玻璃背景、渐变头像
- `web/src/components/layout/app-layout.tsx` — 内容区 max-w-7xl 限宽 + 响应式间距

## 历史补丁
