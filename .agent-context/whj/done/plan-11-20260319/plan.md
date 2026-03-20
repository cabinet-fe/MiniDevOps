# 修复构建失败详情缺失与进度误报

> 状态: 已执行

## 目标

修复构建失败后详情页缺少错误信息、构建日志和分支信息的问题，并纠正构建进度在失败场景下错误显示为全部成功的状态。

## 内容

### 步骤 1：定位构建详情数据缺失链路

- 检查构建创建、执行、失败收敛和详情查询链路，确认分支、错误信息、日志路径等字段在失败场景中的写入是否完整。
- 排查构建日志读取接口和 WebSocket 日志推送逻辑，确认失败构建结束后日志是否仍可被详情页读取。

### 步骤 2：修复失败构建状态与详情展示

- 修复后端在失败场景下对构建记录、日志文件、错误字段和阶段状态的持久化问题。
- 修复前端构建详情页对分支、错误信息、日志和阶段进度的展示逻辑，确保失败时准确反映失败节点而不是全部成功。

### 步骤 3：验证修复结果

- 运行相关后端测试和前端类型检查/构建，验证构建详情、日志与进度展示链路无回归。
- 补充必要的回归测试，覆盖失败构建详情和阶段状态计算。

## 影响范围

- `internal/engine/pipeline.go`
- `internal/engine/pipeline_test.go`
- `internal/model/build.go`
- `internal/service/build_service.go`
- `internal/service/build_service_test.go`
- `web/src/components/build-log-viewer.tsx`
- `web/src/lib/api.ts`
- `web/src/pages/builds/detail.tsx`

## 历史补丁
