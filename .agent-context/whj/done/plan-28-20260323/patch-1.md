# 策略模式重构 Git 多平台 Token 认证

## 补丁内容

将 `buildAuthURL` 中硬编码的 token 认证逻辑重构为策略模式，支持 GitHub、GitLab、Gitee、Gitea 四个平台的差异化 Token 认证：

- **GitHub**: 默认使用 `x-access-token` 作为用户名
- **GitLab**: 默认使用 `oauth2` 作为用户名
- **Gitee**: 默认使用 `oauth2`，但建议填写真实用户名以获得最佳兼容
- **Gitea**: 默认使用 `oauth2`
- **通用**: 未识别平台使用 `oauth2` 兜底

平台通过仓库 URL 的 hostname 自动检测，支持自建实例（如 `gitlab.mycompany.com` 自动识别为 GitLab）。用户在凭证中填写的 username 始终优先于平台默认值。

前端同步调整：token 类型凭证新增可选的用户名输入框（Gitee 等平台需要），列表页也展示 token 凭证的用户名。

## 影响范围

- 新增文件: `internal/engine/git_platform.go`
- 新增文件: `internal/engine/git_test.go`
- 修改文件: `internal/engine/git.go`
- 修改文件: `web/src/pages/credentials/list.tsx`
