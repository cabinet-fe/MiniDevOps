import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
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
import { api } from '@/lib/api'
import { AUTH_TYPES } from '@/lib/constants'

interface ServerPayload {
  name: string
  host: string
  port: number
  username: string
  auth_type: string
  password: string
  private_key: string
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
  username: 'root',
  auth_type: 'password',
  password: '',
  private_key: '',
  description: '',
  tags: '',
}

export function ServerFormPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isEdit = !!id

  const [loading, setLoading] = useState(isEdit)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [form, setForm] = useState<ServerPayload>(DEFAULT_FORM)

  useEffect(() => {
    if (!isEdit || !id) return

    const fetchServer = async () => {
      setLoading(true)
      try {
        const res = await api.get<ServerDetail>(`/servers/${id}`)
        if (res.code !== 0 || !res.data) {
          throw new Error(res.message || '加载服务器失败')
        }
        setForm({
          name: res.data.name || '',
          host: res.data.host || '',
          port: res.data.port || 22,
          username: res.data.username || '',
          auth_type: res.data.auth_type || 'password',
          password: '',
          private_key: '',
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
  }, [id, isEdit])

  const setField = <K extends keyof ServerPayload>(key: K, value: ServerPayload[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }))
  }

  const validate = () => {
    if (!form.name.trim()) return '请输入服务器名称'
    if (!form.host.trim()) return '请输入主机地址'
    if (form.port < 1 || form.port > 65535) return '端口范围为 1-65535'
    if (!form.username.trim()) return '请输入用户名'
    if (!isEdit) {
      if (form.auth_type === 'password' && !form.password) return '请输入密码'
      if (form.auth_type === 'key' && !form.private_key.trim()) return '请输入 SSH 私钥'
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
        port: form.port,
        username: form.username.trim(),
        auth_type: form.auth_type,
        password: form.auth_type === 'password' ? form.password : '',
        private_key: form.auth_type === 'key' ? form.private_key : '',
        description: form.description.trim(),
        tags: form.tags.trim(),
      }

      if (isEdit && id) {
        const res = await api.put<ServerDetail>(`/servers/${id}`, payload)
        if (res.code !== 0) {
          throw new Error(res.message || '更新服务器失败')
        }
        toast.success('服务器已更新')
        navigate('/servers', { replace: true })
        return
      }

      const res = await api.post<ServerDetail>('/servers', payload)
      if (res.code !== 0) {
        throw new Error(res.message || '创建服务器失败')
      }
      toast.success('服务器创建成功')
      navigate('/servers', { replace: true })
    } catch (err) {
      const message = err instanceof Error ? err.message : '提交失败'
      setError(message)
      toast.error(message)
    } finally {
      setSubmitting(false)
    }
  }

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
          <h1 className="text-2xl font-bold tracking-tight">
            {isEdit ? '编辑服务器' : '新建服务器'}
          </h1>
          <p className="mt-1 text-sm text-zinc-500">
            {isEdit ? '更新服务器连接配置' : '添加新的部署目标服务器'}
          </p>
        </div>
        <Link to="/servers">
          <Button variant="outline">返回</Button>
        </Link>
      </div>

      <Card className="border-zinc-200 dark:border-zinc-800">
        <CardHeader>
          <CardTitle>服务器配置</CardTitle>
          <CardDescription>带 * 的字段为必填项</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-5">
            {error && (
              <div className="rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400">
                {error}
              </div>
            )}

            <div className="grid gap-5 md:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="name">服务器名称 *</Label>
                <Input
                  id="name"
                  value={form.name}
                  onChange={(e) => setField('name', e.target.value)}
                  placeholder="例如：production-web-01"
                  maxLength={100}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="host">主机地址 *</Label>
                <Input
                  id="host"
                  value={form.host}
                  onChange={(e) => setField('host', e.target.value)}
                  placeholder="IP 地址或域名"
                  maxLength={200}
                />
              </div>
            </div>

            <div className="grid gap-5 md:grid-cols-3">
              <div className="space-y-2">
                <Label htmlFor="port">SSH 端口 *</Label>
                <Input
                  id="port"
                  type="number"
                  min={1}
                  max={65535}
                  value={form.port}
                  onChange={(e) => setField('port', Math.max(1, Number(e.target.value) || 22))}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="username">用户名 *</Label>
                <Input
                  id="username"
                  value={form.username}
                  onChange={(e) => setField('username', e.target.value)}
                  placeholder="root"
                  maxLength={100}
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

            {form.auth_type === 'password' && (
              <div className="space-y-2">
                <Label htmlFor="password">
                  {isEdit ? '密码（留空表示不变）' : '密码 *'}
                </Label>
                <Input
                  id="password"
                  type="password"
                  value={form.password}
                  onChange={(e) => setField('password', e.target.value)}
                  placeholder={isEdit ? '如不修改可留空' : '请输入 SSH 登录密码'}
                />
              </div>
            )}

            {form.auth_type === 'key' && (
              <div className="space-y-2">
                <Label htmlFor="private-key">
                  {isEdit ? 'SSH 私钥（留空表示不变）' : 'SSH 私钥 *'}
                </Label>
                <Textarea
                  id="private-key"
                  value={form.private_key}
                  onChange={(e) => setField('private_key', e.target.value)}
                  placeholder="-----BEGIN OPENSSH PRIVATE KEY-----"
                  rows={6}
                  className="font-mono text-xs"
                />
              </div>
            )}

            <div className="space-y-2">
              <Label htmlFor="description">描述</Label>
              <Textarea
                id="description"
                value={form.description}
                onChange={(e) => setField('description', e.target.value)}
                placeholder="服务器用途说明"
                rows={2}
                maxLength={500}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="tags">标签</Label>
              <Input
                id="tags"
                value={form.tags}
                onChange={(e) => setField('tags', e.target.value)}
                placeholder="多个标签用逗号分隔，如：web,production"
                maxLength={500}
              />
            </div>

            <div className="flex justify-end gap-2 pt-2">
              <Link to="/servers">
                <Button type="button" variant="outline">
                  取消
                </Button>
              </Link>
              <Button type="submit" disabled={submitting}>
                {submitting
                  ? isEdit ? '保存中...' : '创建中...'
                  : isEdit ? '保存' : '创建服务器'}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
