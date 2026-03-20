# 数据字典模块与项目标签重构

> 状态: 已执行

## 目标

新增数据字典管理模块，支持定义可复用的列表选项（如标签列表）。同时移除项目中的「分组」功能，将标签改为从数据字典中选取，标签输入组件改为选择器。

## 内容

### 步骤 1：后端 — 数据字典模型与数据库迁移

在 `internal/model/` 中创建数据字典相关模型：

- **Dictionary** 模型：`id`, `name`（显示名）, `code`（唯一标识，如 `project_tags`）, `description`, `created_at`, `updated_at`。
- **DictItem** 模型：`id`, `dictionary_id`（外键）, `label`（显示文本）, `value`（存储值）, `sort_order`（排序）, `enabled`（启用/禁用）, `created_at`, `updated_at`。
- GORM AutoMigrate 自动建表。
- 程序启动时自动初始化 `project_tags` 字典（code 为 `project_tags`），如已存在则跳过。

### 步骤 2：后端 — 数据字典 Repository / Service / Handler

按项目分层架构实现完整 CRUD：

- **Repository**（`internal/repository/dict_repo.go`）：
  - `ListDictionaries`, `FindDictionaryByID`, `FindDictionaryByCode`
  - `CreateDictionary`, `UpdateDictionary`, `DeleteDictionary`
  - `ListItemsByDictID`, `CreateItem`, `UpdateItem`, `DeleteItem`, `ReorderItems`
- **Service**（`internal/service/dict_service.go`）：
  - 业务编排，删除字典时级联删除字典项
  - `GetItemsByCode(code)` — 前端通过 code 获取可用选项列表
- **Handler**（`internal/handler/dict_handler.go`）：
  - RESTful API 端点：
    - `GET /api/v1/dictionaries` — 列表
    - `POST /api/v1/dictionaries` — 创建（admin）
    - `GET /api/v1/dictionaries/:id` — 详情
    - `PUT /api/v1/dictionaries/:id` — 更新（admin）
    - `DELETE /api/v1/dictionaries/:id` — 删除（admin）
    - `GET /api/v1/dictionaries/:id/items` — 字典项列表
    - `POST /api/v1/dictionaries/:id/items` — 创建字典项（admin）
    - `PUT /api/v1/dictionaries/:id/items/:itemId` — 更新字典项（admin）
    - `DELETE /api/v1/dictionaries/:id/items/:itemId` — 删除字典项（admin）
    - `GET /api/v1/dictionaries/code/:code/items` — 按 code 获取启用的字典项（所有角色可访问）
- 路由注册到 `cmd/server/main.go`。

### 步骤 3：后端 — 项目模型重构

- 从 `Project` 模型中移除 `GroupName` 字段（`internal/model/project.go`）。
- 使用 GORM 的 `Migrator().DropColumn()` 在迁移时安全移除数据库列。
- 更新 `ProjectRepo`（`internal/repository/project_repo.go`）：移除 `ListAll` 中按 `group_name` 排序的逻辑。
- 更新 `ProjectService`（`internal/service/project_service.go`）：移除导入/导出中的 `GroupName` 字段。
- 更新 `ProjectHandler`（`internal/handler/project_handler.go`）：移除创建/更新请求中的 `group_name` 字段。
- `Tags` 字段保留为逗号分隔字符串存储（与当前一致），但值限制为数据字典中定义的标签。

### 步骤 4：前端 — 数据字典管理页面

- 创建数据字典管理页面（`web/src/pages/dictionaries/`）：
  - 字典列表页：展示所有字典，支持新增、编辑、删除。
  - 字典详情/编辑页：管理字典项列表，支持新增、编辑、删除、排序、启用/禁用。
- 在侧边栏（`web/src/components/layout/sidebar.tsx`）系统管理区域添加「数据字典」菜单项。
- 在 `App.tsx` 中注册路由。

### 步骤 5：前端 — 项目页面重构

- **项目列表页**（`web/src/pages/projects/list.tsx`）：
  - 移除按 `group_name` 分组展示的逻辑（`groupedProjects`、折叠展开等）。
  - 改为平铺列表或卡片展示，保留标签筛选功能。
  - 标签筛选改为从 `/api/v1/dictionaries/code/project_tags/items` 获取选项。
- **项目表单**（`web/src/pages/projects/form.tsx`）：
  - 移除「项目分组」输入框。
  - 「标签」从自由文本输入改为多选选择器组件（基于 shadcn/ui 的 MultiSelect 或 Combobox），选项来源为 `project_tags` 字典。
- **项目详情页**（`web/src/pages/projects/detail.tsx`）：
  - 移除分组相关的 Badge 和 CompactMeta 展示。
  - 标签展示保留（Badge 形式）。

## 影响范围

- `internal/model/dictionary.go` — 新增 Dictionary、DictItem 模型
- `internal/model/database.go` — AutoMigrate 新模型 + DropColumn group_name + 种子 project_tags 字典
- `internal/model/project.go` — 移除 GroupName 字段
- `internal/repository/dict_repo.go` — 新增字典数据访问层
- `internal/repository/project_repo.go` — ListAll 移除 group_name 排序
- `internal/service/dict_service.go` — 新增字典业务层
- `internal/service/project_service.go` — 移除 Export/Import 中 GroupName
- `internal/service/build_service.go` — GroupSummary 改为 TagSummary
- `internal/handler/dict_handler.go` — 新增字典 HTTP handler
- `internal/handler/project_handler.go` — 创建/更新请求移除 group_name
- `cmd/server/main.go` — 注册字典 repo/service/handler 及路由
- `web/src/pages/dictionaries/list.tsx` — 新增数据字典管理页面
- `web/src/pages/projects/list.tsx` — 移除分组逻辑，改为平铺展示 + 字典标签筛选
- `web/src/pages/projects/form.tsx` — 移除分组输入，标签改为多选选择器
- `web/src/pages/projects/detail.tsx` — 移除分组 Badge 和 CompactMeta
- `web/src/components/layout/sidebar.tsx` — 添加「数据字典」菜单项
- `web/src/App.tsx` — 注册 /dictionaries 路由

## 历史补丁

- patch-1: 修复标签选择器嵌套 button 违规
