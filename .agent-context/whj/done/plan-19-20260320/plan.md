# 全局布局美化与仪表板紧凑优化

> 状态: 已执行

## 目标

将 BuildFlow 从"仪表板暗色赛博 + 其他页面 shadcn 默认白"的割裂状态，统一为一致的深色工业化主题。同时将仪表板的信息密度大幅提升、卡片面积大幅缩减，使整个应用在视觉层面形成高度一致的专业 DevOps 控制台风格。

## 内容

### 步骤 1: 全局 CSS 主题与字体统一

**目标**：建立统一的深色色板，替换默认字体，让所有页面共享同一套视觉基底。

- 修改 `web/index.html`：引入 Google Fonts —— IBM Plex Sans（正文）+ JetBrains Mono（数据/代码）
- 修改 `web/src/index.css`：
  - 将 `:root` 变量直接设置为深色值（不再区分 light/dark，固定深色）
  - 色板定义：底色 zinc-950/zinc-900，主色 emerald-500，强调 cyan-400
  - 添加 `font-family` 变量使用 IBM Plex Sans / JetBrains Mono
  - body 强制 `class="dark"` 或直接用深色 CSS 变量
- 修改 `web/src/components/layout/app-layout.tsx`：移除 `bg-zinc-50 dark:bg-zinc-950` 双模式，固定为深色背景

**完成标准**：所有页面在无 `.dark` class 时也呈现深色主题，字体全局生效。

### 步骤 2: 应用布局重构（Sidebar + Header）

**目标**：精炼侧栏和顶栏，减少视觉噪音，建立品牌辨识度。

#### Sidebar (`web/src/components/layout/sidebar.tsx`)
- Logo 区：保持 Rocket 图标，渐变色改为 emerald 系（与主色一致）
- 导航项：active 状态使用 emerald 左边指示条 + 轻微 bg 提亮（替代当前的 `bg-white/10`）
- 分组标签：保持精简，字号缩小
- 整体宽度：展开态 `200px`（从 220px 缩减），收起态保持 `60px`

#### Header (`web/src/components/layout/header.tsx`)
- 高度从 `h-12` 缩为 `h-11`
- 背景：固定深色 `bg-zinc-900/90 backdrop-blur`，移除 light 模式样式
- 面包屑保持，右侧用户区保持
- 移除所有 `dark:` 前缀条件

#### AppLayout (`web/src/components/layout/app-layout.tsx`)
- 固定深色背景
- 增加 `overflow-x-hidden` 防止水平溢出

**完成标准**：侧栏和顶栏在所有页面视觉统一，品牌色一致。

### 步骤 3: 仪表板页面紧凑重构

**目标**：在同等视口下展示更多信息，移除视觉赘肉，卡片面积缩减 50%+。

#### 3.1 移除/简化
- 删除顶部标题横幅 section（页面标题由 Header 面包屑承担）
- 删除 `DARK_CARD` 常量及其赛博风格样式
- 删除底部「项目分组概览」section（此信息在项目列表页已有，仪表板不重复）

#### 3.2 统计指标行
- 将 4 个独立 Card 改为一个紧凑的 flex 行：每个指标只占一个 `flex item`
- 每个指标：上方小字标签 + 下方大字数值，无图标无详情描述
- 整行用一个薄 border-b 分隔，不需要独立卡片容器

#### 3.3 系统资源行
- 3 个资源（CPU/内存/磁盘）压缩为单行 3 列紧凑 meter
- 每个 meter：左侧 label + 数值，右侧细长进度条（h-1.5）
- 右上角显示轮询时间戳
- 不再使用嵌套子卡片

#### 3.4 构建趋势图
- 高度从 280px 缩减为 180px
- 使用统一的深色卡片样式（不再用 DARK_CARD）

#### 3.5 活跃构建 + 最近构建
- 活跃构建：若为空则显示单行占位而非 card 空状态
- 最近构建：改为紧凑的表格行（不再是每条一个子卡片），每行显示构建号、项目名、状态 badge、分支、耗时、时间
- 两区域并排，左侧活跃构建占较小宽度，右侧最近构建占较大宽度

#### 3.6 轮询策略确认
- 初始加载：一次性获取 stats、active-builds、recent-builds、trend（保持不变）
- 定时轮询：仅 `/dashboard/system-resources`（已是当前行为，保持不变）

**完成标准**：仪表板在 1080p 视口下不需要滚动即可看到所有核心信息；视觉风格与全局深色主题统一。

### 步骤 4: 其他页面样式统一

**目标**：所有列表页、表单页与深色主题一致，无 light 模式残留。

- 所有页面移除 `dark:` 前缀（因为已固定深色）
- Card 边框统一为 `border-zinc-800`（移除 `border-zinc-200 dark:border-zinc-800` 双模式）
- 页面标题 `h1` 统一颜色为 `text-white`
- 副标题 `p` 统一为 `text-zinc-400`
- loading spinner 统一样式
- 涉及文件：
  - `web/src/pages/projects/list.tsx`
  - `web/src/pages/servers/list.tsx`
  - `web/src/pages/users/list.tsx`
  - `web/src/pages/audit-logs.tsx`
  - `web/src/pages/settings.tsx`
  - `web/src/pages/dictionaries/list.tsx`
  - `web/src/pages/login.tsx`（登录页保持独立全屏深色风格，只调品牌色为 emerald）

**完成标准**：切换任意页面无风格跳变，所有卡片/表格/按钮/badge 视觉一致。

## 影响范围

- `web/index.html` — Google Fonts、favicon 引用 buildflow.svg、`class="dark"` 默认
- `web/src/index.css` — `:root` 浅色变量 + `.dark` 深色变量（双主题）
- `web/src/App.tsx` — ThemeProvider 包裹、语义色 spinner
- `web/src/components/layout/app-layout.tsx` — `bg-background` 语义色
- `web/src/components/layout/sidebar.tsx` — `bg-sidebar`、`border-border`、`text-foreground` 语义色
- `web/src/components/layout/header.tsx` — 主题切换按钮、语义色面包屑
- `web/src/components/notification-bell.tsx` — 语义色 hover/read 状态
- `web/src/components/build-log-viewer.tsx` — 外层 border 语义色
- `web/src/components/dashboard/build-trend-chart.tsx` — 180px、emerald 成功色
- `web/src/pages/dashboard.tsx` — 全量语义色替换
- `web/src/pages/projects/list.tsx` — 语义色边框/背景/文字
- `web/src/pages/projects/detail.tsx` — 两大块布局 + 语义色
- `web/src/pages/projects/form.tsx` — 语义色边框/文字
- `web/src/pages/projects/environment-form.tsx` — CodeMirror 主题修复、`border-border`、`text-muted-foreground`
- `web/src/pages/servers/list.tsx` — 语义色
- `web/src/pages/servers/form.tsx` — 语义色 spinner
- `web/src/pages/users/list.tsx` — 语义色
- `web/src/pages/audit-logs.tsx` — 语义色
- `web/src/pages/settings.tsx` — 语义色
- `web/src/pages/dictionaries/list.tsx` — 语义色
- `web/src/pages/builds/detail.tsx` — 语义色
- `web/src/pages/login.tsx` — 双主题渐变背景、语义色卡片
- `web/public/buildflow.svg` — 新增 BuildFlow logo

## 历史补丁

- patch-1: 项目详情页紧凑重构与风格统一
- patch-2: 浅色主题支持、CodeMirror 深色修复与 Logo 替换
- patch-3: 仪表板字体放大、图表样式变更与资源轮询加速
