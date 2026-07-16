# 前端约定（web）

改 `web/`、Vue / UI / HTTP 客户端时阅读本文。领域产品规则见 [docs/DESIGN.md](../docs/DESIGN.md)；仓库入口与命令见 [AGENTS.md](../AGENTS.md)。

## 技术栈

| 项 | 技术 |
| -- | ---- |
| 框架 | Vue 3.5+ / TypeScript / Pinia / Vue Router |
| 构建 | Vite+（`vite-plus` / `vp`） |
| UI | `@veltra/desktop` + styles / icons / utils / directives / compositions |
| 工具库 | `@cat-kit/core`、`@cat-kit/fe`、`@cat-kit/http`、`@cat-kit/tsconfig` |
| 包管理 | bun（可由 `vp install` 包装） |
| 目录 | `web/` |

## 必读技能（渐进）

写代码前按需检索，勿整包预加载：

| 场景 | Skill |
| ---- | ----- |
| Vue SFC / Composition | `.agents/skills/vue-best-practices` |
| UI 组件 / 样式 / 图标 | `.agents/skills/veltra-ui`（尤其 `packages/desktop/`） |
| HTTP / 工具库 API | `.agents/skills/cat-kit` |
| 提交 | `.agents/skills/git-commit` |

## 规范

- 路径别名 `@` → `web/src/`。
- **组件与工具必须优先 `@veltra/*` 与 `@cat-kit/*`**。仅确认库内无合适能力后才自研或引入第三方。
- 本地组件目录：`web/src/components/<name>/`，必须含 `index.ts` + `<name>.vue`；类型 / `defineXxx` 等放可选 `helper.ts`。调用方从 `@/components/<name>` 导入（勿写 `.vue` 路径）。
- 字段与交互形态对齐 `@veltra/desktop` 契约（查 `components/<name>/types.d.ts` 与 `api.md`）。例如：
  - 侧栏：`UNav` / `UDualNav` 的 `NavItem`（`title` / `path` / `icon` / `children` 等）
  - 表格列：裸 `UTable` 用 `defineTableColumns`；`ProTable` 用 `defineProTableColumns`；分页：`UPaginator`；表单：`UForm` / `UFormItem`
- API/DTO 可与后端 `snake_case` 并存；映射到组件 props 时保持类型兼容。
- HTTP **只走** `@cat-kit/http` 封装客户端（含 refresh）；禁止页面内散落 `fetch`（除非 DESIGN 标明的特例并抽 helper）。
- 状态：Pinia；权限辅助：composables。
- Token：`access_token` → `@cat-kit/fe` `storage.local` + Bearer；`refresh_token` → 服务端 `Set-Cookie`（HttpOnly，默认跟随 `jwt.refresh_ttl` / 7 天，**不设 Secure**）。HTTP 客户端 `credentials: true`；401 时 POST `/auth/refresh`（带 Cookie）换发 access 并重试原请求。
- 枚举与后端 `snake_case` JSON 字段保持一致。
- **禁止**硬编码全量菜单/权限表作为真源；菜单由后端裁剪下发（见 DESIGN）。

## 登录（前端侧）

- 只提交 `password_cipher`；失败不回退明文密码。
- 密钥优先 `window.__BEDROCK_ENCRYPTION_KEY__`（Go 注入），否则环境变量。
- 安全上下文用 Web Crypto；非安全上下文用兼容库。

## 列表与分页（ProTable）

封装：`web/src/components/pro-table/`（`index.ts` + `pro-table.vue` + `helper.ts`）。页面勿各自复制请求 + 表格 + 分页样板。

```ts
import ProTable, { defineProTableColumns } from "@/components/pro-table";
```

调用方传入：

- **列**：`defineProTableColumns`（支持 `sortable`；勿对 ProTable 再用 Veltra `defineTableColumns`）
- **API**：`url`（相对 `/api/v1`，经 `@cat-kit/http`；信封由 envelope 插件解包）+ `v-model:query`；`dataPath` 默认 `items`（`o(body).get`）
- **filters 插槽**：过滤表单；`autoQueryFields` 控制哪些字段变更自动查询（文本框走查询按钮 / Enter）
- **模式**：`pagination` / `tree` 互斥，或二者皆无（纯列表）；列/行基于 `UTable`
- 组件内部：请求、加载态（`v-loading`）、空态（`UEmpty`）、分页时 `UPaginator`
- 暴露 `search()` / `reload()` 供手动刷新

后端分页信封字段约定见 [.agents/be.md](be.md)。

## 代码风格

- TypeScript / Vue：`vp check`（或项目 lint）
- SFC 文件名：`kebab-case`
- 前端常量：`UPPER_SNAKE_CASE`（对象键 `snake_case`）
- JSON：`snake_case`

## 测试

- 关键路径优先 Playwright 冒烟（登录、菜单、构建日志等）：`web/e2e/`，`bunx playwright test`
- API 级冒烟：`make smoke-api-e2e`
- 不做容量/延迟 SLO 验收（见 ROADMAP）

## 禁止事项（前端）

1. 硬编码全量菜单/权限表作为真源后再隐藏。
2. 绕过统一 HTTP 客户端；或绕过 OpenAPI 契约私自加前后端字段（契约改动走后端 OpenAPI 流程）。
3. 未检索 Veltra / cat-kit 就引入重复能力的第三方 UI/工具库。
