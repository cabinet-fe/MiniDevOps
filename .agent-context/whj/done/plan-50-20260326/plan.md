# 更新项目依赖

> 状态: 已执行

## 目标

检索当前项目的所有依赖，包括后端（Go）和前端（Node/Bun）。确保依赖是最新的，并移除了所有未被使用的依赖。如果包含大版本的跳跃更新，需检查并修正因此产生的破坏性变更。

## 内容

1. **后端依赖（Go）更新与清理**：
   - 分析 `go.mod`。
   - 运行 `go get -t -u ./...` 升级所有直接依赖和间接依赖。
   - 运行 `go mod tidy` 清理和同步依赖。
   - 运行 `go test ./...` 或 `go build ./...` 确认依赖更新未引发破坏性变更，若有则进行代码修正。

2. **前端依赖（React / Bun）更新与清理**：
   - 进入 `web` 目录，安装 `npm-check-updates` 和 `depcheck`。
   - 运行 `depcheck`，检查是否存在未使用的依赖并从 `package.json` 中移除。
   - 运行 `npx npm-check-updates -u` 强行将 `package.json` 中的依赖全部更新到最新版本。
   - 运行 `bun install` 生成新的 lock 文件。
   - 运行 `bun run build` 及 `bun run lint`，检查是否有新版本的破坏性变更导致编译或类型检查失败。若有报错，逐一修正对应代码。

3. **集成验证**：
   - 根目录下运行 `make build`，确保整个项目的前后端均可正确构建和内嵌。

## 影响范围

- \`go.mod\`
- \`go.sum\`
- \`web/package.json\`
- \`web/bun.lock\`
- \`web/src/pages/project-manual.tsx\`

## 历史补丁
