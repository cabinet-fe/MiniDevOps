# 环境变量增强、项目分组与多平台 Webhook

> 状态: 未执行

## 目标

增强环境变量管理能力（加密变量、变量组复用），增加项目分组/标签功能改善项目管理体验，扩展 Webhook 支持更多 Git 平台。

## 内容

### 步骤 1：加密环境变量

- 重构环境变量存储方式，从 JSON 字符串改为独立的 `EnvVar` 模型：
  - 字段：`ID, EnvironmentID, Key, Value, IsSecret, CreatedAt, UpdatedAt`。
  - `IsSecret` 为 true 时，`Value` 使用 AES 加密存储，API 返回时值显示为 `***`。
- 新增环境变量 CRUD 接口：
  - `GET /projects/:id/envs/:envId/vars`
  - `POST /projects/:id/envs/:envId/vars`
  - `PUT /projects/:id/envs/:envId/vars/:varId`
  - `DELETE /projects/:id/envs/:envId/vars/:varId`
- Pipeline 执行时从 `EnvVar` 表读取变量并注入。
- 前端环境变量从文本框改为键值对列表 UI，支持添加/删除/编辑，每个变量可标记"加密"。

### 步骤 2：变量组复用

- 新增 `VarGroup` 模型和 `VarGroupItem` 模型，支持创建全局变量组。
- `Environment` 可关联多个变量组。
- 构建时合并变量组和环境专有变量（环境变量优先级高于变量组）。
- 新增变量组管理页面（系统设置下或独立页面）。

### 步骤 3：项目分组与标签

- `Project` 模型新增 `Tags string`（逗号分隔标签）和 `GroupName string` 字段。
- 项目列表页增加：
  - 按标签筛选的 Filter。
  - 按分组聚合显示（可折叠的分组视图）。
  - 搜索框（项目名称、描述、标签模糊搜索）。
- 项目创建/编辑表单增加标签和分组输入。
- Dashboard 统计按分组展示概览。

### 步骤 4：多 Git 平台 Webhook 支持

- 扩展 `webhook_handler.go`，增加以下平台的 push payload 解析：
  - **Gitea**：格式与 GitHub 类似但有差异（`X-Gitea-Event` header）。
  - **Bitbucket**：使用 `X-Event-Key` header，payload 结构不同。
  - **通用 JSON Webhook**：允许用户自定义 payload 中 ref、commit 字段的 JSONPath 映射。
- 根据请求 Header 自动检测 Git 平台类型。
- 前端 Webhook 配置区域展示各平台的配置指引。

### 步骤 5：验证

- 加密变量存储值在数据库中为密文，API 返回 `***`，构建时正确注入明文。
- 变量组正确合并到构建环境。
- 项目标签筛选和搜索功能正常。
- Gitea/Bitbucket Webhook 可正确触发构建。

## 影响范围

## 历史补丁
