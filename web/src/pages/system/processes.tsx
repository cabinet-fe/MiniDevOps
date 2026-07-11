import { useEffect, useState } from 'react'
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  createColumnHelper,
  type ColumnDef,
} from '@tanstack/react-table'
import {
  Activity,
  ArrowDown,
  ArrowUp,
  ArrowUpDown,
  ChevronLeft,
  ChevronRight,
  Filter,
  Search,
  Trash2,
} from 'lucide-react'
import { toast } from 'sonner'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
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
import { cn } from '@/lib/utils'
import { api } from '@/lib/api'
import type { PaginatedData } from '@/lib/api'
import { useAuthStore } from '@/stores/auth-store'

interface ProcessInfo {
  pid: number
  name: string
  memory_bytes: number
  cpu_percent: number
  ports: number[]
  username: string
  num_threads: number
  cmdline?: string
  status?: string
  create_time?: number
}

type ProcessSort = 'cpu' | 'memory' | 'name'
type SortOrder = 'desc' | 'asc'

function formatBytes(bytes: number) {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let value = bytes
  let unitIndex = 0
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex += 1
  }
  const digits = value >= 100 || unitIndex === 0 ? 0 : value >= 10 ? 1 : 2
  return `${value.toFixed(digits)} ${units[unitIndex]}`
}

function formatPorts(ports: number[] | undefined) {
  if (!ports || ports.length === 0) return '-'
  return ports.join(', ')
}

function formatCreateTime(ms: number | undefined) {
  if (!ms) return '-'
  return new Date(ms).toLocaleString('zh-CN')
}

function SortHeader({
  label,
  field,
  activeField,
  order,
  onToggle,
}: {
  label: string
  field: ProcessSort
  activeField: ProcessSort | null
  order: SortOrder | null
  onToggle: (field: ProcessSort) => void
}) {
  const active = activeField === field
  const Icon = !active ? ArrowUpDown : order === 'asc' ? ArrowUp : ArrowDown
  const title = !active
    ? `按${label}降序`
    : order === 'desc'
      ? `按${label}升序`
      : `取消${label}排序`

  return (
    <div className="flex items-center gap-1">
      <span>{label}</span>
      <button
        type="button"
        className={cn(
          'inline-flex size-6 items-center justify-center rounded-md transition-colors',
          active
            ? 'text-foreground hover:bg-accent'
            : 'text-muted-foreground/60 hover:bg-accent hover:text-muted-foreground',
        )}
        onClick={() => onToggle(field)}
        title={title}
        aria-label={title}
      >
        <Icon className="size-3.5" />
      </button>
    </div>
  )
}

export function SystemProcessesPage() {
  const user = useAuthStore((s) => s.user)
  const isAdmin = user?.role === 'admin'

  const [processes, setProcesses] = useState<ProcessInfo[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [loading, setLoading] = useState(false)

  const [q, setQ] = useState('')
  const [pid, setPid] = useState('')
  const [port, setPort] = useState('')
  const [sort, setSort] = useState<ProcessSort | null>(null)
  const [order, setOrder] = useState<SortOrder | null>(null)

  const [killTarget, setKillTarget] = useState<ProcessInfo | null>(null)
  const [killing, setKilling] = useState(false)

  const pageSize = 20

  const fetchProcesses = async (
    nextPage = page,
    nextSort: ProcessSort | null = sort,
    nextOrder: SortOrder | null = order,
  ) => {
    setLoading(true)
    try {
      const params = new URLSearchParams()
      params.set('page', String(nextPage))
      params.set('page_size', String(pageSize))
      if (nextSort && nextOrder) {
        params.set('sort', nextSort)
        params.set('order', nextOrder)
      }
      if (q.trim()) params.set('q', q.trim())
      if (pid.trim()) params.set('pid', pid.trim())
      if (port.trim()) params.set('port', port.trim())

      const res = await api.get<PaginatedData<ProcessInfo>>(`/system/processes?${params}`)
      if (res.code === 0 && res.data) {
        const data = res.data as PaginatedData<ProcessInfo>
        setProcesses(data.items || [])
        setTotal(data.total ?? 0)
        setPage(data.page ?? nextPage)
        setTotalPages(data.total_pages || Math.ceil((data.total ?? 0) / pageSize) || 1)
      } else {
        toast.error(res.message || '查询失败')
      }
    } catch {
      toast.error('查询失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void fetchProcesses(1)
    // 进入页面默认查询一次
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const handleQuery = () => {
    setPage(1)
    void fetchProcesses(1)
  }

  const toggleSort = (field: ProcessSort) => {
    let nextSort: ProcessSort | null
    let nextOrder: SortOrder | null
    if (sort !== field) {
      nextSort = field
      nextOrder = 'desc'
    } else if (order === 'desc') {
      nextSort = field
      nextOrder = 'asc'
    } else {
      nextSort = null
      nextOrder = null
    }
    setSort(nextSort)
    setOrder(nextOrder)
    setPage(1)
    void fetchProcesses(1, nextSort, nextOrder)
  }

  const goToPage = (next: number) => {
    setPage(next)
    void fetchProcesses(next)
  }

  const handleKill = async () => {
    if (!killTarget) return
    setKilling(true)
    try {
      const res = await api.delete(`/system/processes/${killTarget.pid}`)
      if (res.code === 0) {
        toast.success(`已终止进程 ${killTarget.name || killTarget.pid}`)
        setKillTarget(null)
        fetchProcesses(page)
      } else {
        toast.error(res.message || '终止失败')
      }
    } catch {
      toast.error('终止失败')
    } finally {
      setKilling(false)
    }
  }

  const columnHelper = createColumnHelper<ProcessInfo>()
  const columns: ColumnDef<ProcessInfo, any>[] = [
    columnHelper.accessor('name', {
      header: () => (
        <SortHeader
          label="名称"
          field="name"
          activeField={sort}
          order={order}
          onToggle={toggleSort}
        />
      ),
      cell: ({ getValue }) => (
        <span className="max-w-[140px] truncate block font-medium">{String(getValue() || '-')}</span>
      ),
    }),
    columnHelper.accessor('pid', {
      header: 'PID',
      cell: ({ getValue }) => <span className="font-mono text-xs">{getValue()}</span>,
    }),
    columnHelper.accessor('memory_bytes', {
      header: () => (
        <SortHeader
          label="内存"
          field="memory"
          activeField={sort}
          order={order}
          onToggle={toggleSort}
        />
      ),
      cell: ({ getValue }) => (
        <span className="font-mono text-xs text-muted-foreground">{formatBytes(Number(getValue() || 0))}</span>
      ),
    }),
    columnHelper.accessor('cpu_percent', {
      header: () => (
        <SortHeader
          label="CPU%"
          field="cpu"
          activeField={sort}
          order={order}
          onToggle={toggleSort}
        />
      ),
      cell: ({ getValue }) => (
        <span className="font-mono text-xs text-muted-foreground">
          {Number(getValue() ?? 0).toFixed(1)}
        </span>
      ),
    }),
    columnHelper.accessor('ports', {
      header: '端口',
      cell: ({ getValue }) => (
        <span className="max-w-[100px] truncate block font-mono text-xs text-muted-foreground">
          {formatPorts(getValue() as number[] | undefined)}
        </span>
      ),
    }),
    columnHelper.accessor('username', {
      header: '用户',
      cell: ({ getValue }) => (
        <span className="text-sm text-muted-foreground">{String(getValue() || '-')}</span>
      ),
    }),
    columnHelper.accessor('num_threads', {
      header: '线程',
      cell: ({ getValue }) => (
        <span className="font-mono text-xs text-muted-foreground">{getValue() ?? '-'}</span>
      ),
    }),
    columnHelper.accessor('status', {
      header: '状态',
      cell: ({ getValue }) => {
        const s = String(getValue() || '')
        return s ? (
          <Badge variant="outline" className="font-mono text-[11px]">
            {s}
          </Badge>
        ) : (
          <span className="text-muted-foreground">-</span>
        )
      },
    }),
    columnHelper.accessor('create_time', {
      header: '启动时间',
      cell: ({ getValue }) => (
        <span className="text-xs text-muted-foreground whitespace-nowrap">
          {formatCreateTime(getValue() as number | undefined)}
        </span>
      ),
    }),
    columnHelper.accessor('cmdline', {
      header: '命令行',
      cell: ({ getValue }) => {
        const cmd = String(getValue() || '')
        return cmd ? (
          <span className="max-w-[220px] truncate block font-mono text-xs text-muted-foreground" title={cmd}>
            {cmd}
          </span>
        ) : (
          <span className="text-muted-foreground">-</span>
        )
      },
    }),
  ]

  if (isAdmin) {
    columns.push(
      columnHelper.display({
        id: 'actions',
        header: '操作',
        cell: ({ row }) => (
          <Button
            variant="ghost"
            size="icon-sm"
            className="text-red-500 hover:text-red-600"
            onClick={() => setKillTarget(row.original)}
            title="终止进程"
          >
            <Trash2 className="size-4" />
          </Button>
        ),
      }),
    )
  }

  const table = useReactTable({
    data: processes,
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-foreground">系统管理</h1>
        <p className="mt-1 text-sm text-muted-foreground">查询主机进程，管理员可终止进程</p>
      </div>

      <Card className="border-border">
        <CardHeader className="pb-4">
          <div className="flex items-center gap-2">
            <Filter className="size-4 text-muted-foreground" />
            <CardTitle className="text-base">筛选条件</CardTitle>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap items-end gap-4">
            <div className="space-y-1.5">
              <Label className="text-xs font-medium text-muted-foreground">关键字</Label>
              <Input
                className="w-[200px]"
                placeholder="进程名 / 命令行"
                value={q}
                onChange={(e) => setQ(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleQuery()}
              />
            </div>
            <div className="space-y-1.5">
              <Label className="text-xs font-medium text-muted-foreground">PID</Label>
              <Input
                className="w-[120px]"
                placeholder="进程 ID"
                value={pid}
                onChange={(e) => setPid(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleQuery()}
              />
            </div>
            <div className="space-y-1.5">
              <Label className="text-xs font-medium text-muted-foreground">端口</Label>
              <Input
                className="w-[120px]"
                placeholder="监听端口"
                value={port}
                onChange={(e) => setPort(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleQuery()}
              />
            </div>
            <Button onClick={handleQuery} disabled={loading} className="gap-2">
              <Search className="size-4" />
              {loading ? '查询中...' : '查询'}
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card className="border-border">
        <CardHeader className="pb-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Activity className="size-4 text-muted-foreground" />
              <CardTitle className="text-base">进程列表</CardTitle>
            </div>
            <span className="text-sm text-muted-foreground">共 {total} 条</span>
          </div>
        </CardHeader>
        <CardContent>
          {loading && processes.length === 0 ? (
            <div className="flex h-48 items-center justify-center">
              <div className="border-muted size-8 animate-spin rounded-full border-2 border-t-foreground" />
            </div>
          ) : processes.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
              <Activity className="mb-3 size-10 opacity-40" />
              <p className="text-sm">未找到匹配进程</p>
            </div>
          ) : (
            <>
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    {table.getHeaderGroups().map((hg) => (
                      <TableRow key={hg.id}>
                        {hg.headers.map((h) => (
                          <TableHead key={h.id}>
                            {flexRender(h.column.columnDef.header, h.getContext())}
                          </TableHead>
                        ))}
                      </TableRow>
                    ))}
                  </TableHeader>
                  <TableBody>
                    {table.getRowModel().rows.map((row) => (
                      <TableRow key={row.id}>
                        {row.getVisibleCells().map((cell) => (
                          <TableCell key={cell.id}>
                            {flexRender(cell.column.columnDef.cell, cell.getContext())}
                          </TableCell>
                        ))}
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>

              <div className="mt-4 flex items-center justify-between border-t border-border pt-4">
                <p className="text-sm text-muted-foreground">
                  第 {page} / {totalPages} 页，共 {total} 条
                </p>
                <div className="flex gap-1">
                  <Button
                    variant="outline"
                    size="icon"
                    className="size-8"
                    onClick={() => goToPage(Math.max(1, page - 1))}
                    disabled={page <= 1 || loading}
                  >
                    <ChevronLeft className="size-4" />
                  </Button>
                  <Button
                    variant="outline"
                    size="icon"
                    className="size-8"
                    onClick={() => goToPage(Math.min(totalPages, page + 1))}
                    disabled={page >= totalPages || loading}
                  >
                    <ChevronRight className="size-4" />
                  </Button>
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      <Dialog open={!!killTarget} onOpenChange={(open) => !open && !killing && setKillTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认终止进程</DialogTitle>
            <DialogDescription>
              确定要终止进程{' '}
              <span className="font-mono text-foreground">
                {killTarget?.name || '-'} (PID {killTarget?.pid})
              </span>{' '}
              吗？此操作不可撤销。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setKillTarget(null)} disabled={killing}>
              取消
            </Button>
            <Button variant="destructive" onClick={handleKill} disabled={killing}>
              {killing ? '终止中...' : '终止'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
