import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router'
import { GitBranch, Layers, Loader2, Play, Search } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
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
import { cn } from '@/lib/utils'

interface EnvironmentRow {
  id: number
  project_id: number
  name: string
  branch: string
  sort_order: number
}

interface ProjectBrief {
  id: number
  name: string
  environments?: EnvironmentRow[]
}

type EnvFlatRow = {
  projectId: number
  projectName: string
  env: EnvironmentRow
}

export function EnvironmentListPage() {
  const [projects, setProjects] = useState<ProjectBrief[]>([])
  const [loading, setLoading] = useState(true)
  const [projectFilter, setProjectFilter] = useState<string>('all')
  const [nameQuery, setNameQuery] = useState('')
  const [triggeringKey, setTriggeringKey] = useState<string | null>(null)

  const fetchAllProjects = useCallback(async () => {
    setLoading(true)
    try {
      const acc: ProjectBrief[] = []
      for (let page = 1; ; page += 1) {
        const res = await api.get<PaginatedData<ProjectBrief>>(
          `/projects?page=${page}&page_size=100`,
        )
        if (res.code !== 0 || !res.data) break
        const data = res.data as PaginatedData<ProjectBrief>
        const items = data.items ?? []
        acc.push(...items)
        if (items.length < 100 || page >= (data.total_pages ?? 1)) break
      }
      setProjects(acc)
    } catch {
      toast.error('加载失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchAllProjects()
  }, [fetchAllProjects])

  const flatRows: EnvFlatRow[] = useMemo(() => {
    const rows: EnvFlatRow[] = []
    for (const p of projects) {
      const envs = p.environments ?? []
      for (const env of envs) {
        rows.push({ projectId: p.id, projectName: p.name, env })
      }
    }
    rows.sort((a, b) => {
      const pn = a.projectName.localeCompare(b.projectName, 'zh-CN')
      if (pn !== 0) return pn
      if (a.env.sort_order !== b.env.sort_order) return a.env.sort_order - b.env.sort_order
      return a.env.name.localeCompare(b.env.name, 'zh-CN')
    })
    return rows
  }, [projects])

  const filteredRows = useMemo(() => {
    const pid =
      projectFilter === 'all' ? null : Number.parseInt(projectFilter, 10)
    const q = nameQuery.trim().toLowerCase()
    return flatRows.filter((row) => {
      if (pid !== null && !Number.isNaN(pid) && row.projectId !== pid) return false
      if (q && !row.env.name.toLowerCase().includes(q)) return false
      return true
    })
  }, [flatRows, projectFilter, nameQuery])

  const triggerBuild = async (projectId: number, envId: number) => {
    const key = `${projectId}:${envId}`
    setTriggeringKey(key)
    try {
      const res = await api.post<{ id: number }>(`/projects/${projectId}/builds`, {
        environment_id: envId,
      })
      if (res.code !== 0 || !res.data) {
        toast.error(res.message || '触发失败')
        return
      }
      toast.success('构建已触发')
    } catch {
      toast.error('触发失败')
    } finally {
      setTriggeringKey(null)
    }
  }

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="size-6 animate-spin rounded-full border-2 border-border border-t-emerald-500" />
      </div>
    )
  }

  return (
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
                共 {flatRows.length} 个环境；筛选后 {filteredRows.length} 条
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-end">
            <div className="space-y-1.5 sm:w-[220px]">
              <Label className="text-muted-foreground text-xs">项目</Label>
              <Select value={projectFilter} onValueChange={setProjectFilter}>
                <SelectTrigger>
                  <SelectValue placeholder="选择项目" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">全部项目</SelectItem>
                  {projects
                    .slice()
                    .sort((a, b) => a.name.localeCompare(b.name, 'zh-CN'))
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
                  <TableHead className="min-w-[140px]">项目</TableHead>
                  <TableHead className="min-w-[120px]">环境</TableHead>
                  <TableHead className="w-[160px]">分支</TableHead>
                  <TableHead className="w-[120px] text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredRows.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={4}
                      className="text-muted-foreground py-10 text-center text-sm"
                    >
                      无匹配环境
                    </TableCell>
                  </TableRow>
                ) : (
                  filteredRows.map((row) => {
                    const tKey = `${row.projectId}:${row.env.id}`
                    const busy = triggeringKey === tKey
                    return (
                      <TableRow key={tKey}>
                        <TableCell>
                          <Link
                            to={`/projects/${row.projectId}`}
                            className={cn(
                              'font-medium text-emerald-600 hover:text-emerald-500',
                              'dark:text-emerald-400 dark:hover:text-emerald-300',
                            )}
                          >
                            {row.projectName}
                          </Link>
                        </TableCell>
                        <TableCell className="font-medium">{row.env.name}</TableCell>
                        <TableCell>
                          <span className="text-muted-foreground inline-flex items-center gap-1.5 font-mono text-xs">
                            <GitBranch className="size-3.5 shrink-0" />
                            {row.env.branch || '-'}
                          </span>
                        </TableCell>
                        <TableCell className="text-right">
                          <Button
                            size="sm"
                            onClick={() => triggerBuild(row.projectId, row.env.id)}
                            disabled={busy}
                          >
                            {busy ? (
                              <Loader2 className="size-3.5 animate-spin" />
                            ) : (
                              <Play className="size-3.5" />
                            )}
                            {busy ? '触发中' : '触发构建'}
                          </Button>
                        </TableCell>
                      </TableRow>
                    )
                  })
                )}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
