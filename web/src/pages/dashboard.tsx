import { startTransition, useEffect, useState } from "react";
import { Link } from "react-router";
import {
  ArrowUpRight,
  Clock3,
  Cpu,
  HardDrive,
  MemoryStick,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { BuildTrendChart, type BuildTrendPoint } from "@/components/dashboard/build-trend-chart";
import { api } from "@/lib/api";
import { BUILD_STATUSES } from "@/lib/constants";
import { cn } from "@/lib/utils";

interface DashboardSystemResources {
  cpu_usage_percent: number;
  memory_used_bytes: number;
  memory_total_bytes: number;
  memory_usage_percent: number;
  disk_free_bytes: number;
  disk_total_bytes: number;
  disk_usage_percent: number;
}

interface DashboardStats {
  total_projects: number;
  today_builds: number;
  success_rate: number;
  active_count: number;
  system_resources: DashboardSystemResources;
}

interface DashboardBuild {
  id: number;
  project_id: number;
  environment_id: number;
  build_number: number;
  status: string;
  current_stage: string;
  trigger_type: string;
  branch: string;
  commit_hash: string;
  commit_message: string;
  duration_ms: number;
  created_at: string;
  project_name: string;
  environment_name: string;
}

interface BuildTrendItem {
  date: string;
  status: string;
  count: number;
}

const EMPTY_RESOURCES: DashboardSystemResources = {
  cpu_usage_percent: 0,
  memory_used_bytes: 0,
  memory_total_bytes: 0,
  memory_usage_percent: 0,
  disk_free_bytes: 0,
  disk_total_bytes: 0,
  disk_usage_percent: 0,
};

const STAGE_LABELS: Record<string, string> = {
  pending: "等待",
  cloning: "拉取",
  building: "构建",
  deploying: "部署",
  distributing: "分发",
  success: "完成",
  failed: "失败",
  cancelled: "取消",
};

const POLL_INTERVAL_MS = 2000;

function formatBytes(bytes: number) {
  if (!bytes) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let value = bytes;
  let unitIndex = 0;
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }
  const digits = value >= 100 || unitIndex === 0 ? 0 : value >= 10 ? 1 : 2;
  return `${value.toFixed(digits)} ${units[unitIndex]}`;
}

function formatDuration(durationMs: number) {
  if (!durationMs) return "-";
  const totalSeconds = Math.round(durationMs / 1000);
  if (totalSeconds < 60) return `${totalSeconds}s`;
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  if (minutes < 60) return `${minutes}m ${seconds}s`;
  const hours = Math.floor(minutes / 60);
  return `${hours}h ${minutes % 60}m`;
}

function formatTimestamp(date: string) {
  return new Date(date).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function CircularProgress({
  percent,
  color,
  size = 48,
  strokeWidth = 4,
  children,
}: {
  percent: number;
  color: string;
  size?: number;
  strokeWidth?: number;
  children?: React.ReactNode;
}) {
  const radius = (size - strokeWidth) / 2;
  const circumference = radius * 2 * Math.PI;
  const offset = circumference - (percent / 100) * circumference;
  return (
    <div
      className="relative inline-flex items-center justify-center shrink-0"
      style={{ width: size, height: size }}
    >
      <svg width={size} height={size} className="rotate-[-90deg]">
        {/* Background */}
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke="currentColor"
          strokeWidth={strokeWidth}
          className="text-muted"
        />
        {/* Progress */}
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke={color}
          strokeWidth={strokeWidth}
          strokeDasharray={circumference}
          strokeDashoffset={offset}
          strokeLinecap="round"
          className="transition-all duration-500 ease-in-out"
        />
      </svg>
      {/* Optional: Icon inside */}
      <div className="absolute inset-0 flex items-center justify-center">{children}</div>
    </div>
  );
}

function getStatusInfo(status: string) {
  return (
    BUILD_STATUSES[status as keyof typeof BUILD_STATUSES] ?? {
      label: status,
      color: "bg-slate-500",
    }
  );
}

function buildTrendData(items: BuildTrendItem[]) {
  const byDate = new Map<string, { success: number; failed: number; total: number }>();
  for (const item of items) {
    const current = byDate.get(item.date) ?? { success: 0, failed: 0, total: 0 };
    current.total += Number(item.count);
    if (item.status === "success") current.success += Number(item.count);
    if (item.status === "failed") current.failed += Number(item.count);
    byDate.set(item.date, current);
  }
  return Array.from(byDate.entries())
    .sort((a, b) => a[0].localeCompare(b[0]))
    .map(([date, value]) => ({
      date: date.slice(5),
      success: value.success,
      failed: value.failed,
      total: value.total,
    }));
}

export function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [resources, setResources] = useState<DashboardSystemResources>(EMPTY_RESOURCES);
  const [activeBuilds, setActiveBuilds] = useState<DashboardBuild[]>([]);
  const [recentBuilds, setRecentBuilds] = useState<DashboardBuild[]>([]);
  const [trend, setTrend] = useState<BuildTrendPoint[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let cancelled = false;
    const fetchDashboard = async () => {
      try {
        const [statsRes, activeRes, recentRes, trendRes] = await Promise.all([
          api.get<DashboardStats>("/dashboard/stats"),
          api.get<DashboardBuild[]>("/dashboard/active-builds"),
          api.get<DashboardBuild[]>("/dashboard/recent-builds?limit=8"),
          api.get<BuildTrendItem[]>("/dashboard/trend?days=7"),
        ]);
        if (cancelled) return;
        const nextStats =
          statsRes.code === 0 && statsRes.data ? (statsRes.data as DashboardStats) : null;
        const nextActive =
          activeRes.code === 0 && Array.isArray(activeRes.data)
            ? (activeRes.data as DashboardBuild[])
            : [];
        const nextRecent =
          recentRes.code === 0 && Array.isArray(recentRes.data)
            ? (recentRes.data as DashboardBuild[])
            : [];
        const nextTrend =
          trendRes.code === 0 && Array.isArray(trendRes.data)
            ? buildTrendData(trendRes.data as BuildTrendItem[])
            : [];
        startTransition(() => {
          setStats(nextStats);
          setResources(nextStats?.system_resources ?? EMPTY_RESOURCES);
          setActiveBuilds(nextActive);
          setRecentBuilds(nextRecent);
          setTrend(nextTrend);
        });
      } catch {
        if (!cancelled) setError("仪表盘数据加载失败");
      } finally {
        if (!cancelled) setLoading(false);
      }
    };
    fetchDashboard();
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    const timer = window.setInterval(async () => {
      try {
        const res = await api.get<DashboardSystemResources>("/dashboard/system-resources");
        if (res.code === 0 && res.data) {
          setResources(res.data as DashboardSystemResources);
        }
      } catch {
        /* keep last snapshot */
      }
    }, POLL_INTERVAL_MS);
    return () => window.clearInterval(timer);
  }, []);

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="size-6 animate-spin rounded-full border-2 border-border border-t-emerald-500" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-lg border border-red-500/20 bg-red-500/8 px-4 py-3 text-sm text-red-400">
        {error}
      </div>
    );
  }

  const summaryItems = [
    { label: "项目总数", value: String(stats?.total_projects ?? 0) },
    { label: "今日构建", value: String(stats?.today_builds ?? 0) },
    { label: "成功率", value: `${(stats?.success_rate ?? 0).toFixed(1)}%` },
    { label: "运行中", value: String(stats?.active_count ?? 0) },
  ];

  const meters = [
    {
      key: "cpu",
      label: "CPU",
      value: `${resources.cpu_usage_percent.toFixed(1)}%`,
      percent: resources.cpu_usage_percent,
      icon: Cpu,
      color: "#34d399",
    },
    {
      key: "memory",
      label: "系统使用内存",
      value: resources.memory_total_bytes
        ? `${formatBytes(resources.memory_used_bytes)} / 总计 ${formatBytes(resources.memory_total_bytes)}`
        : "未采集",
      percent: resources.memory_usage_percent,
      icon: MemoryStick,
      color: "#22d3ee",
    },
    {
      key: "disk",
      label: "磁盘使用",
      value: resources.disk_total_bytes
        ? `${formatBytes(resources.disk_total_bytes - resources.disk_free_bytes)} / 总计 ${formatBytes(resources.disk_total_bytes)}`
        : "未采集",
      percent: resources.disk_total_bytes > 0 ? resources.disk_usage_percent : 0,
      icon: HardDrive,
      color: "#a78bfa",
    },
  ];

  return (
    <div className="space-y-4">
      {/* Stats row */}
      <div className="flex items-stretch border-b border-border pb-4">
        {summaryItems.map(({ label, value }, i) => (
          <div key={label} className={cn("flex-1 px-4", i > 0 && "border-l border-border")}>
            <p className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
              {label}
            </p>
            <p className="mt-1 font-mono text-3xl font-semibold text-foreground">{value}</p>
          </div>
        ))}
      </div>

      {/* System resources */}
      <div className="flex rounded-lg border border-border bg-muted/50 px-4 py-4">
        <div className="flex flex-1 flex-wrap gap-x-8 gap-y-4 justify-between items-center sm:grid sm:grid-cols-2 lg:flex">
          {meters.map(({ key, label, value, percent, icon: Icon, color }) => (
            <div key={key} className="flex items-center gap-4">
              <CircularProgress percent={percent} color={color} size={48} strokeWidth={4}>
                <Icon className="size-4 shrink-0" style={{ color }} />
              </CircularProgress>
              <div className="min-w-0 flex flex-col justify-center">
                <p className="text-[13px] font-medium text-muted-foreground">{label}</p>
                <p className="font-mono text-sm font-semibold text-foreground">{value}</p>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Trend chart */}
      <Card className="border-border">
        <CardHeader className="px-4 pb-2 pt-3">
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm font-medium text-muted-foreground">构建趋势</CardTitle>
            <span className="text-xs text-muted-foreground">近 7 天</span>
          </div>
        </CardHeader>
        <CardContent className="px-4 pb-3">
          <BuildTrendChart data={trend} />
        </CardContent>
      </Card>

      {/* Active builds + Recent builds */}
      <div className="grid gap-4 xl:grid-cols-[280px_1fr]">
        <Card className="border-border">
          <CardHeader className="px-4 pb-2 pt-3">
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm font-medium text-muted-foreground">运行中</CardTitle>
              <Badge variant="secondary" className="font-mono text-xs">
                {activeBuilds.length}
              </Badge>
            </div>
          </CardHeader>
          <CardContent className="px-4 pb-3">
            {activeBuilds.length === 0 ? (
              <p className="py-3 text-center text-xs text-muted-foreground">无运行中构建</p>
            ) : (
              <div className="space-y-2">
                {activeBuilds.map((build) => {
                  const statusInfo = getStatusInfo(build.status);
                  return (
                    <Link
                      key={build.id}
                      to={`/builds/${build.id}`}
                      className="flex items-center justify-between rounded-md border border-border px-3 py-2 transition-colors hover:bg-muted/50"
                    >
                      <div className="min-w-0">
                        <p className="truncate text-sm font-medium text-foreground">
                          #{build.build_number} {build.project_name}
                        </p>
                        <p className="mt-0.5 text-xs text-muted-foreground">
                          {build.environment_name ? `${build.environment_name} · ` : ""}
                          {STAGE_LABELS[build.current_stage] ?? build.current_stage}
                        </p>
                      </div>
                      <Badge
                        className={cn("ml-2 shrink-0 text-[11px] text-white", statusInfo.color)}
                      >
                        {statusInfo.label}
                      </Badge>
                    </Link>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>

        <Card className="border-border">
          <CardHeader className="px-4 pb-2 pt-3">
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm font-medium text-muted-foreground">最近构建</CardTitle>
              <Button
                asChild
                variant="ghost"
                size="sm"
                className="h-6 text-xs text-muted-foreground hover:text-foreground"
              >
                <Link to="/projects">
                  查看全部
                  <ArrowUpRight className="ml-1 size-3" />
                </Link>
              </Button>
            </div>
          </CardHeader>
          <CardContent className="px-4 pb-3">
            {recentBuilds.length === 0 ? (
              <p className="py-3 text-center text-xs text-muted-foreground">暂无构建记录</p>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-border text-left text-xs text-muted-foreground">
                      <th className="pb-2 pr-3 font-medium">构建</th>
                      <th className="pb-2 pr-3 font-medium">环境</th>
                      <th className="pb-2 pr-3 font-medium">状态</th>
                      <th className="pb-2 pr-3 font-medium">分支</th>
                      <th className="pb-2 pr-3 font-medium text-right">耗时</th>
                      <th className="pb-2 font-medium text-right">时间</th>
                    </tr>
                  </thead>
                  <tbody>
                    {recentBuilds.map((build) => {
                      const statusInfo = getStatusInfo(build.status);
                      return (
                        <tr key={build.id} className="border-b border-border/50 last:border-0">
                          <td className="py-2 pr-3">
                            <Link to={`/builds/${build.id}`} className="text-foreground">
                              <span className="font-mono">#{build.build_number}</span>{" "}
                              <span className="text-muted-foreground">{build.project_name}</span>
                            </Link>
                          </td>
                          <td className="py-2 pr-3 text-muted-foreground">
                            {build.environment_name || "-"}
                          </td>
                          <td className="py-2 pr-3">
                            <Badge className={cn("text-[11px] text-white", statusInfo.color)}>
                              {statusInfo.label}
                            </Badge>
                          </td>
                          <td className="py-2 pr-3 font-mono text-muted-foreground">
                            {build.branch || "-"}
                          </td>
                          <td className="py-2 pr-3 text-right font-mono text-muted-foreground">
                            <span className="inline-flex items-center gap-1">
                              <Clock3 className="size-3" />
                              {formatDuration(build.duration_ms)}
                            </span>
                          </td>
                          <td className="py-2 text-right text-muted-foreground">
                            {formatTimestamp(build.created_at)}
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
