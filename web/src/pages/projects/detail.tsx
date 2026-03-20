import type { ReactNode } from "react";
import { useCallback, useEffect, useState } from "react";
import { Link, useParams } from "react-router";
import {
  Clock,
  Copy,
  Download,
  ExternalLink,
  GitBranch,
  Layers3,
  Loader2,
  Package,
  Pencil,
  Play,
  Plus,
  Radio,
  Rocket,
  RotateCcw,
  Server,
  Settings2,
  Sparkles,
} from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
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
import { api } from "@/lib/api";
import type { PaginatedData } from "@/lib/api";
import { ARTIFACT_FORMATS, BUILD_SCRIPT_TYPES, BUILD_STATUSES } from "@/lib/constants";
import { cn } from "@/lib/utils";
import { ProjectFormDialog } from "@/pages/projects/form";
import { EnvironmentFormDialog } from "@/pages/projects/environment-form";
import { useNotificationStore } from "@/stores/notification-store";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";

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
  group_name: string;
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

const SIGNAL_TOKENS: {
  shell: string;
  hero: string;
  glow: string;
  panel: string;
  stat: string;
  envCard: string;
  detailButton: string;
  detailButtonGhost: string;
} = {
  shell: "space-y-4 pb-6",
  hero: "relative overflow-hidden rounded-[26px] border border-zinc-200/80 bg-[linear-gradient(135deg,rgba(255,255,255,0.98),rgba(230,236,248,0.95))] p-5 shadow-[0_24px_80px_-48px_rgba(15,23,42,0.55)] dark:border-zinc-800 dark:bg-[linear-gradient(135deg,rgba(12,17,27,0.98),rgba(22,30,46,0.96))]",
  glow: "pointer-events-none absolute inset-x-0 top-0 h-40 bg-[radial-gradient(circle_at_top_right,rgba(56,189,248,0.18),transparent_50%),radial-gradient(circle_at_top_left,rgba(250,204,21,0.16),transparent_45%)]",
  panel:
    "border-zinc-200/80 bg-white/82 shadow-[0_18px_60px_-44px_rgba(15,23,42,0.5)] dark:border-zinc-800 dark:bg-zinc-950/72",
  stat: "rounded-[24px] border border-white/70 bg-white/72 p-3 shadow-sm backdrop-blur dark:border-zinc-800/80 dark:bg-zinc-900/72",
  envCard:
    "rounded-[24px] border border-zinc-200/80 bg-white/90 p-4 shadow-[0_18px_60px_-48px_rgba(15,23,42,0.6)] backdrop-blur dark:border-zinc-800 dark:bg-zinc-950/78",
  detailButton:
    "bg-zinc-950 text-white hover:bg-zinc-800 dark:bg-cyan-300 dark:text-zinc-950 dark:hover:bg-cyan-200",
  detailButtonGhost:
    "border-cyan-500/20 bg-cyan-500/8 text-cyan-700 hover:bg-cyan-500/14 dark:text-cyan-200",
};

const REPO_URL_VALUE_CLASS = "break-all whitespace-normal leading-5";

function formatRepoAuthType(value: string): string {
  if (!value) return "无需认证";
  if (value === "none") return "无需认证";
  return value;
}

const WEBHOOK_GUIDES: Record<string, { title: string; headers: string[]; sample: string }> = {
  auto: {
    title: "自动识别",
    headers: [
      "GitHub: `X-GitHub-Event: push`",
      "GitLab: `X-Gitlab-Event: Push Hook`",
      "Gitea: `X-Gitea-Event: push`",
      "Bitbucket: `X-Event-Key: repo:push`",
    ],
    sample: "服务端会按请求头自动识别平台并解析 push payload。",
  },
  github: {
    title: "GitHub",
    headers: ["Header: `X-GitHub-Event: push`"],
    sample: "仓库 Webhook 选择 JSON，触发事件勾选 push。",
  },
  gitlab: {
    title: "GitLab",
    headers: ["Header: `X-Gitlab-Event: Push Hook`"],
    sample: "GitLab 项目集成里启用 Push events 即可。",
  },
  gitea: {
    title: "Gitea",
    headers: ["Header: `X-Gitea-Event: push`"],
    sample: "Gitea Webhook 选择 Gitea/GitHub 风格 JSON，触发 push。",
  },
  bitbucket: {
    title: "Bitbucket",
    headers: ["Header: `X-Event-Key: repo:push`"],
    sample: "Bitbucket Cloud/Server 推送事件会使用 `push.changes[0]` 结构。",
  },
  generic: {
    title: "通用 JSON",
    headers: ["自定义 JSONPath: `$.ref`、`$.head_commit.id`、`$.head_commit.message`"],
    sample: "支持 `$.field` 与 `$.list[0].field` 形式的路径。",
  },
};

function formatDuration(ms: number): string {
  if (!ms) return "-";
  const seconds = Math.floor(ms / 1000);
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  return `${minutes}m ${seconds % 60}s`;
}

function formatDateTime(date: string): string {
  return new Date(date).toLocaleString("zh-CN");
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

function EnvironmentMeta({
  label,
  value,
  mono = false,
  icon,
  className,
  valueClassName,
}: {
  label: string;
  value: string;
  mono?: boolean;
  icon?: ReactNode;
  className?: string;
  valueClassName?: string;
}) {
  return (
    <div
      className={cn(
        "min-w-0 rounded-[20px] border border-black/5 bg-black/[0.03] px-3 py-2.5 dark:border-white/5 dark:bg-white/[0.03]",
        className,
      )}
    >
      <p className="text-[10px] uppercase tracking-[0.24em] text-zinc-500 dark:text-zinc-400">
        {label}
      </p>
      <div className="mt-1.5 flex min-w-0 items-start gap-2">
        {icon}
        <p
          className={cn(
            "min-w-0 text-sm font-medium text-zinc-900 dark:text-zinc-100",
            mono && "font-mono",
            valueClassName,
          )}
        >
          {value || "-"}
        </p>
      </div>
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
  const latestNotification = useNotificationStore((state) => state.latestNotification);

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

    setProject(res.data);
    setBuildsByEnv(Object.fromEntries(buildPairs));
    setLoading(false);
  }, [projectId]);

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

  const copyWebhook = () => {
    if (!project?.webhook_secret) return;
    const url = `${window.location.origin}/api/v1/webhook/${project.id}/${project.webhook_secret}`;
    navigator.clipboard.writeText(url);
    toast.success("Webhook URL 已复制");
  };

  if (loading || !project) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-2 border-zinc-600 border-t-zinc-300" />
      </div>
    );
  }

  const tokens = SIGNAL_TOKENS;
  const webhookUrl = project.webhook_secret
    ? `${window.location.origin}/api/v1/webhook/${project.id}/${project.webhook_secret}`
    : "";
  const tagList = (project.tags || "")
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
  const webhookGuide = WEBHOOK_GUIDES[project.webhook_type || "auto"] ?? WEBHOOK_GUIDES.auto;
  const successCount = Object.values(buildsByEnv)
    .flat()
    .filter((item) => item.status === "success").length;

  return (
    <div className={tokens.shell}>
      <section className={tokens.hero}>
        <div className={tokens.glow} />
        <div className="relative flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div className="space-y-3">
            <div className="flex flex-wrap items-center gap-2">
              <Badge
                variant="outline"
                className="rounded-full border-current/20 bg-white/60 px-3 py-1 text-[11px] uppercase tracking-[0.24em] dark:bg-white/5"
              >
                Project Control Room
              </Badge>
              <Badge variant="secondary" className="rounded-full">
                {project.group_name || "未分组"}
              </Badge>
              <Badge variant="outline" className="rounded-full">
                {getArtifactLabel(project.artifact_format)}
              </Badge>
            </div>
            <div>
              <h1 className="text-3xl font-semibold tracking-tight text-zinc-950 dark:text-white">
                {project.name}
              </h1>
              <p className="mt-1.5 max-w-3xl text-sm leading-6 text-zinc-600 dark:text-zinc-300">
                {project.description ||
                  "当前项目尚未填写描述，可直接在右上角编辑项目补充背景信息。"}
              </p>
            </div>
            <div className="flex flex-wrap gap-2">
              {tagList.length === 0 ? (
                <Badge variant="outline" className="rounded-full">
                  暂无标签
                </Badge>
              ) : (
                tagList.map((tag) => (
                  <Badge key={tag} variant="outline" className="rounded-full">
                    {tag}
                  </Badge>
                ))
              )}
            </div>
          </div>

          <div className="grid gap-3 lg:min-w-[340px]">
            <div className="grid grid-cols-2 gap-2.5">
              <div className={tokens.stat}>
                <p className="text-[11px] uppercase tracking-[0.24em] text-zinc-500 dark:text-zinc-400">
                  环境
                </p>
                <p className="mt-1 text-xl font-semibold text-zinc-950 dark:text-white">
                  {project.environments.length}
                </p>
              </div>
              <div className={tokens.stat}>
                <p className="text-[11px] uppercase tracking-[0.24em] text-zinc-500 dark:text-zinc-400">
                  成功构建
                </p>
                <p className="mt-1 text-xl font-semibold text-zinc-950 dark:text-white">
                  {successCount}
                </p>
              </div>
              <div className={tokens.stat}>
                <p className="text-[11px] uppercase tracking-[0.24em] text-zinc-500 dark:text-zinc-400">
                  保留数量
                </p>
                <p className="mt-1 text-xl font-semibold text-zinc-950 dark:text-white">
                  {project.max_artifacts}
                </p>
              </div>
              <div className={tokens.stat}>
                <p className="text-[11px] uppercase tracking-[0.24em] text-zinc-500 dark:text-zinc-400">
                  Webhook
                </p>
                <p className="mt-1 text-xl font-semibold text-zinc-950 capitalize dark:text-white">
                  {project.webhook_type || "auto"}
                </p>
              </div>
            </div>
            <div className="flex flex-wrap justify-end gap-2">
              <Button variant="outline" onClick={() => setEditDialogOpen(true)}>
                <Pencil className="size-4" />
                编辑项目
              </Button>
              <Link to="/projects">
                <Button variant="outline">返回列表</Button>
              </Link>
            </div>
          </div>
        </div>
      </section>

      <div className="grid gap-4 xl:grid-cols-[1.3fr_0.9fr]">
        <Card className={tokens.panel}>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2">
              <Layers3 className="size-4" />
              仓库与归档策略
            </CardTitle>
            <CardDescription>仓库来源、认证方式与构建物保留规则。</CardDescription>
          </CardHeader>
          <CardContent className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
            <EnvironmentMeta
              label="仓库地址"
              value={project.repo_url}
              mono
              icon={<Package className="mt-0.5 size-4 shrink-0 text-zinc-500" />}
              className="md:col-span-2 xl:col-span-2"
              valueClassName={REPO_URL_VALUE_CLASS}
            />
            <EnvironmentMeta label="认证方式" value={formatRepoAuthType(project.repo_auth_type)} />
            <EnvironmentMeta label="构建物格式" value={getArtifactLabel(project.artifact_format)} />
            <EnvironmentMeta label="保留数量" value={`${project.max_artifacts} 个`} />
          </CardContent>
        </Card>

        <Card className={tokens.panel}>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2">
              <Radio className="size-4" />
              Webhook 控制面板
            </CardTitle>
            <CardDescription>{webhookGuide.title} 配置指引与当前入口。</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {webhookUrl ? (
              <div>
                <p className="text-xs uppercase tracking-[0.24em] text-zinc-500 dark:text-zinc-400">
                  Webhook URL
                </p>
                <div className="mt-2 flex items-center gap-2">
                  <code className="min-w-0 flex-1 break-all rounded-2xl bg-black/5 px-3 py-2 font-mono text-sm leading-5 whitespace-normal dark:bg-white/5">
                    {webhookUrl}
                  </code>
                  <Button variant="outline" size="icon" onClick={copyWebhook}>
                    <Copy className="size-4" />
                  </Button>
                </div>
              </div>
            ) : (
              <p className="rounded-2xl border border-dashed border-zinc-300 px-4 py-3 text-sm text-zinc-500 dark:border-zinc-700 dark:text-zinc-400">
                当前项目尚未生成 Webhook 地址。
              </p>
            )}
            <div className="rounded-3xl border border-black/5 bg-black/[0.03] p-3.5 dark:border-white/5 dark:bg-white/[0.03]">
              <div className="space-y-1 text-sm text-zinc-600 dark:text-zinc-300">
                {webhookGuide.headers.map((item) => (
                  <p key={item}>{item}</p>
                ))}
              </div>
              <p className="mt-3 text-sm text-zinc-500 dark:text-zinc-400">{webhookGuide.sample}</p>
            </div>
          </CardContent>
        </Card>
      </div>

      <section className="space-y-3">
        <div className="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <p className="text-[11px] uppercase tracking-[0.3em] text-zinc-500 dark:text-zinc-400">
              Environment Grid
            </p>
            <h2 className="mt-1 text-2xl font-semibold tracking-tight text-zinc-950 dark:text-white">
              项目环境
            </h2>
            <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
              统一展示分支、部署、定时与最近构建，并直接在表格内执行常用操作。
            </p>
          </div>
          <Button
            onClick={() => {
              setEditingEnv(null);
              setEnvFormOpen(true);
            }}
          >
            <Plus className="size-4" />
            新建环境
          </Button>
        </div>

        {project.environments.length === 0 ? (
          <div
            className={cn(
              tokens.envCard,
              "flex min-h-48 flex-col items-center justify-center text-center",
            )}
          >
            <Sparkles className="size-8 text-zinc-400" />
            <p className="mt-4 text-lg font-semibold text-zinc-900 dark:text-zinc-100">
              还没有环境
            </p>
            <p className="mt-1 max-w-md text-sm text-zinc-500 dark:text-zinc-400">
              先创建至少一个环境，再配置分支、构建脚本、部署目标和定时策略。
            </p>
          </div>
        ) : (
          project.environments.map((env) => (
            <div key={env.id} className={tokens.envCard}>
              <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                <div className="min-w-0 space-y-2">
                  <div className="flex flex-wrap items-center gap-2">
                    <h3 className="text-lg font-semibold text-zinc-950 dark:text-zinc-50">
                      {env.name}
                    </h3>
                    <Badge variant="outline" className="rounded-full">
                      {getScriptTypeLabel(env.build_script_type)}
                    </Badge>
                    {env.cron_enabled && env.cron_expression ? (
                      <Badge variant="secondary" className="rounded-full">
                        定时中
                      </Badge>
                    ) : null}
                  </div>
                </div>

                <div className="flex shrink-0 flex-wrap gap-2 lg:justify-end">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setEditingEnv(env);
                      setEnvFormOpen(true);
                    }}
                  >
                    <Pencil className="size-4" />
                    编辑环境
                  </Button>
                  <Button variant="outline" size="sm" onClick={() => setTriggerDialogEnv(env)}>
                    <Settings2 className="size-4" />
                    高级触发
                  </Button>
                  <Button
                    size="sm"
                    onClick={() => triggerBuild(env.id)}
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
                </div>
              </div>

              <div className="mt-3 grid gap-2 md:grid-cols-2 xl:grid-cols-4 2xl:grid-cols-5">
                <EnvironmentMeta
                  label="分支"
                  value={env.branch || "-"}
                  mono
                  icon={<GitBranch className="mt-0.5 size-4 shrink-0 text-zinc-500" />}
                />
                <EnvironmentMeta
                  label="构建脚本"
                  value={
                    env.build_script
                      ? `${getScriptTypeLabel(env.build_script_type)} · 已配置`
                      : "未配置"
                  }
                />
                <EnvironmentMeta
                  label="产物目录"
                  value={env.build_output_dir || "未配置"}
                  mono
                  icon={<Package className="mt-0.5 size-4 shrink-0 text-zinc-500" />}
                />
                <EnvironmentMeta
                  label="部署方式"
                  value={env.deploy_method || "未部署"}
                  icon={<Server className="mt-0.5 size-4 shrink-0 text-zinc-500" />}
                />
                <EnvironmentMeta
                  label="部署路径"
                  value={env.deploy_path || "未配置"}
                  mono
                  valueClassName="break-all whitespace-normal leading-5"
                />
                <EnvironmentMeta label="变量组" value={`${env.var_group_ids?.length ?? 0} 个`} />
                <EnvironmentMeta
                  label="Cron 表达式"
                  value={env.cron_enabled ? env.cron_expression || "已启用" : "未启用"}
                  mono
                  icon={<Clock className="mt-0.5 size-4 shrink-0 text-zinc-500" />}
                />
              </div>

              <div className="mt-3 rounded-[22px] border border-black/6 bg-black/[0.025] p-1.5 dark:border-white/6 dark:bg-white/[0.02]">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-20">编号</TableHead>
                      <TableHead className="w-28">状态</TableHead>
                      <TableHead className="w-32">Commit</TableHead>
                      <TableHead>提交信息</TableHead>
                      <TableHead className="w-32">触发</TableHead>
                      <TableHead className="w-36">耗时 / 时间</TableHead>
                      <TableHead className="w-[320px]">快捷操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {(buildsByEnv[env.id] ?? []).length === 0 ? (
                      <TableRow>
                        <TableCell
                          colSpan={7}
                          className="py-10 text-center text-zinc-500 dark:text-zinc-400"
                        >
                          暂无构建记录
                        </TableCell>
                      </TableRow>
                    ) : (
                      (buildsByEnv[env.id] ?? []).map((build) => {
                        const statusInfo = BUILD_STATUSES[
                          build.status as keyof typeof BUILD_STATUSES
                        ] ?? {
                          label: build.status,
                          color: "bg-zinc-500",
                        };
                        const canDownload =
                          build.status === "success" && Boolean(build.artifact_path);
                        const canDeploy = build.status === "success";
                        const canRetry = ["failed", "cancelled"].includes(build.status);
                        const downloadKey = `download:${build.id}`;
                        const deployKey = `deploy:${build.id}`;
                        const retryKey = `retry:${build.id}`;

                        return (
                          <TableRow key={build.id}>
                            <TableCell className="py-2.5 font-semibold text-zinc-900 dark:text-zinc-100">
                              #{build.build_number}
                            </TableCell>
                            <TableCell className="py-2.5">
                              <Badge className={cn(statusInfo.color, "text-white")}>
                                {statusInfo.label}
                              </Badge>
                            </TableCell>
                            <TableCell className="py-2.5 font-mono text-xs text-zinc-600 dark:text-zinc-300">
                              {build.commit_hash ? build.commit_hash.slice(0, 7) : "-"}
                            </TableCell>
                            <TableCell className="max-w-[280px] py-2.5">
                              <div className="truncate text-sm text-zinc-700 dark:text-zinc-200">
                                {build.commit_message || "-"}
                              </div>
                              <div className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
                                {build.branch || env.branch}
                              </div>
                            </TableCell>
                            <TableCell className="py-2.5 text-sm text-zinc-600 dark:text-zinc-300">
                              {build.trigger_type || "-"}
                            </TableCell>
                            <TableCell className="py-2.5 text-sm text-zinc-600 dark:text-zinc-300">
                              <p>{formatDuration(build.duration_ms)}</p>
                              <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
                                {formatDateTime(build.created_at)}
                              </p>
                            </TableCell>
                            <TableCell className="py-2.5">
                              <div className="flex flex-wrap gap-2">
                                {canDownload ? (
                                  <Button
                                    size="sm"
                                    variant="outline"
                                    disabled={buildActionLoading === downloadKey}
                                    onClick={() => handleBuildAction("download", build)}
                                  >
                                    {buildActionLoading === downloadKey ? (
                                      <Loader2 className="size-4 animate-spin" />
                                    ) : (
                                      <Download className="size-4" />
                                    )}
                                    下载
                                  </Button>
                                ) : null}
                                {canDeploy ? (
                                  <Button
                                    size="sm"
                                    variant="outline"
                                    disabled={buildActionLoading === deployKey}
                                    onClick={() => handleBuildAction("deploy", build)}
                                  >
                                    {buildActionLoading === deployKey ? (
                                      <Loader2 className="size-4 animate-spin" />
                                    ) : (
                                      <Rocket className="size-4" />
                                    )}
                                    部署
                                  </Button>
                                ) : null}
                                {canRetry ? (
                                  <Button
                                    size="sm"
                                    variant="outline"
                                    disabled={buildActionLoading === retryKey}
                                    onClick={() => handleBuildAction("retry", build)}
                                  >
                                    {buildActionLoading === retryKey ? (
                                      <Loader2 className="size-4 animate-spin" />
                                    ) : (
                                      <RotateCcw className="size-4" />
                                    )}
                                    重建
                                  </Button>
                                ) : null}
                                <Link to={`/builds/${build.id}`}>
                                  <Button size="sm" className={tokens.detailButton}>
                                    <ExternalLink className="size-4" />
                                    详情
                                  </Button>
                                </Link>
                              </div>
                            </TableCell>
                          </TableRow>
                        );
                      })
                    )}
                  </TableBody>
                </Table>
              </div>
            </div>
          ))
        )}
      </section>

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
