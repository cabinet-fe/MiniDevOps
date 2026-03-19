import { useState, useEffect, useCallback } from 'react'
import { Link } from 'react-router'
import { ExternalLink, RefreshCw } from 'lucide-react'
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  createColumnHelper,
} from '@tanstack/react-table'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
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
import { cn } from '@/lib/utils'

interface BuildItem {
  id: number
  build_number: number
  status: string
  trigger_type: string
  commit_hash: string
  commit_message: string
  duration_ms: number
  created_at: string
  project_id: number
  project_name: string
  environment_name: string
}

const columnHelper = createColumnHelper<BuildItem>()

const columns = [
  columnHelper.accessor('build_number', {
    header: '#',
    cell: ({ getValue }) => `#${getValue()}`,
  }),
  columnHelper.accessor('project_name', {
    header: '项目',
    cell: ({ row }) => (
      <Link
        to={`/projects/${row.original.project_id}`}
        className="text-blue-500 hover:underline"
      >
        {row.original.project_name}
      </Link>
    ),
  }),
  columnHelper.accessor('environment_name', { header: '环境' }),
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
    cell: ({ getValue }) => (
      <span className="max-w-[200px] truncate block">{getValue() || '-'}</span>
    ),
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
    cell: ({ row }) => (
      <Link to={`/builds/${row.original.id}`}>
        <Button variant="ghost" size="icon-sm">
          <ExternalLink className="size-4" />
        </Button>
      </Link>
    ),
  }),
]

export function BuildListPage() {
  const [builds, setBuilds] = useState<BuildItem[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)

  const fetchBuilds = useCallback(async () => {
    setLoading(true)
    try {
      const res = await api.get<PaginatedData<BuildItem>>(`/builds?page=${page}&page_size=20`)
      if (res.code === 0 && res.data) {
        const data = res.data as PaginatedData<BuildItem>
        setBuilds(data.items || [])
        setTotalPages(data.total_pages || 1)
      }
    } finally {
      setLoading(false)
    }
  }, [page])

  useEffect(() => {
    fetchBuilds()
  }, [fetchBuilds])

  const table = useReactTable({
    data: builds,
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">构建记录</h1>
          <p className="mt-1 text-sm text-zinc-500">查看所有项目的构建历史</p>
        </div>
        <Button variant="outline" onClick={fetchBuilds} disabled={loading}>
          <RefreshCw className={cn('size-4', loading && 'animate-spin')} />
          刷新
        </Button>
      </div>

      <Card className="border-zinc-200 dark:border-zinc-800">
        <CardHeader>
          <CardTitle>全部构建</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex h-32 items-center justify-center">
              <div className="size-8 animate-spin rounded-full border-2 border-zinc-600 border-t-zinc-300" />
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
                  {table.getRowModel().rows.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={columns.length} className="text-center text-zinc-500 py-8">
                        暂无构建记录
                      </TableCell>
                    </TableRow>
                  ) : (
                    table.getRowModel().rows.map((row) => (
                      <TableRow key={row.id}>
                        {row.getVisibleCells().map((cell) => (
                          <TableCell key={cell.id}>
                            {flexRender(cell.column.columnDef.cell, cell.getContext())}
                          </TableCell>
                        ))}
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>

              {totalPages > 1 && (
                <div className="mt-4 flex items-center justify-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={page <= 1}
                    onClick={() => setPage((p) => p - 1)}
                  >
                    上一页
                  </Button>
                  <span className="text-sm text-zinc-500">
                    {page} / {totalPages}
                  </span>
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={page >= totalPages}
                    onClick={() => setPage((p) => p + 1)}
                  >
                    下一页
                  </Button>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
