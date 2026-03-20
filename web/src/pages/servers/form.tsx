import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
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
import { api } from '@/lib/api'
import { AUTH_TYPES, OS_TYPES } from '@/lib/constants'

interface ServerPayload {
  name: string
  host: string
  port: number
  os_type: string
  username: string
  auth_type: string
  password: string
  private_key: string
  agent_url: string
  agent_token: string
  description: string
  tags: string
}

interface ServerDetail extends Omit<ServerPayload, 'password' | 'private_key'> {
  id: number
}

const DEFAULT_FORM: ServerPayload = {
  name: '',
  host: '',
  port: 22,
  os_type: 'linux',
  username: 'root',
  auth_type: 'password',
  password: '',
  private_key: '',
  agent_url: '',
  agent_token: '',
  description: '',
  tags: '',
}

interface ServerFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  editId?: number | null
  onSuccess?: () => void
}

export function ServerFormDialog({
  open,
  onOpenChange,
  editId,
  onSuccess,
}: ServerFormDialogProps) {
  const isEdit = !!editId

  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [form, setForm] = useState<ServerPayload>(DEFAULT_FORM)

  useEffect(() => {
    if (!open) {
      setForm(DEFAULT_FORM)
      setError('')
      return
    }

    if (!isEdit || !editId) return

    const fetchServer = async () => {
      setLoading(true)
      try {
        const res = await api.get<ServerDetail>(`/servers/${editId}`)
        if (res.code !== 0 || !res.data) {
          throw new Error(res.message || '加载服务器失败')
        }
        setForm({
          name: res.data.name || '',
          host: res.data.host || '',
          port: res.data.port || 22,
          os_type: res.data.os_type || 'linux',
          username: res.data.username || '',
          auth_type: res.data.auth_type || 'password',
          password: '',
          private_key: '',
          agent_url: res.data.agent_url || '',
          agent_token: '',
          description: res.data.description || '',
          tags: res.data.tags || '',
        })
      } catch (err) {
        const message = err instanceof Error ? err.message : '加载服务器失败'
        setError(message)
        toast.error(message)
      } finally {
        setLoading(false)
      }
    }

    fetchServer()
  }, [open, editId, isEdit])

  const setField = <K extends keyof ServerPayload>(key: K, value: ServerPayload[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }))
  }

  const validate = () => {
    if (!form.name.trim()) return '请输入服务器名称'
    if (form.auth_type !== 'agent') {
      if (!form.host.trim()) return '请输入主机地址'
      if (form.port < 1 || form.port > 65535) return '端口范围为 1-65535'
      if (!form.username.trim()) return '请输入用户名'
    }
    if (!isEdit) {
      if (form.auth_type === 'password' && !form.password) return '请输入密码'
      if (form.auth_type === 'key' && !form.private_key.trim()) return '请输入 SSH 私钥'
      if (form.auth_type === 'agent' && !form.agent_token.trim()) return '请输入 Agent Token'
    }
    if (form.auth_type === 'agent' && !form.agent_url.trim()) {
      return '请输入 Agent URL'
    }
    return ''
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const validationError = validate()
    if (validationError) {
      setError(validationError)
      return
    }

    setError('')
    setSubmitting(true)

    try {
      const payload: ServerPayload = {
        name: form.name.trim(),
        host: form.host.trim(),
        port: form.auth_type === 'agent' ? 0 : form.port,
        os_type: form.os_type,
        username: form.auth_type === 'agent' ? '' : form.username.trim(),
        auth_type: form.auth_type,
        password: form.auth_type === 'password' ? form.password : '',
        private_key: form.auth_type === 'key' ? form.private_key : '',
        agent_url: form.auth_type === 'agent' ? form.agent_url.trim() : '',
        agent_token: form.auth_type === 'agent' ? form.agent_token : '',
        description: form.description.trim(),
        tags: form.tags.trim(),
      }

      if (isEdit && editId) {
        const res = await api.put<ServerDetail>(`/servers/${editId}`, payload)
        if (res.code !== 0) {
          throw new Error(res.message || '更新服务器失败')
        }
        toast.success('服务器已更新')
      } else {
        const res = await api.post<ServerDetail>('/servers', payload)
        if (res.code !== 0) {
          throw new Error(res.message || '创建服务器失败')
        }
        toast.success('服务器创建成功')
      }

      onOpenChange(false)
      onSuccess?.()
    } catch (err) {
      const message = err instanceof Error ? err.message : '提交失败'
      setError(message)
      toast.error(message)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[560px]">
        <DialogHeader>
          <DialogTitle>{isEdit ? '编辑服务器' : '新建服务器'}</DialogTitle>
          <DialogDescription>
            {isEdit ? '更新服务器连接配置' : '添加新的部署目标服务器'}
          </DialogDescription>
        </DialogHeader>

        {loading ? (
          <DialogBody className="flex items-center justify-center py-10">
            <div className="size-8 animate-spin rounded-full border-2 border-zinc-600 border-t-zinc-300" />
          </DialogBody>
        ) : (
          <form onSubmit={handleSubmit} className="flex min-h-0 flex-1 flex-col">
            <DialogBody className="space-y-4">
              {error && (
                <div className="rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400">
                  {error}
                </div>
              )}

              <div className="grid gap-4 sm:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="server-name">服务器名称 *</Label>
                  <Input
                    id="server-name"
                    value={form.name}
                    onChange={(e) => setField('name', e.target.value)}
                    placeholder="例如：production-web-01"
                    maxLength={100}
                  />
                </div>
                <div className="space-y-2">
                  <Label>操作系统 *</Label>
                  <Select value={form.os_type} onValueChange={(value) => setField('os_type', value)}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {OS_TYPES.map((t) => (
                        <SelectItem key={t.value} value={t.value}>
                          {t.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="grid gap-4 sm:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="server-host">主机地址 *</Label>
                  <Input
                    id="server-host"
                    value={form.host}
                    onChange={(e) => setField('host', e.target.value)}
                    placeholder={form.auth_type === 'agent' ? '可选，默认从 Agent URL 自动解析' : 'IP 地址或域名'}
                    maxLength={200}
                  />
                </div>
                <div className="space-y-2">
                  <Label>认证方式 *</Label>
                  <Select
                    value={form.auth_type}
                    onValueChange={(value) => {
                      setForm((prev) => ({
                        ...prev,
                        auth_type: value,
                        password: '',
                        private_key: '',
                        agent_url: value === 'agent' ? prev.agent_url : '',
                        agent_token: '',
                      }))
                    }}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {AUTH_TYPES.map((t) => (
                        <SelectItem key={t.value} value={t.value}>
                          {t.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>

              {form.auth_type !== 'agent' && (
                <div className="grid gap-4 sm:grid-cols-2">
                  <div className="space-y-2">
                    <Label htmlFor="server-port">SSH 端口 *</Label>
                    <Input
                      id="server-port"
                      type="number"
                      min={1}
                      max={65535}
                      value={form.port}
                      onChange={(e) => setField('port', Math.max(1, Number(e.target.value) || 22))}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="server-username">用户名 *</Label>
                    <Input
                      id="server-username"
                      value={form.username}
                      onChange={(e) => setField('username', e.target.value)}
                      placeholder="root"
                      maxLength={100}
                    />
                  </div>
                </div>
              )}

              {form.auth_type === 'password' && (
                <div className="space-y-2">
                  <Label htmlFor="server-password">
                    {isEdit ? '密码（留空表示不变）' : '密码 *'}
                  </Label>
                  <Input
                    id="server-password"
                    type="password"
                    value={form.password}
                    onChange={(e) => setField('password', e.target.value)}
                    placeholder={isEdit ? '如不修改可留空' : '请输入 SSH 登录密码'}
                  />
                </div>
              )}

              {form.auth_type === 'agent' && (
                <>
                  <div className="space-y-2">
                    <Label htmlFor="server-agent-url">Agent URL *</Label>
                    <Input
                      id="server-agent-url"
                      value={form.agent_url}
                      onChange={(e) => setField('agent_url', e.target.value)}
                      placeholder="http://server:9091"
                      maxLength={500}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="server-agent-token">
                      {isEdit ? 'Agent Token（留空表示不变）' : 'Agent Token *'}
                    </Label>
                    <Input
                      id="server-agent-token"
                      type="password"
                      value={form.agent_token}
                      onChange={(e) => setField('agent_token', e.target.value)}
                      placeholder={isEdit ? '如不修改可留空' : '请输入 Agent Bearer Token'}
                    />
                  </div>
                </>
              )}

              {form.auth_type === 'key' && (
                <div className="space-y-2">
                  <Label htmlFor="server-private-key">
                    {isEdit ? 'SSH 私钥（留空表示不变）' : 'SSH 私钥 *'}
                  </Label>
                  <Textarea
                    id="server-private-key"
                    value={form.private_key}
                    onChange={(e) => setField('private_key', e.target.value)}
                    placeholder="-----BEGIN OPENSSH PRIVATE KEY-----"
                    rows={5}
                    className="font-mono text-xs"
                  />
                </div>
              )}

              <div className="space-y-2">
                <Label htmlFor="server-description">描述</Label>
                <Textarea
                  id="server-description"
                  value={form.description}
                  onChange={(e) => setField('description', e.target.value)}
                  placeholder="服务器用途说明"
                  rows={2}
                  maxLength={500}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="server-tags">标签</Label>
                <Input
                  id="server-tags"
                  value={form.tags}
                  onChange={(e) => setField('tags', e.target.value)}
                  placeholder="多个标签用逗号分隔，如：web,production"
                  maxLength={500}
                />
              </div>
            </DialogBody>

            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                取消
              </Button>
              <Button type="submit" disabled={submitting}>
                {submitting
                  ? isEdit ? '保存中...' : '创建中...'
                  : isEdit ? '保存' : '创建服务器'}
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  )
}
