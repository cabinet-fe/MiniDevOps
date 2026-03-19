import { useState, useEffect, useCallback, useMemo } from 'react'
import { Link } from 'react-router'
import {
  ChevronDown,
  ChevronRight,
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
  group_name: string
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

export function ProjectListPage() {
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [selectedTag, setSelectedTag] = useState('all')
  const [viewMode, setViewMode] = useState<'card' | 'table'>('card')
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [collapsedGroups, setCollapsedGroups] = useState<Record<string, boolean>>({})

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

  const allTags = useMemo(() => {
    const tags = new Set<string>()
    for (const project of projects) {
      splitTags(project.tags).forEach((tag) => tags.add(tag))
    }
    return Array.from(tags).sort((a, b) => a.localeCompare(b, 'zh-CN'))
  }, [projects])

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
        (project.group_name || '').toLowerCase().includes(keyword) ||
        tagList.some((tag) => tag.toLowerCase().includes(keyword))
      )
    })
  }, [projects, search, selectedTag])

  const groupedProjects = useMemo(() => {
    const groups = new Map<string, Project[]>()
    for (const project of filtered) {
      const groupName = project.group_name || '未分组'
      const current = groups.get(groupName) ?? []
      current.push(project)
      groups.set(groupName, current)
    }
    return Array.from(groups.entries()).sort((a, b) => a[0].localeCompare(b[0], 'zh-CN'))
  }, [filtered])

  const toggleGroup = (groupName: string) => {
    setCollapsedGroups((prev) => ({ ...prev, [groupName]: !prev[groupName] }))
  }

  const columnHelper = createColumnHelper<Project>()
  const columns = [
    columnHelper.accessor('name', { header: '项目名称' }),
    columnHelper.accessor('group_name', {
      header: '分组',
      cell: ({ getValue }) => getValue() || '未分组',
    }),
    columnHelper.accessor('tags', {
      header: '标签',
      cell: ({ getValue }) => {
        const tags = splitTags(String(getValue()))
        if (tags.length === 0) return '-'
        return (
          <div className="flex flex-wrap gap-1">
            {tags.map((tag) => (
              <Badge key={tag} variant="secondary">
                {tag}
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
        <div className="size-8 animate-spin rounded-full border-2 border-zinc-600 border-t-zinc-300" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">项目</h1>
          <p className="mt-1 text-sm text-zinc-500">按分组、标签与关键字管理构建项目</p>
        </div>
        <Button className="gap-2" onClick={openCreate}>
          <Plus className="size-4" />
          新建项目
        </Button>
      </div>

      <div className="flex flex-col gap-4">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-zinc-500" />
            <Input
              placeholder="搜索项目名称、描述、分组或标签..."
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
                {tag}
              </Button>
            ))}
          </div>
        )}
      </div>

      {error && (
        <div className="rounded-lg border border-red-500/30 bg-red-500/10 p-4 text-red-400">{error}</div>
      )}

      {viewMode === 'card' ? (
        <div className="space-y-4">
          {groupedProjects.map(([groupName, items]) => {
            const collapsed = !!collapsedGroups[groupName]
            return (
              <Card key={groupName} className="border-zinc-200 dark:border-zinc-800">
                <CardHeader className="pb-3">
                  <button
                    type="button"
                    className="flex items-center justify-between text-left"
                    onClick={() => toggleGroup(groupName)}
                  >
                    <div>
                      <CardTitle className="flex items-center gap-2 text-lg">
                        {collapsed ? <ChevronRight className="size-4" /> : <ChevronDown className="size-4" />}
                        {groupName}
                      </CardTitle>
                      <CardDescription className="mt-1">{items.length} 个项目</CardDescription>
                    </div>
                  </button>
                </CardHeader>
                {!collapsed && (
                  <CardContent className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
                    {items.map((project) => (
                      <div key={project.id} className="group relative">
                        <Link to={`/projects/${project.id}`}>
                          <Card className="h-full border-zinc-200 transition-colors hover:border-zinc-400 dark:border-zinc-800 dark:hover:border-zinc-600">
                            <CardHeader className="pb-2">
                              <div className="flex items-start gap-3">
                                <div className="rounded-lg bg-zinc-100 p-2 dark:bg-zinc-800">
                                  <FolderGit2 className="size-5 text-zinc-600 dark:text-zinc-400" />
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
                              <p className="truncate text-xs text-zinc-500">{project.repo_url}</p>
                              <div className="flex flex-wrap gap-1">
                                {splitTags(project.tags).map((tag) => (
                                  <Badge key={tag} variant="secondary">
                                    {tag}
                                  </Badge>
                                ))}
                                {splitTags(project.tags).length === 0 && <span className="text-xs text-zinc-500">无标签</span>}
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
                  </CardContent>
                )}
              </Card>
            )
          })}
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
