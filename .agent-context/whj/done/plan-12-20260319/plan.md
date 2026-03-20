# 补齐构建通知并重构项目环境页与构建物策略

> 状态: 已执行

## 目标

补齐构建成功后的站内实时通知与前端刷新链路，优化项目环境页的布局和快捷操作，新增项目级构建物格式配置并同步调整 agent 端处理策略，同时统一弹框固定标题与底部操作区，并为项目详情页提供可切换的整体视觉风格。

## 内容

### 步骤 1：梳理并补齐构建通知与构建物格式能力

- 检查后端构建完成后的通知创建、WebSocket 推送和前端通知订阅链路，补齐构建成功场景下的站内自动通知以及相关列表刷新能力。
- 为项目模型、接口和构建流程新增构建物格式配置，支持 `zip` 与 `gzip` 两种格式，并同步调整 `cmd/agent/main.go` 的打包或传输策略。

### 步骤 2：重构项目环境页和弹框交互

- 调整项目详情页环境区域的数据呈现，隐藏具体构建脚本文本，统一信息对齐方式，并在表格中增加下载构建物、部署、重新构建和高亮详情入口等快捷操作，按构建状态控制可用性。
- 统一现有弹框的结构，使标题区和底部操作区在长表单滚动时保持固定可见。

### 步骤 3：用可切换风格重新设计项目页面布局并完成验证

- 基于 `frontend-design` 技能为项目详情页建立完整的页面重设计方案，提供至少两种可切换风格供选择，并确保桌面端与移动端均可正常工作。
- 运行相关后端测试、前端 lint 与构建，确认新增通知、构建格式、页面布局和快捷操作链路可用且无明显回归。

## 影响范围

- `cmd/agent/main.go`
- `internal/deployer/agent.go`
- `internal/deployer/deployer.go`
- `internal/engine/pipeline.go`
- `internal/engine/pipeline_test.go`
- `internal/handler/project_handler.go`
- `internal/model/project.go`
- `internal/service/project_service.go`
- `web/src/components/notification-bell.tsx`
- `web/src/components/ui/dialog.tsx`
- `web/src/hooks/use-websocket.ts`
- `web/src/lib/constants.ts`
- `web/src/pages/projects/detail.tsx`
- `web/src/pages/projects/form.tsx`
- `web/src/stores/notification-store.ts`

## 历史补丁
- patch-1: 收敛项目详情页为 Signal Deck 并压缩环境布局
