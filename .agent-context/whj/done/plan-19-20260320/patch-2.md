# 浅色主题支持、CodeMirror 深色修复与 Logo 替换

## 补丁内容

### 1. 浅色主题支持

在 plan-19 统一深色主题的基础上，恢复并完善了双主题（浅色 + 深色）支持：

- **CSS 变量双套**：`:root` 定义浅色主题变量，`.dark` 定义深色主题变量。浅色以白底灰文 + emerald 主色为基调，深色保持 plan-19 的 zinc-950 底色 + emerald 品牌色
- **ThemeProvider**：在 `App.tsx` 中接入 `next-themes` 的 `ThemeProvider`，默认深色，支持切换
- **主题切换按钮**：Header 右侧新增 Sun/Moon 图标按钮，一键切换浅色/深色

### 2. 全局语义色替换

将所有页面和布局组件中的硬编码 zinc 灰阶替换为 CSS 变量驱动的语义色：

- `text-white` → `text-foreground`
- `text-zinc-400/500` → `text-muted-foreground`
- `bg-zinc-950` → `bg-background`
- `bg-zinc-900` → `bg-card`
- `border-zinc-800` → `border-border`
- Sidebar/Header 使用 `bg-sidebar`、`bg-card/90` 等语义 token

### 3. CodeMirror 深色模式修复

- `environment-form.tsx` 中的 CodeMirror 组件此前在深色模式下显示白色背景，原因是缺少 ThemeProvider 导致 `useTheme` 无法正确获取主题状态
- 接入 ThemeProvider 后 `cmTheme` 能正确返回 `'dark'`/`'light'`，CodeMirror 内置主题自动适配
- 额外通过 CSS 选择器 `[&_.cm-editor]:!bg-transparent` 确保编辑器背景与外层容器一致
- 表单内所有 `border-zinc-200 dark:border-zinc-800` 统一为 `border-border`

### 4. Logo 替换

- 新建 `web/public/buildflow.svg`：emerald→teal 渐变底色 + 白色火箭 + 管道箭头图案
- `index.html` favicon 引用从 `vite.svg` 改为 `buildflow.svg`

## 影响范围

- 新增文件: `web/public/buildflow.svg`
- 修改文件: `web/src/index.css`
- 修改文件: `web/index.html`
- 修改文件: `web/src/App.tsx`
- 修改文件: `web/src/components/layout/app-layout.tsx`
- 修改文件: `web/src/components/layout/sidebar.tsx`
- 修改文件: `web/src/components/layout/header.tsx`
- 修改文件: `web/src/components/notification-bell.tsx`
- 修改文件: `web/src/components/build-log-viewer.tsx`
- 修改文件: `web/src/pages/login.tsx`
- 修改文件: `web/src/pages/dashboard.tsx`
- 修改文件: `web/src/pages/projects/list.tsx`
- 修改文件: `web/src/pages/projects/detail.tsx`
- 修改文件: `web/src/pages/projects/form.tsx`
- 修改文件: `web/src/pages/projects/environment-form.tsx`
- 修改文件: `web/src/pages/servers/list.tsx`
- 修改文件: `web/src/pages/servers/form.tsx`
- 修改文件: `web/src/pages/users/list.tsx`
- 修改文件: `web/src/pages/audit-logs.tsx`
- 修改文件: `web/src/pages/settings.tsx`
- 修改文件: `web/src/pages/dictionaries/list.tsx`
- 修改文件: `web/src/pages/builds/detail.tsx`
