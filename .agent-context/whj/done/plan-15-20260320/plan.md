# 仪表盘资源与趋势重构

> 状态: 已执行

## 目标

为仪表盘增加系统资源展示能力，包括 CPU 利用率、内存使用情况与剩余硬盘容量；将现有构建趋势图从 Recharts 重构为 ECharts 6；基于“未来感指挥中心”风格重新设计整个仪表盘页面，使页面在桌面端与移动端都具备更强的信息层级、视觉辨识度与可读性。

## 内容

1. 梳理现有仪表盘前后端数据流与页面结构，补充系统资源数据模型与 API，确保接口返回 CPU、内存、磁盘剩余空间等仪表盘所需字段。
2. 在后端实现系统资源采集与路由接入，保持响应格式与现有 `/api/v1/dashboard/*` 接口一致，并补充必要的容错逻辑。
3. 在前端引入并接入 ECharts 6，替换现有构建趋势图实现，重构趋势数据映射与图表配置，使趋势卡片能清晰表达近 7 天构建成功、失败与总量走势。
4. 按“未来感指挥中心”视觉方向重做仪表盘页面布局、卡片与图表样式，加入系统资源展示模块，并兼顾移动端排版与现有设计系统的一致性。
5. 运行前端与相关后端验证命令，修复发现的问题；完成后更新计划状态与影响范围。

## 影响范围

- cmd/server/main.go
- internal/handler/build_handler.go
- internal/service/build_service.go
- internal/service/dashboard_metrics.go
- internal/service/dashboard_metrics_test.go
- web/package.json
- web/bun.lock
- web/src/components/dashboard/build-trend-chart.tsx
- web/src/pages/dashboard.tsx

## 历史补丁

- patch-1: 仪表盘收敛与资源轮询修正
