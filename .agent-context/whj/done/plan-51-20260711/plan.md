# 操作手册迁移与列表排序交互

> 状态: 已执行

## 目标

1. 将「项目手册」更名为「操作手册」，并从「概览」移至「管理」导航分组。
2. 进入系统管理页时默认发起一次进程列表请求，无需先点「查询」。
3. 所有带排序的列表：排序控件放在表头文本右侧，点击循环「不排序 → 降序 → 升序 → 不排序」。

## 内容

1. **操作手册导航与文案**
   - `sidebar.tsx`：从「概览」移除手册项，加入「管理」分组，标签改为「操作手册」。
   - `header.tsx`：面包屑文案改为「操作手册」。
   - `project-manual.tsx`：页面主标题改为「操作手册」。
   - 路由仍为 `/manual`，无需改路径。

2. **系统管理默认查询**
   - `system/processes.tsx`：挂载时 `useEffect` 调用一次 `fetchProcesses(1)`，去掉「必须先点查询」的空态依赖。

3. **表头排序交互（系统进程列表）**
   - 移除筛选区「排序 / 顺序」两个 Select。
   - 在可排序列（CPU%、内存，以及后端已支持的名称）表头文本右侧放置排序按钮。
   - 状态机：`null`（不排序）→ `desc` → `asc` → `null`；点击另一列时从该列 `desc` 开始。
   - 排序变化后自动重新请求第 1 页。
   - 后端 `sortProcesses`：当 `sort` 为空或不识别且非默认时跳过排序；handler 在未传 `sort` 时不强制默认 `cpu`（或前端不传 sort/order 表示不排序）。
   - 仪表盘「系统进程」为 Top-N 指标切换（非三态排序），保持现有「按 CPU / 按内存」切换，不纳入本交互。

## 影响范围

- `web/src/components/layout/sidebar.tsx`
- `web/src/components/layout/header.tsx`
- `web/src/pages/project-manual.tsx`
- `web/src/pages/system/processes.tsx`
- `internal/service/process_service.go`
- `internal/service/process_service_test.go`
- `internal/handler/system_handler.go`

## 历史补丁
