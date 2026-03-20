# 修复开发环境下 WebSocket 提前关闭告警

## 补丁内容

调整 `useWebSocket` 的连接生命周期管理：将消息/打开/关闭回调改为稳定引用，避免组件重渲染时反复重建连接；同时将首次建连延后到 effect 确认存活后执行，规避 React StrictMode 在开发环境下首轮挂载后立即清理导致的 “WebSocket is closed before the connection is established” 告警。

## 影响范围

- 修改文件: `web/src/hooks/use-websocket.ts`
