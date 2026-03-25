import { useEffect, useRef } from "react";
import * as echarts from "echarts";

export interface BuildTrendPoint {
  date: string;
  success: number;
  failed: number;
  total: number;
}

interface BuildTrendChartProps {
  data: BuildTrendPoint[];
}

export function BuildTrendChart({ data }: BuildTrendChartProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const chartRef = useRef<echarts.ECharts | null>(null);

  useEffect(() => {
    if (!containerRef.current || data.length === 0) return;

    const chart =
      echarts.getInstanceByDom(containerRef.current) ?? echarts.init(containerRef.current);
    chartRef.current = chart;

    const observer = new ResizeObserver(() => chart.resize());
    observer.observe(containerRef.current);

    return () => {
      observer.disconnect();
      chart.dispose();
      chartRef.current = null;
    };
  }, [data.length]);

  useEffect(() => {
    if (!containerRef.current || data.length === 0) return;

    const chart = chartRef.current ?? echarts.init(containerRef.current);
    chartRef.current = chart;

    const axisColor = "rgba(148, 163, 184, 0.72)";
    const splitLineColor = "rgba(52, 211, 153, 0.12)";
    const tooltipBg = "rgba(2, 6, 23, 0.94)";
    const tooltipBorder = "rgba(56, 189, 248, 0.24)";
    const textColor = "#e2e8f0";

    chart.setOption({
      animationDuration: 600,
      animationEasing: "elasticOut",
      grid: { top: 40, right: 16, bottom: 24, left: 8, containLabel: true },
      legend: {
        top: 0,
        right: 0,
        itemWidth: 14,
        itemHeight: 8,
        itemGap: 16,
        textStyle: {
          color: axisColor,
          fontSize: 12,
        },
      },
      tooltip: {
        trigger: "axis",
        backgroundColor: tooltipBg,
        borderColor: tooltipBorder,
        borderWidth: 1,
        textStyle: {
          color: textColor,
          fontSize: 12,
        },
        axisPointer: {
          type: "cross",
          crossStyle: { color: "rgba(148, 163, 184, 0.3)" },
        },
      },
      xAxis: {
        type: "category",
        data: data.map((item) => item.date),
        axisLine: {
          lineStyle: { color: splitLineColor },
        },
        axisTick: { show: false },
        axisLabel: {
          color: axisColor,
          fontSize: 12,
        },
      },
      yAxis: {
        type: "value",
        splitNumber: 4,
        axisLabel: {
          color: axisColor,
          fontSize: 12,
        },
        axisLine: { show: false },
        axisTick: { show: false },
        splitLine: {
          lineStyle: {
            color: splitLineColor,
            type: "dashed",
          },
        },
      },
      series: [
        {
          name: "成功",
          type: "bar",
          barMaxWidth: 28,
          barGap: "30%",
          itemStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: "#6ee7b7" },
              { offset: 1, color: "#059669" },
            ]),
            borderRadius: [4, 4, 0, 0],
          },
          emphasis: {
            itemStyle: {
              color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                { offset: 0, color: "#a7f3d0" },
                { offset: 1, color: "#34d399" },
              ]),
              shadowColor: "rgba(52, 211, 153, 0.4)",
              shadowBlur: 12,
            },
          },
          data: data.map((item) => item.success),
        },
        {
          name: "失败",
          type: "bar",
          barMaxWidth: 28,
          barGap: "30%",
          itemStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: "#fda4af" },
              { offset: 1, color: "#e11d48" },
            ]),
            borderRadius: [4, 4, 0, 0],
          },
          emphasis: {
            itemStyle: {
              color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                { offset: 0, color: "#fecdd3" },
                { offset: 1, color: "#fb7185" },
              ]),
              shadowColor: "rgba(251, 113, 133, 0.4)",
              shadowBlur: 12,
            },
          },
          data: data.map((item) => item.failed),
        },
      ],
    });
  }, [data]);

  if (data.length === 0) {
    return (
      <div className="flex h-[180px] items-center justify-center rounded-lg border border-border bg-muted/30 text-sm text-muted-foreground">
        暂无趋势数据
      </div>
    );
  }

  return <div ref={containerRef} className="h-[180px] w-full" />;
}
