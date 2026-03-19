import { useState, useEffect, useCallback } from 'react'
import { useParams, Link } from 'react-router'
import { Copy, Play, ExternalLink, Clock, Settings2, Plus, Pencil, Loader2 } from 'lucide-react'
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  createColumnHelper,
} from '@tanstack/react-table'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { api } from '@/lib/api'
import type { PaginatedData } from '@/lib/api'
import { BUILD_STATUSES } from '@/lib/constants'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'
import { ProjectFormDialog } from '@/pages/projects/form'
import { EnvironmentFormDialog } from '@/pages/projects/environment-form'
import { BUILD_SCRIPT_TYPES } from '@/lib/constants'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'

interface Environment {
  id: number
  project_id: number
  name: string
  branch: string
  build_script: string
  build_script_type: string
  build_output_dir: string
  deploy_server_id: number | null
  deploy_method: string
  deploy_path: string
  post_deploy_script: string
  cron_expression: string
  cron_enabled: boolean
  sort_order: number
  var_group_ids: number[]
}

interface Project {
  id: number
  name: string
  description: string
  group_name: string
  tags: string
  repo_url: string
  webhook_secret?: string
  webhook_type: string
  webhook_ref_path: string
  webhook_commit_path: string
  webhook_message_path: string
  environments: Environment[]
}

interface Build {
  id: number
  build_number: number
  status: string
  trigger_type: string
  branch: string
  commit_hash: string
  commit_message: string
  duration_ms: number
  created_at: string
}

const WEBHOOK_GUIDES: Record<string, { title: string; headers: string[]; sample: string }> = {
  auto: {
    title: '自动识别',
    headers: ['GitHub: `X-GitHub-Event: push`', 'GitLab: `X-Gitlab-Event: Push Hook`', 'Gitea: `X-Gitea-Event: push`', 'Bitbucket: `X-Event-Key: repo:push`'],
    sample: '服务端会按请求头自动识别平台并解析 push payload。',
  },
  github: {
    title: 'GitHub',
    headers: ['Header: `X-GitHub-Event: push`'],
    sample: '仓库 Webhook 选择 JSON，触发事件勾选 push。',
  },
  gitlab: {
    title: 'GitLab',
    headers: ['Header: `X-Gitlab-Event: Push Hook`'],
    sample: 'GitLab 项目集成里启用 Push events 即可。',
  },
  gitea: {
    title: 'Gitea',
    headers: ['Header: `X-Gitea-Event: push`'],
    sample: 'Gitea Webhook 选择 Gitea/GitHub 风格 JSON，触发 push。',
  },
  bitbucket: {
    title: 'Bitbucket',
    headers: ['Header: `X-Event-Key: repo:push`'],
    sample: 'Bitbucket Cloud/Server 推送事件会使用 `push.changes[0]` 结构。',
  },
  generic: {
    title: '通用 JSON',
    headers: ['自定义 JSONPath: `$.ref`、`$.head_commit.id`、`$.head_commit.message`'],
    sample: '支持 `$.field` 与 `$.list[0].field` 形式的路径。',
  },
}

export function ProjectDetailPage() {
  const { id } = useParams<{ id: string }>()
  const [project, setProject] = useState<Project | null>(null)
  const [buildsByEnv, setBuildsByEnv] = useState<Record<number, Build[]>>({})
  const [loading, setLoading] = useState(true)
  const [triggering, setTriggering] = useState<number | null>(null)
  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [triggerDialogEnv, setTriggerDialogEnv] = useState<Environment | null>(null)
  const [envFormOpen, setEnvFormOpen] = useState(false)
  const [editingEnv, setEditingEnv] = useState<Environment | null>(null)

  const fetchProject = useCallback(async () => {
    if (!id) return
    try {
      const res = await api.get<Project>(`/projects/${id}`)
      if (res.code === 0 && res.data) {
        setProject(res.data)
        const envs = res.data.environments || []
        for (const env of envs) {
          const br = await api.get<PaginatedData<Build>>(`/projects/${id}/builds?environment_id=${env.id}&page_size=20`)
          if (br.code === 0 && br.data) {
            const items = (br.data as PaginatedData<Build>).items || []
            setBuildsByEnv((prev) => ({ ...prev, [env.id]: items }))
          }
        }
      }
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => {
    fetchProject()
  }, [fetchProject])

  const triggerBuild = async (envId: number, branch?: string, commitHash?: string) => {
    if (!id) return
    setTriggering(envId)
    try {
      const payload: Record<string, unknown> = { environment_id: envId }
      if (branch) payload.branch = branch
      if (commitHash) payload.commit_hash = commitHash
      const res = await api.post<Build>(`/projects/${id}/builds`, payload)
      if (res.code === 0 && res.data) {
        toast.success('构建已触发')
        setBuildsByEnv((prev) => ({
          ...prev,
          [envId]: [res.data!, ...(prev[envId] || [])],
        }))
      } else {
        toast.error(res.message || '触发失败')
      }
    } catch {
      toast.error('触发失败')
    } finally {
      setTriggering(null)
    }
  }

  const copyWebhook = () => {
    if (!project?.webhook_secret) return
    const base = window.location.origin
    const url = `${base}/api/v1/webhook/${project.id}/${project.webhook_secret}`
    navigator.clipboard.writeText(url)
    toast.success('Webhook URL 已复制')
  }

  if (loading || !project) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-2 border-zinc-600 border-t-zinc-300" />
      </div>
    )
  }

  const webhookUrl = project.webhook_secret
    ? `${window.location.origin}/api/v1/webhook/${project.id}/${project.webhook_secret}`
    : ''
  const tagList = (project.tags || '').split(',').map((item) => item.trim()).filter(Boolean)
  const webhookGuide = WEBHOOK_GUIDES[project.webhook_type || 'auto'] ?? WEBHOOK_GUIDES.auto

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{project.name}</h1>
          <p className="mt-1 text-sm text-zinc-500">{project.description || '暂无描述'}</p>
          <div className="mt-3 flex flex-wrap gap-2">
            <Badge variant="outline">{project.group_name || '未分组'}</Badge>
            {tagList.map((tag) => (
              <Badge key={tag} variant="secondary">
                {tag}
              </Badge>
            ))}
          </div>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => setEditDialogOpen(true)}>编辑</Button>
          <Link to="/projects">
            <Button variant="outline">返回列表</Button>
          </Link>
        </div>
      </div>

      <Card className="border-zinc-200 dark:border-zinc-800">
        <CardHeader>
          <CardTitle>项目信息</CardTitle>
          <CardDescription>仓库与 Webhook 配置</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <p className="text-sm font-medium text-zinc-500">仓库地址</p>
            <p className="mt-1 font-mono text-sm">{project.repo_url}</p>
          </div>
          {webhookUrl && (
            <div>
              <p className="text-sm font-medium text-zinc-500">Webhook URL</p>
              <div className="mt-1 flex items-center gap-2">
                <code className="flex-1 truncate rounded bg-zinc-100 dark:bg-zinc-800 px-2 py-1.5 text-sm">
                  {webhookUrl}
                </code>
                <Button variant="outline" size="icon" onClick={copyWebhook}>
                  <Copy className="size-4" />
                </Button>
              </div>
            </div>
          )}
          <div className="rounded-xl border border-zinc-200 bg-zinc-50/80 p-4 dark:border-zinc-800 dark:bg-zinc-900/50">
            <p className="text-sm font-medium">{webhookGuide.title} 配置指引</p>
            <div className="mt-2 space-y-1 text-sm text-zinc-500">
              {webhookGuide.headers.map((header) => (
                <p key={header}>{header}</p>
              ))}
            </div>
            <p className="mt-3 text-sm text-zinc-600 dark:text-zinc-300">{webhookGuide.sample}</p>
            {project.webhook_type === 'generic' && (
              <div className="mt-3 grid gap-2 text-xs text-zinc-500 sm:grid-cols-3">
                <div>
                  <p className="font-medium text-zinc-700 dark:text-zinc-200">Ref</p>
                  <p className="font-mono">{project.webhook_ref_path || '-'}</p>
                </div>
                <div>
                  <p className="font-medium text-zinc-700 dark:text-zinc-200">Commit</p>
                  <p className="font-mono">{project.webhook_commit_path || '-'}</p>
                </div>
                <div>
                  <p className="font-medium text-zinc-700 dark:text-zinc-200">Message</p>
                  <p className="font-mono">{project.webhook_message_path || '-'}</p>
                </div>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      <Card className="border-zinc-200 dark:border-zinc-800">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>环境</CardTitle>
              <CardDescription>按环境查看构建历史</CardDescription>
            </div>
            <Button size="sm" onClick={() => { setEditingEnv(null); setEnvFormOpen(true) }}>
              <Plus className="size-4" />
              新建环境
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue={project.environments?.[0]?.id?.toString() ?? '0'}>
            <TabsList className="mb-4">
              {project.environments?.map((env) => (
                <TabsTrigger key={env.id} value={String(env.id)}>
                  {env.name}
                </TabsTrigger>
              ))}
            </TabsList>
            {project.environments?.map((env) => (
              <TabsContent key={env.id} value={String(env.id)}>
                <div className="space-y-4">
                  <div className="flex flex-wrap gap-4 rounded-lg border border-zinc-200 dark:border-zinc-800 p-4">
                    <div>
                      <p className="text-xs text-zinc-500">分支</p>
                      <p className="font-medium">{env.branch}</p>
                    </div>
                    <div>
                      <p className="text-xs text-zinc-500">构建脚本</p>
                      <div className="flex items-center gap-1.5">
                        <Badge variant="outline" className="text-[10px] px-1.5 py-0">
                          {BUILD_SCRIPT_TYPES.find((t) => t.value === env.build_script_type)?.label || 'Bash'}
                        </Badge>
                        <p className="font-medium font-mono text-sm">{env.build_script || '-'}</p>
                      </div>
                    </div>
                    <div>
                      <p className="text-xs text-zinc-500">产物目录</p>
                      <p className="font-medium font-mono text-sm">{env.build_output_dir || '-'}</p>
                    </div>
                    <div>
                      <p className="text-xs text-zinc-500">部署方式</p>
                      <p className="font-medium">{env.deploy_method || '-'}</p>
                    </div>
                    <div>
                      <p className="text-xs text-zinc-500">变量组</p>
                      <p className="font-medium">{env.var_group_ids?.length ?? 0} 个</p>
                    </div>
                    <div>
                      <p className="text-xs text-zinc-500">部署路径</p>
                      <p className="font-medium">{env.deploy_path || '-'}</p>
                    </div>
                    {env.cron_enabled && env.cron_expression && (
                      <div>
                        <p className="text-xs text-zinc-500">定时构建</p>
                        <div className="flex items-center gap-1.5">
                          <Clock className="size-3.5 text-blue-500" />
                          <p className="font-medium font-mono text-sm">{env.cron_expression}</p>
                          <Badge className="bg-blue-500 text-white text-[10px] px-1.5 py-0">已启用</Badge>
                        </div>
                      </div>
                    )}
                    <div className="flex items-end gap-2 ml-auto">
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        onClick={() => { setEditingEnv(env); setEnvFormOpen(true) }}
                      >
                        <Pencil className="size-4" />
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setTriggerDialogEnv(env)}
                      >
                        <Settings2 className="size-4" />
                        高级触发
                      </Button>
                      <Button
                        onClick={() => triggerBuild(env.id)}
                        disabled={triggering === env.id}
                      >
                        {triggering === env.id ? '触发中...' : (
                          <>
                            <Play className="size-4" />
                            触发构建
                          </>
                        )}
                      </Button>
                    </div>
                  </div>

                  <BuildHistoryTable
                    builds={buildsByEnv[env.id] || []}
                  />
                </div>
              </TabsContent>
            ))}
          </Tabs>
        </CardContent>
      </Card>

      <ProjectFormDialog
        open={editDialogOpen}
        onOpenChange={setEditDialogOpen}
        editId={Number(id)}
        onSuccess={() => fetchProject()}
      />

      <EnvironmentFormDialog
        open={envFormOpen}
        onOpenChange={setEnvFormOpen}
        projectId={Number(id)}
        editEnv={editingEnv}
        onSuccess={() => fetchProject()}
      />

      <TriggerBuildDialog
        env={triggerDialogEnv}
        open={!!triggerDialogEnv}
        onOpenChange={(open) => { if (!open) setTriggerDialogEnv(null) }}
        onTrigger={(envId, branch, commitHash) => {
          setTriggerDialogEnv(null)
          triggerBuild(envId, branch, commitHash)
        }}
        triggering={triggering}
        projectId={Number(id)}
      />
    </div>
  )
}

function TriggerBuildDialog({
  env,
  open,
  onOpenChange,
  onTrigger,
  triggering,
  projectId,
}: {
  env: Environment | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onTrigger: (envId: number, branch?: string, commitHash?: string) => void
  triggering: number | null
  projectId: number
}) {
  const [branch, setBranch] = useState('')
  const [commitHash, setCommitHash] = useState('')
  const [branches, setBranches] = useState<string[]>([])
  const [branchesLoading, setBranchesLoading] = useState(false)
  const [branchPopoverOpen, setBranchPopoverOpen] = useState(false)

  useEffect(() => {
    if (!open) {
      setBranch('')
      setCommitHash('')
      setBranches([])
      return
    }
    setBranchesLoading(true)
    api.get<string[]>(`/projects/${projectId}/branches`).then((res) => {
      if (res.code === 0 && res.data) {
        setBranches(Array.isArray(res.data) ? res.data : [])
      }
    }).finally(() => setBranchesLoading(false))
  }, [open, projectId])

  if (!env) return null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[420px]">
        <DialogHeader>
          <DialogTitle>触发构建 - {env.name}</DialogTitle>
          <DialogDescription>
            可指定分支或 Commit，留空则使用环境默认分支（{env.branch}）
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-2">
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
                  {branchesLoading && <Loader2 className="size-4 animate-spin ml-2" />}
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-[--radix-popover-trigger-width] p-0" align="start">
                <Command>
                  <CommandInput
                    placeholder="搜索或输入分支名..."
                    value={branch}
                    onValueChange={(v: string) => setBranch(v)}
                  />
                  <CommandList>
                    <CommandEmpty>
                      {branchesLoading ? '加载中...' : '无匹配分支，可直接输入'}
                    </CommandEmpty>
                    <CommandGroup>
                      {branches.map((b) => (
                        <CommandItem
                          key={b}
                          value={b}
                          onSelect={(v: string) => {
                            setBranch(v)
                            setBranchPopoverOpen(false)
                          }}
                        >
                          {b}
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
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>取消</Button>
          <Button
            onClick={() => onTrigger(env.id, branch || undefined, commitHash || undefined)}
            disabled={triggering === env.id}
          >
            {triggering === env.id ? '触发中...' : (
              <>
                <Play className="size-4" />
                触发构建
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function BuildHistoryTable({ builds }: { builds: Build[] }) {
  const columnHelper = createColumnHelper<Build>()
  const columns = [
    columnHelper.accessor('build_number', { header: '#', cell: ({ getValue }) => `#${getValue()}` }),
    columnHelper.accessor('status', {
      header: '状态',
      cell: ({ getValue }) => {
        const s = String(getValue())
        const info = BUILD_STATUSES[s as keyof typeof BUILD_STATUSES] ?? { label: s, color: 'bg-gray-500' }
        return <Badge className={cn(info.color, 'text-white')}>{info.label}</Badge>
      },
    }),
    columnHelper.accessor('commit_hash', {
      header: 'Commit',
      cell: ({ getValue }) => (
        <span className="font-mono text-xs">{String(getValue()).slice(0, 7) || '-'}</span>
      ),
    }),
    columnHelper.accessor('commit_message', {
      header: '提交信息',
      cell: ({ getValue }) => <span className="max-w-[200px] truncate block">{getValue() || '-'}</span>,
    }),
    columnHelper.accessor('trigger_type', { header: '触发' }),
    columnHelper.accessor('duration_ms', {
      header: '耗时',
      cell: ({ getValue }) => {
        const ms = Number(getValue())
        return ms ? `${(ms / 1000).toFixed(1)}s` : '-'
      },
    }),
    columnHelper.accessor('created_at', {
      header: '时间',
      cell: ({ getValue }) => new Date(String(getValue())).toLocaleString('zh-CN'),
    }),
    columnHelper.display({
      id: 'actions',
      header: '操作',
      cell: ({ row }) => {
        const b = row.original
        return (
          <div className="flex gap-1">
            <Link to={`/builds/${b.id}`}>
              <Button variant="ghost" size="icon-sm">
                <ExternalLink className="size-4" />
              </Button>
            </Link>
          </div>
        )
      },
    }),
  ]

  const table = useReactTable({
    data: builds,
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  return (
    <Table>
      <TableHeader>
        {table.getHeaderGroups().map((hg) => (
          <TableRow key={hg.id}>
            {hg.headers.map((h) => (
              <TableHead key={h.id}>{flexRender(h.column.columnDef.header, h.getContext())}</TableHead>
            ))}
          </TableRow>
        ))}
      </TableHeader>
      <TableBody>
        {table.getRowModel().rows.length === 0 ? (
          <TableRow>
            <TableCell colSpan={8} className="text-center text-zinc-500 py-8">
              暂无构建记录
            </TableCell>
          </TableRow>
        ) : (
          table.getRowModel().rows.map((row) => (
            <TableRow key={row.id}>
              {row.getVisibleCells().map((cell) => (
                <TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>
              ))}
            </TableRow>
          ))
        )}
      </TableBody>
    </Table>
  )
}
