import { useCallback, useEffect, useRef, useState } from "react";
import { Link, useParams } from "react-router";
import { Copy, ExternalLink, GitBranch, Loader2, Pencil, Play, Plus, Sparkles } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { TooltipProvider } from "@/components/ui/tooltip";
import { api } from "@/lib/api";
import type { PaginatedData } from "@/lib/api";
import { EnvironmentBuildsTable, extractFileName } from "@/components/environment-builds-table";
import { ARTIFACT_FORMATS, BUILD_SCRIPT_TYPES } from "@/lib/constants";
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

const BUILD_PAGE_SIZE = 20;

function formatRepoAuthType(value: string): string {
  if (!value || value === "none") return "无需认证";
  if (value === "credential") return "凭证";
  return value;
}

function getArtifactLabel(format: string): string {
  return ARTIFACT_FORMATS.find((item) => item.value === format)?.label ?? "Gzip (.tar.gz)";
}

function getScriptTypeLabel(type: string): string {
  return BUILD_SCRIPT_TYPES.find((item) => item.value === type)?.label ?? "Bash";
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
      {label} <span className={cn("text-foreground", mono && "font-mono")}>{value || "-"}</span>
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
  const [buildPageByEnv, setBuildPageByEnv] = useState<Record<number, number>>({});
  const [buildPaginationByEnv, setBuildPaginationByEnv] = useState<
    Record<number, { total: number; total_pages: number }>
  >({});
  const [buildsLoading, setBuildsLoading] = useState(false);
  const [loading, setLoading] = useState(true);
  const [triggering, setTriggering] = useState<number | null>(null);
  const [buildActionLoading, setBuildActionLoading] = useState<string | null>(null);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
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
    setProject({ ...res.data, environments: envs });
    setLoading(false);
  }, [projectId]);

  const fetchBuildsForEnv = useCallback(
    async (envId: number, page: number, opts?: { silent?: boolean }) => {
      if (!projectId) return;
      if (!opts?.silent) setBuildsLoading(true);
      try {
        const res = await api.get<PaginatedData<Build>>(
          `/projects/${projectId}/builds?environment_id=${envId}&page=${page}&page_size=${BUILD_PAGE_SIZE}`,
        );
        if (res.code !== 0 || !res.data) return;
        const data = res.data as PaginatedData<Build>;
        const items = data.items ?? [];
        setBuildsByEnv((prev) => ({ ...prev, [envId]: items }));
        setBuildPaginationByEnv((prev) => ({
          ...prev,
          [envId]: {
            total: data.total ?? 0,
            total_pages: Math.max(
              1,
              data.total_pages ?? (Math.ceil((data.total ?? 0) / BUILD_PAGE_SIZE) || 1),
            ),
          },
        }));
      } finally {
        if (!opts?.silent) setBuildsLoading(false);
      }
    },
    [projectId],
  );

  const activeTabRef = useRef(activeTab);
  const buildPageByEnvRef = useRef(buildPageByEnv);
  activeTabRef.current = activeTab;
  buildPageByEnvRef.current = buildPageByEnv;

  const activeEnvId = activeTab ? Number(activeTab) : 0;
  const currentBuildPage = activeEnvId ? (buildPageByEnv[activeEnvId] ?? 1) : 1;

  useEffect(() => {
    if (!projectId || !activeEnvId) return;
    void fetchBuildsForEnv(activeEnvId, currentBuildPage);
  }, [projectId, activeEnvId, currentBuildPage, fetchBuildsForEnv]);

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
    const interval = window.setInterval(
      () => {
        void fetchProject();
        const tab = activeTabRef.current;
        if (tab) {
          const eid = Number(tab);
          void fetchBuildsForEnv(eid, buildPageByEnvRef.current[eid] ?? 1, { silent: true });
        }
      },
      hasActiveBuilds ? 4000 : 15000,
    );
    return () => window.clearInterval(interval);
  }, [projectId, project, hasActiveBuilds, fetchProject, fetchBuildsForEnv]);

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
      setBuildPageByEnv((prev) => ({ ...prev, [envId]: 1 }));
      await fetchBuildsForEnv(envId, 1);
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
      toast.success(action === "deploy" ? "开始部署" : "已重新构建");
      await fetchProject();
      const tab = activeTabRef.current;
      if (tab) {
        const eid = Number(tab);
        await fetchBuildsForEnv(eid, buildPageByEnvRef.current[eid] ?? 1, { silent: true });
      }
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
                const pagination = buildPaginationByEnv[env.id] ?? {
                  total: 0,
                  total_pages: 1,
                };
                const envBuildPage = buildPageByEnv[env.id] ?? 1;
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
                        {env.deploy_path && <MetaItem label="路径" value={env.deploy_path} mono />}
                        <MetaItem
                          label="Cron"
                          value={env.cron_enabled ? env.cron_expression || "已启用" : "未启用"}
                          mono
                        />
                        <MetaItem label="变量组" value={`${env.var_group_ids?.length ?? 0} 个`} />
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
                          {triggering === env.id ? "构建中" : "构建"}
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

                      <EnvironmentBuildsTable
                        env={env}
                        builds={builds}
                        loading={buildsLoading && env.id === activeEnvId}
                        page={envBuildPage}
                        totalPages={Math.max(1, pagination.total_pages)}
                        total={pagination.total}
                        onPageChange={(p) => {
                          setBuildPageByEnv((prev) => ({ ...prev, [env.id]: p }));
                        }}
                        buildActionLoading={buildActionLoading}
                        onBuildAction={handleBuildAction}
                      />
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
      </div>
    </TooltipProvider>
  );
}
