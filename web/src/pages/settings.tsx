import { useState, useRef, useEffect } from 'react'
import { Download, Upload, FileJson, HardDrive, Trash2, Database, FolderArchive, Pencil, Plus } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
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
import { toast } from 'sonner'
import { api } from '@/lib/api'

interface WorkspaceInfo {
  project_id: number
  project_name: string
  workspace_size: number
  cache_size: number
}

interface VarGroupItem {
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
  items: VarGroupItem[]
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`
}

export function SettingsPage() {
  const [restoreConfirm, setRestoreConfirm] = useState(false)
  const [restoring, setRestoring] = useState(false)
  const restoreInputRef = useRef<HTMLInputElement>(null)
  const importInputRef = useRef<HTMLInputElement>(null)

  const [workspaces, setWorkspaces] = useState<WorkspaceInfo[]>([])
  const [loadingWorkspaces, setLoadingWorkspaces] = useState(false)
  const [cleaningId, setCleaningId] = useState<number | null>(null)
  const [cleanType, setCleanType] = useState<'workspace' | 'cache' | null>(null)
  const [varGroups, setVarGroups] = useState<VarGroup[]>([])
  const [varGroupDialogOpen, setVarGroupDialogOpen] = useState(false)
  const [varGroupSubmitting, setVarGroupSubmitting] = useState(false)
  const [editingVarGroupId, setEditingVarGroupId] = useState<number | null>(null)
  const [varGroupForm, setVarGroupForm] = useState<{
    name: string
    description: string
    items: VarGroupItem[]
  }>({ name: '', description: '', items: [] })

  const loadWorkspaces = async () => {
    setLoadingWorkspaces(true)
    try {
      const res = await api.get<WorkspaceInfo[]>('/system/workspaces')
      if (res.code === 0 && res.data) {
        setWorkspaces(Array.isArray(res.data) ? res.data : [])
      }
    } catch {
      // ignore
    } finally {
      setLoadingWorkspaces(false)
    }
  }

  const loadVarGroups = async () => {
    try {
      const res = await api.get<VarGroup[]>('/var-groups')
      if (res.code === 0 && res.data) {
        setVarGroups(Array.isArray(res.data) ? res.data : [])
      }
    } catch {
      // ignore
    }
  }

  useEffect(() => {
    loadWorkspaces()
    loadVarGroups()
  }, [])

  const handleCleanWorkspace = async (projectId: number) => {
    setCleaningId(projectId)
    setCleanType('workspace')
    try {
      const res = await api.delete(`/system/workspaces/${projectId}`)
      if (res.code === 0) {
        toast.success('工作空间已清理')
        loadWorkspaces()
      } else {
        toast.error(res.message || '清理失败')
      }
    } catch {
      toast.error('清理失败')
    } finally {
      setCleaningId(null)
      setCleanType(null)
    }
  }

  const handleCleanCache = async (projectId: number) => {
    setCleaningId(projectId)
    setCleanType('cache')
    try {
      const res = await api.delete(`/system/caches/${projectId}`)
      if (res.code === 0) {
        toast.success('构建缓存已清理')
        loadWorkspaces()
      } else {
        toast.error(res.message || '清理失败')
      }
    } catch {
      toast.error('清理失败')
    } finally {
      setCleaningId(null)
      setCleanType(null)
    }
  }

  const totalWorkspaceSize = workspaces.reduce((sum, w) => sum + w.workspace_size, 0)
  const totalCacheSize = workspaces.reduce((sum, w) => sum + w.cache_size, 0)

  const handleBackup = async () => {
    try {
      const token = localStorage.getItem('access_token')
      const res = await fetch('/api/v1/system/backup', {
        method: 'POST',
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      })
      if (!res.ok) {
        const data = await res.json().catch(() => ({}))
        toast.error(data.message || '导出失败')
        return
      }
      const blob = await res.blob()
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `buildflow-backup-${new Date().toISOString().slice(0, 10)}.tar.gz`
      a.click()
      URL.revokeObjectURL(url)
      toast.success('备份已下载')
    } catch {
      toast.error('导出失败')
    }
  }

  const handleRestore = async () => {
    const file = restoreInputRef.current?.files?.[0]
    if (!file) {
      toast.error('请选择备份文件')
      return
    }
    setRestoring(true)
    try {
      const formData = new FormData()
      formData.append('file', file)
      const token = localStorage.getItem('access_token')
      const res = await fetch('/api/v1/system/restore', {
        method: 'POST',
        headers: token ? { Authorization: `Bearer ${token}` } : {},
        body: formData,
      })
      const data = await res.json().catch(() => ({}))
      if (data.code === 0) {
        toast.success('恢复完成，请重启服务')
        setRestoreConfirm(false)
      } else {
        toast.error(data.message || '恢复失败')
      }
    } catch {
      toast.error('恢复失败')
    } finally {
      setRestoring(false)
    }
  }

  const handleImport = async () => {
    const file = importInputRef.current?.files?.[0]
    if (!file) {
      toast.error('请选择导入文件')
      return
    }
    try {
      const formData = new FormData()
      formData.append('file', file)
      const token = localStorage.getItem('access_token')
      const res = await fetch('/api/v1/projects/import', {
        method: 'POST',
        headers: token ? { Authorization: `Bearer ${token}` } : {},
        body: formData,
      })
      const data = await res.json().catch(() => ({}))
      if (data.code === 0) {
        toast.success('项目已导入')
      } else {
        toast.error(data.message || '导入失败')
      }
    } catch {
      toast.error('导入失败')
    }
  }

  const openCreateVarGroup = () => {
    setEditingVarGroupId(null)
    setVarGroupForm({ name: '', description: '', items: [] })
    setVarGroupDialogOpen(true)
  }

  const openEditVarGroup = (group: VarGroup) => {
    setEditingVarGroupId(group.id)
    setVarGroupForm({
      name: group.name,
      description: group.description || '',
      items: (group.items || []).map((item) => ({
        id: item.id,
        key: item.key,
        value: '',
        is_secret: item.is_secret,
        keep_value: item.is_secret,
      })),
    })
    setVarGroupDialogOpen(true)
  }

  const updateVarGroupItem = (index: number, patch: Partial<VarGroupItem>) => {
    setVarGroupForm((prev) => ({
      ...prev,
      items: prev.items.map((item, current) => current === index ? { ...item, ...patch } : item),
    }))
  }

  const saveVarGroup = async () => {
    if (!varGroupForm.name.trim()) {
      toast.error('变量组名称不能为空')
      return
    }
    if (varGroupForm.items.some((item) => !item.key.trim())) {
      toast.error('变量项 key 不能为空')
      return
    }
    setVarGroupSubmitting(true)
    try {
      const payload = {
        name: varGroupForm.name.trim(),
        description: varGroupForm.description.trim(),
        items: varGroupForm.items.map((item) => ({
          id: item.id,
          key: item.key.trim(),
          value: item.value,
          is_secret: item.is_secret,
          keep_value: item.keep_value ?? false,
        })),
      }
      const res = editingVarGroupId
        ? await api.put(`/var-groups/${editingVarGroupId}`, payload)
        : await api.post('/var-groups', payload)
      if (res.code !== 0) {
        throw new Error(res.message || '保存变量组失败')
      }
      toast.success(editingVarGroupId ? '变量组已更新' : '变量组已创建')
      setVarGroupDialogOpen(false)
      loadVarGroups()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '保存变量组失败')
    } finally {
      setVarGroupSubmitting(false)
    }
  }

  const deleteVarGroup = async (groupId: number) => {
    try {
      const res = await api.delete(`/var-groups/${groupId}`)
      if (res.code !== 0) {
        throw new Error(res.message || '删除变量组失败')
      }
      toast.success('变量组已删除')
      loadVarGroups()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '删除变量组失败')
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-foreground">系统设置</h1>
        <p className="mt-1 text-sm text-muted-foreground">备份、恢复、导入与存储管理</p>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <Card className="border-border">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Download className="size-5" />
              系统备份
            </CardTitle>
            <CardDescription>
              导出数据库和配置文件为 tar.gz 压缩包
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button onClick={handleBackup}>
              <Download className="size-4" />
              导出备份
            </Button>
          </CardContent>
        </Card>

        <Card className="border-border">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Upload className="size-5" />
              系统恢复
            </CardTitle>
            <CardDescription>
              从备份文件恢复系统数据，恢复后需重启服务
            </CardDescription>
          </CardHeader>
          <CardContent>
            <input
              ref={restoreInputRef}
              type="file"
              accept=".tar.gz,.gz"
              className="hidden"
              onChange={() => setRestoreConfirm(true)}
            />
            <Button variant="outline" onClick={() => restoreInputRef.current?.click()}>
              <Upload className="size-4" />
              选择备份文件
            </Button>
          </CardContent>
        </Card>

        <Card className="border-border md:col-span-2">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <FileJson className="size-5" />
              项目导入
            </CardTitle>
            <CardDescription>
              从 JSON 文件导入项目配置
            </CardDescription>
          </CardHeader>
          <CardContent>
            <input
              ref={importInputRef}
              type="file"
              accept=".json"
              className="hidden"
              onChange={() => importInputRef.current?.files?.[0] && handleImport()}
            />
            <Button variant="outline" onClick={() => importInputRef.current?.click()}>
              <FileJson className="size-4" />
              选择 JSON 文件
            </Button>
          </CardContent>
        </Card>
      </div>

      <Card className="border-border">
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle>变量组</CardTitle>
            <CardDescription>维护可复用的全局变量组，环境可直接关联使用。</CardDescription>
          </div>
          <Button onClick={openCreateVarGroup}>
            <Plus className="size-4" />
            新建变量组
          </Button>
        </CardHeader>
        <CardContent>
          {varGroups.length === 0 ? (
            <div className="rounded-lg border border-dashed border-border p-6 text-sm text-muted-foreground">
              暂无变量组
            </div>
          ) : (
            <div className="space-y-3">
              {varGroups.map((group) => (
                <div key={group.id} className="flex items-center justify-between rounded-lg border border-border p-4">
                  <div>
                    <p className="font-medium">{group.name}</p>
                    <p className="mt-1 text-sm text-muted-foreground">{group.description || '暂无描述'}</p>
                    <p className="mt-2 text-xs text-muted-foreground">{group.items?.length ?? 0} 个变量项</p>
                  </div>
                  <div className="flex gap-2">
                    <Button variant="outline" size="sm" onClick={() => openEditVarGroup(group)}>
                      <Pencil className="size-4" />
                      编辑
                    </Button>
                    <Button variant="ghost" size="sm" onClick={() => deleteVarGroup(group.id)}>
                      <Trash2 className="size-4 text-muted-foreground" />
                      删除
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Storage Management Section */}
      <div>
        <h2 className="text-lg font-semibold tracking-tight mb-4 flex items-center gap-2">
          <HardDrive className="size-5" />
          存储管理
        </h2>

        <div className="grid gap-4 md:grid-cols-2 mb-4">
          <Card className="border-border">
            <CardContent className="pt-6">
              <div className="flex items-center gap-3">
                <div className="flex items-center justify-center size-10 rounded-lg bg-blue-500/10 text-blue-500">
                  <Database className="size-5" />
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">工作空间总占用</p>
                  <p className="text-xl font-semibold">{formatBytes(totalWorkspaceSize)}</p>
                </div>
              </div>
            </CardContent>
          </Card>
          <Card className="border-border">
            <CardContent className="pt-6">
              <div className="flex items-center gap-3">
                <div className="flex items-center justify-center size-10 rounded-lg bg-amber-500/10 text-amber-500">
                  <FolderArchive className="size-5" />
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">构建缓存总占用</p>
                  <p className="text-xl font-semibold">{formatBytes(totalCacheSize)}</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        <Card className="border-border">
          <CardHeader>
            <CardTitle className="text-base">项目存储详情</CardTitle>
            <CardDescription>
              各项目的工作空间和构建缓存占用情况
            </CardDescription>
          </CardHeader>
          <CardContent>
            {loadingWorkspaces ? (
              <div className="flex items-center justify-center py-8">
                <div className="border-muted size-6 animate-spin rounded-full border-2 border-t-foreground" />
              </div>
            ) : workspaces.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground text-sm">
                暂无项目存储数据
              </div>
            ) : (
              <div className="divide-y divide-border">
                {workspaces.map((w) => (
                  <div key={w.project_id} className="flex items-center justify-between py-3">
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium truncate">
                        {w.project_name || `项目 #${w.project_id}`}
                      </p>
                      <div className="flex gap-4 mt-1 text-xs text-muted-foreground">
                        <span>工作空间: {formatBytes(w.workspace_size)}</span>
                        <span>缓存: {formatBytes(w.cache_size)}</span>
                      </div>
                    </div>
                    <div className="flex gap-2 ml-4">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleCleanWorkspace(w.project_id)}
                        disabled={cleaningId === w.project_id}
                        title="清理工作空间"
                      >
                        <Trash2 className="size-4 text-muted-foreground" />
                        {cleaningId === w.project_id && cleanType === 'workspace' ? '清理中...' : '清理工作空间'}
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleCleanCache(w.project_id)}
                        disabled={cleaningId === w.project_id}
                        title="清理缓存"
                      >
                        <Trash2 className="size-4 text-muted-foreground" />
                        {cleaningId === w.project_id && cleanType === 'cache' ? '清理中...' : '清理缓存'}
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            )}
            <div className="mt-4 pt-4 border-t border-border">
              <Button variant="outline" size="sm" onClick={loadWorkspaces} disabled={loadingWorkspaces}>
                刷新
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>

      <Dialog open={restoreConfirm} onOpenChange={setRestoreConfirm}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认恢复</DialogTitle>
            <DialogDescription>
              恢复将覆盖当前数据。确定要执行恢复吗？恢复完成后请重启服务。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRestoreConfirm(false)}>取消</Button>
            <Button variant="destructive" onClick={handleRestore} disabled={restoring}>
              {restoring ? '恢复中...' : '确认恢复'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={varGroupDialogOpen} onOpenChange={setVarGroupDialogOpen}>
        <DialogContent className="sm:max-w-[760px]">
          <DialogHeader>
            <DialogTitle>{editingVarGroupId ? '编辑变量组' : '新建变量组'}</DialogTitle>
            <DialogDescription>变量组中的密文项会加密存储，环境关联后可直接参与构建。</DialogDescription>
          </DialogHeader>
          <DialogBody className="space-y-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="var-group-name">名称</Label>
                <Input
                  id="var-group-name"
                  value={varGroupForm.name}
                  onChange={(e) => setVarGroupForm((prev) => ({ ...prev, name: e.target.value }))}
                  placeholder="例如：frontend-common"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="var-group-description">描述</Label>
                <Input
                  id="var-group-description"
                  value={varGroupForm.description}
                  onChange={(e) => setVarGroupForm((prev) => ({ ...prev, description: e.target.value }))}
                  placeholder="描述此变量组的用途"
                />
              </div>
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <Label>变量项</Label>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setVarGroupForm((prev) => ({
                    ...prev,
                    items: [...prev.items, { key: '', value: '', is_secret: false }],
                  }))}
                >
                  <Plus className="size-4" />
                  添加变量
                </Button>
              </div>
              {varGroupForm.items.length === 0 ? (
                <div className="rounded-lg border border-dashed border-border p-4 text-sm text-muted-foreground">
                  暂无变量项
                </div>
              ) : (
                <div className="space-y-3">
                  {varGroupForm.items.map((item, index) => (
                    <div key={`${item.id ?? 'new'}-${index}`} className="rounded-lg border border-border p-3">
                      <div className="grid gap-3 sm:grid-cols-[1fr_1.3fr_auto_auto]">
                        <Input
                          value={item.key}
                          onChange={(e) => updateVarGroupItem(index, { key: e.target.value })}
                          placeholder="变量名"
                        />
                        <Input
                          type={item.is_secret ? 'password' : 'text'}
                          value={item.value}
                          onChange={(e) => updateVarGroupItem(index, { value: e.target.value, keep_value: false })}
                          placeholder={item.keep_value ? '已存储密文，留空则保持不变' : '变量值'}
                        />
                        <div className="flex items-center gap-2 rounded-md border border-border px-3">
                          <Switch
                            checked={item.is_secret}
                            onCheckedChange={(checked) => updateVarGroupItem(index, {
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
                          onClick={() => setVarGroupForm((prev) => ({
                            ...prev,
                            items: prev.items.filter((_, current) => current !== index),
                          }))}
                        >
                          <Trash2 className="size-4" />
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </DialogBody>
          <DialogFooter>
            <Button variant="outline" onClick={() => setVarGroupDialogOpen(false)}>取消</Button>
            <Button onClick={saveVarGroup} disabled={varGroupSubmitting}>
              {varGroupSubmitting ? '保存中...' : editingVarGroupId ? '保存修改' : '创建变量组'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
