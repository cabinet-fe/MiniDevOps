# 开放接口（常用摘录）

本文摘录最常用的开放 HTTP 约定与接口，便于脚本与外部系统对接。完整契约见仓库 `api/` 目录（按域拆分）；通用约定见 `.agents/api.md`。

**Base URL**：`{host}/api/v1`  
**WebSocket**：`{host}/ws`（路径前缀不是 `/api/v1`）

---

## 1. 通用约定

### 1.1 响应信封

所有 JSON 响应使用：

```json
{ "code": 0, "message": "success", "data": {} }
```

| 情况 | `code`           | 说明                                             |
| ---- | ---------------- | ------------------------------------------------ |
| 成功 | `0`              | `message` 为 `success`（200）或 `created`（201） |
| 失败 | 等于 HTTP 状态码 | 如 400 / 401 / 403；另含 `request_id`            |

字段均为 `snake_case`。异步创建常返回 `202`，并在 `data` 中给出资源 ID 与状态。

### 1.2 分页

请求：`page`、`page_size`  
响应 `data`：`items`、`total`、`page`、`page_size`、`total_pages`

### 1.3 认证方式

| 方式                | 用法                                                                                             |
| ------------------- | ------------------------------------------------------------------------------------------------ |
| 登录 JWT            | `Authorization: Bearer <access_token>`                                                           |
| 个人访问令牌（PAT） | `Authorization: Bearer br_...`（明文仅创建时返回一次）                                           |
| 刷新令牌            | HttpOnly Cookie `refresh_token`；非浏览器客户端可在 `POST /auth/refresh` body 传 `refresh_token` |

收到 `401` 后调用 `POST /auth/refresh`，成功后重试原请求。

---

## 2. 认证

### POST /auth/login — 登录

- 认证：不需要
- 请求：`{ "username": "...", "password_cipher": "..." }`（Web 优先密文；调试才可用 `password`）
- 响应：`data.access_token`、`user`、`permissions`、`menus`；`refresh_token` 仅 Set-Cookie

### POST /auth/refresh — 刷新

- 认证：不需要（依赖 Cookie 或 body 中的 `refresh_token`）
- 响应：新的 `access_token`，并轮换 Cookie

### POST /auth/logout — 登出

- 清除 `refresh_token` Cookie

### GET /auth/me — 当前用户

- 响应：`user`、`permissions`（功能 `full_code` 列表）、`menus`（两层侧栏）

**调用示例（登录）：**

```bash
curl -fsS -X POST "$HOST/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"your-password"}' \
  -c cookies.txt
# 取出 data.access_token 后：
curl -fsS "$HOST/api/v1/auth/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

---

## 3. 个人访问令牌（PAT）

管理接口在资源域；鉴权仍走 Bearer。

| 方法   | 路径                    | 权限                     | 说明                               |
| ------ | ----------------------- | ------------------------ | ---------------------------------- |
| GET    | `/resource/tokens`      | `resource_tokens:view`   | 列出本人令牌元数据（分页）         |
| POST   | `/resource/tokens`      | `resource_tokens:create` | 创建；响应含明文 `token`（仅一次） |
| DELETE | `/resource/tokens/{id}` | `resource_tokens:delete` | 删除                               |

创建请求示例字段：`name*`、`scopes*`、`expires_at?` 或 `expires_in_days?`（`30` / `90` / `180` / `365`）。  
常用 scope：`skills:read`、`agents:run`、`docs:write`、`docs:publish`。

```bash
curl -fsS -X POST "$HOST/api/v1/resource/tokens" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"ci","scopes":["agents:run"],"expires_in_days":90}'
```

---

## 4. 健康检查

安装与探活常用：

```bash
curl -fsS "$HOST/api/v1/health"
```

---

## 5. CI/CD（节选）

路径相对 `/api/v1`。

| 方法 | 路径                      | 权限                      | 说明            |
| ---- | ------------------------- | ------------------------- | --------------- |
| GET  | `/build-jobs`             | `cicd_build_jobs:view`    | 列出构建任务    |
| POST | `/build-jobs`             | `cicd_build_jobs:create`  | 创建构建任务    |
| POST | `/build-jobs/{id}/runs`   | `cicd_build_jobs:execute` | 入队构建（202） |
| GET  | `/build-runs`             | `cicd_build_runs:view`    | 列出构建运行    |
| GET  | `/build-runs/{id}`        | `cicd_build_runs:view`    | 运行详情        |
| POST | `/build-runs/{id}/cancel` | `cicd_build_jobs:execute` | 取消运行        |

手动触发示例：

```bash
curl -fsS -X POST "$HOST/api/v1/build-jobs/1/runs" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"trigger_type":"manual"}'
```

---

## 6. AI（节选）

| 方法 | 路径                       | 权限 / 说明                               |
| ---- | -------------------------- | ----------------------------------------- |
| GET  | `/ai/agents`               | `ai_agents:view` — 列出智能体             |
| POST | `/ai/agents/{id}/runs`     | `ai_agents:execute` — 手动触发运行        |
| POST | `/ai/agents/{id}/api-runs` | 需 PAT scope（如 `agents:run`）— API 触发 |
| GET  | `/ai/runs`                 | `ai_runs:view` — 列出运行记录             |
| GET  | `/ai/runs/{id}`            | `ai_runs:view` — 运行详情                 |
| POST | `/ai/runs/{id}/cancel`     | 取消运行                                  |
| GET  | `/ai/skills`               | `ai_skills:view` — 列出技能               |

请求体与字段以 `api/ai.md` 为准。

---

## 7. 域索引（完整契约）

| 域    | 仓库文件          | 内容                             |
| ----- | ----------------- | -------------------------------- |
| 认证  | `api/auth.md`     | login / refresh / logout / me    |
| 系统  | `api/system.md`   | 用户、角色、RBAC、字典、操作日志 |
| 资源  | `api/resource.md` | 仓库、服务器、凭证、PAT          |
| CI/CD | `api/cicd.md`     | 构建任务、运行、Webhook          |
| 运维  | `api/ops.md`      | 仪表盘卡片、进程、开发环境       |
| 项目  | `api/project.md`  | 项目、需求、文档发布             |
| AI    | `api/ai.md`       | Agents、Runs、Skills             |
