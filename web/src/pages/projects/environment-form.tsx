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
import { Switch } from '@/components/ui/switch'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { api } from '@/lib/api'
import { DEPLOY_METHODS } from '@/lib/constants'

interface EnvironmentPayload {
  name: string
  branch: string
  build_script: string
  build_output_dir: string
  deploy_server_id: number | null
  deploy_path: string
  deploy_method: string
  post_deploy_script: string
  env_vars: string
  cron_expression: string
  cron_enabled: boolean
  sort_order: number
}

interface EnvironmentDetail extends EnvironmentPayload {
  id: number
  project_id: number
}

interface Server {
  id: number
  name: string
  host: string
}

const DEFAULT_FORM: EnvironmentPayload = {
  name: '',
  branch: 'main',
  build_script: '',
  build_output_dir: '',
  deploy_server_id: null,
  deploy_path: '',
  deploy_method: 'rsync',
  post_deploy_script: '',
  env_vars: '',
  cron_expression: '',
  cron_enabled: false,
  sort_order: 0,
}

// Common cron presets for easy selection
const CRON_PRESETS = [
  { label: '每小时', value: '0 * * * *' },
  { label: '每天 02:00', value: '0 2 * * *' },
  { label: '每天 08:00', value: '0 8 * * *' },
  { label: '工作日 09:00', value: '0 9 * * 1-5' },
  { label: '每周一 03:00', value: '0 3 * * 1' },
  { label: '自定义', value: '' },
]

interface EnvironmentFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  projectId: number
  editEnv?: EnvironmentDetail | null
  onSuccess?: () => void
}

export function EnvironmentFormDialog({
  open,
  onOpenChange,
  projectId,
  editEnv,
  onSuccess,
}: EnvironmentFormDialogProps) {
  const isEdit = !!editEnv

  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [form, setForm] = useState<EnvironmentPayload>(DEFAULT_FORM)
  const [servers, setServers] = useState<Server[]>([])

  useEffect(() => {
    if (!open) {
      setForm(DEFAULT_FORM)
      setError('')
      return
    }

    // Load servers
    api.get<Server[]>('/servers').then((res) => {
      if (res.code === 0 && res.data) {
        setServers(Array.isArray(res.data) ? res.data : [])
      }
    })

    if (isEdit && editEnv) {
      setForm({
        name: editEnv.name || '',
        branch: editEnv.branch || 'main',
        build_script: editEnv.build_script || '',
        build_output_dir: editEnv.build_output_dir || '',
        deploy_server_id: editEnv.deploy_server_id,
        deploy_path: editEnv.deploy_path || '',
        deploy_method: editEnv.deploy_method || 'rsync',
        post_deploy_script: editEnv.post_deploy_script || '',
        env_vars: editEnv.env_vars || '',
        cron_expression: editEnv.cron_expression || '',
        cron_enabled: editEnv.cron_enabled || false,
        sort_order: editEnv.sort_order || 0,
      })
    }
  }, [open, editEnv, isEdit])

  const setField = <K extends keyof EnvironmentPayload>(key: K, value: EnvironmentPayload[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }))
  }

  const validate = () => {
    if (!form.name.trim()) return '请输入环境名称'
    if (form.cron_enabled && !form.cron_expression.trim()) return '启用定时构建时必须填写 Cron 表达式'
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
      if (isEdit && editEnv) {
        const res = await api.put<EnvironmentDetail>(
          `/projects/${projectId}/envs/${editEnv.id}`,
          form,
        )
        if (res.code !== 0) {
          throw new Error(res.message || '更新环境失败')
        }
        toast.success('环境已更新')
      } else {
        const res = await api.post<EnvironmentDetail>(
          `/projects/${projectId}/envs`,
          form,
        )
        if (res.code !== 0) {
          throw new Error(res.message || '创建环境失败')
        }
        toast.success('环境已创建')
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

  const cronPresetValue = CRON_PRESETS.find((p) => p.value === form.cron_expression)?.value ?? ''

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[560px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{isEdit ? '编辑环境' : '新建环境'}</DialogTitle>
          <DialogDescription>
            {isEdit ? '更新环境构建与部署配置' : '为项目创建新的构建环境'}
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400">
              {error}
            </div>
          )}

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="env-name">环境名称 *</Label>
              <Input
                id="env-name"
                value={form.name}
                onChange={(e) => setField('name', e.target.value)}
                placeholder="例如：production"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="env-branch">分支</Label>
              <Input
                id="env-branch"
                value={form.branch}
                onChange={(e) => setField('branch', e.target.value)}
                placeholder="main"
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="env-build-script">构建脚本</Label>
            <Textarea
              id="env-build-script"
              value={form.build_script}
              onChange={(e) => setField('build_script', e.target.value)}
              placeholder="npm install && npm run build"
              rows={3}
              className="font-mono text-sm"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="env-build-output">产物目录</Label>
            <Input
              id="env-build-output"
              value={form.build_output_dir}
              onChange={(e) => setField('build_output_dir', e.target.value)}
              placeholder="dist"
            />
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label>部署方式</Label>
              <Select
                value={form.deploy_method}
                onValueChange={(v) => setField('deploy_method', v)}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {DEPLOY_METHODS.map((m) => (
                    <SelectItem key={m.value} value={m.value}>
                      {m.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>部署服务器</Label>
              <Select
                value={form.deploy_server_id ? String(form.deploy_server_id) : 'none'}
                onValueChange={(v) => setField('deploy_server_id', v === 'none' ? null : Number(v))}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">不部署</SelectItem>
                  {servers.map((s) => (
                    <SelectItem key={s.id} value={String(s.id)}>
                      {s.name} ({s.host})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="env-deploy-path">部署路径</Label>
            <Input
              id="env-deploy-path"
              value={form.deploy_path}
              onChange={(e) => setField('deploy_path', e.target.value)}
              placeholder="/var/www/html"
            />
          </div>

          {/* Cron 定时构建 */}
          <div className="rounded-lg border border-zinc-200 dark:border-zinc-800 p-4 space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium">定时构建</p>
                <p className="text-xs text-zinc-500 mt-0.5">
                  使用 Cron 表达式配置定时自动构建
                </p>
              </div>
              <Switch
                checked={form.cron_enabled}
                onCheckedChange={(checked) => setField('cron_enabled', checked)}
              />
            </div>

            {form.cron_enabled && (
              <div className="space-y-3">
                <div className="space-y-2">
                  <Label>预设</Label>
                  <Select
                    value={cronPresetValue}
                    onValueChange={(v) => {
                      if (v) setField('cron_expression', v)
                    }}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="选择预设或自定义" />
                    </SelectTrigger>
                    <SelectContent>
                      {CRON_PRESETS.map((p) => (
                        <SelectItem key={p.value || 'custom'} value={p.value || 'custom'}>
                          {p.label}{p.value ? ` (${p.value})` : ''}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="env-cron-expr">Cron 表达式</Label>
                  <Input
                    id="env-cron-expr"
                    value={form.cron_expression}
                    onChange={(e) => setField('cron_expression', e.target.value)}
                    placeholder="0 2 * * *"
                    className="font-mono"
                  />
                  <p className="text-xs text-zinc-500">
                    标准 5 段格式：分 时 日 月 周（例如 <code className="bg-zinc-100 dark:bg-zinc-800 px-1 rounded">0 2 * * *</code> 表示每天 02:00）
                  </p>
                </div>
              </div>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="env-env-vars">环境变量 (JSON)</Label>
            <Textarea
              id="env-env-vars"
              value={form.env_vars}
              onChange={(e) => setField('env_vars', e.target.value)}
              placeholder='{"NODE_ENV": "production"}'
              rows={2}
              className="font-mono text-sm"
            />
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              取消
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting ? (isEdit ? '保存中...' : '创建中...') : isEdit ? '保存' : '创建环境'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
