# 操作手册（2.0 GA）

面向运维与首次部署。产品行为以 [DESIGN.md](./DESIGN.md) 为准；分期 Gate 见 [ROADMAP.md](./ROADMAP.md)。

---

## 1. 全新安装（默认 SQLite）

2.0 **只支持全新安装**，**不提供** 1.x → 2.0 数据迁移。

```bash
# 1. 取得发布包（示例：Linux amd64）
#    bedrock-linux-amd64 + bedrock-agent-linux-amd64（+ .sha256）

# 2. 准备空数据目录与配置
mkdir -p ./data
cp config.example.yaml config.yaml   # 或参考仓库 config.yaml
# database.driver: sqlite
# database.path: ./data/bedrock.sqlite
# encryption.key: 64 hex（生产务必更换）
# admin.username / admin.password: 首启种子超管

# 3. 启动（空目录 → migration → 种子超管）
./bedrock-linux-amd64 --config ./config.yaml

# 4. 验证
curl -fsS http://127.0.0.1:8080/api/v1/health
# 浏览器打开 http://host:8080 使用超管登录
```

可复现冒烟：

```bash
make smoke-fresh-install
```

---

## 2. 多数据库配置

支持 `sqlite`（默认）、`postgres` / `postgresql`、`mysql`。

- **改 driver ≠ 搬迁数据**。切换引擎前请自行用目标库工具完成数据搬迁（平台不提供跨库迁移）。
- 错误的连通性配置会导致 **拒绝启动**。
- 合同测试 / 冒烟：

```bash
# 单元级三驱动合同（需 DSN 时设置环境变量）
go test ./internal/platform/db/... -tags=contract
# BEDROCK_CONTRACT_POSTGRES_DSN / BEDROCK_CONTRACT_MYSQL_DSN

# 进程级冒烟（SQLite 必跑；Postgres/MySQL 可选）
make smoke-three-db
# BEDROCK_SMOKE_POSTGRES=1 BEDROCK_DB_*...
# BEDROCK_SMOKE_MYSQL=1 BEDROCK_DB_*...
```

---

## 3. 备份指引（不假装统一物理备份）

| 驱动 | 建议 |
| --- | --- |
| SQLite | 停写或使用备份命令/文件复制 `database.path`；同时备份 `build.*` / `storage.root` 等数据目录 |
| Postgres | `pg_dump` / PITR 等官方工具 |
| MySQL | `mysqldump` / 官方备份方案 |

平台可提供备份**指引**与（若有）SQLite 辅助命令；**不会**声称跨引擎统一物理备份。

工作区、制品、日志、对象存储目录需按业务 RPO 一并纳入备份范围。

---

## 4. 已接受风险（产品内可见 + 文档）

1. **HTTP + 浏览器会话存储**：`access_token`（Web Storage）可能被同机脚本读取；`refresh_token` 为 HttpOnly Cookie（不设 Secure）；`password_cipher` **不替代** TLS。生产强烈建议 HTTPS。
2. **同 UID 执行**：构建脚本、AI CLI、自定义超管命令与 Bedrock 进程同一 OS 用户；RBAC **不是** OS 沙箱。
3. **自定义超管命令 / 开发环境脚本**：仅超管；任意命令执行，须审计与最小授权。

---

## 5. 前端 embed 与回滚

- 默认 `FRONTEND_DIR=web-v2`；Release 将 `web-v2/dist` 拷入 `cmd/server/dist` 后 `go build` embed。
- 旧 `web/`（或上一版前端产物）**保留至少一个发布周期**便于回滚。
- 回滚步骤见 [release-checklist.md](./release-checklist.md#前端-embed-回滚)。

---

## 6. 发布包回退

1. 停止当前 Server / Agent 进程。
2. 换回上一版本二进制（校验 SHA256）。
3. **不要**对 2.0 schema 期望兼容更旧的未声明迁移；回退前确认 migration 版本与备份。
4. 数据目录从备份还原（若二进制回退伴随破坏性 schema 变更）。

---

## 7. Deploy Agent

独立二进制与 Server **同版本**发布：`bedrock-agent-linux-amd64` / `bedrock-agent-linux-arm64` 等。Agent 部署在目标机，不嵌入 Server。
