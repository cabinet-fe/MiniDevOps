# shadcn/ui 4.0 迁移

> 状态: 已执行

## 目标

将 shadcn/ui 从当前 CLI v3.8.4 迁移到 v4.0，更新所有依赖和组件代码以匹配新版本要求。

## 内容

### 步骤 1：调研 shadcn 4.0 变更

- 查阅 shadcn/ui 4.0 changelog 和迁移指南
- 确认 components.json schema 变化、依赖变化、组件 API 变化
- 确认当前 18 个 UI 组件中哪些需要调整

### 步骤 2：升级 shadcn CLI 与核心依赖

- 升级 `shadcn` devDependency 到 v4.0+
- 根据迁移指南更新 `radix-ui` 等相关依赖版本
- 更新 `web/components.json` 配置格式（如有变化）

### 步骤 3：迁移 UI 组件

- 对 `web/src/components/ui/` 下全部 18 个组件逐一检查和更新
- 优先使用 `shadcn diff` 或 `shadcn add --overwrite` 自动迁移
- 手动处理自动迁移无法覆盖的 breaking changes

### 步骤 4：修复引用与类型

- 全局检查组件 import 路径是否需要调整
- 修复因 API 变更导致的类型错误和运行时问题

### 步骤 5：验证

- TypeScript 编译通过：`cd web && bun run build`
- Lint 通过：`cd web && bun run lint`
- 各页面 UI 渲染正常，无视觉回退

## 影响范围

- `web/package.json` — `shadcn` devDependency 从 `^3.8.4` 升级到 `^4.0.8`
- `web/bun.lock` — 锁文件自动更新

## 历史补丁
