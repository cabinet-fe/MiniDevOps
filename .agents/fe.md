# 前端约定（web）

## 技术栈

| 项     | 技术                                                                   |
| ------ | ---------------------------------------------------------------------- |
| 框架   | Vue 3.5+ / TypeScript / Pinia / Vue Router                             |
| 构建   | Vite+（`vite-plus` / `vp`）                                            |
| UI     | `@veltra/desktop` + styles / icons / utils / directives / compositions |
| 工具库 | `@cat-kit/core`、`@cat-kit/fe`、`@cat-kit/http`、`@cat-kit/tsconfig`   |
| 包管理 | bun                                                                    |
| 目录   | `web/`                                                                 |

## 规范

- 路径别名 `@` → `web/src/`。
- **组件与工具必须优先 `@veltra/*` 与 `@cat-kit/*`**。仅确认库内无合适能力后才自研或引入第三方。
- 本地组件目录：`web/src/components/<name>/`，必须含 `index.ts` + `<name>.vue`；类型 / `defineXxx` 等放可选 `helper.ts`。调用方从 `@/components/<name>` 导入（勿写 `.vue` 路径）。
- 字段与交互形态对齐 `@veltra/desktop` 契约（查 `components/<name>/types.d.ts` 与 `api.md`）。例如：
  - 侧栏：`UGroupNav` 的 `GroupNavGroup`（`title` + `children: NavItem[]`，叶子含 `title` / `path` / `icon`）；由登录 / `/auth/me` 的两层 `menus` 映射，无折叠
  - 表格列：裸 `UTable` 用 `defineTableColumns`；`ProTable` 用 `defineProTableColumns`；分页：`UPaginator`；表单：`UForm` / `UFormItem`
- 权限码：功能 `full_code`（如 `system_users:view`，点号已改为下划线）；`hasPermission` / 路由 `meta.permission` 只认此形态
- 状态：Pinia；权限辅助：composables。
- Token：`access_token` → `@cat-kit/fe` `storage.local` + Bearer。HTTP / 信封 / refresh / JSON 字段约定见 [.agents/api.md](api.md)。
- **禁止**硬编码全量菜单/权限表作为真源；菜单由 `/auth/me` 两层 `menus` 下发（含 `super_admin_only` 过滤）。

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
- **API**：`url`（相对 `/api/v1`，经 `@cat-kit/http`；信封由 envelope 插件解包）+ `:query`（单向 prop，父组件 `reactive` 对象，表格就地读写字段，勿用 `v-model:query`）；`dataPath` 默认 `items`（`o(body).get`）
- **filters 插槽**：过滤表单（勿再写「查询」按钮，组件内置）；`autoQueryFields` 控制哪些字段变更自动查询（文本框走 Enter 或内置查询按钮）
- **模式**：`pagination` / `tree` 互斥，或二者皆无（纯列表）；列/行基于 `UTable`
- 组件内部：请求、加载态（`v-loading`）、空态（`UEmpty`）、分页时 `UPaginator`
- 暴露 `search()` / `reload()` 供手动刷新

分页信封与 query 字段约定见 [.agents/api.md](api.md)。

## 代码风格

- TypeScript / Vue：`vp check`（或项目 lint）
- SFC 文件名：`kebab-case`
- 前端常量：`UPPER_SNAKE_CASE`（对象键 `snake_case`）

## 开发偏好

### 追求代码简洁性

前端表单的字段的定义始终是和后端保持一致的, 不要结构字段再赋值, 非常啰嗦.

```ts
// good
o(form).extend(row);

// bad
o(form).extend({
  field1: row.field1,
  field2: row.field2,
});
```

## 测试

- 关键路径优先 Playwright 冒烟（登录、菜单、构建日志等）：`web/e2e/`，`bunx playwright test`
- API 级冒烟：`make smoke-api-e2e`
- 不做容量/延迟 SLO 验收（见 ROADMAP）

## 禁止事项（前端）

1. 硬编码全量菜单/权限表作为真源后再隐藏。
2. 绕过统一 HTTP 客户端（见 [.agents/api.md](api.md)）。
3. 未检索 Veltra / cat-kit 就引入重复能力的第三方 UI/工具库。
