# 登录优化、首页按需加载与 Gzip

> 状态: 已执行

## 目标

1. 重新设计登录页视觉（区别于通用渐变卡片），并支持「记住账号密码」（本地存储，用户可选）。
2. 通过 React 路由级懒加载减小首屏 JS；生产环境由 Gin 对 HTTP 响应启用 gzip 压缩静态与 API 文本。

## 内容

1. **登录页**：工业风深色界面（与现有 IBM Plex / JetBrains 字体体系一致）、表单区与背景层次；勾选「记住账号密码」时用 `localStorage` 读写用户名与密码，取消勾选或成功登录后按选项清除密码；未勾选不写入密码。
2. **路由**：`App.tsx` 对除根布局外的页面使用 `React.lazy` + `Suspense`，提供统一加载占位。
3. **Vite**：`build.rollupOptions.output.manualChunks` 适度拆分 `react`/`react-dom`/`react-router` 等 vendor，配合懒加载。
4. **后端**：引入 `gin-contrib/gzip`，在 `main.go` 注册中间件（在 API 与静态资源之前），确保不与 WebSocket 升级冲突。

## 影响范围

- `web/src/lib/login-persist.ts`：记住登录凭据（localStorage）
- `web/src/pages/login.tsx`：登录页样式与记住账号密码
- `web/src/App.tsx`：路由懒加载与 Suspense
- `web/src/main.tsx`：移除重复的 ThemeProvider
- `web/vite.config.ts`：manualChunks 拆分 vendor
- `cmd/server/main.go`：gzip 中间件
- `go.mod` / `go.sum`：gin-contrib/gzip
- `web/src/index.css`：自托管 IBM Plex Sans / JetBrains Mono（@font-face）

## 历史补丁

- patch-1: 登录页调亮
- patch-2: 自托管字体（替代 Google Fonts）
- patch-3: 登录页车间 / 工业警示风格（F）
