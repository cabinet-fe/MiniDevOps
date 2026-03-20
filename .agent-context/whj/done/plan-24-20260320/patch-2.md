# 前端登录加密改为 crypto-es + Web Crypto，移除明文回退

## 补丁内容

1. **依赖**：移除 `aes-js`，改用 **`crypto-es`**；在非安全上下文（无 `crypto.subtle`，如 `http://` 内网 IP）下用 `crypto-es` 做 AES-256-CBC，与后端格式一致。
2. **双实现路径**：**安全上下文**（浏览器对 `SubtleCrypto` 可用，含 HTTPS、`localhost` 等）下使用 **Web Crypto** `encrypt(AES-CBC)`；否则使用 **crypto-es**（`AES` + `CBC` + `Pkcs7`）。仅保留这一条加密路径，不再在失败或未配置时回退为 JSON 中的明文 `password`。
3. **API**：`encryptLoginPassword` 改为 `async`，校验 64 位 hex 密钥，失败直接抛错；`auth-store` 登录体固定为 `{ username, password_cipher }`。
4. **文档**：`AGENTS.md` 同步上述行为。

## 影响范围

- 修改文件: `/home/whj/codes/dev-ops/web/package.json`
- 修改文件: `/home/whj/codes/dev-ops/web/bun.lock`
- 修改文件: `/home/whj/codes/dev-ops/web/src/lib/login-crypto.ts`
- 修改文件: `/home/whj/codes/dev-ops/web/src/stores/auth-store.ts`
- 修改文件: `/home/whj/codes/dev-ops/AGENTS.md`
