# 修复弹框主体滚动与固定头尾的布局异常

> 状态: 已执行

## 目标

修复当前所有基于通用 `Dialog` 的弹框显示异常问题，使弹框在内容未达到上限时保持自然高度，只有在达到最大高度后才由主体内容区域滚动，同时保留标题区和底部操作区的稳定显示。

## 内容

### 步骤 1：重构通用弹框容器的滚动结构

- 检查 `web/src/components/ui/dialog.tsx` 里当前 `DialogContent`、`DialogHeader`、`DialogFooter` 的布局关系，改为由弹框容器负责高度约束、由中间主体区域负责滚动，而不是让整个弹框根节点滚动。
- 保持现有头部和底部的视觉样式与关闭按钮行为，避免影响短内容弹框的尺寸和交互。

### 步骤 2：收敛页面级弹框的重复滚动样式并验证典型场景

- 检查项目、环境、服务器、设置等长内容弹框的调用方式，移除与新结构冲突的 `overflow-y-auto` / `max-h` 类覆盖，必要时补上主体滚动容器。
- 运行前端 lint 与构建，确认通用弹框、长表单弹框和确认类弹框均可通过检查且没有明显回归。

## 影响范围

- `web/src/components/ui/dialog.tsx`
- `web/src/pages/settings.tsx`
- `web/src/pages/servers/form.tsx`
- `web/src/pages/projects/environment-form.tsx`
- `web/src/pages/projects/form.tsx`
- `web/src/pages/users/list.tsx`
- `web/src/pages/projects/detail.tsx`
- `web/src/hooks/use-websocket.ts`

## 历史补丁
- patch-1: 修复开发环境下 WebSocket 提前关闭告警
