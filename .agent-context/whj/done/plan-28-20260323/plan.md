# 新增仓库凭证模块

> 状态: 已执行

## 目标

将项目的仓库认证方式从「直接填写用户名/密码/Token」改为「选择无需认证或关联已有凭证」。新增独立的凭证管理模块，支持凭证的 CRUD 操作，凭证敏感信息 AES-GCM 加密存储。完成数据库迁移，包括线上版本的迁移指导。

**权限模型**：
- 所有角色均可管理自己的凭证（增删改查）
- 管理员可查看和选择所有人的凭证（用于项目关联），但不能编辑/删除他人凭证
- 非管理员仅能查看和选择自己的凭证
- 凭证敏感字段（密码/Token/密钥）永远不通过 API 返回明文，仅返回是否已设置的标志

## 内容

### 步骤 1：后端 - 凭证模型与数据库迁移

1. 创建 `internal/model/credential.go`，定义 `Credential` 模型：
   - `ID` (uint, PK)
   - `Name` (string, 100, not null) — 凭证名称，用于展示
   - `Type` (string, 20, not null) — 凭证类型：`password`（用户名/密码）、`token`（Token）
   - `Username` (string, 200) — 用户名（password 类型使用）
   - `Password` (string, json:"-", 1000) — 密码或 Token，AES-GCM 加密存储
   - `Description` (string, 500) — 描述
   - `CreatedBy` (uint, not null) — 创建者 ID
   - `CreatedAt` / `UpdatedAt` (time.Time)
   - 唯一约束：`(name, created_by)` 同一用户下凭证名不重复

2. 在 `internal/model/database.go` 的 `AutoMigrate` 中添加 `&Credential{}`。

3. 添加数据迁移逻辑（在 `InitDB` 中 AutoMigrate 之后）：
   - 扫描所有 `repo_auth_type != 'none'` 且 `repo_password != ''` 的 Project
   - 为每个项目创建一个 Credential 记录（名称使用 `项目名-仓库凭证`，去重处理）
   - 将 Project 的 `credential_id` 指向新创建的凭证
   - 清空 Project 的 `repo_username` 和 `repo_password` 字段

4. 修改 `internal/model/project.go` 的 `Project` 模型：
   - 新增 `CredentialID` (*uint, json:"credential_id", nullable) — 关联凭证
   - 保留 `RepoAuthType` 字段，但取值改为 `none` / `credential`
   - 标记 `RepoUsername` 和 `RepoPassword` 为废弃（数据迁移后不再使用，可在后续版本删除）

### 步骤 2：后端 - 凭证 Repository 和 Service

1. 创建 `internal/repository/credential_repo.go`：
   - `Create(credential *Credential) error`
   - `Update(credential *Credential) error`
   - `Delete(id uint) error`
   - `FindByID(id uint) (*Credential, error)`
   - `FindByCreator(createdBy uint) ([]Credential, error)`
   - `FindAll() ([]Credential, error)`
   - `FindByIDs(ids []uint) ([]Credential, error)`

2. 创建 `internal/service/credential_service.go`：
   - `Create(credential *Credential) error` — 加密 Password 字段后存储
   - `Update(credential *Credential) error` — 智能处理密码更新（空则保持不变）
   - `Delete(id uint) error` — 检查是否被项目引用，被引用则拒绝删除
   - `GetByID(id uint) (*Credential, error)` — 不返回解密后的密码
   - `GetByIDWithSecret(id uint) (*Credential, error)` — 内部使用，返回解密后的密码
   - `ListByUser(userID uint, role string) ([]Credential, error)` — admin 返回所有，其他返回自己的
   - `ListForSelect(userID uint, role string) ([]CredentialOption, error)` — 用于项目表单下拉，仅返回 id+name+type

### 步骤 3：后端 - 凭证 Handler 和路由

1. 创建 `internal/handler/credential_handler.go`：
   - `List` GET `/api/v1/credentials` — 返回凭证列表（脱敏）
   - `Create` POST `/api/v1/credentials` — 创建凭证
   - `GetByID` GET `/api/v1/credentials/:id` — 获取单个凭证（脱敏）
   - `Update` PUT `/api/v1/credentials/:id` — 更新凭证（仅创建者可操作）
   - `Delete` DELETE `/api/v1/credentials/:id` — 删除凭证（仅创建者可操作，且未被引用）
   - `ListForSelect` GET `/api/v1/credentials/select` — 项目表单用的精简列表
   - 权限检查：所有接口需认证；编辑/删除操作需校验 `created_by == 当前用户ID`

2. 在 `cmd/server/main.go` 中注册凭证路由，注入依赖。

3. 修改项目相关逻辑：
   - `project_service.go`：创建/更新项目时，若 `repo_auth_type == "credential"`，验证 credential_id 存在
   - `project_handler.go`：项目创建/更新接口接受 `credential_id` 参数
   - `engine/git.go`：克隆/拉取时，根据 `credential_id` 查询凭证获取认证信息
   - `project_handler.go` 中 `ListBranches`：同上，通过凭证获取认证信息

### 步骤 4：前端 - 凭证管理页面

1. 创建 `web/src/pages/credentials/list.tsx`：
   - 凭证列表页面，展示名称、类型、描述、创建者、创建时间
   - 支持新建、编辑、删除操作
   - 非管理员仅显示自己的凭证
   - 管理员可看到所有凭证，但编辑/删除按钮仅自己的凭证可用
   - 删除前确认，若被项目引用则提示无法删除

2. 在 `web/src/pages/credentials/list.tsx` 中内嵌凭证表单 Dialog：
   - 凭证名称（必填）
   - 凭证类型选择：用户名/密码、Token
   - 用户名（password 类型显示）
   - 密码/Token（必填，编辑时留空保持不变）
   - 描述（可选）

3. 更新 `web/src/lib/constants.ts`：新增凭证类型常量。

### 步骤 5：前端 - 更新项目表单和菜单

1. 修改 `web/src/pages/projects/form.tsx`：
   - `repo_auth_type` 选项改为：`无需认证` / `凭证`
   - 选择「凭证」时，显示凭证下拉选择框（调用 `/api/v1/credentials/select`）
   - 移除直接输入用户名/密码/Token 的表单字段
   - 提交时传 `credential_id` 而非 `repo_username` / `repo_password`

2. 修改 `web/src/components/layout/sidebar.tsx`：
   - 在「资源」分组中添加凭证菜单项：`{ path: '/credentials', label: '凭证', icon: KeyRound, roles: [] }`

3. 修改 `web/src/App.tsx`：添加凭证页面路由。

### 步骤 6：数据库迁移与线上部署指导

1. 确保 AutoMigrate 正确处理：
   - 新增 `Credential` 表
   - Project 表新增 `credential_id` 字段
   - 数据迁移：已有项目的仓库凭证自动转为 Credential 记录

2. 线上迁移方案：
   - GORM AutoMigrate 在应用启动时自动执行，无需手动 SQL
   - 数据迁移逻辑也写在 InitDB 中，幂等执行（通过检查是否已迁移来避免重复）
   - 部署步骤：停旧版本 → 备份数据库 → 部署新版本 → 启动（自动迁移）
   - 回滚方案：恢复备份的数据库文件 + 回滚到旧版本二进制

## 影响范围

- `README.md`
- `cmd/server/main.go`
- `internal/model/credential.go`
- `internal/model/project.go`
- `internal/model/database.go`
- `internal/repository/credential_repo.go`
- `internal/repository/project_repo.go`
- `internal/service/credential_service.go`
- `internal/service/project_service.go`
- `internal/handler/credential_handler.go`
- `internal/handler/project_handler.go`
- `internal/engine/pipeline.go`
- `internal/engine/git.go`
- `internal/engine/git_platform.go`
- `internal/engine/git_test.go`
- `web/src/lib/constants.ts`
- `web/src/pages/credentials/list.tsx`
- `web/src/pages/projects/form.tsx`
- `web/src/pages/projects/detail.tsx`
- `web/src/components/layout/sidebar.tsx`
- `web/src/App.tsx`

## 历史补丁

- patch-1: 策略模式重构 Git 多平台 Token 认证
