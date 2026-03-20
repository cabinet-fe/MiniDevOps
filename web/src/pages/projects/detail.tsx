import { useCallback, useEffect, useState } from "react";
import { Link, useParams } from "react-router";
import {
  Copy,
  Download,
  ExternalLink,
  GitBranch,
  Loader2,
  Pencil,
  Play,
  Plus,
  Rocket,
  RotateCcw,
  Settings2,
  Sparkles,
} from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import { api } from "@/lib/api";
import type { PaginatedData } from "@/lib/api";
import { ARTIFACT_FORMATS, BUILD_SCRIPT_TYPES, BUILD_STATUSES } from "@/lib/constants";
import { cn } from "@/lib/utils";
import { ProjectFormDialog } from "@/pages/projects/form";
import { EnvironmentFormDialog } from "@/pages/projects/environment-form";
import { useNotificationStore } from "@/stores/notification-store";

interface Environment {
  id: number;
  project_id: number;
  name: string;
  branch: string;
  build_script: string;
  build_script_type: string;
  build_output_dir: string;
  deploy_server_id: number | null;
  deploy_method: string;
  deploy_path: string;
  post_deploy_script: string;
  cron_expression: string;
  cron_enabled: boolean;
  sort_order: number;
  var_group_ids: number[];
  cache_paths?: string;
}

interface Project {
  id: number;
  name: string;
  description: string;
  tags: string;
  repo_url: string;
  repo_auth_type: string;
  max_artifacts: number;
  artifact_format: string;
  webhook_secret?: string;
  webhook_type: string;
  webhook_ref_path: string;
  webhook_commit_path: string;
  webhook_message_path: string;
  environments: Environment[];
}

interface Build {
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

const RUNNING_BUILD_STATUSES = new Set(["pending", "cloning", "building", "deploying"]);

function formatRepoAuthType(value: string): string {
  if (!value || value === "none") return "无需认证";
  return value;
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

function getArtifactLabel(format: string): string {
  return ARTIFACT_FORMATS.find((item) => item.value === format)?.label ?? "Gzip (.tar.gz)";
}

function getScriptTypeLabel(type: string): string {
  return BUILD_SCRIPT_TYPES.find((item) => item.value === type)?.label ?? "Bash";
}

function extractFileName(path: string, fallback: string): string {
  const segments = path.split("/").filter(Boolean);
  return segments.at(-1) ?? fallback;
}

function getExternalHref(value: string): string | undefined {
  return /^https?:\/\//i.test(value) ? value : undefined;
}

function tagDictLabel(value: string, dict: { label: string; value: string }[]) {
  return dict.find((d) => d.value === value)?.label ?? value;
}

function MetaItem({
  label,
  value,
  mono = false,
}: {
  label: string;
  value: string;
  mono?: boolean;
}) {
  return (
    <span className="text-muted-foreground">
      {label}{" "}
      <span className={cn("text-foreground", mono && "font-mono")}>{value || "-"}</span>
    </span>
  );
}

function UrlRow({
  label,
  value,
  href,
  copyLabel,
}: {
  label: string;
  value: string;
  href?: string;
  copyLabel: string;
}) {
  const handleCopy = async () => {
    if (!value) return;
    await navigator.clipboard.writeText(value);
    toast.success(copyLabel);
  };

  return (
    <div className="bg-muted/50 flex items-center gap-2 rounded-md border border-border px-3 py-2">
      <span className="text-muted-foreground shrink-0 text-[11px] font-medium uppercase tracking-wider">
        {label}
      </span>
      <span
        className={cn(
          "min-w-0 flex-1 truncate font-mono text-xs",
          value ? "text-muted-foreground" : "text-muted-foreground/50",
        )}
        title={value || undefined}
      >
        {value || "未配置"}
      </span>
      {value && (
        <div className="flex shrink-0 gap-0.5">
          <Button variant="ghost" size="icon-xs" onClick={handleCopy}>
            <Copy className="size-3" />
          </Button>
          {href && (
            <Button variant="ghost" size="icon-xs" asChild>
              <a href={href} target="_blank" rel="noreferrer">
                <ExternalLink className="size-3" />
              </a>
            </Button>
          )}
        </div>
      )}
    </div>
  );
}

export function ProjectDetailPage() {
  const { id } = useParams<{ id: string }>();
  const projectId = Number(id);
  const [project, setProject] = useState<Project | null>(null);
  const [buildsByEnv, setBuildsByEnv] = useState<Record<number, Build[]>>({});
  const [loading, setLoading] = useState(true);
  const [triggering, setTriggering] = useState<number | null>(null);
  const [buildActionLoading, setBuildActionLoading] = useState<string | null>(null);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [triggerDialogEnv, setTriggerDialogEnv] = useState<Environment | null>(null);
  const [envFormOpen, setEnvFormOpen] = useState(false);
  const [editingEnv, setEditingEnv] = useState<Environment | null>(null);
  const [activeTab, setActiveTab] = useState<string>("");
  const [dictTags, setDictTags] = useState<{ label: string; value: string }[]>([]);
  const latestNotification = useNotificationStore((state) => state.latestNotification);

  useEffect(() => {
    api
      .get<{ label: string; value: string }[]>("/dictionaries/code/project_tags/items")
      .then((res) => {
        if (res.code === 0 && res.data) {
          setDictTags(res.data);
        }
      });
  }, []);

  const fetchProject = useCallback(async () => {
    if (!projectId) return;
    const res = await api.get<Project>(`/projects/${projectId}`);
    if (res.code !== 0 || !res.data) {
      setLoading(false);
      return;
    }
    const envs = res.data.environments ?? [];
    const buildPairs = await Promise.all(
      envs.map(async (env) => {
        const buildRes = await api.get<PaginatedData<Build>>(
          `/projects/${projectId}/builds?environment_id=${env.id}&page_size=20`,
        );
        const items = buildRes.code === 0 && buildRes.data ? (buildRes.data.items ?? []) : [];
        return [env.id, items] as const;
      }),
    );
    setProject({ ...res.data, environments: envs });
    setBuildsByEnv(Object.fromEntries(buildPairs));
    setLoading(false);
  }, [projectId]);

  useEffect(() => {
    if (!project) return;
    const envIds = project.environments.map((e) => String(e.id));
    if (envIds.length > 0 && (!activeTab || !envIds.includes(activeTab))) {
      setActiveTab(envIds[0]);
    }
  }, [project, activeTab]);

  useEffect(() => {
    fetchProject();
  }, [fetchProject]);

  useEffect(() => {
    if (!latestNotification || latestNotification.project_id !== projectId) return;
    fetchProject();
  }, [latestNotification, projectId, fetchProject]);

  const hasActiveBuilds = Object.values(buildsByEnv).some((items) =>
    items.some((item) => RUNNING_BUILD_STATUSES.has(item.status)),
  );

  useEffect(() => {
    if (!projectId || !project) return;
    const interval = window.setInterval(fetchProject, hasActiveBuilds ? 4000 : 15000);
    return () => window.clearInterval(interval);
  }, [projectId, project, hasActiveBuilds, fetchProject]);

  const triggerBuild = async (envId: number, branch?: string, commitHash?: string) => {
    if (!projectId) return;
    setTriggering(envId);
    try {
      const payload: Record<string, unknown> = { environment_id: envId };
      if (branch) payload.branch = branch;
      if (commitHash) payload.commit_hash = commitHash;
      const res = await api.post<Build>(`/projects/${projectId}/builds`, payload);
      if (res.code !== 0 || !res.data) {
        toast.error(res.message || "触发失败");
        return;
      }
      toast.success("构建已触发");
      setBuildsByEnv((prev) => ({
        ...prev,
        [envId]: [res.data!, ...(prev[envId] ?? [])],
      }));
    } catch {
      toast.error("触发失败");
    } finally {
      setTriggering(null);
    }
  };

  const handleBuildAction = async (action: "download" | "deploy" | "retry", build: Build) => {
    const actionKey = `${action}:${build.id}`;
    setBuildActionLoading(actionKey);
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
      toast.success(action === "deploy" ? "部署已触发" : "已重新触发构建");
      await fetchProject();
    } catch {
      toast.error("操作失败");
    } finally {
      setBuildActionLoading(null);
    }
  };

  if (loading || !project) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="size-6 animate-spin rounded-full border-2 border-border border-t-emerald-500" />
      </div>
    );
  }

  const webhookUrl = project.webhook_secret
    ? `${window.location.origin}/api/v1/webhook/${project.id}/${project.webhook_secret}`
    : "";
  const tagList = (project.tags || "")
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
  const repoHref = getExternalHref(project.repo_url);
  const successCount = Object.values(buildsByEnv)
    .flat()
    .filter((item) => item.status === "success").length;

  return (
    <TooltipProvider>
      <div className="space-y-4">
        {/* ── Block 1: Project Info ── */}
        <Card className="border-border">
          <CardHeader className="pb-3">
            <div className="flex items-start justify-between gap-4">
              <div className="min-w-0 space-y-1.5">
                <div className="flex flex-wrap items-center gap-2">
                  <CardTitle className="text-xl tracking-tight">{project.name}</CardTitle>
                  <Badge variant="outline" className="font-mono text-[11px]">
                    #{project.id}
                  </Badge>
                  {tagList.map((tag) => (
                    <Badge key={tag} variant="secondary">
                      {tagDictLabel(tag, dictTags)}
                    </Badge>
                  ))}
                </div>
                <p className="text-muted-foreground max-w-3xl text-sm">
                  {project.description || "未填写项目说明"}
                </p>
              </div>
              <div className="flex shrink-0 gap-2">
                <Button variant="outline" size="sm" onClick={() => setEditDialogOpen(true)}>
                  <Pencil className="size-3.5" />
                  编辑
                </Button>
                <Link to="/projects">
                  <Button variant="ghost" size="sm">
                    返回
                  </Button>
                </Link>
              </div>
            </div>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-stretch border-b border-border/80 pb-3">
              {[
                { label: "环境", value: `${project.environments.length}` },
                { label: "成功构建", value: `${successCount}` },
                { label: "保留制品", value: `${project.max_artifacts}` },
                { label: "认证", value: formatRepoAuthType(project.repo_auth_type) },
                { label: "格式", value: getArtifactLabel(project.artifact_format) },
                { label: "Webhook", value: project.webhook_type || "auto" },
              ].map(({ label, value }, i) => (
                <div
                  key={label}
                  className={cn("flex-1 px-3", i > 0 && "border-l border-border/80")}
                >
                  <p className="text-muted-foreground text-[10px] font-medium uppercase tracking-widest">
                    {label}
                  </p>
                  <p className="text-foreground mt-0.5 text-sm font-medium">{value}</p>
                </div>
              ))}
            </div>

            <div className="grid gap-2 lg:grid-cols-2">
              <UrlRow
                label="仓库"
                value={project.repo_url}
                href={repoHref}
                copyLabel="仓库地址已复制"
              />
              <UrlRow
                label="Webhook"
                value={webhookUrl}
                href={webhookUrl || undefined}
                copyLabel="Webhook URL 已复制"
              />
            </div>

            {project.webhook_type === "generic" && (
              <div className="flex flex-wrap gap-x-5 gap-y-1 text-xs">
                <MetaItem label="Ref:" value={project.webhook_ref_path || "$.ref"} mono />
                <MetaItem
                  label="Commit:"
                  value={project.webhook_commit_path || "$.head_commit.id"}
                  mono
                />
                <MetaItem
                  label="Message:"
                  value={project.webhook_message_path || "$.head_commit.message"}
                  mono
                />
              </div>
            )}
          </CardContent>
        </Card>

        {/* ── Block 2: Environments ── */}
        <Card className="border-border">
          {project.environments.length === 0 ? (
            <div className="flex min-h-48 flex-col items-center justify-center p-6 text-center">
              <Sparkles className="text-muted-foreground/50 size-8" />
              <p className="text-muted-foreground mt-3 text-sm font-medium">还没有环境</p>
              <p className="text-muted-foreground mt-1 text-xs">
                先创建至少一个环境，再配置分支、构建脚本和部署目标。
              </p>
              <Button
                size="sm"
                className="mt-4"
                onClick={() => {
                  setEditingEnv(null);
                  setEnvFormOpen(true);
                }}
              >
                <Plus className="size-3.5" />
                新建环境
              </Button>
            </div>
          ) : (
            <Tabs value={activeTab} onValueChange={setActiveTab}>
              <CardHeader className="pb-0">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-lg">环境与构建</CardTitle>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setEditingEnv(null);
                      setEnvFormOpen(true);
                    }}
                  >
                    <Plus className="size-3.5" />
                    新建环境
                  </Button>
                </div>
                <TabsList variant="line" className="mt-2 w-full justify-start">
                  {project.environments.map((env) => (
                    <TabsTrigger key={env.id} value={String(env.id)} className="gap-1.5">
                      {env.name}
                      {env.cron_enabled && env.cron_expression && (
                        <span className="size-1.5 rounded-full bg-emerald-500" />
                      )}
                    </TabsTrigger>
                  ))}
                </TabsList>
              </CardHeader>

              {project.environments.map((env) => {
                const builds = buildsByEnv[env.id] ?? [];
                return (
                  <TabsContent key={env.id} value={String(env.id)}>
                    <CardContent className="space-y-3 pt-4">
                      <div className="flex flex-wrap items-center gap-x-5 gap-y-1.5 text-xs">
                        <span className="text-muted-foreground inline-flex items-center gap-1">
                          <GitBranch className="size-3" />
                          <span className="text-foreground font-mono">{env.branch || "-"}</span>
                        </span>
                        <MetaItem
                          label="脚本"
                          value={
                            env.build_script
                              ? `${getScriptTypeLabel(env.build_script_type)} · 已配置`
                              : "未配置"
                          }
                        />
                        <MetaItem label="产物" value={env.build_output_dir || "未配置"} mono />
                        <MetaItem label="部署" value={env.deploy_method || "未部署"} />
                        {env.deploy_path && (
                          <MetaItem label="路径" value={env.deploy_path} mono />
                        )}
                        <MetaItem
                          label="Cron"
                          value={
                            env.cron_enabled
                              ? env.cron_expression || "已启用"
                              : "未启用"
                          }
                          mono
                        />
                        <MetaItem
                          label="变量组"
                          value={`${env.var_group_ids?.length ?? 0} 个`}
                        />
                      </div>

                      <div className="flex gap-2">
                        <Button
                          size="sm"
                          onClick={() => triggerBuild(env.id)}
                          disabled={triggering === env.id}
                        >
                          {triggering === env.id ? (
                            <Loader2 className="size-3.5 animate-spin" />
                          ) : (
                            <Play className="size-3.5" />
                          )}
                          {triggering === env.id ? "触发中" : "触发构建"}
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => setTriggerDialogEnv(env)}
                        >
                          <Settings2 className="size-3.5" />
                          高级触发
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => {
                            setEditingEnv(env);
                            setEnvFormOpen(true);
                          }}
                        >
                          <Pencil className="size-3.5" />
                          编辑环境
                        </Button>
                      </div>

                      <div className="max-h-[420px] overflow-y-auto rounded-md border border-border">
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
                                  onAction={handleBuildAction}
                                />
                              ))
                            )}
                          </TableBody>
                        </Table>
                      </div>
                    </CardContent>
                  </TabsContent>
                );
              })}
            </Tabs>
          )}
        </Card>

        <ProjectFormDialog
          open={editDialogOpen}
          onOpenChange={setEditDialogOpen}
          editId={projectId}
          onSuccess={() => fetchProject()}
        />
        <EnvironmentFormDialog
          open={envFormOpen}
          onOpenChange={setEnvFormOpen}
          projectId={projectId}
          editEnv={editingEnv}
          onSuccess={() => fetchProject()}
        />
        <TriggerBuildDialog
          env={triggerDialogEnv}
          open={!!triggerDialogEnv}
          onOpenChange={(open) => {
            if (!open) setTriggerDialogEnv(null);
          }}
          onTrigger={(envId, branch, commitHash) => {
            setTriggerDialogEnv(null);
            triggerBuild(envId, branch, commitHash);
          }}
          triggering={triggering}
          projectId={projectId}
        />
      </div>
    </TooltipProvider>
  );
}

function BuildRow({
  build,
  env,
  actionLoading,
  onAction,
}: {
  build: Build;
  env: Environment;
  actionLoading: string | null;
  onAction: (action: "download" | "deploy" | "retry", build: Build) => void;
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

function TriggerBuildDialog({
  env,
  open,
  onOpenChange,
  onTrigger,
  triggering,
  projectId,
}: {
  env: Environment | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onTrigger: (envId: number, branch?: string, commitHash?: string) => void;
  triggering: number | null;
  projectId: number;
}) {
  const [branch, setBranch] = useState("");
  const [commitHash, setCommitHash] = useState("");
  const [branches, setBranches] = useState<string[]>([]);
  const [branchesLoading, setBranchesLoading] = useState(false);
  const [branchPopoverOpen, setBranchPopoverOpen] = useState(false);

  useEffect(() => {
    if (!open) {
      setBranch("");
      setCommitHash("");
      setBranches([]);
      return;
    }
    setBranchesLoading(true);
    api
      .get<string[]>(`/projects/${projectId}/branches`)
      .then((res) => {
        if (res.code === 0 && res.data) {
          setBranches(Array.isArray(res.data) ? res.data : []);
        }
      })
      .finally(() => setBranchesLoading(false));
  }, [open, projectId]);

  if (!env) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[460px]">
        <DialogHeader>
          <DialogTitle>高级触发 · {env.name}</DialogTitle>
          <DialogDescription>
            可指定分支或 Commit，留空则使用环境默认分支（{env.branch}）。
          </DialogDescription>
        </DialogHeader>
        <DialogBody className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="trigger-branch">分支（可选）</Label>
            <Popover open={branchPopoverOpen} onOpenChange={setBranchPopoverOpen}>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  role="combobox"
                  aria-expanded={branchPopoverOpen}
                  className="w-full justify-between font-normal"
                >
                  {branch || `默认: ${env.branch}`}
                  {branchesLoading && <Loader2 className="ml-2 size-4 animate-spin" />}
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-[--radix-popover-trigger-width] p-0" align="start">
                <Command>
                  <CommandInput
                    placeholder="搜索或输入分支名..."
                    value={branch}
                    onValueChange={(value: string) => setBranch(value)}
                  />
                  <CommandList>
                    <CommandEmpty>
                      {branchesLoading ? "加载中..." : "无匹配分支，可直接输入"}
                    </CommandEmpty>
                    <CommandGroup>
                      {branches.map((item) => (
                        <CommandItem
                          key={item}
                          value={item}
                          onSelect={(value: string) => {
                            setBranch(value);
                            setBranchPopoverOpen(false);
                          }}
                        >
                          {item}
                        </CommandItem>
                      ))}
                    </CommandGroup>
                  </CommandList>
                </Command>
              </PopoverContent>
            </Popover>
          </div>
          <div className="space-y-2">
            <Label htmlFor="trigger-commit">Commit Hash（可选）</Label>
            <Input
              id="trigger-commit"
              value={commitHash}
              onChange={(e) => setCommitHash(e.target.value)}
              placeholder="例如: a1b2c3d"
              className="font-mono"
            />
          </div>
        </DialogBody>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button
            onClick={() => onTrigger(env.id, branch || undefined, commitHash || undefined)}
            disabled={triggering === env.id}
          >
            {triggering === env.id ? (
              <>
                <Loader2 className="size-4 animate-spin" />
                触发中
              </>
            ) : (
              <>
                <Play className="size-4" />
                触发构建
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
