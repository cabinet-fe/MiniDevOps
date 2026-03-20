# 以 aes-js 替代 crypto-js 并修复开发环境密文登录

## 补丁内容

1. **依赖**：移除 `crypto-js` / `@types/crypto-js`，改用体积更小的 **`aes-js`**（仅 AES 与 PKCS#7），登录加密仍输出 `hex(IV||CBC 密文)`，与后端 `DecryptLoginPasswordCipher` 一致；明文 UTF-8 使用 `TextEncoder`。
2. **密文未生效原因**：Vite 仅在存在 `VITE_*` 环境变量时才会启用密文分支；仓库此前未提供 `web/.env*`，开发时 `import.meta.env.VITE_BUILDFLOW_ENCRYPTION_KEY` 为空，始终回退明文。**新增 `web/.env.development`**，与根目录 `config.yaml` 默认 `encryption.key` 一致，使 `make dev` / `bun run dev` 默认走 `password_cipher`。**新增 `web/.env.example`** 供复制自定义。
3. **文档**：`AGENTS.md`、`README.md` 同步依赖名与开发说明。

## 影响范围

- 新增文件: `/home/whj/codes/dev-ops/web/.env.development`
- 新增文件: `/home/whj/codes/dev-ops/web/.env.example`
- 修改文件: `/home/whj/codes/dev-ops/web/src/lib/login-crypto.ts`
- 修改文件: `/home/whj/codes/dev-ops/web/package.json`
- 修改文件: `/home/whj/codes/dev-ops/web/bun.lock`
- 修改文件: `/home/whj/codes/dev-ops/AGENTS.md`
- 修改文件: `/home/whj/codes/dev-ops/README.md`
