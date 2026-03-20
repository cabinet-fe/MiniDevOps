import { useState, useEffect, useCallback } from 'react'
import { Plus, Pencil, Trash2, Network, Loader2 } from 'lucide-react'
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  createColumnHelper,
} from '@tanstack/react-table'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { api } from '@/lib/api'
import type { PaginatedData } from '@/lib/api'
import { useAuth } from '@/hooks/use-auth'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'
import { ServerFormDialog } from '@/pages/servers/form'

interface Server {
  id: number
  name: string
  host: string
  port: number
  username: string
  status: string
  description: string
  tags: string
}

export function ServerListPage() {
  const { isOps } = useAuth()
  const [servers, setServers] = useState<Server[]>([])
  const [loading, setLoading] = useState(true)
  const [tagFilter, setTagFilter] = useState('')
  const [allTags, setAllTags] = useState<string[]>([])
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [formDialogOpen, setFormDialogOpen] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [testingId, setTestingId] = useState<number | null>(null)

  const fetchServers = useCallback(async () => {
    try {
      const res = await api.get<PaginatedData<Server>>(
        `/servers?page=1&page_size=100${tagFilter ? `&tag=${tagFilter}` : ''}`
      )
      if (res.code === 0 && res.data) {
        const data = res.data as PaginatedData<Server>
        setServers(data.items || [])
        const tags = new Set<string>()
        for (const s of data.items || []) {
          (s.tags || '').split(',').forEach((t) => t.trim() && tags.add(t.trim()))
        }
        setAllTags(Array.from(tags))
      }
    } catch {
      toast.error('加载失败')
    } finally {
      setLoading(false)
    }
  }, [tagFilter])

  useEffect(() => {
    fetchServers()
  }, [fetchServers])

  const openCreate = () => {
    setEditingId(null)
    setFormDialogOpen(true)
  }

  const openEdit = (id: number) => {
    setEditingId(id)
    setFormDialogOpen(true)
  }

  const testConnection = async (serverId: number) => {
    setTestingId(serverId)
    try {
      const res = await api.post<{ message: string }>(`/servers/${serverId}/test`)
      if (res.code === 0) {
        toast.success(res.data?.message || '连接成功')
        fetchServers()
      } else {
        toast.error(res.message || '连接失败')
        fetchServers()
      }
    } catch {
      toast.error('连接失败')
      fetchServers()
    } finally {
      setTestingId(null)
    }
  }

  const handleDelete = async () => {
    if (!deleteId) return
    setDeleting(true)
    try {
      const res = await api.delete(`/servers/${deleteId}`)
      if (res.code === 0) {
        toast.success('已删除')
        setDeleteId(null)
        fetchServers()
      } else {
        toast.error(res.message || '删除失败')
      }
    } catch {
      toast.error('删除失败')
    } finally {
      setDeleting(false)
    }
  }

  const columnHelper = createColumnHelper<Server>()
  const columns = [
    columnHelper.accessor('name', { header: '名称' }),
    columnHelper.accessor('host', { header: '主机' }),
    columnHelper.accessor('port', { header: '端口' }),
    columnHelper.accessor('status', {
      header: '状态',
      cell: ({ getValue }) => {
        const s = String(getValue())
        const config: Record<string, { label: string; class: string }> = {
          online: { label: '在线', class: 'bg-emerald-500/15 text-emerald-600 border-emerald-500/25' },
          offline: { label: '离线', class: 'bg-red-500/15 text-red-600 border-red-500/25' },
        }
        const c = config[s] ?? { label: '未检测', class: 'bg-muted/50 text-muted-foreground border-border/50' }
        return (
          <Badge variant="outline" className={cn('font-medium', c.class)}>
            {c.label}
          </Badge>
        )
      },
    }),
    columnHelper.accessor('tags', {
      header: '标签',
      cell: ({ getValue }) => {
        const tags = (String(getValue()) || '').split(',').map((t) => t.trim()).filter(Boolean)
        return (
          <div className="flex flex-wrap gap-1">
            {tags.map((t) => (
              <Badge key={t} variant="outline" className="text-xs">
                {t}
              </Badge>
            ))}
          </div>
        )
      },
    }),
    ...(isOps()
      ? [
          columnHelper.display({
            id: 'actions',
            header: '操作',
            cell: ({ row }) => {
              const isTesting = testingId === row.original.id
              return (
              <div className="flex gap-1">
                <Button
                  variant="ghost"
                  size="icon-sm"
                  onClick={() => testConnection(row.original.id)}
                  disabled={isTesting}
                  title="测试连接"
                >
                  {isTesting ? <Loader2 className="size-4 animate-spin" /> : <Network className="size-4" />}
                </Button>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  onClick={() => openEdit(row.original.id)}
                >
                  <Pencil className="size-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  onClick={() => setDeleteId(row.original.id)}
                >
                  <Trash2 className="size-4" />
                </Button>
              </div>
              )
            },
          }),
        ]
      : []),
  ]

  const table = useReactTable({
    data: servers,
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="border-muted size-8 animate-spin rounded-full border-2 border-t-foreground" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground">服务器</h1>
          <p className="mt-1 text-sm text-muted-foreground">管理部署目标服务器</p>
        </div>
        {isOps() && (
          <Button className="gap-2" onClick={openCreate}>
            <Plus className="size-4" />
            新建服务器
          </Button>
        )}
      </div>

      {allTags.length > 0 && (
        <div className="flex flex-wrap gap-2">
          <Button
            variant={!tagFilter ? 'secondary' : 'outline'}
            size="sm"
            onClick={() => setTagFilter('')}
          >
            全部
          </Button>
          {allTags.map((t) => (
            <Button
              key={t}
              variant={tagFilter === t ? 'secondary' : 'outline'}
              size="sm"
              onClick={() => setTagFilter(tagFilter === t ? '' : t)}
            >
              {t}
            </Button>
          ))}
        </div>
      )}

      <Card className="border-border">
        <CardHeader>
          <CardTitle>服务器列表</CardTitle>
          <CardDescription>共 {servers.length} 台服务器</CardDescription>
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
              {table.getRowModel().rows.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center text-muted-foreground py-8">
                    暂无服务器
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
        </CardContent>
      </Card>

      <Dialog open={!!deleteId} onOpenChange={(o) => !o && setDeleteId(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认删除</DialogTitle>
            <DialogDescription>确定要删除此服务器吗？此操作不可撤销。</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteId(null)}>取消</Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleting}>
              {deleting ? '删除中...' : '删除'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ServerFormDialog
        open={formDialogOpen}
        onOpenChange={setFormDialogOpen}
        editId={editingId}
        onSuccess={() => fetchServers()}
      />
    </div>
  )
}
