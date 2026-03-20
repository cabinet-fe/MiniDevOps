# 仪表板字体放大、图表样式变更与资源轮询加速

## 补丁内容

### 1. 字体大小提升

将仪表板页面各处过小字体统一上调一档：

- 统计行：标签 `10px → 12px (text-xs)`，数值 `text-2xl → text-3xl`
- 系统资源：标签 `11px → 12px`，数值 `text-xs → text-sm`
- 运行中构建：名称 `text-xs → text-sm`，阶段 `11px → 12px`
- 最近构建表格：正文 `text-xs → text-sm`，表头 `11px → 12px`
- 各处辅助文字和 Badge 同步微调

### 2. 构建趋势图表样式变更

- 从堆叠窄柱 (barWidth: 10, stacked) 改为并排分组宽柱 (barMaxWidth: 28, grouped)
- 柱体填充从纯色改为顶浅底深的线性渐变（成功: emerald 渐变, 失败: rose 渐变）
- 圆角从全圆 (999px) 改为 4px 矩形倒角
- 移除总量折线系列，保持双柱清晰对比
- Tooltip 指示器从阴影改为十字线 (cross)
- 分割线改为虚线风格
- 图例、轴标签字号从 11px 提升至 12px
- 空态容器改用语义色 token

### 3. 系统资源轮询间隔

`POLL_INTERVAL_MS` 从 5000ms 缩短为 2000ms，提升资源监控实时性。

## 影响范围

- 修改文件: `web/src/pages/dashboard.tsx`
- 修改文件: `web/src/components/dashboard/build-trend-chart.tsx`
