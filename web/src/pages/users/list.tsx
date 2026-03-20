import { useState, useEffect } from 'react'
import { Plus, Pencil, Trash2 } from 'lucide-react'
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
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
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
  DialogBody,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { api } from '@/lib/api'
import type { PaginatedData } from '@/lib/api'
import { ROLES } from '@/lib/constants'
import { useAuthStore } from '@/stores/auth-store'
import { toast } from 'sonner'

interface User {
  id: number
  username: string
  display_name: string
  role: string
  email: string
  is_active: boolean
  created_at: string
}

export function UserListPage() {
  const currentUser = useAuthStore((s) => s.user)
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [createOpen, setCreateOpen] = useState(false)
  const [editUser, setEditUser] = useState<User | null>(null)
  const [deleteId, setDeleteId] = useState<number | null>(null)

  const [createForm, setCreateForm] = useState({
    username: '',
    password: '',
    display_name: '',
    role: 'dev',
    email: '',
  })
  const [editForm, setEditForm] = useState({
    display_name: '',
    role: 'dev',
    email: '',
    is_active: true,
  })

  const [ submitting, setSubmitting] = useState(false)

  useEffect(() => {
    fetchUsers()
  }, [])

  const fetchUsers = async () => {
    try {
      const res = await api.get<PaginatedData<User>>('/users?page=1&page_size=100')
      if (res.code === 0 && res.data) {
        setUsers((res.data as PaginatedData<User>).items || [])
      }
    } catch {
      toast.error('加载失败')
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!createForm.username || !createForm.password) {
      toast.error('请填写用户名和密码')
      return
    }
    setSubmitting(true)
    try {
      const res = await api.post<User>('/users', createForm)
      if (res.code === 0) {
        toast.success('用户已创建')
        setCreateOpen(false)
        setCreateForm({ username: '', password: '', display_name: '', role: 'dev', email: '' })
        fetchUsers()
      } else toast.error(res.message || '创建失败')
    } catch {
      toast.error('创建失败')
    } finally {
      setSubmitting(false)
    }
  }

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!editUser) return
    setSubmitting(true)
    try {
      const res = await api.put<User>(`/users/${editUser.id}`, editForm)
      if (res.code === 0) {
        toast.success('已更新')
        setEditUser(null)
        fetchUsers()
      } else toast.error(res.message || '更新失败')
    } catch {
      toast.error('更新失败')
    } finally {
      setSubmitting(false)
    }
  }

  const handleDelete = async () => {
    if (!deleteId) return
    setSubmitting(true)
    try {
      const res = await api.delete(`/users/${deleteId}`)
      if (res.code === 0) {
        toast.success('已删除')
        setDeleteId(null)
        fetchUsers()
      } else toast.error(res.message || '删除失败')
    } catch {
      toast.error('删除失败')
    } finally {
      setSubmitting(false)
    }
  }

  const columnHelper = createColumnHelper<User>()
  const columns: ColumnDef<User, any>[] = [
    columnHelper.accessor('username', { header: '用户名' }),
    columnHelper.accessor('display_name', { header: '昵称', cell: ({ getValue }) => getValue() || '-' }),
    columnHelper.accessor('role', {
      header: '角色',
      cell: ({ getValue }) => {
        const r = String(getValue())
        const info = ROLES.find((x) => x.value === r)
        return <Badge variant="secondary">{info?.label ?? r}</Badge>
      },
    }),
    columnHelper.accessor('email', { header: '邮箱', cell: ({ getValue }) => getValue() || '-' }),
    columnHelper.accessor('is_active', {
      header: '状态',
      cell: ({ getValue }) => (
        <Badge variant={getValue() ? 'default' : 'secondary'}>
          {getValue() ? '活跃' : '禁用'}
        </Badge>
      ),
    }),
    columnHelper.display({
      id: 'actions',
      header: '操作',
      cell: ({ row }) => (
        <div className="flex gap-1">
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={() => {
              setEditUser(row.original)
              setEditForm({
                display_name: row.original.display_name || '',
                role: row.original.role,
                email: row.original.email || '',
                is_active: row.original.is_active,
              })
            }}
          >
            <Pencil className="size-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={() => setDeleteId(row.original.id)}
            disabled={row.original.id === currentUser?.id}
          >
            <Trash2 className="size-4" />
          </Button>
        </div>
      ),
    }),
  ]

  const table = useReactTable({
    data: users,
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
          <h1 className="text-2xl font-bold tracking-tight">用户管理</h1>
          <p className="mt-1 text-sm text-zinc-500">管理系统用户与角色</p>
        </div>
        <Button onClick={() => setCreateOpen(true)} className="gap-2">
          <Plus className="size-4" />
          新建用户
        </Button>
      </div>

      <Card className="border-zinc-200 dark:border-zinc-800">
        <CardHeader>
          <CardTitle>用户列表</CardTitle>
          <CardDescription>共 {users.length} 个用户</CardDescription>
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

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <form onSubmit={handleCreate} className="flex min-h-0 flex-1 flex-col">
            <DialogHeader>
              <DialogTitle>新建用户</DialogTitle>
              <DialogDescription>创建新的系统用户</DialogDescription>
            </DialogHeader>
            <DialogBody className="space-y-4">
              <div>
                <Label>用户名 *</Label>
                <Input value={createForm.username} onChange={(e) => setCreateForm((f) => ({ ...f, username: e.target.value }))} placeholder="用户名" className="mt-1" />
              </div>
              <div>
                <Label>密码 *</Label>
                <Input type="password" value={createForm.password} onChange={(e) => setCreateForm((f) => ({ ...f, password: e.target.value }))} placeholder="密码" className="mt-1" />
              </div>
              <div>
                <Label>昵称</Label>
                <Input value={createForm.display_name} onChange={(e) => setCreateForm((f) => ({ ...f, display_name: e.target.value }))} placeholder="显示名称" className="mt-1" />
              </div>
              <div>
                <Label>角色</Label>
                <Select value={createForm.role} onValueChange={(v) => setCreateForm((f) => ({ ...f, role: v }))}>
                  <SelectTrigger className="mt-1">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {ROLES.map((r) => (
                      <SelectItem key={r.value} value={r.value}>{r.label}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label>邮箱</Label>
                <Input type="email" value={createForm.email} onChange={(e) => setCreateForm((f) => ({ ...f, email: e.target.value }))} placeholder="email@example.com" className="mt-1" />
              </div>
            </DialogBody>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setCreateOpen(false)}>取消</Button>
              <Button type="submit" disabled={submitting}>{submitting ? '创建中...' : '创建'}</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={!!editUser} onOpenChange={(o) => !o && setEditUser(null)}>
        <DialogContent>
          <form onSubmit={handleUpdate} className="flex min-h-0 flex-1 flex-col">
            <DialogHeader>
              <DialogTitle>编辑用户 {editUser?.username}</DialogTitle>
              <DialogDescription>修改用户信息</DialogDescription>
            </DialogHeader>
            <DialogBody className="space-y-4">
              <div>
                <Label>昵称</Label>
                <Input value={editForm.display_name} onChange={(e) => setEditForm((f) => ({ ...f, display_name: e.target.value }))} placeholder="显示名称" className="mt-1" />
              </div>
              <div>
                <Label>角色</Label>
                <Select value={editForm.role} onValueChange={(v) => setEditForm((f) => ({ ...f, role: v }))}>
                  <SelectTrigger className="mt-1">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {ROLES.map((r) => (
                      <SelectItem key={r.value} value={r.value}>{r.label}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label>邮箱</Label>
                <Input type="email" value={editForm.email} onChange={(e) => setEditForm((f) => ({ ...f, email: e.target.value }))} placeholder="email@example.com" className="mt-1" />
              </div>
              <div className="flex items-center justify-between">
                <Label>启用</Label>
                <Switch checked={editForm.is_active} onCheckedChange={(v) => setEditForm((f) => ({ ...f, is_active: v }))} />
              </div>
            </DialogBody>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setEditUser(null)}>取消</Button>
              <Button type="submit" disabled={submitting}>{submitting ? '保存中...' : '保存'}</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={!!deleteId} onOpenChange={(o) => !o && setDeleteId(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认删除</DialogTitle>
            <DialogDescription>确定要删除此用户吗？此操作不可撤销。不能删除自己的账号。</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteId(null)}>取消</Button>
            <Button variant="destructive" onClick={handleDelete} disabled={submitting}>
              {submitting ? '删除中...' : '删除'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
