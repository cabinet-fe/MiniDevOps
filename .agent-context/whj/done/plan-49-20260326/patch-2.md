# 修复搜索框文字颜色与搜索匹配计数

## 补丁内容

修复日志查看器搜索功能的两个问题：

1. **搜索框输入文字不可见**：搜索框 `<Input>` 组件在 `bg-transparent` 背景上没有显式指定文字颜色，继承了 shadcn/ui 默认的深色文字，在 `bg-zinc-800` 深色背景上几乎看不清。添加 `text-zinc-200` 和 `placeholder:text-zinc-500` 类名修复。

2. **搜索匹配计数错误**：原实现通过 `findNext()`/`findPrevious()` 的 `boolean` 返回值手动递增 `matchCount` 和 `currentMatch`，导致计数完全不准确（如显示 "1/1"、"2/1" 等）。改为使用 `SearchAddon` 的 `onDidChangeResults` 事件回调获取精确的 `resultCount` 和 `resultIndex`，同时启用 `decorations` 选项以触发该事件并高亮所有匹配项（深金色标记匹配，蓝色标记当前活跃匹配）。

## 影响范围

- 修改文件: `web/src/components/build-log-viewer.tsx`
