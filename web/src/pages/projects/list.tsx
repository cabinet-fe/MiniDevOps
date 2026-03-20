import { useState, useEffect, useCallback, useMemo } from 'react'
import { Link } from 'react-router'
import {
  FolderGit2,
  LayoutGrid,
  List,
  Pencil,
  Plus,
  Search,
} from 'lucide-react'
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
  tags: string
  repo_url: string
  environments?: { id: number; name: string }[]
}

function splitTags(tags: string) {
  return (tags || '')
    .split(',')
    .map((tag) => tag.trim())
    .filter(Boolean)
}

function tagDictLabel(value: string, dict: { label: string; value: string }[]) {
  return dict.find((d) => d.value === value)?.label ?? value
}

export function ProjectListPage() {
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [selectedTag, setSelectedTag] = useState('all')
  const [viewMode, setViewMode] = useState<'card' | 'table'>('card')
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)

  const fetchProjects = useCallback(async () => {
    setLoading(true)
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

  const [dictTags, setDictTags] = useState<{ label: string; value: string }[]>([])

  useEffect(() => {
    api
      .get<{ label: string; value: string }[]>('/dictionaries/code/project_tags/items')
      .then((res) => {
        if (res.code === 0 && res.data) {
          setDictTags(res.data)
        }
      })
  }, [])

  const allTags = useMemo(() => {
    if (dictTags.length > 0) {
      return dictTags.map((t) => t.value)
    }
    const tags = new Set<string>()
    for (const project of projects) {
      splitTags(project.tags).forEach((tag) => tags.add(tag))
    }
    return Array.from(tags).sort((a, b) => a.localeCompare(b, 'zh-CN'))
  }, [projects, dictTags])

  const filtered = useMemo(() => {
    const keyword = search.trim().toLowerCase()
    return projects.filter((project) => {
      const tagList = splitTags(project.tags)
      const matchTag = selectedTag === 'all' || tagList.includes(selectedTag)
      if (!matchTag) return false
      if (!keyword) return true
      return (
        project.name.toLowerCase().includes(keyword) ||
        (project.description || '').toLowerCase().includes(keyword) ||
        tagList.some((tag) => {
          if (tag.toLowerCase().includes(keyword)) return true
          if (dictTags.length > 0) {
            return tagDictLabel(tag, dictTags).toLowerCase().includes(keyword)
          }
          return false
        })
      )
    })
  }, [projects, search, selectedTag, dictTags])

  const columnHelper = createColumnHelper<Project>()
  const columns = [
    columnHelper.accessor('name', { header: '项目名称' }),
    columnHelper.accessor('tags', {
      header: '标签',
      cell: ({ getValue }) => {
        const tags = splitTags(String(getValue()))
        if (tags.length === 0) return '-'
        return (
          <div className="flex flex-wrap gap-1">
            {tags.map((tag) => (
              <Badge key={tag} variant="secondary">
                {tagDictLabel(tag, dictTags)}
              </Badge>
            ))}
          </div>
        )
      },
    }),
    columnHelper.accessor('description', { header: '描述', cell: ({ getValue }) => getValue() || '-' }),
    columnHelper.accessor((row) => row.environments?.length ?? 0, { header: '环境数', id: 'env_count' }),
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
            <Button variant="outline" size="sm">
              查看
            </Button>
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
        <div className="border-muted size-8 animate-spin rounded-full border-2 border-t-foreground" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-foreground text-2xl font-bold tracking-tight">项目</h1>
          <p className="text-muted-foreground mt-1 text-sm">按标签与关键字管理构建项目</p>
        </div>
        <Button className="gap-2" onClick={openCreate}>
          <Plus className="size-4" />
          新建项目
        </Button>
      </div>

      <div className="flex flex-col gap-4">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
          <div className="relative flex-1">
            <Search className="text-muted-foreground absolute left-3 top-1/2 size-4 -translate-y-1/2" />
            <Input
              placeholder="搜索项目名称、描述或标签..."
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

        {allTags.length > 0 && (
          <div className="flex flex-wrap gap-2">
            <Button
              variant={selectedTag === 'all' ? 'secondary' : 'outline'}
              size="sm"
              onClick={() => setSelectedTag('all')}
            >
              全部标签
            </Button>
            {allTags.map((tag) => (
              <Button
                key={tag}
                variant={selectedTag === tag ? 'secondary' : 'outline'}
                size="sm"
                onClick={() => setSelectedTag(tag)}
              >
                {tagDictLabel(tag, dictTags)}
              </Button>
            ))}
          </div>
        )}
      </div>

      {error && (
        <div className="rounded-lg border border-red-500/30 bg-red-500/10 p-4 text-red-400">{error}</div>
      )}

      {viewMode === 'card' ? (
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {filtered.map((project) => (
            <div key={project.id} className="group relative">
              <Link to={`/projects/${project.id}`}>
                <Card className="h-full border-border transition-colors">
                  <CardHeader className="pb-2">
                    <div className="flex items-start gap-3">
                      <div className="bg-muted rounded-lg p-2">
                        <FolderGit2 className="text-muted-foreground size-5" />
                      </div>
                      <div className="min-w-0 flex-1">
                        <CardTitle className="truncate text-lg">{project.name}</CardTitle>
                        <CardDescription className="line-clamp-2 mt-1">
                          {project.description || '暂无描述'}
                        </CardDescription>
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-3">
                    <p className="text-muted-foreground truncate text-xs">{project.repo_url}</p>
                    <div className="flex flex-wrap gap-1">
                      {splitTags(project.tags).map((tag) => (
                        <Badge key={tag} variant="secondary">
                          {tagDictLabel(tag, dictTags)}
                        </Badge>
                      ))}
                      {splitTags(project.tags).length === 0 && (
                        <span className="text-muted-foreground text-xs">无标签</span>
                      )}
                    </div>
                    <Badge variant="outline">{project.environments?.length ?? 0} 个环境</Badge>
                  </CardContent>
                </Card>
              </Link>
              <Button
                variant="ghost"
                size="icon-sm"
                className="absolute right-3 top-3 opacity-0 transition-opacity group-hover:opacity-100"
                onClick={(e) => {
                  e.preventDefault()
                  e.stopPropagation()
                  openEdit(project.id)
                }}
              >
                <Pencil className="size-4" />
              </Button>
            </div>
          ))}
        </div>
      ) : (
        <Card className="border-border">
          <CardHeader>
            <CardTitle>项目列表</CardTitle>
            <CardDescription>共 {filtered.length} 个项目</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                {table.getHeaderGroups().map((headerGroup) => (
                  <TableRow key={headerGroup.id}>
                    {headerGroup.headers.map((header) => (
                      <TableHead key={header.id}>
                        {flexRender(header.column.columnDef.header, header.getContext())}
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
