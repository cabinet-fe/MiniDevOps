import { Link } from "react-router";
import {
  ChevronLeft,
  ChevronRight,
  Download,
  ExternalLink,
  Loader2,
  Rocket,
  RotateCcw,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { BUILD_STATUSES } from "@/lib/constants";
import { cn } from "@/lib/utils";

export interface BuildListEntry {
  id: number;
  environment_id: number;
  build_number: number;
  status: string;
  trigger_type: string;
  branch: string;
  commit_hash: string;
  commit_message: string;
  artifact_path: string;
  duration_ms: number;
  created_at: string;
}

export interface EnvBranchRef {
  branch: string;
}

function formatDuration(ms: number): string {
  if (!ms) return "-";
  const seconds = Math.floor(ms / 1000);
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  return `${minutes}m ${seconds % 60}s`;
}

function formatDateTime(date: string): string {
  return new Date(date).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function extractFileName(path: string, fallback: string): string {
  const segments = path.split("/").filter(Boolean);
  return segments.at(-1) ?? fallback;
}

function BuildRow({
  build,
  env,
  actionLoading,
  onAction,
}: {
  build: BuildListEntry;
  env: EnvBranchRef;
  actionLoading: string | null;
  onAction: (action: "download" | "deploy" | "retry", build: BuildListEntry) => void;
}) {
  const statusInfo = BUILD_STATUSES[build.status as keyof typeof BUILD_STATUSES] ?? {
    label: build.status,
    color: "bg-zinc-500",
  };
  const canDownload = build.status === "success" && Boolean(build.artifact_path);
  const canDeploy = build.status === "success";
  const canRetry = ["failed", "cancelled"].includes(build.status);
  const branchDiffers = build.branch && build.branch !== env.branch;

  return (
    <TableRow>
      <TableCell className="text-foreground py-1.5 font-mono text-xs font-semibold">
        #{build.build_number}
      </TableCell>
      <TableCell className="py-1.5">
        <Badge className={cn("text-[10px] text-white", statusInfo.color)}>
          {statusInfo.label}
        </Badge>
      </TableCell>
      <TableCell className="py-1.5">
        <span className="text-muted-foreground font-mono text-[11px]">
          {build.commit_hash ? build.commit_hash.slice(0, 7) : "-"}
        </span>
        {branchDiffers && (
          <span className="text-muted-foreground ml-1.5 text-[10px]">{build.branch}</span>
        )}
      </TableCell>
      <TableCell className="max-w-[240px] py-1.5">
        <p className="text-muted-foreground truncate text-xs">{build.commit_message || "-"}</p>
      </TableCell>
      <TableCell className="text-muted-foreground py-1.5 text-[11px]">
        <p className="font-mono">{formatDuration(build.duration_ms)}</p>
        <p className="text-muted-foreground">{formatDateTime(build.created_at)}</p>
      </TableCell>
      <TableCell className="py-1.5">
        <div className="flex justify-end gap-0.5">
          {canDownload && (
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon-xs"
                  disabled={actionLoading === `download:${build.id}`}
                  onClick={() => onAction("download", build)}
                >
                  {actionLoading === `download:${build.id}` ? (
                    <Loader2 className="size-3 animate-spin" />
                  ) : (
                    <Download className="size-3" />
                  )}
                </Button>
              </TooltipTrigger>
              <TooltipContent>下载制品</TooltipContent>
            </Tooltip>
          )}
          {canDeploy && (
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon-xs"
                  disabled={actionLoading === `deploy:${build.id}`}
                  onClick={() => onAction("deploy", build)}
                >
                  {actionLoading === `deploy:${build.id}` ? (
                    <Loader2 className="size-3 animate-spin" />
                  ) : (
                    <Rocket className="size-3" />
                  )}
                </Button>
              </TooltipTrigger>
              <TooltipContent>重新部署</TooltipContent>
            </Tooltip>
          )}
          {canRetry && (
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon-xs"
                  disabled={actionLoading === `retry:${build.id}`}
                  onClick={() => onAction("retry", build)}
                >
                  {actionLoading === `retry:${build.id}` ? (
                    <Loader2 className="size-3 animate-spin" />
                  ) : (
                    <RotateCcw className="size-3" />
                  )}
                </Button>
              </TooltipTrigger>
              <TooltipContent>重新构建</TooltipContent>
            </Tooltip>
          )}
          <Tooltip>
            <TooltipTrigger asChild>
              <Button variant="ghost" size="icon-xs" asChild>
                <Link to={`/builds/${build.id}`}>
                  <ExternalLink className="size-3" />
                </Link>
              </Button>
            </TooltipTrigger>
            <TooltipContent>查看详情</TooltipContent>
          </Tooltip>
        </div>
      </TableCell>
    </TableRow>
  );
}

export { extractFileName };

export function EnvironmentBuildsTable({
  env,
  builds,
  loading = false,
  page,
  totalPages,
  total,
  onPageChange,
  buildActionLoading,
  onBuildAction,
}: {
  env: EnvBranchRef;
  builds: BuildListEntry[];
  loading?: boolean;
  page: number;
  totalPages: number;
  total: number;
  onPageChange: (page: number) => void;
  buildActionLoading: string | null;
  onBuildAction: (action: "download" | "deploy" | "retry", build: BuildListEntry) => void;
}) {
  return (
    <div className="space-y-3">
      <div className="max-h-[min(420px,60vh)] overflow-y-auto rounded-md border border-border">
        {loading ? (
          <div className="flex h-40 items-center justify-center">
            <Loader2 className="text-muted-foreground size-6 animate-spin" />
          </div>
        ) : (
          <Table>
            <TableHeader className="bg-card sticky top-0 z-10">
              <TableRow>
                <TableHead className="w-16">编号</TableHead>
                <TableHead className="w-20">状态</TableHead>
                <TableHead className="w-28">Commit</TableHead>
                <TableHead>提交信息</TableHead>
                <TableHead className="w-28">耗时 / 时间</TableHead>
                <TableHead className="w-28 text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {builds.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={6}
                    className="text-muted-foreground py-8 text-center text-xs"
                  >
                    暂无构建记录
                  </TableCell>
                </TableRow>
              ) : (
                builds.map((build) => (
                  <BuildRow
                    key={build.id}
                    build={build}
                    env={env}
                    actionLoading={buildActionLoading}
                    onAction={onBuildAction}
                  />
                ))
              )}
            </TableBody>
          </Table>
        )}
      </div>

      {!loading && total > 0 && (
        <div className="flex items-center justify-between border-t border-border pt-3">
          <p className="text-muted-foreground text-sm">
            第 {page} / {totalPages} 页，共 {total} 条
          </p>
          <div className="flex gap-1">
            <Button
              variant="outline"
              size="icon"
              className="size-8"
              onClick={() => onPageChange(Math.max(1, page - 1))}
              disabled={page <= 1}
            >
              <ChevronLeft className="size-4" />
            </Button>
            <Button
              variant="outline"
              size="icon"
              className="size-8"
              onClick={() => onPageChange(Math.min(totalPages, page + 1))}
              disabled={page >= totalPages}
            >
              <ChevronRight className="size-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
