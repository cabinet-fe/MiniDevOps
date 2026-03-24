import { useState, useEffect } from 'react'
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  createColumnHelper,
  type ColumnDef,
} from '@tanstack/react-table'
import { format, isValid, parse } from 'date-fns'
import { zhCN } from 'date-fns/locale/zh-CN'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Calendar } from '@/components/ui/calendar'
import { Badge } from '@/components/ui/badge'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { api } from '@/lib/api'
import type { PaginatedData } from '@/lib/api'
import { toast } from 'sonner'
import {
  Plus,
  Pencil,
  Trash2,
  LogIn,
  Activity,
  FlaskConical,
  ShieldCheck,
  ChevronLeft,
  ChevronRight,
  FileText,
  Filter,
  CalendarIcon,
} from 'lucide-react'
import { cn } from '@/lib/utils'

interface AuditLog {
  id: number
  user_id: number
  action: string
  resource_type: string
  resource_id: number
  details: string
  ip_address: string
  created_at: string
}

const ACTION_CONFIG: Record<string, { label: string; color: string; icon: typeof Plus }> = {
  create: { label: '创建', color: 'bg-emerald-500/15 text-emerald-600 border-emerald-500/20', icon: Plus },
  update: { label: '更新', color: 'bg-blue-500/15 text-blue-600 border-blue-500/20', icon: Pencil },
  delete: { label: '删除', color: 'bg-red-500/15 text-red-600 border-red-500/20', icon: Trash2 },
  login: { label: '登录', color: 'bg-violet-500/15 text-violet-600 border-violet-500/20', icon: LogIn },
  test: { label: '测试', color: 'bg-amber-500/15 text-amber-600 border-amber-500/20', icon: FlaskConical },
  auth: { label: '认证', color: 'bg-violet-500/15 text-violet-600 border-violet-500/20', icon: ShieldCheck },
}

const RESOURCE_LABELS: Record<string, string> = {
  project: '项目',
  server: '服务器',
  build: '构建',
  environment: '环境',
  user: '用户',
  settings: '设置',
  system: '系统',
  auth: '认证',
}

function AuditDateField({
  label,
  value,
  onChange,
}: {
  label: string
  value: string
  onChange: (next: string) => void
}) {
  const parsed = value ? parse(value, 'yyyy-MM-dd', new Date()) : undefined
  const selected = parsed && isValid(parsed) ? parsed : undefined

  return (
    <div className="space-y-1.5">
      <label className="text-xs font-medium text-muted-foreground">{label}</label>
      <Popover>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            className={cn(
              'w-[160px] justify-start text-left font-normal',
              !selected && 'text-muted-foreground'
            )}
          >
            <CalendarIcon className="mr-2 size-4 shrink-0 opacity-60" />
            {selected ? format(selected, 'yyyy-MM-dd') : '选择日期'}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0" align="start">
          <Calendar
            mode="single"
            selected={selected}
            onSelect={(d) => onChange(d ? format(d, 'yyyy-MM-dd') : '')}
            locale={zhCN}
            captionLayout="dropdown"
          />
        </PopoverContent>
      </Popover>
    </div>
  )
}

function ActionBadge({ action }: { action: string }) {
  const config = ACTION_CONFIG[action] ?? {
    label: action,
    color: 'bg-muted/50 text-muted-foreground border-border/50',
    icon: Activity,
  }
  const Icon = config.icon
  return (
    <Badge variant="outline" className={cn('gap-1 font-medium', config.color)}>
      <Icon className="size-3" />
      {config.label}
    </Badge>
  )
}

export function AuditLogsPage() {
  const [logs, setLogs] = useState<AuditLog[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const pageSize = 20

  const [actionFilter, setActionFilter] = useState('')
  const [resourceFilter, setResourceFilter] = useState('')
  const [fromDate, setFromDate] = useState('')
  const [toDate, setToDate] = useState('')

  useEffect(() => {
    fetchLogs()
  }, [page, actionFilter, resourceFilter, fromDate, toDate])

  const fetchLogs = async () => {
    setLoading(true)
    try {
      const params = new URLSearchParams()
      params.set('page', String(page))
      params.set('page_size', String(pageSize))
      if (actionFilter) params.set('action', actionFilter)
      if (resourceFilter) params.set('resource_type', resourceFilter)
      if (fromDate) params.set('from', fromDate)
      if (toDate) params.set('to', toDate)

      const res = await api.get<PaginatedData<AuditLog>>(`/system/audit-logs?${params}`)
      if (res.code === 0 && res.data) {
        const data = res.data as PaginatedData<AuditLog>
        setLogs(data.items || [])
        setTotal(data.total ?? 0)
      }
    } catch {
      toast.error('加载失败')
    } finally {
      setLoading(false)
    }
  }

  const columnHelper = createColumnHelper<AuditLog>()
  const columns: ColumnDef<AuditLog, any>[] = [
    columnHelper.accessor('action', {
      header: '操作',
      cell: ({ getValue }) => <ActionBadge action={String(getValue())} />,
    }),
    columnHelper.accessor('resource_type', {
      header: '资源类型',
      cell: ({ getValue }) => {
        const rt = String(getValue())
        return rt ? (
          <span className="text-sm">{RESOURCE_LABELS[rt] ?? rt}</span>
        ) : (
          <span className="text-muted-foreground">-</span>
        )
      },
    }),
    columnHelper.accessor('resource_id', {
      header: '资源 ID',
      cell: ({ getValue }) => {
        const id = Number(getValue())
        return id > 0 ? (
          <span className="font-mono text-xs text-muted-foreground">#{id}</span>
        ) : (
          <span className="text-muted-foreground">-</span>
        )
      },
    }),
    columnHelper.accessor('details', {
      header: '详情',
      cell: ({ getValue }) => {
        const d = String(getValue() || '')
        return d ? (
          <span className="max-w-[260px] truncate block font-mono text-xs text-muted-foreground">{d}</span>
        ) : (
          <span className="text-muted-foreground">-</span>
        )
      },
    }),
    columnHelper.accessor('ip_address', {
      header: 'IP 地址',
      cell: ({ getValue }) => (
        <span className="font-mono text-xs">{getValue() || '-'}</span>
      ),
    }),
    columnHelper.accessor('created_at', {
      header: '时间',
      cell: ({ getValue }) => (
        <span className="text-sm text-muted-foreground">
          {new Date(String(getValue())).toLocaleString('zh-CN')}
        </span>
      ),
    }),
  ]

  const table = useReactTable({
    data: logs,
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  const totalPages = Math.ceil(total / pageSize) || 1

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-foreground">审计日志</h1>
        <p className="mt-1 text-sm text-muted-foreground">追踪系统中所有状态变更操作的记录</p>
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
              <label className="text-xs font-medium text-muted-foreground">操作类型</label>
              <Select value={actionFilter || 'all'} onValueChange={(v) => { setActionFilter(v === 'all' ? '' : v); setPage(1) }}>
                <SelectTrigger className="w-[140px]">
                  <SelectValue placeholder="全部" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">全部</SelectItem>
                  <SelectItem value="create">创建</SelectItem>
                  <SelectItem value="update">更新</SelectItem>
                  <SelectItem value="delete">删除</SelectItem>
                  <SelectItem value="login">登录</SelectItem>
                  <SelectItem value="test">测试</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-muted-foreground">资源类型</label>
              <Select value={resourceFilter || 'all'} onValueChange={(v) => { setResourceFilter(v === 'all' ? '' : v); setPage(1) }}>
                <SelectTrigger className="w-[140px]">
                  <SelectValue placeholder="全部" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">全部</SelectItem>
                  <SelectItem value="project">项目</SelectItem>
                  <SelectItem value="server">服务器</SelectItem>
                  <SelectItem value="build">构建</SelectItem>
                  <SelectItem value="environment">环境</SelectItem>
                  <SelectItem value="user">用户</SelectItem>
                  <SelectItem value="system">系统</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <AuditDateField
              label="开始日期"
              value={fromDate}
              onChange={(v) => { setFromDate(v); setPage(1) }}
            />
            <AuditDateField
              label="结束日期"
              value={toDate}
              onChange={(v) => { setToDate(v); setPage(1) }}
            />
            {(actionFilter || resourceFilter || fromDate || toDate) && (
              <Button
                variant="ghost"
                size="sm"
                className="text-muted-foreground"
                onClick={() => {
                  setActionFilter('')
                  setResourceFilter('')
                  setFromDate('')
                  setToDate('')
                  setPage(1)
                }}
              >
                清除筛选
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      <Card className="border-border">
        <CardHeader className="pb-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <FileText className="size-4 text-muted-foreground" />
              <CardTitle className="text-base">操作记录</CardTitle>
            </div>
            <span className="text-sm text-muted-foreground">共 {total} 条</span>
          </div>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex h-48 items-center justify-center">
              <div className="border-muted size-8 animate-spin rounded-full border-2 border-t-foreground" />
            </div>
          ) : logs.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
              <FileText className="mb-3 size-10 opacity-40" />
              <p className="text-sm">暂无审计记录</p>
            </div>
          ) : (
            <>
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

              <div className="mt-4 flex items-center justify-between border-t border-border pt-4">
                <p className="text-sm text-muted-foreground">
                  第 {page} / {totalPages} 页，共 {total} 条
                </p>
                <div className="flex gap-1">
                  <Button
                    variant="outline"
                    size="icon"
                    className="size-8"
                    onClick={() => setPage((p) => Math.max(1, p - 1))}
                    disabled={page <= 1}
                  >
                    <ChevronLeft className="size-4" />
                  </Button>
                  <Button
                    variant="outline"
                    size="icon"
                    className="size-8"
                    onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                    disabled={page >= totalPages}
                  >
                    <ChevronRight className="size-4" />
                  </Button>
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
