# Windows 构建脚本执行

> 状态: 已执行

## 目标

在 `runtime.GOOS == windows` 时不再默认调用不存在的 `sh`，使本机构建可执行；为无法用 Bash 的环境提供 PowerShell / CMD 脚本类型，语义上接近 `set -e` 与多行命令顺序执行（失败即停）。

## 内容

1. 在 `internal/engine` 中抽取/实现构建脚本解析：Unix 仍为 `sh -c`；Windows 上 `bash` 默认先 `LookPath("bash")` 再 `LookPath("sh")`，均失败则返回明确错误提示安装 Git Bash 或改用 powershell/cmd。
2. 新增 `powershell`：`powershell.exe` 或 `pwsh`（LookPath），用临时 `.ps1` + `-File` 执行（支持多行），参数含 `-NoProfile`、`-NonInteractive`、`-ExecutionPolicy` `Bypass`。
3. 新增 `cmd`：`cmd.exe /C` + 临时 `.cmd`（多行），或在单行场景下直接 `/C`（实现时统一用临时文件以简化引号与换行）。
4. 更新 `web/src/lib/constants.ts` 的 `BUILD_SCRIPT_TYPES` 与 `environment-form.tsx` 的 CodeMirror 扩展（powershell 可用 javascript 或纯文本）；`detail.tsx` 的 `getScriptTypeLabel`。
5. 运行 `go test ./internal/engine/...`，必要时为 `newBuildScriptCommand` 添加 `_test.go`（build tag 或纯逻辑测试）。

## 影响范围

- `internal/engine/build_script_cmd.go`（新建）
- `internal/engine/build_script_cmd_test.go`（新建）
- `internal/engine/pipeline.go`
- `web/src/lib/constants.ts`
- `web/src/pages/projects/environment-form.tsx`

## 历史补丁
