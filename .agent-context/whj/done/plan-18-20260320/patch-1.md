# 修复标签选择器嵌套 button 违规

## 补丁内容

项目表单的标签多选组件中，`PopoverTrigger asChild` 包裹了 `<Button>`（渲染为 `<button>`），而 Badge 内的删除按钮也是 `<button>`，导致 HTML 规范违规：`<button> cannot be a descendant of <button>`，React 控制台持续报错。

将 `PopoverTrigger` 的子元素从 `<Button>` 改为 `<div>`，用 CSS 模拟输入框外观（`border-input bg-background` 等 token），保留 `role="combobox"` 和 `tabIndex={0}` 维持可访问性，内部 Badge 删除按钮不受影响。

## 影响范围

- 修改文件: `web/src/pages/projects/form.tsx`
