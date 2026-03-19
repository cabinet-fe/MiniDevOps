# 前端工具链迁移：Vite 8 + oxlint + bun

> 状态: 已执行

## 目标

将前端工程的核心工具链统一升级：Vite 7 → 8.0、ESLint → oxlint、npm → bun。迁移后前端项目可正常 dev / build / lint，后端嵌入构建流程不受影响。

## 内容

### 步骤 1：包管理器迁移 npm → bun

1. 删除 `web/node_modules/` 和 `web/package-lock.json`（如存在）。
2. 在 `web/` 下运行 `bun install` 生成 `bun.lock`。
3. 更新 `web/.gitignore`，确认 `node_modules/` 已忽略、`bun.lock` 不被忽略。
4. 更新 `web/package.json` 的 `scripts`，将需要 npm 特定行为的命令适配 bun（当前脚本均为 vite/tsc 直接调用，预计无需改动）。
5. 验证 `bun run dev` 和 `bun run build` 正常运行。

### 步骤 2：Vite 7 → 8.0

1. 运行 `bun add -D vite@^8.0.0`，升级 Vite 到 8.0。
2. 同步升级 Vite 生态插件至兼容版本：
   - `@vitejs/plugin-react` → 适配 Vite 8 的最新版本。
   - `@tailwindcss/vite` → 适配 Vite 8 的最新版本。
3. 查阅 Vite 8 迁移指南，检查 `vite.config.ts` 是否存在破坏性变更需要适配：
   - `defineConfig` API 变化。
   - `server.proxy` 配置格式变化。
   - `resolve.alias` 配置变化。
4. 检查 `tsconfig.app.json` 中 `"types": ["vite/client"]` 是否仍然有效。
5. 验证 `bun run dev` 启动正常，`bun run build` 产物正确。

### 步骤 3：ESLint → oxlint

1. 卸载 ESLint 全家桶：
   - `eslint`、`@eslint/js`、`eslint-plugin-react-hooks`、`eslint-plugin-react-refresh`、`typescript-eslint`、`globals`
2. 安装 oxlint：`bun add -D oxlint`。
3. 删除 `web/eslint.config.js`。
4. 创建 oxlint 配置文件（如需要，oxlint 默认规则已覆盖常见场景，可零配置启用；如需定制则创建 `oxlintrc.json`）。
5. 更新 `web/package.json` 中 `scripts.lint`：`"lint": "oxlint"` 替代原有 `"lint": "eslint ."`。
6. 运行 `bun run lint` 验证无阻断性错误。

### 步骤 4：更新项目元文件

1. 更新 `Makefile`：所有 `npm run` 替换为 `bun run`、`npm` 替换为 `bun`。
2. 更新 `AGENTS.md`：
   - 技术栈表格中 Lint 行改为 oxlint，构建工具 Vite 版本改为 8.x，新增 bun 作为包管理器。
   - 常用命令中 `npm` 替换为 `bun`。
   - 前端规范中的 ESLint 引用改为 oxlint。
3. 更新 `web/README.md`（如涉及 npm/eslint 引用）。

### 步骤 5：最终验证

1. `bun run lint` — 无阻断性错误。
2. `bun run build` — TypeScript 编译 + Vite 构建成功，产物输出到 `web/dist/`。
3. `make build` — 完整构建流程（前端 → 嵌入 → Go 二进制）成功。

## 影响范围

- `web/bun.lock`
- `web/package.json`
- `web/.gitignore`
- `web/eslint.config.js`（删除）
- `web/vite.config.ts`
- `web/README.md`
- `Makefile`
- `AGENTS.md`
- `web/dist/`
- `cmd/server/dist/`
- `buildflow`

## 历史补丁
