# 修复构建日志搜索异常

## 补丁内容

在 `build-log-viewer.tsx` 中初始化 `Terminal` 时增加 `allowProposedApi: true` 配置，以修复开启使用 decorations 特性的搜索功能时由于使用了实验性 API 导致的 `Uncaught Error: You must set the allowProposedApi option to true to use proposed API` 错误。

## 影响范围

- 修改文件: `web/src/components/build-log-viewer.tsx`
