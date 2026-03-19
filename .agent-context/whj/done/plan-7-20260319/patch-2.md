# 优化环境弹框宽度与代码输入高亮

## 补丁内容

- 调整 `EnvironmentFormDialog` 弹框的宽度，从 `sm:max-w-[560px]` 放大至 `sm:max-w-[800px]`。
- 重新排列网格列布局（采用 3 列布局或 `[160px_1fr]` 的特定宽比例），以便一行能显示更多配置项。
- 引入使用 `@uiw/react-codemirror` 及对应的语法高亮插件。
- 替换原有 `Textarea`，采用 `CodeMirror` 组件支持 `build_script` 脚本（支持 Bash、Node、Python 语法高亮）和 `env_vars` 环境变量（支持 JSON 高亮）的代码级输入体验。

## 影响范围

- 修改文件: `web/src/pages/projects/environment-form.tsx`
- 修改文件: `web/package.json`
