# 禁用 Cookie，refresh_token 改为 JSON body 传递

> 状态: 已执行

## 目标

移除所有 HTTP 请求中的 `credentials: 'include'`，将 refresh_token 从 Cookie 改为 localStorage 存储 + JSON body 传递，彻底禁用 Cookie。

## 内容

1. **后端 `auth_handler.go`**：
   - Login：移除 `c.SetCookie`，将 `refresh_token` 写入响应体。
   - Logout：移除 `c.SetCookie`。
   - Refresh：从 JSON body 读取 `refresh_token`（替代 `c.Cookie`），将新 `refresh_token` 写入响应体（替代 `c.SetCookie`）。

2. **前端 `api.ts`**：
   - 移除所有 `credentials: 'include'`。
   - refresh 请求改为 POST JSON body 传递 `refresh_token`（从 localStorage 读取）。
   - 成功 refresh 后同时更新 localStorage 中的 `refresh_token`。

3. **前端 `auth-store.ts`**：
   - login 成功后存储 `refresh_token` 到 localStorage。
   - logout 时清除 `refresh_token`。

4. **前端 `settings.tsx`**：
   - 移除三处直接 `fetch` 调用中的 `credentials: 'include'`。

## 影响范围

- `internal/handler/auth_handler.go` — 移除 Cookie 读写，refresh_token 改为 JSON body 传递/返回
- `web/src/lib/api.ts` — 移除所有 `credentials: 'include'`，refresh 请求改为 POST body 传递 refresh_token
- `web/src/stores/auth-store.ts` — login 存储 refresh_token，logout/异常清除 refresh_token
- `web/src/pages/settings.tsx` — 三处直接 fetch 调用移除 `credentials: 'include'`

## 历史补丁
