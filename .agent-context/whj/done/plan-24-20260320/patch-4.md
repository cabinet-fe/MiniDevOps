# 运行时 encryption.key 与嵌入前端对齐

## 补丁内容

**问题**：后端 `encryption.key` 来自运行时 `config.yaml`（可改），而前端此前仅依赖 Vite 编译时 `VITE_BUILDFLOW_ENCRYPTION_KEY`，改后端密钥后需重编前端才能密文登录。

**方案**：

1. **`cmd/server/embed_prod.go`**：在返回嵌入的 `dist/index.html` 时，于 `</head>` 前插入  
   `<script>window.__BUILDFLOW_ENCRYPTION_KEY__=…</script>`（值为 JSON 转义后的当前 `encryption.key` 字符串）。
2. **`serveSPA`** 增加参数 `encryptionKeyHex`，`main.go` 传入 `cfg.Encryption.Key`；`embed_dev.go` 保留占位参数（开发模式仍用 Vite，不注入）。
3. **`web/src/lib/login-crypto.ts`**：读取密钥时 **优先** `window.__BUILDFLOW_ENCRYPTION_KEY__`，否则回退 `import.meta.env.VITE_BUILDFLOW_ENCRYPTION_KEY`。
4. **`web/src/vite.env.d.ts`**：声明 `Window.__BUILDFLOW_ENCRYPTION_KEY__`。
5. **文档**：`AGENTS.md`、`README.md` 说明嵌入部署与 dev 的差异。

## 影响范围

- 修改文件: `/home/whj/codes/dev-ops/cmd/server/embed_prod.go`
- 修改文件: `/home/whj/codes/dev-ops/cmd/server/embed_dev.go`
- 修改文件: `/home/whj/codes/dev-ops/cmd/server/main.go`
- 修改文件: `/home/whj/codes/dev-ops/web/src/lib/login-crypto.ts`
- 修改文件: `/home/whj/codes/dev-ops/web/src/vite.env.d.ts`
- 修改文件: `/home/whj/codes/dev-ops/AGENTS.md`
- 修改文件: `/home/whj/codes/dev-ops/README.md`
