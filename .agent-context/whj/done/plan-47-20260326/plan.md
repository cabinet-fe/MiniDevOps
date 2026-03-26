# 控制台可访问性修复与 Webhook 手册

> 状态: 已执行

## 目标

消除浏览器中与 Dialog/Sheet、密码输入、ECharts 相关的控制台警告；在 `AGENTS.md` 中补充 Webhook 使用说明。

## 内容

1. 为移动端侧栏 `SheetContent` 补充 `SheetTitle` / `SheetDescription`（或与 Radix 要求一致的 `aria-describedby`），消除 Radix Dialog 可访问性警告。
2. 为未标注的 `type="password"` 输入补充合适的 `autoComplete`（如 `current-password`、`new-password`）。
3. 在 ECharts 初始化中使用 `renderer: 'svg'` 或官方支持的配置，减轻非 passive 的 `wheel` 监听告警（若仍无法消除则采用文档化说明的最小改动）。
4. 在 `AGENTS.md` 的 API 章节扩展 Webhook：URL 格式、认证、查询参数 `environment_id`、与分支匹配行为。

## 影响范围

- `web/src/components/layout/sidebar.tsx`：移动端 Sheet 可访问性标题与描述。
- `web/src/pages/credentials/list.tsx`：加载态对话框标题/描述；凭证密码 `autoComplete`。
- `web/src/pages/users/list.tsx`、`web/src/pages/servers/form.tsx`、`web/src/pages/settings.tsx`：密码类输入 `autoComplete`。
- `web/src/components/dashboard/build-trend-chart.tsx`：ECharts 使用 SVG 渲染初始化。
- `web/src/pages/project-manual.tsx`：Webhook 章节与目录。
- `AGENTS.md`：API 规范中 Webhook 说明扩展。

## 历史补丁
