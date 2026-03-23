# 自托管字体（替代 Google Fonts）

## 补丁内容

在 `web/src/index.css` 中为 IBM Plex Sans（400 / 500 / 700）与 JetBrains Mono（400 / 500）添加 `@font-face`，`src` 指向 `web/src/assets/fonts/` 下已提交的 woff2 文件，与现有 `--font-sans` / `--font-mono` 名称一致，构建时由 Vite 打包为同源资源，不再依赖 `fonts.googleapis.com` / `fonts.gstatic.com`。

## 影响范围

- 修改文件: `web/src/index.css`
