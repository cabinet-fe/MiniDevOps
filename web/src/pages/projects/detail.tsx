import { useState, useEffect, useCallback } from 'react'
import { useParams, Link } from 'react-router'
import { Copy, Play, ExternalLink } from 'lucide-react'
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
import { api } from '@/lib/api'
import type { PaginatedData } from '@/lib/api'
import { BUILD_STATUSES } from '@/lib/constants'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'
import { ProjectFormDialog } from '@/pages/projects/form'

interface Environment {
  id: number
  name: string
  branch: string
  build_script: string
  build_output_dir: string
  deploy_method: string
  deploy_path: string
}

interface Project {
  id: number
  name: string
  description: string
  repo_url: string
  webhook_secret?: string
  environments: Environment[]
}

interface Build {
  id: number
  build_number: number
  status: string
  trigger_type: string
  commit_hash: string
  commit_message: string
  duration_ms: number
  created_at: string
}

export function ProjectDetailPage() {
  const { id } = useParams<{ id: string }>()
  const [project, setProject] = useState<Project | null>(null)
  const [buildsByEnv, setBuildsByEnv] = useState<Record<number, Build[]>>({})
  const [loading, setLoading] = useState(true)
  const [triggering, setTriggering] = useState<number | null>(null)
  const [editDialogOpen, setEditDialogOpen] = useState(false)

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

  const triggerBuild = async (envId: number) => {
    if (!id) return
    setTriggering(envId)
    try {
      const res = await api.post<Build>(`/projects/${id}/builds`, { environment_id: envId })
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

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{project.name}</h1>
          <p className="mt-1 text-sm text-zinc-500">{project.description || '暂无描述'}</p>
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
        </CardContent>
      </Card>

      <Card className="border-zinc-200 dark:border-zinc-800">
        <CardHeader>
          <CardTitle>环境</CardTitle>
          <CardDescription>按环境查看构建历史</CardDescription>
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
                      <p className="text-xs text-zinc-500">部署方式</p>
                      <p className="font-medium">{env.deploy_method || '-'}</p>
                    </div>
                    <div>
                      <p className="text-xs text-zinc-500">部署路径</p>
                      <p className="font-medium">{env.deploy_path || '-'}</p>
                    </div>
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

                  <BuildHistoryTable
                    builds={buildsByEnv[env.id] || []}
                    projectId={Number(id)}
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
    </div>
  )
}

function BuildHistoryTable({ builds, projectId }: { builds: Build[]; projectId: number }) {
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
            <Link to={`/projects/${projectId}/builds/${b.id}`}>
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
