import { useState, useEffect } from 'react'
import { Link } from 'react-router'
import {
  FolderGit2,
  Zap,
  TrendingUp,
  Activity,
  ChevronRight,
  ExternalLink,
} from 'lucide-react'
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  createColumnHelper,
} from '@tanstack/react-table'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { api } from '@/lib/api'
import { BUILD_STATUSES } from '@/lib/constants'
import { cn } from '@/lib/utils'

interface DashboardStats {
  total_projects: number
  today_builds: number
  success_rate: number
  active_count: number
  group_summary: {
    group_name: string
    project_count: number
    environment_count: number
  }[]
}

interface Build {
  id: number
  project_id: number
  environment_id: number
  build_number: number
  status: string
  trigger_type: string
  commit_hash: string
  commit_message: string
  duration_ms: number
  created_at: string
}

interface BuildTrendItem {
  date: string
  status: string
  count: number
}

interface ChartTrendItem {
  date: string
  success: number
  failed: number
  total: number
}

const STAT_CARDS: Array<{
  key: keyof DashboardStats
  label: string
  icon: React.ElementType
  color: string
  bgColor: string
  suffix?: string
}> = [
  {
    key: 'total_projects',
    label: '项目总数',
    icon: FolderGit2,
    color: 'from-blue-500 to-blue-600',
    bgColor: 'bg-blue-500/10',
  },
  {
    key: 'today_builds',
    label: '今日构建',
    icon: Zap,
    color: 'from-amber-500 to-amber-600',
    bgColor: 'bg-amber-500/10',
  },
  {
    key: 'success_rate',
    label: '成功率',
    icon: TrendingUp,
    color: 'from-emerald-500 to-emerald-600',
    bgColor: 'bg-emerald-500/10',
    suffix: '%',
  },
  {
    key: 'active_count',
    label: '运行中',
    icon: Activity,
    color: 'from-violet-500 to-violet-600',
    bgColor: 'bg-violet-500/10',
  },
]

export function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [activeBuilds, setActiveBuilds] = useState<Build[]>([])
  const [recentBuilds, setRecentBuilds] = useState<Build[]>([])
  const [trend, setTrend] = useState<ChartTrendItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    const fetch = async () => {
      try {
        const [statsRes, activeRes, recentRes, trendRes] = await Promise.all([
          api.get<DashboardStats>('/dashboard/stats'),
          api.get<Build[]>('/dashboard/active-builds'),
          api.get<Build[]>('/dashboard/recent-builds?limit=10'),
          api.get<BuildTrendItem[]>('/dashboard/trend?days=7'),
        ])

        if (statsRes.code === 0 && statsRes.data) setStats(statsRes.data as DashboardStats)
        if (activeRes.code === 0 && activeRes.data) setActiveBuilds(Array.isArray(activeRes.data) ? (activeRes.data as Build[]) : [])
        if (recentRes.code === 0 && recentRes.data) setRecentBuilds(Array.isArray(recentRes.data) ? (recentRes.data as Build[]) : [])

        if (trendRes.code === 0 && trendRes.data) {
          const byDate = new Map<string, { success: number; failed: number; total: number }>()
          for (const item of trendRes.data as BuildTrendItem[]) {
            const cur = byDate.get(item.date) ?? { success: 0, failed: 0, total: 0 }
            cur.total += Number(item.count)
            if (item.status === 'success') cur.success += Number(item.count)
            else if (item.status === 'failed') cur.failed += Number(item.count)
            byDate.set(item.date, cur)
          }
          setTrend(
            Array.from(byDate.entries())
              .sort((a, b) => a[0].localeCompare(b[0]))
              .map(([date, v]) => ({ date: date.slice(5), success: v.success, failed: v.failed, total: v.total }))
          )
        }
      } catch {
        setError('加载失败')
      } finally {
        setLoading(false)
      }
    }
    fetch()
  }, [])

  const getStatValue = (key: keyof DashboardStats) => {
    if (!stats) return '-'
    const val = stats[key]
    if (key === 'success_rate') return val.toFixed(1)
    return String(val)
  }

  const columnHelper = createColumnHelper<Build>()
  const columns = [
    columnHelper.accessor('project_id', { header: '项目' }),
    columnHelper.accessor('environment_id', { header: '环境' }),
    columnHelper.accessor('status', {
      header: '状态',
      cell: ({ getValue }) => {
        const s = String(getValue())
        const info = BUILD_STATUSES[s as keyof typeof BUILD_STATUSES] ?? { label: s, color: 'bg-gray-500' }
        return <Badge className={cn('text-xs', info.color, 'text-white')}>{info.label}</Badge>
      },
    }),
    columnHelper.accessor('trigger_type', { header: '触发' }),
    columnHelper.accessor('created_at', {
      header: '时间',
      cell: ({ getValue }) => new Date(String(getValue())).toLocaleString('zh-CN'),
    }),
    columnHelper.display({
      id: 'actions',
      header: '',
      cell: ({ row }) => (
        <Link to={`/builds/${row.original.id}`}>
          <Button variant="ghost" size="icon-sm">
            <ExternalLink className="size-4" />
          </Button>
        </Link>
      ),
    }),
  ]

  const table = useReactTable({
    data: recentBuilds,
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-2 border-zinc-600 border-t-zinc-300" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded-lg border border-red-500/30 bg-red-500/10 p-4 text-red-400">{error}</div>
    )
  }

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-zinc-900 dark:text-white">仪表盘</h1>
        <p className="mt-1 text-sm text-zinc-500">构建流水线概览</p>
      </div>

      {/* Stats */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {STAT_CARDS.map(({ key, label, icon: Icon, color, bgColor, suffix = '' }) => (
          <Card key={key} className="border-zinc-200 dark:border-zinc-800">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-zinc-500">{label}</CardTitle>
              <div className={cn('rounded-lg p-2', bgColor)}>
                <Icon className={cn('size-5', color.includes('blue') && 'text-blue-500', color.includes('amber') && 'text-amber-500', color.includes('emerald') && 'text-emerald-500', color.includes('violet') && 'text-violet-500')} />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {getStatValue(key)}
                {suffix}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Active builds */}
        <Card className="lg:col-span-1 border-zinc-200 dark:border-zinc-800">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Activity className="size-5" />
              运行中的构建
            </CardTitle>
            <CardDescription>当前正在执行的构建任务</CardDescription>
          </CardHeader>
          <CardContent>
            {activeBuilds.length === 0 ? (
              <p className="text-sm text-zinc-500">暂无运行中的构建</p>
            ) : (
              <ul className="space-y-2">
                {activeBuilds.map((b) => {
                  const info = BUILD_STATUSES[b.status as keyof typeof BUILD_STATUSES] ?? { label: b.status, color: 'bg-gray-500' }
                  return (
                    <li key={b.id} className="flex items-center justify-between rounded-lg border border-zinc-200 dark:border-zinc-800 p-3">
                      <div>
                        <p className="font-medium">#{b.build_number}</p>
                        <p className="text-xs text-zinc-500">项目 {b.project_id}</p>
                      </div>
                      <Badge className={cn(info.color, 'text-white')}>{info.label}</Badge>
                      <Link to={`/builds/${b.id}`}>
                        <Button variant="ghost" size="icon-sm">
                          <ChevronRight className="size-4" />
                        </Button>
                      </Link>
                    </li>
                  )
                })}
              </ul>
            )}
          </CardContent>
        </Card>

        {/* Chart */}
        <Card className="lg:col-span-2 border-zinc-200 dark:border-zinc-800">
          <CardHeader>
            <CardTitle>构建趋势</CardTitle>
            <CardDescription>近 7 天构建统计</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-[220px]">
              {trend.length === 0 ? (
                <div className="flex h-full items-center justify-center text-zinc-500 text-sm">暂无数据</div>
              ) : (
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={trend}>
                    <CartesianGrid strokeDasharray="3 3" className="stroke-zinc-700" />
                    <XAxis dataKey="date" className="text-xs" />
                    <YAxis className="text-xs" />
                    <Tooltip
                      contentStyle={{ backgroundColor: 'rgb(24 24 27)', border: '1px solid rgb(63 63 70)' }}
                      labelStyle={{ color: 'rgb(161 161 170)' }}
                    />
                    <Area type="monotone" dataKey="success" stackId="1" stroke="#10b981" fill="#10b981" fillOpacity={0.6} name="成功" />
                    <Area type="monotone" dataKey="failed" stackId="1" stroke="#ef4444" fill="#ef4444" fillOpacity={0.6} name="失败" />
                  </AreaChart>
                </ResponsiveContainer>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      <Card className="border-zinc-200 dark:border-zinc-800">
        <CardHeader>
          <CardTitle>项目分组概览</CardTitle>
          <CardDescription>按分组查看项目与环境规模</CardDescription>
        </CardHeader>
        <CardContent>
          {!stats?.group_summary?.length ? (
            <div className="text-sm text-zinc-500">暂无分组数据</div>
          ) : (
            <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
              {stats.group_summary.map((group) => (
                <div key={group.group_name} className="rounded-xl border border-zinc-200 p-4 dark:border-zinc-800">
                  <p className="font-medium">{group.group_name}</p>
                  <div className="mt-3 flex items-center gap-3 text-sm text-zinc-500">
                    <span>{group.project_count} 个项目</span>
                    <span>{group.environment_count} 个环境</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Recent builds table */}
      <Card className="border-zinc-200 dark:border-zinc-800">
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle>最近构建</CardTitle>
            <CardDescription>最近的构建记录</CardDescription>
          </div>
          <Link to="/projects">
            <Button variant="outline" size="sm">查看全部</Button>
          </Link>
        </CardHeader>
        <CardContent>
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
              {table.getRowModel().rows.map((row) => (
                <TableRow key={row.id}>
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>
                  ))}
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
