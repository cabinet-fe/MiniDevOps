# 侧栏菜单补全与项目构建配置展示

## 补丁内容

侧栏缺少「构建」菜单入口，用户无法直接访问全局构建列表。项目详情页环境信息区域未显示 build_script 和 build_output_dir 字段。

修改内容：
1. 后端新增 `GET /api/v1/builds` 全局构建列表接口，返回带项目名和环境名的构建记录分页列表
2. 前端新增构建列表页面 `/builds`，展示所有项目的构建历史
3. 侧栏「资源」分组新增「构建」菜单项（Hammer 图标）
4. 项目详情页环境信息区域新增「构建脚本」和「产物目录」字段展示

## 影响范围

- 新增文件: `web/src/pages/builds/list.tsx`
- 修改文件: `internal/repository/build_repo.go`
- 修改文件: `internal/service/build_service.go`
- 修改文件: `internal/handler/build_handler.go`
- 修改文件: `cmd/server/main.go`
- 修改文件: `web/src/App.tsx`
- 修改文件: `web/src/components/layout/sidebar.tsx`
- 修改文件: `web/src/pages/projects/detail.tsx`
