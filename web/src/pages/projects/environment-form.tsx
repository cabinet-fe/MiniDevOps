import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import { Loader2, Plus, Trash2 } from 'lucide-react'
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
  DialogBody,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { api } from '@/lib/api'
import type { PaginatedData } from '@/lib/api'
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
import { javascript } from '@codemirror/lang-javascript'
import { python } from '@codemirror/lang-python'
import { StreamLanguage } from '@codemirror/language'
import { shell } from '@codemirror/legacy-modes/mode/shell'

interface EnvVarRow {
  id?: number
  key: string
  value: string
  is_secret: boolean
  keep_value?: boolean
}

interface VarGroup {
  id: number
  name: string
  description: string
}

export interface EnvironmentPayload {
  name: string
  branch: string
  build_script: string
  build_script_type: string
  build_output_dir: string
  deploy_server_id: number | null
  deploy_path: string
  deploy_method: string
  post_deploy_script: string
  cache_paths?: string
  cron_expression: string
  cron_enabled: boolean
  sort_order: number
  var_group_ids: number[]
}

export interface EnvironmentDetail extends EnvironmentPayload {
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
  cache_paths: '',
  cron_expression: '',
  cron_enabled: false,
  sort_order: 0,
  var_group_ids: [],
}

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
  const [varGroups, setVarGroups] = useState<VarGroup[]>([])
  const [envVars, setEnvVars] = useState<EnvVarRow[]>([])
  const [initialEnvVars, setInitialEnvVars] = useState<EnvVarRow[]>([])

  useEffect(() => {
    if (!open) {
      setForm(DEFAULT_FORM)
      setError('')
      setBranches([])
      setEnvVars([])
      setInitialEnvVars([])
      setVarGroups([])
      return
    }

    api.get<PaginatedData<Server>>('/servers?page=1&page_size=100').then((res) => {
      if (res.code === 0 && res.data) {
        setServers(Array.isArray(res.data.items) ? res.data.items : [])
      }
    })

    api.get<VarGroup[]>('/var-groups').then((res) => {
      if (res.code === 0 && res.data) {
        setVarGroups(Array.isArray(res.data) ? res.data : [])
      }
    })

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
        cache_paths: editEnv.cache_paths || '',
        cron_expression: editEnv.cron_expression || '',
        cron_enabled: editEnv.cron_enabled || false,
        sort_order: editEnv.sort_order || 0,
        var_group_ids: editEnv.var_group_ids || [],
      })

      api.get<EnvVarRow[]>(`/projects/${projectId}/envs/${editEnv.id}/vars`).then((res) => {
        if (res.code !== 0 || !res.data) return
        const rows = (Array.isArray(res.data) ? res.data : []).map((item) => ({
          id: item.id,
          key: item.key,
          value: '',
          is_secret: item.is_secret,
          keep_value: item.is_secret,
        }))
        setEnvVars(rows)
        setInitialEnvVars(rows)
      })
    } else {
      setEnvVars([])
      setInitialEnvVars([])
    }
  }, [open, editEnv, isEdit, projectId])

  const setField = <K extends keyof EnvironmentPayload>(key: K, value: EnvironmentPayload[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }))
  }

  const updateEnvVar = (index: number, patch: Partial<EnvVarRow>) => {
    setEnvVars((prev) => prev.map((item, current) => {
      if (current !== index) return item
      return { ...item, ...patch }
    }))
  }

  const validate = () => {
    if (!form.name.trim()) return '请输入环境名称'
    if (form.cron_enabled && !form.cron_expression.trim()) return '启用定时构建时必须填写 Cron 表达式'
    if (envVars.some((item) => !item.key.trim())) return '环境变量 key 不能为空'
    return ''
  }

  const syncEnvVars = async (envId: number) => {
    const current = envVars.filter((item) => item.key.trim())
    const initialIds = new Set(initialEnvVars.map((item) => item.id).filter(Boolean))
    const currentIds = new Set(current.map((item) => item.id).filter(Boolean))

    for (const initial of initialEnvVars) {
      if (initial.id && !currentIds.has(initial.id)) {
        await api.delete(`/projects/${projectId}/envs/${envId}/vars/${initial.id}`)
      }
    }

    for (const item of current) {
      if (item.id) {
        const res = await api.put(
          `/projects/${projectId}/envs/${envId}/vars/${item.id}`,
          {
            key: item.key.trim(),
            value: item.value,
            is_secret: item.is_secret,
            keep_value: item.keep_value ?? false,
          },
        )
        if (res.code !== 0) {
          throw new Error(res.message || `更新变量 ${item.key} 失败`)
        }
        initialIds.delete(item.id)
      } else {
        const res = await api.post(
          `/projects/${projectId}/envs/${envId}/vars`,
          {
            key: item.key.trim(),
            value: item.value,
            is_secret: item.is_secret,
          },
        )
        if (res.code !== 0) {
          throw new Error(res.message || `创建变量 ${item.key} 失败`)
        }
      }
    }
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
      const payload = {
        ...form,
        name: form.name.trim(),
        branch: form.branch.trim(),
        build_output_dir: form.build_output_dir.trim(),
        deploy_path: form.deploy_path.trim(),
        cache_paths: (form.cache_paths ?? '').trim(),
      }

      let envId = editEnv?.id

      if (isEdit && editEnv) {
        const res = await api.put<EnvironmentDetail>(
          `/projects/${projectId}/envs/${editEnv.id}`,
          payload,
        )
        if (res.code !== 0) {
          throw new Error(res.message || '更新环境失败')
        }
        envId = editEnv.id
        toast.success('环境已更新')
      } else {
        const res = await api.post<EnvironmentDetail>(
          `/projects/${projectId}/envs`,
          payload,
        )
        if (res.code !== 0 || !res.data) {
          throw new Error(res.message || '创建环境失败')
        }
        envId = res.data.id
        toast.success('环境已创建')
      }

      if (!envId) {
        throw new Error('环境 ID 缺失')
      }
      await syncEnvVars(envId)

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

  const cronPresetValue = CRON_PRESETS.find((item) => item.value === form.cron_expression)?.value ?? ''
  const currentScriptType = BUILD_SCRIPT_TYPES.find((item) => item.value === form.build_script_type)

  const getScriptExtensions = () => {
    switch (form.build_script_type) {
      case 'node':
        return [javascript()]
      case 'python':
        return [python()]
      case 'bash':
      default:
        return [StreamLanguage.define(shell)]
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[860px]">
        <form onSubmit={handleSubmit} className="flex min-h-0 flex-1 flex-col">
          <DialogHeader>
            <DialogTitle>{isEdit ? '编辑环境' : '新建环境'}</DialogTitle>
            <DialogDescription>
              {isEdit ? '更新环境构建、部署与变量配置' : '为项目创建新的构建环境'}
            </DialogDescription>
          </DialogHeader>

          <DialogBody className="space-y-4">
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
                      {branchesLoading && <Loader2 className="ml-2 size-4 animate-spin" />}
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent className="w-[--radix-popover-trigger-width] p-0" align="start">
                    <Command>
                      <CommandInput
                        placeholder="搜索或输入分支名..."
                        value={form.branch}
                        onValueChange={(value: string) => setField('branch', value)}
                      />
                      <CommandList>
                        <CommandEmpty>{branchesLoading ? '加载中...' : '无匹配分支，可直接输入'}</CommandEmpty>
                        <CommandGroup>
                          {branches.map((branch) => (
                            <CommandItem
                              key={branch}
                              value={branch}
                              onSelect={(value: string) => {
                                setField('branch', value)
                                setBranchPopoverOpen(false)
                              }}
                            >
                              {branch}
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
                <Select value={form.build_script_type} onValueChange={(value) => setField('build_script_type', value)}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {BUILD_SCRIPT_TYPES.map((item) => (
                      <SelectItem key={item.value} value={item.value}>
                        {item.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>构建脚本</Label>
                <div className="overflow-hidden rounded-md border border-border">
                  <CodeMirror
                    value={form.build_script}
                    height="160px"
                    theme={cmTheme}
                    extensions={getScriptExtensions()}
                    onChange={(value) => setField('build_script', value)}
                    placeholder={currentScriptType?.placeholder || 'npm install && npm run build'}
                    className="text-sm font-mono [&_.cm-editor]:!bg-transparent [&_.cm-gutters]:!bg-transparent"
                  />
                </div>
              </div>
            </div>

            <div className="grid gap-4 sm:grid-cols-3">
              <div className="space-y-2">
                <Label>部署方式</Label>
                <Select value={form.deploy_method} onValueChange={(value) => setField('deploy_method', value)}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {DEPLOY_METHODS.map((method) => (
                      <SelectItem key={method.value} value={method.value}>
                        {method.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>部署服务器</Label>
                <Select
                  value={form.deploy_server_id ? String(form.deploy_server_id) : 'none'}
                  onValueChange={(value) => setField('deploy_server_id', value === 'none' ? null : Number(value))}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="none">不部署</SelectItem>
                    {servers.map((server) => (
                      <SelectItem key={server.id} value={String(server.id)}>
                        {server.name} ({server.host})
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

            <div className="rounded-lg border border-border p-4 space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium">定时构建</p>
                  <p className="mt-0.5 text-xs text-muted-foreground">使用 Cron 表达式配置定时自动构建</p>
                </div>
                <Switch checked={form.cron_enabled} onCheckedChange={(checked) => setField('cron_enabled', checked)} />
              </div>

              {form.cron_enabled && (
                <div className="space-y-3">
                  <div className="space-y-2">
                    <Label>预设</Label>
                    <Select
                      value={cronPresetValue}
                      onValueChange={(value) => {
                        if (value && value !== 'custom') setField('cron_expression', value)
                      }}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="选择预设或自定义" />
                      </SelectTrigger>
                      <SelectContent>
                        {CRON_PRESETS.map((item) => (
                          <SelectItem key={item.value || 'custom'} value={item.value || 'custom'}>
                            {item.label}{item.value ? ` (${item.value})` : ''}
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
                  </div>
                </div>
              )}
            </div>

            <div className="space-y-2">
              <Label>变量组</Label>
              {varGroups.length === 0 ? (
                <div className="rounded-lg border border-dashed border-border p-4 text-sm text-muted-foreground">
                  暂无可用变量组，可在系统设置中创建。
                </div>
              ) : (
                <div className="flex flex-wrap gap-2">
                  {varGroups.map((group) => {
                    const selected = form.var_group_ids.includes(group.id)
                    return (
                      <Button
                        key={group.id}
                        type="button"
                        variant={selected ? 'secondary' : 'outline'}
                        size="sm"
                        onClick={() => {
                          if (selected) {
                            setField('var_group_ids', form.var_group_ids.filter((id) => id !== group.id))
                            return
                          }
                          setField('var_group_ids', [...form.var_group_ids, group.id])
                        }}
                      >
                        {group.name}
                      </Button>
                    )
                  })}
                </div>
              )}
            </div>

            <div className="space-y-3 rounded-lg border border-border p-4">
              <div className="flex items-center justify-between">
                <div>
                  <Label>环境变量</Label>
                  <p className="mt-1 text-xs text-muted-foreground">键值对形式维护，支持按变量逐个加密。</p>
                </div>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setEnvVars((prev) => [...prev, { key: '', value: '', is_secret: false }])}
                >
                  <Plus className="size-4" />
                  添加变量
                </Button>
              </div>
              {envVars.length === 0 ? (
                <div className="rounded-lg border border-dashed border-border p-4 text-sm text-muted-foreground">
                  暂未配置环境变量
                </div>
              ) : (
                <div className="space-y-3">
                  {envVars.map((item, index) => (
                    <div key={`${item.id ?? 'new'}-${index}`} className="rounded-lg border border-border p-3">
                      <div className="grid gap-3 sm:grid-cols-[1fr_1.4fr_auto_auto]">
                        <Input
                          value={item.key}
                          onChange={(e) => updateEnvVar(index, { key: e.target.value })}
                          placeholder="变量名，例如 NODE_ENV"
                        />
                        <Input
                          type={item.is_secret ? 'password' : 'text'}
                          value={item.value}
                          onChange={(e) => updateEnvVar(index, { value: e.target.value, keep_value: false })}
                          placeholder={item.keep_value ? '已存储密文，留空则保持不变' : '变量值'}
                        />
                        <div className="flex items-center gap-2 rounded-md border border-border px-3">
                          <Switch
                            checked={item.is_secret}
                            onCheckedChange={(checked) => updateEnvVar(index, {
                              is_secret: checked,
                              keep_value: checked ? item.keep_value : false,
                              value: checked ? item.value : '',
                            })}
                          />
                          <span className="text-sm text-muted-foreground">加密</span>
                        </div>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon-sm"
                          onClick={() => setEnvVars((prev) => prev.filter((_, current) => current !== index))}
                        >
                          <Trash2 className="size-4" />
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            <div className="space-y-2">
              <Label>构建缓存路径</Label>
              <div className="overflow-hidden rounded-md border border-border">
                <CodeMirror
                  value={form.cache_paths ?? ''}
                  height="80px"
                  theme={cmTheme}
                  extensions={[StreamLanguage.define(shell)]}
                  onChange={(value) => setField('cache_paths', value)}
                  placeholder={'node_modules\n.npm\nvendor'}
                  className="text-sm font-mono [&_.cm-editor]:!bg-transparent [&_.cm-gutters]:!bg-transparent"
                />
              </div>
              <p className="text-xs text-muted-foreground">每行一个路径，构建后缓存这些目录，下次构建时恢复以加速依赖安装。</p>
            </div>
          </DialogBody>

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
