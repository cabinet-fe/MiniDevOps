# 修复生产构建未注入 VITE 密钥导致登录报错

## 补丁内容

1. **原因**：`vite build` 使用 `production` 模式，**不会**加载 `.env.development`，仅加载 `.env` / `.env.production` 等；此前仅本地存在 `.env.development` 时，嵌入后端的产物里 `VITE_BUILDFLOW_ENCRYPTION_KEY` 为空，登录时 `encryptLoginPassword` 抛错。
2. **新增 `web/.env`**：与根目录 `config.yaml` 默认 `encryption.key` 一致，使 dev 与 `make build` 均能注入该变量（默认密钥与仓库已有 `encryption.key` 公开程度一致；生产需在 CI 覆盖）。
3. **`login-crypto.ts`**：改为静态读取 `import.meta.env.VITE_BUILDFLOW_ENCRYPTION_KEY`，与 Vite 对 `VITE_*` 的注入方式一致。
4. **文档**：`README.md`、`AGENTS.md`、`web/.env.example` 同步说明。

## 影响范围

- 新增文件: `/home/whj/codes/dev-ops/web/.env`
- 修改文件: `/home/whj/codes/dev-ops/web/src/lib/login-crypto.ts`
- 修改文件: `/home/whj/codes/dev-ops/README.md`
- 修改文件: `/home/whj/codes/dev-ops/AGENTS.md`
- 修改文件: `/home/whj/codes/dev-ops/web/.env.example`
