import { useState, useEffect, useCallback } from 'react'
import { Link } from 'react-router'
import { LayoutGrid, List, Plus, Search, FolderGit2, Pencil } from 'lucide-react'
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  createColumnHelper,
} from '@tanstack/react-table'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
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
import { ProjectFormDialog } from '@/pages/projects/form'

interface Project {
  id: number
  name: string
  description: string
  repo_url: string
  environments?: { id: number; name: string }[]
}

export function ProjectListPage() {
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [viewMode, setViewMode] = useState<'card' | 'table'>('card')
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)

  const fetchProjects = useCallback(async () => {
    try {
      const res = await api.get<PaginatedData<Project>>('/projects?page=1&page_size=100')
      if (res.code === 0 && res.data) {
        const data = res.data as PaginatedData<Project>
        setProjects(data.items || [])
      }
    } catch {
      setError('加载失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchProjects()
  }, [fetchProjects])

  const openCreate = () => {
    setEditingId(null)
    setDialogOpen(true)
  }

  const openEdit = (id: number) => {
    setEditingId(id)
    setDialogOpen(true)
  }

  const handleSuccess = () => {
    fetchProjects()
  }

  const filtered = projects.filter(
    (p) =>
      !search ||
      p.name.toLowerCase().includes(search.toLowerCase()) ||
      (p.description || '').toLowerCase().includes(search.toLowerCase())
  )

  const columnHelper = createColumnHelper<Project>()
  const columns = [
    columnHelper.accessor('name', { header: '项目名称' }),
    columnHelper.accessor('description', { header: '描述', cell: ({ getValue }) => getValue() || '-' }),
    columnHelper.accessor('repo_url', { header: '仓库地址' }),
    columnHelper.accessor((r) => r.environments?.length ?? 0, { header: '环境数', id: 'env_count' }),
    columnHelper.display({
      id: 'actions',
      header: '',
      cell: ({ row }) => (
        <div className="flex gap-1">
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={(e) => {
              e.preventDefault()
              e.stopPropagation()
              openEdit(row.original.id)
            }}
          >
            <Pencil className="size-4" />
          </Button>
          <Link to={`/projects/${row.original.id}`}>
            <Button variant="outline" size="sm">查看</Button>
          </Link>
        </div>
      ),
    }),
  ]

  const table = useReactTable({
    data: filtered,
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

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">项目</h1>
          <p className="mt-1 text-sm text-zinc-500">管理构建项目</p>
        </div>
        <Button className="gap-2" onClick={openCreate}>
          <Plus className="size-4" />
          新建项目
        </Button>
      </div>

      <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-zinc-500" />
          <Input
            placeholder="搜索项目..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
          />
        </div>
        <div className="flex gap-1">
          <Button
            variant={viewMode === 'card' ? 'secondary' : 'ghost'}
            size="icon"
            onClick={() => setViewMode('card')}
          >
            <LayoutGrid className="size-4" />
          </Button>
          <Button
            variant={viewMode === 'table' ? 'secondary' : 'ghost'}
            size="icon"
            onClick={() => setViewMode('table')}
          >
            <List className="size-4" />
          </Button>
        </div>
      </div>

      {error && (
        <div className="rounded-lg border border-red-500/30 bg-red-500/10 p-4 text-red-400">{error}</div>
      )}

      {viewMode === 'card' ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {filtered.map((p) => (
            <div key={p.id} className="group relative">
              <Link to={`/projects/${p.id}`}>
                <Card className="h-full border-zinc-200 transition-colors hover:border-zinc-400 dark:border-zinc-800 dark:hover:border-zinc-600">
                  <CardHeader className="pb-2">
                    <div className="flex items-start justify-between">
                      <div className="flex items-center gap-2">
                        <div className="rounded-lg bg-zinc-100 dark:bg-zinc-800 p-2">
                          <FolderGit2 className="size-5 text-zinc-600 dark:text-zinc-400" />
                        </div>
                        <CardTitle className="text-lg">{p.name}</CardTitle>
                      </div>
                    </div>
                    <CardDescription className="line-clamp-2">{p.description || '暂无描述'}</CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-2">
                    <p className="truncate text-xs text-zinc-500">{p.repo_url}</p>
                    <div className="flex items-center justify-between">
                      <Badge variant="secondary">
                        {(p.environments?.length ?? 0)} 个环境
                      </Badge>
                    </div>
                  </CardContent>
                </Card>
              </Link>
              <Button
                variant="ghost"
                size="icon-sm"
                className="absolute top-3 right-3 opacity-0 transition-opacity group-hover:opacity-100"
                onClick={(e) => {
                  e.preventDefault()
                  e.stopPropagation()
                  openEdit(p.id)
                }}
              >
                <Pencil className="size-4" />
              </Button>
            </div>
          ))}
        </div>
      ) : (
        <Card className="border-zinc-200 dark:border-zinc-800">
          <CardHeader>
            <CardTitle>项目列表</CardTitle>
            <CardDescription>共 {filtered.length} 个项目</CardDescription>
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
      )}

      <ProjectFormDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        editId={editingId}
        onSuccess={handleSuccess}
      />
    </div>
  )
}
