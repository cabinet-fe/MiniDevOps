import { useState, useEffect, useCallback } from 'react'
import { Plus, Pencil, Trash2, Bot, Loader2 } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
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
import { useAuth } from '@/hooks/use-auth'
import { toast } from 'sonner'
import { AGENT_PROXIES } from '@/lib/constants'

interface Agent {
  id: number
  name: string
  prompt: string
  proxy_key: string
  enabled: boolean
  project_ids: number[]
}

interface Project {
  id: number
  name: string
}

interface AgentForm {
  name: string
  prompt: string
  proxy_key: string
  enabled: boolean
  project_ids: number[]
}

const DEFAULT_FORM: AgentForm = {
  name: '',
  prompt: '',
  proxy_key: 'opencode',
  enabled: true,
  project_ids: [],
}

function proxyLabel(key: string) {
  return AGENT_PROXIES.find((p) => p.value === key)?.label || key
}

export function AgentListPage() {
  const { isOps } = useAuth()
  const [agents, setAgents] = useState<Agent[]>([])
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<Agent | null>(null)
  const [form, setForm] = useState<AgentForm>(DEFAULT_FORM)
  const [submitting, setSubmitting] = useState(false)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [deleting, setDeleting] = useState(false)

  const fetchAgents = useCallback(async () => {
    try {
      const res = await api.get<Agent[]>('/agents')
      if (res.code === 0 && res.data) {
        setAgents(Array.isArray(res.data) ? res.data : [])
      }
    } catch {
      toast.error('加载智能体失败')
    } finally {
      setLoading(false)
    }
  }, [])

  const fetchProjects = useCallback(async () => {
    try {
      const res = await api.get<PaginatedData<Project>>('/projects?page=1&page_size=200')
      if (res.code === 0 && res.data) {
        setProjects(res.data.items || [])
      }
    } catch {
      /* ignore */
    }
  }, [])

  useEffect(() => {
    fetchAgents()
    fetchProjects()
  }, [fetchAgents, fetchProjects])

  const openCreate = () => {
    setEditing(null)
    setForm(DEFAULT_FORM)
    setFormOpen(true)
  }

  const openEdit = (agent: Agent) => {
    setEditing(agent)
    setForm({
      name: agent.name,
      prompt: agent.prompt || '',
      proxy_key: agent.proxy_key,
      enabled: agent.enabled,
      project_ids: agent.project_ids || [],
    })
    setFormOpen(true)
  }

  const toggleProject = (id: number) => {
    setForm((prev) => {
      if (prev.project_ids.includes(id)) {
        return { ...prev, project_ids: prev.project_ids.filter((x) => x !== id) }
      }
      return { ...prev, project_ids: [...prev.project_ids, id] }
    })
  }

  const handleSubmit = async () => {
    if (!form.name.trim()) {
      toast.error('请填写名称')
      return
    }
    setSubmitting(true)
    try {
      const payload = {
        name: form.name.trim(),
        prompt: form.prompt,
        proxy_key: form.proxy_key,
        enabled: form.enabled,
        project_ids: form.project_ids,
      }
      const res = editing
        ? await api.put(`/agents/${editing.id}`, payload)
        : await api.post('/agents', payload)
      if (res.code === 0) {
        toast.success(editing ? '已更新' : '已创建')
        setFormOpen(false)
        fetchAgents()
      } else {
        toast.error(res.message || '保存失败')
      }
    } catch {
      toast.error('保存失败')
    } finally {
      setSubmitting(false)
    }
  }

  const handleDelete = async () => {
    if (deleteId == null) return
    setDeleting(true)
    try {
      const res = await api.delete(`/agents/${deleteId}`)
      if (res.code === 0) {
        toast.success('已删除')
        setDeleteId(null)
        fetchAgents()
      } else {
        toast.error(res.message || '删除失败')
      }
    } catch {
      toast.error('删除失败')
    } finally {
      setDeleting(false)
    }
  }

  const projectName = (id: number) => projects.find((p) => p.id === id)?.name || `#${id}`

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0">
          <div>
            <CardTitle className="flex items-center gap-2">
              <Bot className="size-5" />
              智能体管理
            </CardTitle>
            <CardDescription>配置提示词、CLI 代理与项目范围；可在环境中挂载于构建后执行</CardDescription>
          </div>
          {isOps() && (
            <Button onClick={openCreate}>
              <Plus className="mr-2 size-4" />
              新建智能体
            </Button>
          )}
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex justify-center py-12">
              <Loader2 className="size-6 animate-spin text-muted-foreground" />
            </div>
          ) : agents.length === 0 ? (
            <div className="rounded-lg border border-dashed p-8 text-center text-sm text-muted-foreground">
              暂无智能体
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>名称</TableHead>
                  <TableHead>代理</TableHead>
                  <TableHead>项目范围</TableHead>
                  <TableHead>状态</TableHead>
                  {isOps() && <TableHead className="w-[120px]">操作</TableHead>}
                </TableRow>
              </TableHeader>
              <TableBody>
                {agents.map((agent) => (
                  <TableRow key={agent.id}>
                    <TableCell className="font-medium">{agent.name}</TableCell>
                    <TableCell>{proxyLabel(agent.proxy_key)}</TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {(agent.project_ids || []).length === 0 ? (
                          <span className="text-muted-foreground">未选择</span>
                        ) : (
                          agent.project_ids.map((id) => (
                            <Badge key={id} variant="secondary">
                              {projectName(id)}
                            </Badge>
                          ))
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant={agent.enabled ? 'default' : 'outline'}>
                        {agent.enabled ? '启用' : '停用'}
                      </Badge>
                    </TableCell>
                    {isOps() && (
                      <TableCell>
                        <div className="flex gap-1">
                          <Button variant="ghost" size="icon" onClick={() => openEdit(agent)}>
                            <Pencil className="size-4" />
                          </Button>
                          <Button variant="ghost" size="icon" onClick={() => setDeleteId(agent.id)}>
                            <Trash2 className="size-4 text-destructive" />
                          </Button>
                        </div>
                      </TableCell>
                    )}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Dialog open={formOpen} onOpenChange={setFormOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{editing ? '编辑智能体' : '新建智能体'}</DialogTitle>
            <DialogDescription>绑定 CLI 代理与可挂载的项目范围</DialogDescription>
          </DialogHeader>
          <DialogBody className="space-y-4">
            <div className="space-y-2">
              <Label>名称</Label>
              <Input
                value={form.name}
                onChange={(e) => setForm((p) => ({ ...p, name: e.target.value }))}
                placeholder="代码审查"
              />
            </div>
            <div className="space-y-2">
              <Label>代理</Label>
              <Select
                value={form.proxy_key}
                onValueChange={(v) => setForm((p) => ({ ...p, proxy_key: v }))}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {AGENT_PROXIES.map((p) => (
                    <SelectItem key={p.value} value={p.value}>
                      {p.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>提示词</Label>
              <Textarea
                value={form.prompt}
                onChange={(e) => setForm((p) => ({ ...p, prompt: e.target.value }))}
                rows={5}
                placeholder="请审查本次构建产物相关变更…"
              />
            </div>
            <div className="flex items-center justify-between">
              <Label>启用</Label>
              <Switch
                checked={form.enabled}
                onCheckedChange={(v) => setForm((p) => ({ ...p, enabled: v }))}
              />
            </div>
            <div className="space-y-2">
              <Label>项目范围</Label>
              {projects.length === 0 ? (
                <p className="text-sm text-muted-foreground">暂无项目</p>
              ) : (
                <div className="flex max-h-40 flex-wrap gap-2 overflow-y-auto">
                  {projects.map((p) => {
                    const selected = form.project_ids.includes(p.id)
                    return (
                      <Button
                        key={p.id}
                        type="button"
                        size="sm"
                        variant={selected ? 'secondary' : 'outline'}
                        onClick={() => toggleProject(p.id)}
                      >
                        {p.name}
                      </Button>
                    )
                  })}
                </div>
              )}
            </div>
          </DialogBody>
          <DialogFooter>
            <Button variant="outline" onClick={() => setFormOpen(false)}>
              取消
            </Button>
            <Button onClick={handleSubmit} disabled={submitting}>
              {submitting && <Loader2 className="mr-2 size-4 animate-spin" />}
              保存
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={deleteId != null} onOpenChange={(o) => !o && setDeleteId(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>删除智能体</DialogTitle>
            <DialogDescription>删除后环境挂载关系也会清除，确认继续？</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteId(null)}>
              取消
            </Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleting}>
              {deleting && <Loader2 className="mr-2 size-4 animate-spin" />}
              删除
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
