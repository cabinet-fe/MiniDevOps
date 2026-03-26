# 使用 xterm.js 重构构建日志查看器

> 状态: 已执行

## 目标

修复构建日志查看器中长行文字换行后与下一行重叠的 bug，并使用 xterm.js 替代当前的 DOM 渲染方案，实现高性能的构建日志显示，天然支持 ANSI 颜色/样式解析，内建搜索功能。

## 问题分析

当前实现使用 `ansi-to-html` 将 ANSI 转义码转换为 HTML，然后用 `dangerouslySetInnerHTML` 渲染每行日志。存在以下问题：

1. **文字重叠 bug**：每行固定 `height: 24px`（`LINE_HEIGHT = 24`），但使用了 `whitespace-pre-wrap break-all`，当内容超出单行宽度后文字会换行，而固定高度导致换行内容溢出并与下一行重叠。
2. **性能瓶颈**：每行日志对应一个 DOM 节点，大量日志时 DOM 节点爆炸，虽然有 5000 行以上的虚拟滚动，但虚拟滚动实现依赖固定行高，与换行冲突。
3. **ANSI 解析不完整**：`ansi-to-html` 只处理基本颜色，不支持 256 色和 TrueColor。
4. **搜索高亮粗暴**：直接操作 HTML 字符串做正则替换，可能破坏 ANSI 转 HTML 后的标签结构。

## 技术方案

使用 **xterm.js**（`@xterm/xterm`）替换当前方案：

- **渲染**：Canvas 渲染（默认），不创建 DOM 节点，性能远超 DOM 方案
- **ANSI**：天然支持全部 ANSI 转义码（8色/256色/TrueColor/粗体/斜体等）
- **搜索**：`@xterm/addon-search` 提供搜索功能
- **自适应**：`@xterm/addon-fit` 让终端自适应容器尺寸
- **WebGL 加速**（可选）：`@xterm/addon-webgl` 进一步提高渲染性能

## 内容

### 步骤 1：安装依赖

```bash
cd web && bun add @xterm/xterm @xterm/addon-fit @xterm/addon-search @xterm/addon-webgl
```

同时移除不再需要的 `ansi-to-html` 依赖。

### 步骤 2：重写 `build-log-viewer.tsx`

重新实现 `BuildLogViewer` 组件：

- 初始化 xterm `Terminal` 实例，配置只读模式（`disableStdin: true`）
- 加载 `FitAddon`、`SearchAddon`、`WebglAddon`
- 保留现有的 WebSocket 集成（`useWebSocket` hook），通过 `terminal.write()` 写入日志
- 保留 HTTP API 加载历史日志（通过 `api.getText()`）
- 保留顶部工具栏（状态 Badge、搜索、复制、全屏）
- 样式采用 xterm.js 的 CSS（`@xterm/xterm/css/xterm.css`）
- 搜索功能使用 `SearchAddon.findNext()` / `findPrevious()`
- 自动滚动跟随：检测用户滚动行为，自动跟随/取消跟随
- 保持组件的 props 接口不变（`BuildLogViewerProps`），确保与 `detail.tsx` 兼容

### 步骤 3：验证

- 运行 `bun run lint` 检查
- 运行 `bun run build` 确认编译通过

## 影响范围

- `web/src/components/build-log-viewer.tsx` — 使用 xterm.js 完全重写
- `web/src/index.css` — 添加 xterm 容器 CSS 样式
- `web/package.json` — 新增 @xterm/* 依赖，移除 ansi-to-html

## 历史补丁

- patch-1: 修复日志查看器工具栏按钮对比度
- patch-2: 修复搜索框文字颜色与搜索匹配计数
- patch-3: 修复构建日志搜索异常
- patch-4: 优化构建进度 UI
