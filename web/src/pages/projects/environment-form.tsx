import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import { Loader2 } from 'lucide-react'
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
import { DEPLOY_METHODS, BUILD_SCRIPT_TYPES } from '@/lib/constants'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import { useTheme } from 'next-themes'
import CodeMirror from '@uiw/react-codemirror'
import { json } from '@codemirror/lang-json'
import { javascript } from '@codemirror/lang-javascript'
import { python } from '@codemirror/lang-python'
import { StreamLanguage } from '@codemirror/language'
import { shell } from '@codemirror/legacy-modes/mode/shell'

interface EnvironmentPayload {
  name: string
  branch: string
  build_script: string
  build_script_type: string
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
  branch: '',
  build_script: '',
  build_script_type: 'bash',
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
  const { theme, systemTheme } = useTheme()
  const currentTheme = theme === 'system' ? systemTheme : theme
  const cmTheme = currentTheme === 'dark' ? 'dark' : 'light'

  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [form, setForm] = useState<EnvironmentPayload>(DEFAULT_FORM)
  const [servers, setServers] = useState<Server[]>([])
  const [branches, setBranches] = useState<string[]>([])
  const [branchesLoading, setBranchesLoading] = useState(false)
  const [branchPopoverOpen, setBranchPopoverOpen] = useState(false)

  useEffect(() => {
    if (!open) {
      setForm(DEFAULT_FORM)
      setError('')
      setBranches([])
      return
    }

    // Load servers
    api.get<Server[]>('/servers').then((res) => {
      if (res.code === 0 && res.data) {
        setServers(Array.isArray(res.data) ? res.data : [])
      }
    })

    // Load branches
    setBranchesLoading(true)
    api.get<string[]>(`/projects/${projectId}/branches`).then((res) => {
      if (res.code === 0 && res.data) {
        setBranches(Array.isArray(res.data) ? res.data : [])
      }
    }).finally(() => setBranchesLoading(false))

    if (isEdit && editEnv) {
      setForm({
        name: editEnv.name || '',
        branch: editEnv.branch || '',
        build_script: editEnv.build_script || '',
        build_script_type: editEnv.build_script_type || 'bash',
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
  }, [open, editEnv, isEdit, projectId])

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

  // Get current script type info for placeholder
  const currentScriptType = BUILD_SCRIPT_TYPES.find((t) => t.value === form.build_script_type)

  const getScriptExtensions = () => {
    switch (form.build_script_type) {
      case 'node': return [javascript()]
      case 'python': return [python()]
      case 'bash':
      default: return [StreamLanguage.define(shell)]
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[800px] max-h-[90vh] overflow-y-auto">
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

          <div className="grid gap-4 sm:grid-cols-3">
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
              <Label>分支</Label>
              <Popover open={branchPopoverOpen} onOpenChange={setBranchPopoverOpen}>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    role="combobox"
                    aria-expanded={branchPopoverOpen}
                    className="w-full justify-between font-normal"
                  >
                    {form.branch || '选择或输入分支'}
                    {branchesLoading && <Loader2 className="size-4 animate-spin ml-2" />}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-[--radix-popover-trigger-width] p-0" align="start">
                  <Command>
                    <CommandInput
                      placeholder="搜索或输入分支名..."
                      value={form.branch}
                      onValueChange={(v: string) => setField('branch', v)}
                    />
                    <CommandList>
                      <CommandEmpty>
                        {branchesLoading ? '加载中...' : '无匹配分支，可直接输入'}
                      </CommandEmpty>
                      <CommandGroup>
                        {branches.map((b) => (
                          <CommandItem
                            key={b}
                            value={b}
                            onSelect={(v: string) => {
                              setField('branch', v)
                              setBranchPopoverOpen(false)
                            }}
                          >
                            {b}
                          </CommandItem>
                        ))}
                      </CommandGroup>
                    </CommandList>
                  </Command>
                </PopoverContent>
              </Popover>
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
          </div>

          <div className="grid gap-4 sm:grid-cols-[160px_1fr]">
            <div className="space-y-2">
              <Label>脚本类型</Label>
              <Select
                value={form.build_script_type}
                onValueChange={(v) => setField('build_script_type', v)}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {BUILD_SCRIPT_TYPES.map((t) => (
                    <SelectItem key={t.value} value={t.value}>
                      {t.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>构建脚本</Label>
              <div className="border border-zinc-200 dark:border-zinc-800 rounded-md overflow-hidden">
                <CodeMirror
                  value={form.build_script}
                  height="160px"
                  theme={cmTheme}
                  extensions={getScriptExtensions()}
                  onChange={(val) => setField('build_script', val)}
                  placeholder={currentScriptType?.placeholder || 'npm install && npm run build'}
                  className="text-sm font-mono"
                />
              </div>
            </div>
          </div>

          <div className="grid gap-4 sm:grid-cols-3">
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
            <div className="space-y-2">
              <Label htmlFor="env-deploy-path">部署路径</Label>
              <Input
                id="env-deploy-path"
                value={form.deploy_path}
                onChange={(e) => setField('deploy_path', e.target.value)}
                placeholder="/var/www/html"
              />
            </div>
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
            <Label>环境变量 (JSON)</Label>
            <div className="border border-zinc-200 dark:border-zinc-800 rounded-md overflow-hidden">
              <CodeMirror
                value={form.env_vars}
                height="100px"
                theme={cmTheme}
                extensions={[json()]}
                onChange={(val) => setField('env_vars', val)}
                placeholder='{"NODE_ENV": "production"}'
                className="text-sm font-mono"
              />
            </div>
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
