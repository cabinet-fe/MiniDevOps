# 项目手册页面

> 状态: 已执行

## 目标

在前端增加「项目手册」独立页面，以中文向用户说明：如何新建项目（重点讲解仓库认证方式，并给出 GitHub、码云的可操作教程）、如何新建与配置环境、部署流程与原理（含 Rsync/SFTP/SCP/Agent 的差异），以及如何独立运行部署 Agent 服务。页面需纳入主导航，风格与现有仪表盘/设置页一致。

## 内容

1. **路由与入口**
   - 在 `web/src/App.tsx` 注册受保护路由（建议路径 `/manual`），懒加载或同步引入新页面组件。
   - 在 `web/src/components/layout/sidebar.tsx` 的合适分组（建议在「概览」下与仪表盘并列，或单独「帮助」分组）增加「项目手册」菜单项及图标，所有登录角色可见。

2. **页面结构与样式**
   - 新增页面组件（建议 `web/src/pages/project-manual.tsx`，文件名 kebab-case）。
   - 使用现有布局（与 `AppLayout` 一致）、`Card` / 标题层级 / `Separator`，长文可用 `ScrollArea`；正文区域使用 Tailwind `prose`（若项目已依赖 `@tailwindcss/typography`）或沿用现有页面的 `text-muted-foreground`、`space-y` 排版，保证可读性与移动端适配。
   - 提供页内目录（锚点跳转）：「新建项目」「仓库认证（GitHub / 码云）」「新建环境」「部署方式与原理」「独立部署 Agent」。

3. **文档正文要点（须与代码一致）**
   - **新建项目**：说明仓库地址使用 HTTPS、默认分支；引用 `REPO_AUTH_TYPES`（`none` / `用户名+密码` / `Token`）与 `internal/engine/git.go` 中通过 URL 注入凭据的行为（HTTPS 基本认证形式）。
   - **GitHub**：HTTPS + Personal Access Token（Fine-grained 或 classic 的权限范围简述）；在 BuildFlow 中选「Token」时用户名/密码字段如何填写（例如 classic PAT 常用 `username` + `token` 作密码，或平台惯例）；公开仓库可用「无需认证」。
   - **码云（Gitee）**：私人令牌创建入口与在「Token」模式下的填写方式；企业版/私有库注意 HTTPS 与权限。
   - **新建环境**：引导从项目详情进入环境配置；说明分支、构建脚本类型（`BUILD_SCRIPT_TYPES`）、产物格式（`ARTIFACT_FORMATS`）、可选 Cron、变量组等（与 `environment-form` 字段对齐，避免写不存在的能力）。
   - **部署与原理**：简述 Pipeline「克隆 → 构建 → 归档 → 部署」；说明 `DEPLOY_METHODS` 各方式适用场景；**Agent**：主控端通过 HTTP `POST` 上传归档到 Agent 的 `upload` 路径，`Authorization: Bearer <token>`，以及 `X-Target-Path` 等（与 `internal/deployer/agent.go` 一致）。
   - **独立 Agent 服务**：说明 `cmd/agent` 独立二进制、监听地址、必选 token、可选 TLS（`BUILDFLOW_AGENT_ADDR`、`BUILDFLOW_AGENT_TOKEN`、`BUILDFLOW_AGENT_TLS_CERT` / `TLS_KEY` 或与 README/flags 一致）；服务器侧在 BuildFlow 中填写 Agent URL（指向 `upload` 的基地址，与现有 `joinAgentURL` 行为一致）、Agent Token；防火墙与 HTTPS 建议。

4. **验收标准**
   - 登录后侧边栏可进入手册页，路由正确、无控制台报错。
   - 上述章节内容完整、术语与界面中文标签一致，技术细节不误导（尤其认证与 Agent 配置）。
   - `cd web && bun run lint` 与 `bun run build` 通过。

## 影响范围

- `web/src/App.tsx`：注册受保护路由 `/manual`。
- `web/src/components/layout/sidebar.tsx`：「概览」分组增加「项目手册」入口。
- `web/src/pages/project-manual.tsx`：新增项目手册页面（路由、目录锚点与正文）。

## 历史补丁
