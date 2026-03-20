# 登录加密优化（AES-256-CBC）

> 状态: 已执行

## 目标

在配置了前端密钥时，对登录请求中的密码做 **AES-256-CBC** 保护：前端在提交前用 **第三方轻量 AES 库** 加密，后端在 `Authenticate` 前解密，与现有 `encryption.key`（32 字节 hex）对齐，避免明文密码出现在 JSON 体中（降低日志/代理层误记录风险）。

**不再以 Web Crypto（`crypto.subtle`）为唯一实现**：选用 **npm 上的高性能、轻量 AES 实现**（实施时二选一并在代码中固定），使行为不依赖「安全上下文」即可使用 `SubtleCrypto`，与 HTTP 内网部署兼容；未配置 `VITE_*` 密钥时仍回退为明文 `password`，与现有行为一致。

保留短期兼容：仍接受明文 `password` 字段，便于灰度与本地调试。

## 内容

1. **依赖选型（已锁定）**
   - **前端**：使用 **`aes-js`**（体积小、专注 AES，PKCS#7 由库提供）。禁止在业务代码中手写块运算；加解密流程封装在单一模块内。
   - **后端**：使用 Go **标准库** `crypto/aes` + `cipher`（`NewCBCDecrypter` 等）— 这是 Go 官方实现，性能与 `stdlib` 一致，无需再叠一层第三方 C 绑定；解密逻辑集中在 `internal/pkg/crypto.go`，与「第三方前端库 + 约定格式」对齐即可。

2. **后端 `internal/pkg/crypto.go`**
   - 新增 AES-CBC 解密：`ciphertext` 格式为 `IV(16 字节) || PKCS#7 填充后的密文`，整体 **hex** 编码（与现有 GCM 存储区分：可用长度/偶数字节等启发式，解密成功则视为密文）。
   - 使用已有 `encryptionKey`（与 `InitEncryption` 一致），`cipher.NewCBCDecrypter`，解密后校验并去除 PKCS#7。
   - 单元测试：与前端约定格式往返一致（可用 `export_test` 或包内 `EncryptAESCBCHex` 辅助）；错误密文、短输入、错误 padding 返回明确错误。

3. **后端 `internal/handler/auth_handler.go`（Login）**
   - 请求体：`username` 必填；`password`（明文，可选）与 `password_cipher`（hex，可选）— **若 `password_cipher` 非空则只解密该字段**；否则使用 `password`。
   - 解密失败返回 400 与统一错误文案（不泄露细节），与「密码错误」401 区分。

4. **前端**
   - 新增 `web/src/lib/login-crypto.ts`（或等价模块）：使用已选 **npm AES 库** 实现 AES-256-CBC，IV 每次随机 16 字节，输出与后端约定一致的 hex 串。
   - **回退（满足任一即只发明文 `password`，不发 `password_cipher`）**：
     - 未配置 `VITE_BUILDFLOW_ENCRYPTION_KEY`（或与后端同值的注入变量）；
     - 或运行时初始化加密失败（可打日志后回退）。
   - `web/src/stores/auth-store.ts`：仅在「密钥存在且加密成功」时提交 `password_cipher` 并省略明文 `password`；否则沿用明文 `password`。
   - 在 `web/package.json` 中声明上述依赖，**不使用** `crypto.subtle` 作为登录加密的唯一路径。

5. **文档与配置示例**
   - 更新 `AGENTS.md`：说明登录可选 AES-CBC、**前端依赖第三方 AES 库**、环境变量与 `encryption.key` 关系、明文回退条件。
   - 若存在 `config.yaml` / `.env.example` 的前端说明，补 Vite 变量示例。

6. **验收**
   - `go test ./internal/pkg/...` 与 `go test ./internal/handler/...` 通过。
   - `cd web && bun run build` 通过。
   - 手动：相同 hex key 下 **HTTPS / HTTP / localhost / 内网 IP** 均能按预期（有密钥则密文登录，无密钥则明文）；未配置前端 key 时明文登录仍成功。

## 影响范围

- `internal/pkg/crypto.go` — `DecryptLoginPasswordCipher`、PKCS#7、测试用 `encryptAES256CBCHexForTest`
- `internal/pkg/crypto_test.go`（新增）
- `internal/handler/auth_handler.go` — `password` / `password_cipher` 分支与 400 文案
- `web/package.json`、`web/bun.lock` — 依赖 `crypto-es`（patch-2 起；曾用 `aes-js`）
- `web/src/lib/login-crypto.ts`（新增）
- `web/src/vite-env.d.ts`（新增）
- `web/src/stores/auth-store.ts`
- `AGENTS.md`
- `cmd/server/embed_prod.go`、`cmd/server/embed_dev.go`、`cmd/server/main.go` — 嵌入 SPA 注入运行时密钥（patch-4）
- `config.yaml`（注释）
- `README.md`（`encryption.key` 说明）
- `web/.env`、`web/.env.development`、`web/.env.example` — 开发/构建注入 VITE 密钥与示例

## 历史补丁

- patch-1: 以 aes-js 替代 crypto-js，并补充开发环境 VITE 密钥使密文登录默认可用
- patch-2: 前端登录加密改为 crypto-es + Web Crypto，移除明文回退
- patch-3: 新增 `web/.env` 修复 production build 未注入 VITE 密钥导致登录报错
- patch-4: 嵌入二进制在 `index.html` 注入 `window.__BUILDFLOW_ENCRYPTION_KEY__`，与运行时 `encryption.key` 对齐，优先于 VITE 编译时注入

