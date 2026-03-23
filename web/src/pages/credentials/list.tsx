import { useCallback, useEffect, useMemo, useState } from 'react'
import { KeyRound, Pencil, Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
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
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { api } from '@/lib/api'
import { CREDENTIAL_TYPES } from '@/lib/constants'
import { useAuthStore } from '@/stores/auth-store'

interface Credential {
  id: number
  name: string
  type: string
  username: string
  description: string
  created_by: number
  creator_name?: string
  has_secret: boolean
  created_at: string
}

interface CredentialPayload {
  name: string
  type: string
  username: string
  password: string
  description: string
}

const DEFAULT_FORM: CredentialPayload = {
  name: '',
  type: 'password',
  username: '',
  password: '',
  description: '',
}

function formatType(type: string) {
  return CREDENTIAL_TYPES.find((item) => item.value === type)?.label ?? type
}

function formatTime(value: string) {
  if (!value) return '-'
  return new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function CredentialListPage() {
  const user = useAuthStore((state) => state.user)
  const [credentials, setCredentials] = useState<Credential[]>([])
  const [loading, setLoading] = useState(true)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<Credential | null>(null)
  const [deleting, setDeleting] = useState(false)

  const currentUserID = user?.id ?? 0
  const isAdmin = user?.role === 'admin'

  const canManage = useCallback(
    (credential: Credential) => currentUserID > 0 && credential.created_by === currentUserID,
    [currentUserID],
  )

  const fetchCredentials = useCallback(async () => {
    setLoading(true)
    try {
      const res = await api.get<Credential[]>('/credentials')
      if (res.code !== 0 || !res.data) {
        throw new Error(res.message || '加载凭证失败')
      }
      setCredentials(res.data)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '加载凭证失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchCredentials()
  }, [fetchCredentials])

  const summary = useMemo(() => {
    const tokenCount = credentials.filter((item) => item.type === 'token').length
    const passwordCount = credentials.filter((item) => item.type === 'password').length
    return { tokenCount, passwordCount }
  }, [credentials])

  const handleDelete = async () => {
    if (!deleteTarget) return
    setDeleting(true)
    try {
      const res = await api.delete(`/credentials/${deleteTarget.id}`)
      if (res.code !== 0) {
        throw new Error(res.message || '删除失败')
      }
      toast.success('凭证已删除')
      setDeleteTarget(null)
      fetchCredentials()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '删除失败')
    } finally {
      setDeleting(false)
    }
  }

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
          <h1 className="text-foreground text-2xl font-bold tracking-tight">凭证</h1>
          <p className="text-muted-foreground mt-1 text-sm">管理仓库访问凭证，敏感信息仅加密存储</p>
        </div>
        <Button
          className="gap-2"
          onClick={() => {
            setEditingId(null)
            setDialogOpen(true)
          }}
        >
          <Plus className="size-4" />
          新建凭证
        </Button>
      </div>

      <div className="grid gap-4 sm:grid-cols-3">
        <Card className="border-border">
          <CardHeader className="pb-2">
            <CardDescription>总凭证数</CardDescription>
            <CardTitle className="text-2xl">{credentials.length}</CardTitle>
          </CardHeader>
        </Card>
        <Card className="border-border">
          <CardHeader className="pb-2">
            <CardDescription>用户名/密码</CardDescription>
            <CardTitle className="text-2xl">{summary.passwordCount}</CardTitle>
          </CardHeader>
        </Card>
        <Card className="border-border">
          <CardHeader className="pb-2">
            <CardDescription>Token</CardDescription>
            <CardTitle className="text-2xl">{summary.tokenCount}</CardTitle>
          </CardHeader>
        </Card>
      </div>

      <Card className="border-border">
        <CardHeader>
          <CardTitle>凭证列表</CardTitle>
          <CardDescription>
            {isAdmin ? '管理员可查看全部凭证，仅可编辑/删除自己创建的凭证' : '仅展示你创建的凭证'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>名称</TableHead>
                <TableHead>类型</TableHead>
                <TableHead>用户名</TableHead>
                <TableHead>描述</TableHead>
                <TableHead>创建者</TableHead>
                <TableHead>创建时间</TableHead>
                <TableHead className="w-28 text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {credentials.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-muted-foreground py-8 text-center">
                    暂无凭证
                  </TableCell>
                </TableRow>
              ) : (
                credentials.map((credential) => {
                  const editable = canManage(credential)
                  return (
                    <TableRow key={credential.id}>
                      <TableCell className="font-medium">
                        <div className="flex items-center gap-2">
                          <KeyRound className="text-muted-foreground size-4" />
                          {credential.name}
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">{formatType(credential.type)}</Badge>
                      </TableCell>
                      <TableCell className="font-mono text-xs">
                        {credential.username || '-'}
                      </TableCell>
                      <TableCell className="text-muted-foreground max-w-[280px] truncate">
                        {credential.description || '-'}
                      </TableCell>
                      <TableCell>
                        {credential.creator_name || `用户 #${credential.created_by}`}
                      </TableCell>
                      <TableCell>{formatTime(credential.created_at)}</TableCell>
                      <TableCell>
                        <div className="flex justify-end gap-1">
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            disabled={!editable}
                            onClick={() => {
                              setEditingId(credential.id)
                              setDialogOpen(true)
                            }}
                          >
                            <Pencil className="size-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            disabled={!editable}
                            onClick={() => setDeleteTarget(credential)}
                          >
                            <Trash2 className="size-4" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  )
                })
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <CredentialFormDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        editId={editingId}
        onSuccess={fetchCredentials}
      />

      <Dialog open={!!deleteTarget} onOpenChange={(open) => !open && setDeleteTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认删除凭证</DialogTitle>
            <DialogDescription>
              删除后相关项目将无法继续使用该凭证拉取仓库，且该操作不可撤销。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteTarget(null)}>
              取消
            </Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleting}>
              {deleting ? '删除中...' : '删除'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

function CredentialFormDialog({
  open,
  onOpenChange,
  editId,
  onSuccess,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  editId: number | null
  onSuccess: () => void
}) {
  const isEdit = !!editId
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [form, setForm] = useState<CredentialPayload>(DEFAULT_FORM)

  useEffect(() => {
    if (!open) {
      setForm(DEFAULT_FORM)
      setError('')
      return
    }
    if (!isEdit || !editId) {
      setForm(DEFAULT_FORM)
      return
    }

    const fetchCredential = async () => {
      setLoading(true)
      try {
        const res = await api.get<Credential>(`/credentials/${editId}`)
        if (res.code !== 0 || !res.data) {
          throw new Error(res.message || '加载凭证失败')
        }
        setForm({
          name: res.data.name || '',
          type: res.data.type || 'password',
          username: res.data.username || '',
          password: '',
          description: res.data.description || '',
        })
      } catch (err) {
        const message = err instanceof Error ? err.message : '加载凭证失败'
        setError(message)
        toast.error(message)
      } finally {
        setLoading(false)
      }
    }

    fetchCredential()
  }, [open, isEdit, editId])

  const setField = <K extends keyof CredentialPayload>(key: K, value: CredentialPayload[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }))
  }

  const validate = () => {
    if (!form.name.trim()) return '请输入凭证名称'
    if (form.type === 'password' && !form.username.trim()) return '用户名不能为空'
    if (!isEdit && !form.password.trim()) return '请填写密码或 Token'
    return ''
  }

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    const validationError = validate()
    if (validationError) {
      setError(validationError)
      return
    }

    setSubmitting(true)
    setError('')
    try {
      const payload: Partial<CredentialPayload> = {
        name: form.name.trim(),
        type: form.type,
        username: form.username.trim(),
        description: form.description.trim(),
      }
      if (!isEdit || form.password.trim()) {
        payload.password = form.password
      }

      const res = isEdit && editId
        ? await api.put(`/credentials/${editId}`, payload)
        : await api.post('/credentials', payload)
      if (res.code !== 0) {
        throw new Error(res.message || '保存失败')
      }
      toast.success(isEdit ? '凭证已更新' : '凭证已创建')
      onOpenChange(false)
      onSuccess()
    } catch (err) {
      setError(err instanceof Error ? err.message : '保存失败')
      toast.error(err instanceof Error ? err.message : '保存失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[560px]">
        {loading ? (
          <DialogBody className="flex items-center justify-center py-10">
            <div className="border-muted size-8 animate-spin rounded-full border-2 border-t-foreground" />
          </DialogBody>
        ) : (
          <form onSubmit={submit} className="flex min-h-0 flex-1 flex-col">
            <DialogHeader>
              <DialogTitle>{isEdit ? '编辑凭证' : '新建凭证'}</DialogTitle>
              <DialogDescription>
                凭证敏感字段只会加密存储，不会在 API 中返回明文。
              </DialogDescription>
            </DialogHeader>

            <DialogBody className="space-y-4">
              {error && (
                <div className="rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400">
                  {error}
                </div>
              )}

              <div className="grid gap-4 sm:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="credential-name">凭证名称 *</Label>
                  <Input
                    id="credential-name"
                    value={form.name}
                    onChange={(e) => setField('name', e.target.value)}
                    maxLength={100}
                    placeholder="例如：GitLab Token-生产"
                  />
                </div>

                <div className="space-y-2">
                  <Label>凭证类型 *</Label>
                  <Select
                    value={form.type}
                    onValueChange={(value) =>
                      setForm((prev) => ({
                        ...prev,
                        type: value,
                      }))
                    }
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {CREDENTIAL_TYPES.map((item) => (
                        <SelectItem key={item.value} value={item.value}>
                          {item.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="credential-username">
                  用户名{form.type === 'password' ? ' *' : ''}
                </Label>
                <Input
                  id="credential-username"
                  value={form.username}
                  onChange={(e) => setField('username', e.target.value)}
                  maxLength={200}
                  placeholder={
                    form.type === 'token'
                      ? 'Gitee 必填，其他平台可留空自动识别'
                      : '仓库用户名'
                  }
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="credential-password">{form.type === 'token' ? 'Token' : '密码'} *</Label>
                <Input
                  id="credential-password"
                  type="password"
                  value={form.password}
                  onChange={(e) => setField('password', e.target.value)}
                  placeholder={isEdit ? '留空则保持不变' : `请输入${form.type === 'token' ? 'Token' : '密码'}`}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="credential-description">描述</Label>
                <Textarea
                  id="credential-description"
                  value={form.description}
                  onChange={(e) => setField('description', e.target.value)}
                  rows={3}
                  maxLength={500}
                  placeholder="用于哪个代码托管平台或仓库"
                />
              </div>
            </DialogBody>

            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                取消
              </Button>
              <Button type="submit" disabled={submitting}>
                {submitting ? '提交中...' : isEdit ? '保存修改' : '创建凭证'}
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  )
}
