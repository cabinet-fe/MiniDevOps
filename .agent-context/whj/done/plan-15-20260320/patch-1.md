# 仪表盘收敛与资源轮询修正

## 补丁内容

修正上一版仪表盘中存在的三类问题：

1. 去除偏示例化、概念化的英文文案，改回应用仪表盘应有的业务化表达。
2. 收紧整体布局，移除过大的头图区和重复资源卡片，统一为一套深色业务卡片体系，保证视觉语言一致。
3. 新增独立的 `/api/v1/dashboard/system-resources` 接口，前端以 5 秒轮询方式刷新 CPU、内存和磁盘指标，避免资源数据停留在首屏快照。

## 影响范围

- 修改文件: `cmd/server/main.go`
- 修改文件: `internal/handler/build_handler.go`
- 修改文件: `internal/service/build_service.go`
- 修改文件: `web/src/components/dashboard/build-trend-chart.tsx`
- 修改文件: `web/src/pages/dashboard.tsx`
