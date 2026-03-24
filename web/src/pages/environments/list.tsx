import { useCallback, useEffect, useMemo, useState } from "react";
import { Link } from "react-router";
import {
  ChevronLeft,
  ChevronRight,
  GitBranch,
  History,
  Layers,
  Loader2,
  Play,
  Search,
  Trash2,
} from "lucide-react";
import { toast } from "sonner";
import {
  EnvironmentBuildsTable,
  extractFileName,
  type BuildListEntry,
} from "@/components/environment-builds-table";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { TooltipProvider } from "@/components/ui/tooltip";
import { api } from "@/lib/api";
import type { PaginatedData } from "@/lib/api";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth-store";

interface EnvironmentRow {
  id: number;
  project_id: number;
  name: string;
  branch: string;
  sort_order: number;
}

interface EnvironmentListItem extends EnvironmentRow {
  project_name: string;
}

interface ProjectBrief {
  id: number;
  name: string;
}

type EnvFlatRow = {
  projectId: number;
  projectName: string;
  env: EnvironmentRow;
};

const LIST_PAGE_SIZE = 20;
const DIALOG_PAGE_SIZE = 20;

export function EnvironmentListPage() {
  const user = useAuthStore((s) => s.user);
  const canDeleteEnvironment = user?.role === "admin" || user?.role === "ops";
  const [projectOptions, setProjectOptions] = useState<ProjectBrief[]>([]);
  const [projectsLoading, setProjectsLoading] = useState(true);
  const [listItems, setListItems] = useState<EnvironmentListItem[]>([]);
  const [listLoading, setListLoading] = useState(true);
  const [listPage, setListPage] = useState(1);
  const [listTotal, setListTotal] = useState(0);
  const [listTotalPages, setListTotalPages] = useState(1);
  const [projectFilter, setProjectFilter] = useState<string>("all");
  const [nameQuery, setNameQuery] = useState("");
  const [debouncedName, setDebouncedName] = useState("");
  const [triggeringKey, setTriggeringKey] = useState<string | null>(null);
  const [buildsDialogOpen, setBuildsDialogOpen] = useState(false);
  const [buildsDialogRow, setBuildsDialogRow] = useState<EnvFlatRow | null>(null);
  const [dialogBuilds, setDialogBuilds] = useState<BuildListEntry[]>([]);
  const [dialogPage, setDialogPage] = useState(1);
  const [dialogTotal, setDialogTotal] = useState(0);
  const [dialogTotalPages, setDialogTotalPages] = useState(1);
  const [dialogLoading, setDialogLoading] = useState(false);
  const [dialogActionLoading, setDialogActionLoading] = useState<string | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<EnvironmentListItem | null>(null);
  const [deleteSubmitting, setDeleteSubmitting] = useState(false);

  useEffect(() => {
    const t = window.setTimeout(() => {
      const next = nameQuery.trim();
      setDebouncedName((prev) => {
        if (prev !== next) setListPage(1);
        return next;
      });
    }, 300);
    return () => window.clearTimeout(t);
  }, [nameQuery]);

  const fetchProjectOptions = useCallback(async () => {
    setProjectsLoading(true);
    try {
      const acc: ProjectBrief[] = [];
      for (let page = 1; ; page += 1) {
        const res = await api.get<PaginatedData<ProjectBrief>>(
          `/projects?page=${page}&page_size=100`,
        );
        if (res.code !== 0 || !res.data) break;
        const data = res.data as PaginatedData<ProjectBrief>;
        const items = data.items ?? [];
        acc.push(...items);
        if (items.length < 100 || page >= (data.total_pages ?? 1)) break;
      }
      setProjectOptions(acc);
    } catch {
      toast.error("加载项目列表失败");
    } finally {
      setProjectsLoading(false);
    }
  }, []);

  useEffect(() => {
    void fetchProjectOptions();
  }, [fetchProjectOptions]);

  const loadEnvironmentList = useCallback(async () => {
    setListLoading(true);
    try {
      const params = new URLSearchParams();
      params.set("page", String(listPage));
      params.set("page_size", String(LIST_PAGE_SIZE));
      if (projectFilter !== "all") params.set("project_id", projectFilter);
      if (debouncedName) params.set("name", debouncedName);
      const res = await api.get<PaginatedData<EnvironmentListItem>>(
        `/environments?${params.toString()}`,
      );
      if (res.code !== 0 || !res.data) {
        toast.error(res.message || "加载失败");
        return;
      }
      const data = res.data as PaginatedData<EnvironmentListItem>;
      setListItems(data.items ?? []);
      setListTotal(data.total ?? 0);
      setListTotalPages(
        Math.max(1, data.total_pages ?? (Math.ceil((data.total ?? 0) / LIST_PAGE_SIZE) || 1)),
      );
    } catch {
      toast.error("加载失败");
    } finally {
      setListLoading(false);
    }
  }, [listPage, projectFilter, debouncedName]);

  useEffect(() => {
    void loadEnvironmentList();
  }, [loadEnvironmentList]);

  const loadDialogBuilds = useCallback(async () => {
    if (!buildsDialogRow) return;
    setDialogLoading(true);
    try {
      const res = await api.get<PaginatedData<BuildListEntry>>(
        `/projects/${buildsDialogRow.projectId}/builds?environment_id=${buildsDialogRow.env.id}&page=${dialogPage}&page_size=${DIALOG_PAGE_SIZE}`,
      );
      if (res.code !== 0 || !res.data) {
        toast.error(res.message || "加载失败");
        return;
      }
      const data = res.data as PaginatedData<BuildListEntry>;
      setDialogBuilds(data.items ?? []);
      setDialogTotal(data.total ?? 0);
      setDialogTotalPages(
        Math.max(1, data.total_pages ?? (Math.ceil((data.total ?? 0) / DIALOG_PAGE_SIZE) || 1)),
      );
    } catch {
      toast.error("加载失败");
    } finally {
      setDialogLoading(false);
    }
  }, [buildsDialogRow, dialogPage]);

  useEffect(() => {
    if (!buildsDialogOpen || !buildsDialogRow) return;
    void loadDialogBuilds();
  }, [buildsDialogOpen, buildsDialogRow, dialogPage, loadDialogBuilds]);

  const handleDialogBuildAction = async (
    action: "download" | "deploy" | "retry",
    build: BuildListEntry,
  ) => {
    if (!buildsDialogRow) return;
    const actionKey = `${action}:${build.id}`;
    setDialogActionLoading(actionKey);
    try {
      if (action === "download") {
        const blob = await api.download(`/builds/${build.id}/artifact`);
        const link = document.createElement("a");
        link.href = URL.createObjectURL(blob);
        link.download = extractFileName(
          build.artifact_path,
          `build-${build.build_number}-artifact`,
        );
        link.click();
        URL.revokeObjectURL(link.href);
        toast.success("下载开始");
        return;
      }
      const endpoint =
        action === "deploy" ? `/builds/${build.id}/deploy` : `/builds/${build.id}/retry`;
      const res = await api.post(endpoint);
      if (res.code !== 0) {
        toast.error(res.message || "操作失败");
        return;
      }
      toast.success(action === "deploy" ? "开始部署" : "已重新构建");
      await loadDialogBuilds();
    } catch {
      toast.error("操作失败");
    } finally {
      setDialogActionLoading(null);
    }
  };

  const openBuildsDialog = (row: EnvironmentListItem) => {
    setBuildsDialogRow({
      projectId: row.project_id,
      projectName: row.project_name,
      env: {
        id: row.id,
        project_id: row.project_id,
        name: row.name,
        branch: row.branch,
        sort_order: row.sort_order,
      },
    });
    setDialogPage(1);
    setBuildsDialogOpen(true);
  };

  const handleConfirmDeleteEnvironment = async () => {
    if (!deleteTarget) return;
    setDeleteSubmitting(true);
    try {
      const res = await api.delete(`/projects/${deleteTarget.project_id}/envs/${deleteTarget.id}`);
      if (res.code !== 0) {
        toast.error(res.message || "删除失败");
        return;
      }
      toast.success("环境已删除");
      if (
        buildsDialogRow &&
        buildsDialogRow.projectId === deleteTarget.project_id &&
        buildsDialogRow.env.id === deleteTarget.id
      ) {
        setBuildsDialogOpen(false);
        setBuildsDialogRow(null);
      }
      setDeleteTarget(null);
      await loadEnvironmentList();
    } catch {
      toast.error("删除失败");
    } finally {
      setDeleteSubmitting(false);
    }
  };

  const triggerBuild = async (projectId: number, envId: number) => {
    const key = `${projectId}:${envId}`;
    setTriggeringKey(key);
    try {
      const res = await api.post<{ id: number }>(`/projects/${projectId}/builds`, {
        environment_id: envId,
      });
      if (res.code !== 0 || !res.data) {
        toast.error(res.message || "触发失败");
        return;
      }
      toast.success("构建已触发");
    } catch {
      toast.error("触发失败");
    } finally {
      setTriggeringKey(null);
    }
  };

  const filterSummary = useMemo(() => {
    const parts: string[] = [];
    if (projectFilter !== "all") {
      const p = projectOptions.find((x) => String(x.id) === projectFilter);
      if (p) parts.push(`项目：${p.name}`);
    }
    if (debouncedName) parts.push(`名称含「${debouncedName}」`);
    return parts.length ? parts.join("；") : "全部";
  }, [projectFilter, debouncedName, projectOptions]);

  const initialLoading = projectsLoading && projectOptions.length === 0;

  if (initialLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="size-6 animate-spin rounded-full border-2 border-border border-t-emerald-500" />
      </div>
    );
  }

  return (
    <TooltipProvider>
      <div className="space-y-4">
        <div>
          <h1 className="text-foreground text-xl font-semibold tracking-tight">环境</h1>
          <p className="text-muted-foreground mt-1 text-sm">
            跨项目查看环境并快捷触发构建（使用环境默认分支）
          </p>
        </div>

        <Card>
          <CardHeader className="pb-3">
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div>
                <CardTitle className="flex items-center gap-2 text-base">
                  <Layers className="size-4 text-emerald-500/80" />
                  全部环境
                </CardTitle>
                <CardDescription>
                  共 {listTotal} 条（{filterSummary}）
                </CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-end">
              <div className="space-y-1.5 sm:w-[220px]">
                <Label className="text-muted-foreground text-xs">项目</Label>
                <Select
                  value={projectFilter}
                  onValueChange={(v) => {
                    setProjectFilter(v);
                    setListPage(1);
                  }}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="选择项目" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">全部项目</SelectItem>
                    {projectOptions
                      .slice()
                      .sort((a, b) => a.name.localeCompare(b.name, "zh-CN"))
                      .map((p) => (
                        <SelectItem key={p.id} value={String(p.id)}>
                          {p.name}
                        </SelectItem>
                      ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="min-w-0 flex-1 space-y-1.5">
                <Label htmlFor="env-name-filter" className="text-muted-foreground text-xs">
                  环境名称
                </Label>
                <div className="relative">
                  <Search className="text-muted-foreground pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2" />
                  <Input
                    id="env-name-filter"
                    value={nameQuery}
                    onChange={(e) => setNameQuery(e.target.value)}
                    placeholder="过滤环境名称…"
                    className="pl-9"
                  />
                </div>
              </div>
            </div>

            <div className="overflow-x-auto rounded-md border border-border">
              <Table>
                <TableHeader>
                  <TableRow className="bg-muted/40 hover:bg-muted/40">
                    <TableHead className="min-w-[120px]">环境</TableHead>
                    <TableHead className="min-w-[140px]">项目</TableHead>
                    <TableHead className="w-[160px]">分支</TableHead>
                    <TableHead className="min-w-[200px] text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {listLoading ? (
                    <TableRow>
                      <TableCell colSpan={4} className="py-12 text-center">
                        <Loader2 className="text-muted-foreground mx-auto size-6 animate-spin" />
                      </TableCell>
                    </TableRow>
                  ) : listItems.length === 0 ? (
                    <TableRow>
                      <TableCell
                        colSpan={4}
                        className="text-muted-foreground py-10 text-center text-sm"
                      >
                        无匹配环境
                      </TableCell>
                    </TableRow>
                  ) : (
                    listItems.map((row) => {
                      const tKey = `${row.project_id}:${row.id}`;
                      const busy = triggeringKey === tKey;
                      return (
                        <TableRow key={tKey}>
                          <TableCell className="font-medium">{row.name}</TableCell>
                          <TableCell>
                            <Link
                              to={`/projects/${row.project_id}`}
                              className={cn(
                                "font-medium text-emerald-600 hover:text-emerald-500",
                                "dark:text-emerald-400 dark:hover:text-emerald-300",
                              )}
                            >
                              {row.project_name}
                            </Link>
                          </TableCell>
                          <TableCell>
                            <span className="text-muted-foreground inline-flex items-center gap-1.5 font-mono text-xs">
                              <GitBranch className="size-3.5 shrink-0" />
                              {row.branch || "-"}
                            </span>
                          </TableCell>
                          <TableCell className="text-right">
                            <div className="flex flex-wrap items-center justify-end gap-2">
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => openBuildsDialog(row)}
                              >
                                <History className="size-3.5" />
                                查看
                              </Button>
                              <Button
                                size="sm"
                                onClick={() => triggerBuild(row.project_id, row.id)}
                                disabled={busy}
                              >
                                {busy ? (
                                  <Loader2 className="size-3.5 animate-spin" />
                                ) : (
                                  <Play className="size-3.5" />
                                )}
                                {busy ? "构建中" : "构建"}
                              </Button>
                              {canDeleteEnvironment && (
                                <Button
                                  variant="outline"
                                  size="sm"
                                  className="border-destructive/40 text-destructive hover:bg-destructive/10"
                                  onClick={() => setDeleteTarget(row)}
                                >
                                  <Trash2 className="size-3.5" />
                                  删除
                                </Button>
                              )}
                            </div>
                          </TableCell>
                        </TableRow>
                      );
                    })
                  )}
                </TableBody>
              </Table>
            </div>

            {!listLoading && listTotal > 0 && (
              <div className="flex items-center justify-between border-t border-border pt-3">
                <p className="text-muted-foreground text-sm">
                  第 {listPage} / {listTotalPages} 页，共 {listTotal} 条
                </p>
                <div className="flex gap-1">
                  <Button
                    variant="outline"
                    size="icon"
                    className="size-8"
                    onClick={() => setListPage((p) => Math.max(1, p - 1))}
                    disabled={listPage <= 1}
                  >
                    <ChevronLeft className="size-4" />
                  </Button>
                  <Button
                    variant="outline"
                    size="icon"
                    className="size-8"
                    onClick={() => setListPage((p) => Math.min(listTotalPages, p + 1))}
                    disabled={listPage >= listTotalPages}
                  >
                    <ChevronRight className="size-4" />
                  </Button>
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        <Dialog
          open={deleteTarget !== null}
          onOpenChange={(open) => {
            if (!open && !deleteSubmitting) setDeleteTarget(null);
          }}
        >
          <DialogContent showCloseButton={!deleteSubmitting}>
            <DialogHeader>
              <DialogTitle>删除环境</DialogTitle>
              <DialogDescription>
                {deleteTarget
                  ? `将永久删除「${deleteTarget.project_name}」下的环境「${deleteTarget.name}」及其全部构建记录、日志、产物与该环境工作区/缓存目录，且不可恢复。确定继续？`
                  : ""}
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                disabled={deleteSubmitting}
                onClick={() => setDeleteTarget(null)}
              >
                取消
              </Button>
              <Button
                type="button"
                variant="destructive"
                disabled={deleteSubmitting}
                onClick={() => void handleConfirmDeleteEnvironment()}
              >
                {deleteSubmitting ? <Loader2 className="size-4 animate-spin" /> : "删除"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        <Dialog
          open={buildsDialogOpen}
          onOpenChange={(open) => {
            setBuildsDialogOpen(open);
            if (!open) setBuildsDialogRow(null);
          }}
        >
          <DialogContent className="sm:max-w-3xl">
            <DialogHeader>
              <DialogTitle>
                {buildsDialogRow
                  ? `构建记录 · ${buildsDialogRow.projectName} / ${buildsDialogRow.env.name}`
                  : "构建记录"}
              </DialogTitle>
              <DialogDescription>
                该环境下的构建历史与操作，与项目详情中「环境与构建」一致。
              </DialogDescription>
            </DialogHeader>
            <DialogBody className="space-y-0 pt-2">
              {buildsDialogRow && (
                <EnvironmentBuildsTable
                  env={buildsDialogRow.env}
                  builds={dialogBuilds}
                  loading={dialogLoading}
                  page={dialogPage}
                  totalPages={dialogTotalPages}
                  total={dialogTotal}
                  onPageChange={setDialogPage}
                  buildActionLoading={dialogActionLoading}
                  onBuildAction={handleDialogBuildAction}
                />
              )}
            </DialogBody>
          </DialogContent>
        </Dialog>
      </div>
    </TooltipProvider>
  );
}
