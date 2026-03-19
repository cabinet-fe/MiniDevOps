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
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { api } from '@/lib/api'
import { REPO_AUTH_TYPES, WEBHOOK_TYPES } from '@/lib/constants'

interface ProjectPayload {
  name: string
  description: string
  group_name: string
  tags: string
  repo_url: string
  repo_auth_type: string
  repo_username: string
  repo_password: string
  max_artifacts: number
  webhook_type: string
  webhook_ref_path: string
  webhook_commit_path: string
  webhook_message_path: string
}

interface ProjectDetail extends Omit<ProjectPayload, 'repo_password'> {
  id: number
}

const DEFAULT_FORM: ProjectPayload = {
  name: '',
  description: '',
  group_name: '',
  tags: '',
  repo_url: '',
  repo_auth_type: 'none',
  repo_username: '',
  repo_password: '',
  max_artifacts: 5,
  webhook_type: 'auto',
  webhook_ref_path: '$.ref',
  webhook_commit_path: '$.head_commit.id',
  webhook_message_path: '$.head_commit.message',
}

interface ProjectFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  editId?: number | null
  onSuccess?: () => void
}

export function ProjectFormDialog({
  open,
  onOpenChange,
  editId,
  onSuccess,
}: ProjectFormDialogProps) {
  const isEdit = !!editId

  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [form, setForm] = useState<ProjectPayload>(DEFAULT_FORM)

  useEffect(() => {
    if (!open) {
      setForm(DEFAULT_FORM)
      setError('')
      return
    }

    if (!isEdit || !editId) return

    const fetchProject = async () => {
      setLoading(true)
      try {
        const res = await api.get<ProjectDetail>(`/projects/${editId}`)
        if (res.code !== 0 || !res.data) {
          throw new Error(res.message || '加载项目失败')
        }
        setForm({
          name: res.data.name || '',
          description: res.data.description || '',
          group_name: res.data.group_name || '',
          tags: res.data.tags || '',
          repo_url: res.data.repo_url || '',
          repo_auth_type: res.data.repo_auth_type || 'none',
          repo_username: res.data.repo_username || '',
          repo_password: '',
          max_artifacts: res.data.max_artifacts || 5,
          webhook_type: res.data.webhook_type || 'auto',
          webhook_ref_path: res.data.webhook_ref_path || '$.ref',
          webhook_commit_path: res.data.webhook_commit_path || '$.head_commit.id',
          webhook_message_path: res.data.webhook_message_path || '$.head_commit.message',
        })
      } catch (err) {
        const message = err instanceof Error ? err.message : '加载项目失败'
        setError(message)
        toast.error(message)
      } finally {
        setLoading(false)
      }
    }

    fetchProject()
  }, [open, editId, isEdit])

  const setField = <K extends keyof ProjectPayload>(key: K, value: ProjectPayload[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }))
  }

  const validate = () => {
    if (!form.name.trim()) return '请输入项目名称'
    if (!form.repo_url.trim()) return '请输入仓库地址'
    if (form.max_artifacts < 1) return '构建产物保留数量必须大于 0'
    if (!isEdit && form.repo_auth_type !== 'none' && !form.repo_password.trim()) {
      return '请填写仓库认证信息'
    }
    if (form.webhook_type === 'generic' && !form.webhook_ref_path.trim()) {
      return '通用 Webhook 必须配置 ref JSONPath'
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
      const payload: ProjectPayload = {
        name: form.name.trim(),
        description: form.description.trim(),
        group_name: form.group_name.trim(),
        tags: form.tags.trim(),
        repo_url: form.repo_url.trim(),
        repo_auth_type: form.repo_auth_type,
        repo_username: form.repo_auth_type === 'none' ? '' : form.repo_username.trim(),
        repo_password: form.repo_auth_type === 'none' ? '' : form.repo_password,
        max_artifacts: form.max_artifacts,
        webhook_type: form.webhook_type,
        webhook_ref_path: form.webhook_type === 'generic' ? form.webhook_ref_path.trim() : '',
        webhook_commit_path: form.webhook_type === 'generic' ? form.webhook_commit_path.trim() : '',
        webhook_message_path: form.webhook_type === 'generic' ? form.webhook_message_path.trim() : '',
      }

      if (isEdit && editId) {
        const res = await api.put<ProjectDetail>(`/projects/${editId}`, payload)
        if (res.code !== 0) {
          throw new Error(res.message || '更新项目失败')
        }
        toast.success('项目已更新')
      } else {
        const res = await api.post<ProjectDetail>('/projects', payload)
        if (res.code !== 0 || !res.data) {
          throw new Error(res.message || '创建项目失败')
        }
        toast.success('项目创建成功')
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
      <DialogContent className="sm:max-w-[720px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{isEdit ? '编辑项目' : '新建项目'}</DialogTitle>
          <DialogDescription>
            {isEdit ? '更新项目仓库、分组与 Webhook 配置' : '创建新的构建项目并配置仓库信息'}
          </DialogDescription>
        </DialogHeader>

        {loading ? (
          <div className="flex h-32 items-center justify-center">
            <div className="size-8 animate-spin rounded-full border-2 border-zinc-600 border-t-zinc-300" />
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <div className="rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400">
                {error}
              </div>
            )}

            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="project-name">项目名称 *</Label>
                <Input
                  id="project-name"
                  value={form.name}
                  onChange={(e) => setField('name', e.target.value)}
                  placeholder="例如：buildflow-web"
                  maxLength={100}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="project-max-artifacts">构建产物保留数量 *</Label>
                <Input
                  id="project-max-artifacts"
                  type="number"
                  min={1}
                  value={form.max_artifacts}
                  onChange={(e) =>
                    setField('max_artifacts', Math.max(1, Number(e.target.value) || 1))
                  }
                />
              </div>
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="project-group-name">项目分组</Label>
                <Input
                  id="project-group-name"
                  value={form.group_name}
                  onChange={(e) => setField('group_name', e.target.value)}
                  placeholder="例如：客户端、基础设施"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="project-tags">标签</Label>
                <Input
                  id="project-tags"
                  value={form.tags}
                  onChange={(e) => setField('tags', e.target.value)}
                  placeholder="例如：react, prod, internal"
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="project-repo-url">仓库地址 *</Label>
              <Input
                id="project-repo-url"
                value={form.repo_url}
                onChange={(e) => setField('repo_url', e.target.value)}
                placeholder="https://github.com/org/repo.git"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="project-description">描述</Label>
              <Textarea
                id="project-description"
                value={form.description}
                onChange={(e) => setField('description', e.target.value)}
                placeholder="简要描述该项目用途"
                rows={3}
                maxLength={500}
              />
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label>仓库认证方式</Label>
                <Select
                  value={form.repo_auth_type}
                  onValueChange={(value) => {
                    if (value === 'none') {
                      setForm((prev) => ({
                        ...prev,
                        repo_auth_type: value,
                        repo_username: '',
                        repo_password: '',
                      }))
                      return
                    }
                    setField('repo_auth_type', value)
                  }}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {REPO_AUTH_TYPES.map((type) => (
                      <SelectItem key={type.value} value={type.value}>
                        {type.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label>Webhook 平台</Label>
                <Select value={form.webhook_type} onValueChange={(value) => setField('webhook_type', value)}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {WEBHOOK_TYPES.map((type) => (
                      <SelectItem key={type.value} value={type.value}>
                        {type.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            {form.repo_auth_type !== 'none' && (
              <div className="grid gap-4 sm:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="project-repo-username">用户名</Label>
                  <Input
                    id="project-repo-username"
                    value={form.repo_username}
                    onChange={(e) => setField('repo_username', e.target.value)}
                    placeholder="仓库用户名"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="project-repo-password">{isEdit ? '更新凭证' : '凭证'} *</Label>
                  <Input
                    id="project-repo-password"
                    type="password"
                    value={form.repo_password}
                    onChange={(e) => setField('repo_password', e.target.value)}
                    placeholder={isEdit ? '留空则保持不变' : '输入密码或 Token'}
                  />
                </div>
              </div>
            )}

            {form.webhook_type === 'generic' && (
              <div className="space-y-4 rounded-xl border border-zinc-200 bg-zinc-50/70 p-4 dark:border-zinc-800 dark:bg-zinc-900/40">
                <div>
                  <p className="text-sm font-medium">通用 JSONPath 映射</p>
                  <p className="mt-1 text-xs text-zinc-500">
                    支持 `$.field` 与 `$.list[0].field` 这种路径格式。
                  </p>
                </div>
                <div className="grid gap-4 sm:grid-cols-2">
                  <div className="space-y-2">
                    <Label htmlFor="project-webhook-ref">Ref JSONPath *</Label>
                    <Input
                      id="project-webhook-ref"
                      value={form.webhook_ref_path}
                      onChange={(e) => setField('webhook_ref_path', e.target.value)}
                      placeholder="$.ref"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="project-webhook-commit">Commit JSONPath</Label>
                    <Input
                      id="project-webhook-commit"
                      value={form.webhook_commit_path}
                      onChange={(e) => setField('webhook_commit_path', e.target.value)}
                      placeholder="$.head_commit.id"
                    />
                  </div>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="project-webhook-message">Message JSONPath</Label>
                  <Input
                    id="project-webhook-message"
                    value={form.webhook_message_path}
                    onChange={(e) => setField('webhook_message_path', e.target.value)}
                    placeholder="$.head_commit.message"
                  />
                </div>
              </div>
            )}

            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                取消
              </Button>
              <Button type="submit" disabled={submitting}>
                {submitting ? '提交中...' : isEdit ? '保存修改' : '创建项目'}
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  )
}
