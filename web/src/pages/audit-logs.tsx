import { useState, useEffect } from 'react'
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  createColumnHelper,
  type ColumnDef,
} from '@tanstack/react-table'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
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

export function AuditLogsPage() {
  const [logs, setLogs] = useState<AuditLog[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const pageSize = 20

  const [actionFilter, setActionFilter] = useState('')
  const [userFilter, setUserFilter] = useState('')
  const [fromDate, setFromDate] = useState('')
  const [toDate, setToDate] = useState('')

  useEffect(() => {
    fetchLogs()
  }, [page, actionFilter, userFilter, fromDate, toDate])

  const fetchLogs = async () => {
    setLoading(true)
    try {
      const params = new URLSearchParams()
      params.set('page', String(page))
      params.set('page_size', String(pageSize))
      if (actionFilter) params.set('action', actionFilter)
      if (userFilter) params.set('user_id', userFilter)
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
    columnHelper.accessor('action', { header: '操作' }),
    columnHelper.accessor('user_id', { header: '用户 ID' }),
    columnHelper.accessor('resource_type', { header: '资源类型', cell: ({ getValue }) => getValue() || '-' }),
    columnHelper.accessor('resource_id', { header: '资源 ID' }),
    columnHelper.accessor('ip_address', { header: 'IP' }),
    columnHelper.accessor('created_at', {
      header: '时间',
      cell: ({ getValue }) => new Date(String(getValue())).toLocaleString('zh-CN'),
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
        <h1 className="text-2xl font-bold tracking-tight">审计日志</h1>
        <p className="mt-1 text-sm text-zinc-500">系统操作记录</p>
      </div>

      <Card className="border-zinc-200 dark:border-zinc-800">
        <CardHeader>
          <CardTitle>筛选</CardTitle>
          <CardDescription>按条件筛选审计记录</CardDescription>
          <div className="mt-4 flex flex-wrap gap-4">
            <div>
              <label className="text-xs text-zinc-500">操作类型</label>
              <Select value={actionFilter} onValueChange={setActionFilter}>
                <SelectTrigger className="mt-1 w-[180px]">
                  <SelectValue placeholder="全部" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">全部</SelectItem>
                  <SelectItem value="create">创建</SelectItem>
                  <SelectItem value="update">更新</SelectItem>
                  <SelectItem value="delete">删除</SelectItem>
                  <SelectItem value="login">登录</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <label className="text-xs text-zinc-500">用户 ID</label>
              <Input
                value={userFilter}
                onChange={(e) => setUserFilter(e.target.value)}
                placeholder="用户 ID"
                className="mt-1 w-[120px]"
              />
            </div>
            <div>
              <label className="text-xs text-zinc-500">开始日期</label>
              <Input
                type="date"
                value={fromDate}
                onChange={(e) => setFromDate(e.target.value)}
                className="mt-1 w-[160px]"
              />
            </div>
            <div>
              <label className="text-xs text-zinc-500">结束日期</label>
              <Input
                type="date"
                value={toDate}
                onChange={(e) => setToDate(e.target.value)}
                className="mt-1 w-[160px]"
              />
            </div>
          </div>
        </CardHeader>
      </Card>

      <Card className="border-zinc-200 dark:border-zinc-800">
        <CardHeader>
          <CardTitle>日志列表</CardTitle>
          <CardDescription>共 {total} 条记录</CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex h-48 items-center justify-center">
              <div className="size-8 animate-spin rounded-full border-2 border-zinc-600 border-t-zinc-300" />
            </div>
          ) : (
            <>
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
                      <TableCell colSpan={6} className="text-center text-zinc-500 py-8">
                        暂无记录
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
              <div className="mt-4 flex items-center justify-between">
                <p className="text-sm text-zinc-500">
                  第 {page} / {totalPages} 页
                </p>
                <div className="flex gap-2">
                  <Button variant="outline" size="sm" onClick={() => setPage((p) => Math.max(1, p - 1))} disabled={page <= 1}>
                    上一页
                  </Button>
                  <Button variant="outline" size="sm" onClick={() => setPage((p) => Math.min(totalPages, p + 1))} disabled={page >= totalPages}>
                    下一页
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
