# 页面优化 — 审计日志增强、服务器状态修复、构建页面移除

> 状态: 已执行

## 目标

修复并优化三个存在问题的页面：
1. 审计日志页面：中间件记录信息不完整（resource_type/resource_id 为空、action 为原始 METHOD+PATH 不可读），前端筛选与后端数据不匹配，UI 简陋
2. 服务器页面：Status 字段从未被更新，始终为 unknown，TestConnection 不回写状态
3. 构建列表页：与仪表盘功能高度重叠，移除独立入口

## 内容

### 步骤 1：增强审计中间件（后端）

修改 `internal/middleware/audit.go`，从 HTTP 方法和 URL 路径中解析出：
- `action`：基于 HTTP 方法映射（POST→create, PUT/PATCH→update, DELETE→delete），对登录路径特殊处理为 `login`
- `resource_type`：从路径段提取资源类型（projects→project, servers→server, builds→build, environments→environment, users→user 等）
- `resource_id`：从路径段提取数字 ID

### 步骤 2：优化审计日志前端页面

重写 `web/src/pages/audit-logs.tsx`：
- 操作类型筛选与后端实际存储的 action 值对齐
- 增加 resource_type 筛选
- 用 Badge + 颜色区分操作类型（create=绿色, update=蓝色, delete=红色, login=紫色）
- 将 resource_type 显示为中文标签
- 改善表格布局和空状态展示

### 步骤 3：修复服务器状态（后端）

修改 `internal/service/server_service.go` 的 `TestConnection` 方法：
- 连接成功后将 Status 更新为 `online` 并持久化到数据库
- 连接失败后将 Status 更新为 `offline` 并持久化到数据库

需要给 ServerService 添加对 repo 的直接状态更新方法，避免触发完整的 Update 校验和加密流程。
在 `internal/repository/server_repo.go` 添加 `UpdateStatus(id uint, status string)` 方法。

### 步骤 4：优化服务器列表前端

修改 `web/src/pages/servers/list.tsx`：
- 修复表格渲染 bug（TableBody 中错误使用 `cell.column.columnDef.header` 应为 `cell.column.columnDef.cell`）
- 状态 Badge 增加 online（绿色）、offline（红色）、unknown（灰色）三态显示
- 测试连接按钮增加 loading 状态

### 步骤 5：移除构建列表页入口

- 从 `web/src/components/layout/sidebar.tsx` 的 NAV_GROUPS 中移除构建项
- 从 `web/src/App.tsx` 中移除 BuildListPage 路由（保留 BuildDetailPage 路由 `/builds/:id`）
- 删除 `web/src/pages/builds/list.tsx`

## 影响范围

- `internal/middleware/audit.go` — 增强审计中间件，解析 action/resource_type/resource_id/details
- `internal/repository/server_repo.go` — 新增 UpdateStatus 方法
- `internal/service/server_service.go` — TestConnection 重构，成功/失败后回写 Status
- `internal/handler/server_handler.go` — TestConnection 响应增加 message 字段
- `web/src/pages/audit-logs.tsx` — 重写审计日志页面 UI，修复筛选逻辑
- `web/src/pages/servers/list.tsx` — 状态 Badge 三态显示，测试连接 loading 状态
- `web/src/components/layout/sidebar.tsx` — 移除构建列表导航项
- `web/src/App.tsx` — 移除 BuildListPage 路由
- `web/src/pages/builds/list.tsx` — 已删除

## 历史补丁
