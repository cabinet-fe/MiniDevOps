# 修复项目运行时崩溃与未实现功能

> 状态: 已执行

## 目标

修复项目中导致页面崩溃的 bug 和补全未实现的功能页面，使全部路由页面可正常访问和使用。

## 内容

### 步骤 1：修复审计日志页面白屏崩溃

- 文件：`web/src/pages/audit-logs.tsx`
- 问题：第 122 行 `<SelectItem value="">全部</SelectItem>` 使用空字符串作为 value，Radix UI Select 禁止空字符串 value（因为空字符串用于清空选择显示 placeholder），导致组件抛出运行时异常，整个页面白屏。
- 修复方案：将"全部"选项的 value 改为 `"all"`，在 `onValueChange` 回调中把 `"all"` 映射回空字符串（即无筛选），保持原有筛选逻辑不变。

### 步骤 2：实现服务器表单页面（新建/编辑）

- 文件：`web/src/pages/servers/form.tsx`
- 问题：当前仅有占位文本 "Server form coming soon"，新建和编辑服务器功能完全不可用。
- 实现方案：参照 `web/src/pages/projects/form.tsx` 的模式，实现完整的服务器新建/编辑表单。
  - 表单字段（对应后端 API `POST /api/v1/servers` 和 `PUT /api/v1/servers/:id`）：
    - name（名称，必填）
    - host（主机地址，必填）
    - port（端口，默认 22）
    - username（用户名，必填）
    - auth_type（认证方式：password / key，必填）
    - password（密码，auth_type=password 时显示）
    - private_key（私钥，auth_type=key 时显示）
    - description（描述，选填）
    - tags（标签，选填）
  - 编辑模式下通过 `GET /api/v1/servers/:id` 加载数据
  - 使用中文 UI 文案，与项目整体风格一致
  - 使用 `@/lib/constants.ts` 中的 `AUTH_TYPES` 常量

### 步骤 3：验证

- 前端 lint 通过（`bun run lint`）
- 前端 TypeScript 编译通过（`bun run build`）
- 浏览器验证：审计日志页面正常渲染、筛选功能正常；服务器新建/编辑表单页面正常渲染

## 影响范围

- `web/src/pages/audit-logs.tsx` — 修复 SelectItem value="" 导致的运行时崩溃
- `web/src/pages/servers/form.tsx` — 完整实现服务器新建/编辑表单

## 历史补丁
