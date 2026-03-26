# 修复日志查看器工具栏按钮对比度

## 补丁内容

日志查看器右上角的按钮组使用 shadcn ghost variant，其默认 hover 背景为 `accent` 色（偏白/浅灰），在 `bg-zinc-950` 深色背景上对比度极差，导致按钮文字几乎看不清。

修改内容：
- 工具栏容器加上 `bg-zinc-900/80` 背景，增强视觉层次
- 搜索框容器背景从 `bg-zinc-900` 改为 `bg-zinc-800`，与工具栏区分
- 所有 ghost 按钮的文字颜色从 `text-zinc-400` 提升至 `text-zinc-300`
- 所有 ghost 按钮覆盖 hover 背景为 `hover:bg-zinc-700`，替代默认 accent 色
- hover 文字统一为 `hover:text-white`

## 影响范围

- 修改文件: `web/src/components/build-log-viewer.tsx`
