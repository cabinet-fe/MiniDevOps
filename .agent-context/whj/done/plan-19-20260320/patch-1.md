# 项目详情页紧凑重构与风格统一

## 补丁内容

将项目详情页从多卡片分散布局重构为两大块紧凑结构，与全局深色主题统一，大幅提升信息密度和操作可达效率。

### 结构变更

- **两大块布局**：整个页面仅包含「项目信息」和「环境与构建」两个 Card，移除了原来的仓库策略卡片、Webhook 接入卡片、英雄区等多个独立区块
- **Tab 切换环境**：用 Radix Tabs 替代垂直堆叠的环境卡片，通过标签页切换不同环境，解决环境过多时需要滚动寻找的问题
- **固定高度可滚动构建表**：每个环境的构建历史表格限制最大高度 420px 并支持内部滚动，表头 sticky 固定，防止长构建列表撑长页面
- **图标按钮操作列**：构建行操作从文字按钮改为图标按钮 + Tooltip，操作列宽度从 320px 缩减到 112px

### 信息密度提升

- 项目概览改为仪表板风格的紧凑统计栏（环境、成功构建、保留制品、认证、格式、Webhook），与 dashboard 页面风格一致
- 仓库和 Webhook URL 合并为两列紧凑行，内嵌复制和外部链接按钮
- 环境配置从 7 个独立 EnvironmentMeta 卡片改为单行内联 key-value 展示
- 移除了 SIGNAL_TOKENS、OverviewStat、EnvironmentMeta、CompactMeta、InlineLinkPanel 等过度装饰性组件

### 风格统一

- 移除所有 `dark:` 前缀，固定深色主题
- 统一使用 `border-zinc-800` 边框、`text-white`/`text-zinc-200` 文字、`text-zinc-400`/`text-zinc-500` 辅助文字
- Loading spinner 与全局一致（emerald 色）
- 日期格式改为紧凑的月/日 时:分格式

### 代码精简

- 提取 `BuildRow` 为独立组件，提升可读性
- 文件从 1148 行缩减到约 580 行（减少 ~50%）

## 影响范围

- 修改文件: `web/src/pages/projects/detail.tsx`
