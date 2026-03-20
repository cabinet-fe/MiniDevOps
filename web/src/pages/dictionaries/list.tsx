import { useState, useEffect, useCallback } from 'react'
import {
  BookOpen,
  ChevronRight,
  GripVertical,
  Pencil,
  Plus,
  Trash2,
} from 'lucide-react'
import { toast } from 'sonner'
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
import { api } from '@/lib/api'

interface Dictionary {
  id: number
  name: string
  code: string
  description: string
  items?: DictItem[]
}

interface DictItem {
  id: number
  dictionary_id: number
  label: string
  value: string
  sort_order: number
  enabled: boolean
}

export function DictionaryListPage() {
  const [dictionaries, setDictionaries] = useState<Dictionary[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedDict, setSelectedDict] = useState<Dictionary | null>(null)
  const [items, setItems] = useState<DictItem[]>([])
  const [itemsLoading, setItemsLoading] = useState(false)
  const [dictDialogOpen, setDictDialogOpen] = useState(false)
  const [editingDict, setEditingDict] = useState<Dictionary | null>(null)
  const [itemDialogOpen, setItemDialogOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<DictItem | null>(null)

  const fetchDictionaries = useCallback(async () => {
    try {
      const res = await api.get<Dictionary[]>('/dictionaries')
      if (res.code === 0 && res.data) {
        setDictionaries(res.data)
      }
    } catch {
      toast.error('加载字典列表失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchDictionaries()
  }, [fetchDictionaries])

  const fetchItems = useCallback(async (dictId: number) => {
    setItemsLoading(true)
    try {
      const res = await api.get<DictItem[]>(`/dictionaries/${dictId}/items`)
      if (res.code === 0 && res.data) {
        setItems(res.data)
      }
    } catch {
      toast.error('加载字典项失败')
    } finally {
      setItemsLoading(false)
    }
  }, [])

  const selectDict = (dict: Dictionary) => {
    setSelectedDict(dict)
    fetchItems(dict.id)
  }

  const handleDeleteDict = async (id: number) => {
    if (!confirm('确定删除此字典及其所有字典项？')) return
    try {
      const res = await api.delete(`/dictionaries/${id}`)
      if (res.code === 0) {
        toast.success('字典已删除')
        if (selectedDict?.id === id) {
          setSelectedDict(null)
          setItems([])
        }
        fetchDictionaries()
      } else {
        toast.error(res.message || '删除失败')
      }
    } catch {
      toast.error('删除失败')
    }
  }

  const handleDeleteItem = async (itemId: number) => {
    if (!selectedDict) return
    if (!confirm('确定删除此字典项？')) return
    try {
      const res = await api.delete(`/dictionaries/${selectedDict.id}/items/${itemId}`)
      if (res.code === 0) {
        toast.success('字典项已删除')
        fetchItems(selectedDict.id)
      } else {
        toast.error(res.message || '删除失败')
      }
    } catch {
      toast.error('删除失败')
    }
  }

  const handleToggleEnabled = async (item: DictItem) => {
    if (!selectedDict) return
    try {
      const res = await api.put(`/dictionaries/${selectedDict.id}/items/${item.id}`, {
        enabled: !item.enabled,
      })
      if (res.code === 0) {
        fetchItems(selectedDict.id)
      } else {
        toast.error(res.message || '操作失败')
      }
    } catch {
      toast.error('操作失败')
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
          <h1 className="text-2xl font-bold tracking-tight text-foreground">数据字典</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            管理系统中的可复用选项列表，如标签、类别等
          </p>
        </div>
        <Button
          className="gap-2"
          onClick={() => {
            setEditingDict(null)
            setDictDialogOpen(true)
          }}
        >
          <Plus className="size-4" />
          新建字典
        </Button>
      </div>

      <div className="grid gap-6 lg:grid-cols-[320px_1fr]">
        <Card className="border-border">
          <CardHeader className="pb-3">
            <CardTitle className="text-base">字典列表</CardTitle>
            <CardDescription>{dictionaries.length} 个字典</CardDescription>
          </CardHeader>
          <CardContent className="space-y-1 p-3">
            {dictionaries.length === 0 && (
              <p className="px-3 py-8 text-center text-sm text-muted-foreground">
                暂无字典，点击上方按钮新建
              </p>
            )}
            {dictionaries.map((dict) => (
              <div
                key={dict.id}
                className={`group flex cursor-pointer items-center gap-3 rounded-lg px-3 py-2.5 transition-colors ${
                  selectedDict?.id === dict.id
                    ? 'bg-muted'
                    : 'hover:bg-muted/50'
                }`}
                onClick={() => selectDict(dict)}
              >
                <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-muted">
                  <BookOpen className="size-4 text-muted-foreground" />
                </div>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium">{dict.name}</p>
                  <p className="truncate text-xs text-muted-foreground">{dict.code}</p>
                </div>
                <ChevronRight className="size-4 shrink-0 text-muted-foreground" />
              </div>
            ))}
          </CardContent>
        </Card>

        {selectedDict ? (
          <Card className="border-border">
            <CardHeader>
              <div className="flex items-start justify-between">
                <div>
                  <CardTitle className="flex items-center gap-2">
                    {selectedDict.name}
                    <Badge variant="secondary" className="font-mono text-xs">
                      {selectedDict.code}
                    </Badge>
                  </CardTitle>
                  <CardDescription className="mt-1">
                    {selectedDict.description || '暂无描述'}
                  </CardDescription>
                </div>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setEditingDict(selectedDict)
                      setDictDialogOpen(true)
                    }}
                  >
                    <Pencil className="size-4" />
                    编辑
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    className="text-red-500 hover:text-red-600"
                    onClick={() => handleDeleteDict(selectedDict.id)}
                  >
                    <Trash2 className="size-4" />
                    删除
                  </Button>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="mb-4 flex items-center justify-between">
                <p className="text-sm text-muted-foreground">{items.length} 个字典项</p>
                <Button
                  size="sm"
                  onClick={() => {
                    setEditingItem(null)
                    setItemDialogOpen(true)
                  }}
                >
                  <Plus className="size-4" />
                  新增字典项
                </Button>
              </div>
              {itemsLoading ? (
                <div className="flex h-32 items-center justify-center">
                  <div className="border-muted size-6 animate-spin rounded-full border-2 border-t-foreground" />
                </div>
              ) : items.length === 0 ? (
                <div className="flex h-32 items-center justify-center text-sm text-muted-foreground">
                  暂无字典项
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-10" />
                      <TableHead>显示文本</TableHead>
                      <TableHead>存储值</TableHead>
                      <TableHead className="w-20">排序</TableHead>
                      <TableHead className="w-20">状态</TableHead>
                      <TableHead className="w-24">操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {items.map((item) => (
                      <TableRow key={item.id}>
                        <TableCell>
                          <GripVertical className="size-4 text-muted-foreground" />
                        </TableCell>
                        <TableCell className="font-medium">{item.label}</TableCell>
                        <TableCell className="font-mono text-xs text-muted-foreground">
                          {item.value}
                        </TableCell>
                        <TableCell>{item.sort_order}</TableCell>
                        <TableCell>
                          <Switch
                            checked={item.enabled}
                            onCheckedChange={() => handleToggleEnabled(item)}
                            size="sm"
                          />
                        </TableCell>
                        <TableCell>
                          <div className="flex gap-1">
                            <Button
                              variant="ghost"
                              size="icon-sm"
                              onClick={() => {
                                setEditingItem(item)
                                setItemDialogOpen(true)
                              }}
                            >
                              <Pencil className="size-4" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon-sm"
                              className="text-red-500 hover:text-red-600"
                              onClick={() => handleDeleteItem(item.id)}
                            >
                              <Trash2 className="size-4" />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        ) : (
          <Card className="flex min-h-[400px] items-center justify-center border-border">
            <div className="text-center">
              <BookOpen className="mx-auto size-12 text-muted-foreground/50" />
              <p className="mt-4 text-lg font-medium text-muted-foreground">
                选择一个字典查看详情
              </p>
              <p className="mt-1 text-sm text-muted-foreground">
                在左侧列表中选择字典来管理其字典项
              </p>
            </div>
          </Card>
        )}
      </div>

      <DictFormDialog
        open={dictDialogOpen}
        onOpenChange={setDictDialogOpen}
        editDict={editingDict}
        onSuccess={() => {
          fetchDictionaries()
          if (editingDict && selectedDict?.id === editingDict.id) {
            api.get<Dictionary>(`/dictionaries/${editingDict.id}`).then((res) => {
              if (res.code === 0 && res.data) setSelectedDict(res.data)
            })
          }
        }}
      />

      <DictItemFormDialog
        open={itemDialogOpen}
        onOpenChange={setItemDialogOpen}
        dictId={selectedDict?.id ?? 0}
        editItem={editingItem}
        onSuccess={() => {
          if (selectedDict) fetchItems(selectedDict.id)
        }}
      />
    </div>
  )
}

function DictFormDialog({
  open,
  onOpenChange,
  editDict,
  onSuccess,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  editDict: Dictionary | null
  onSuccess: () => void
}) {
  const isEdit = !!editDict
  const [name, setName] = useState('')
  const [code, setCode] = useState('')
  const [description, setDescription] = useState('')
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (open && editDict) {
      setName(editDict.name)
      setCode(editDict.code)
      setDescription(editDict.description)
    } else if (open) {
      setName('')
      setCode('')
      setDescription('')
    }
  }, [open, editDict])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim() || !code.trim()) {
      toast.error('请填写名称和编码')
      return
    }
    setSubmitting(true)
    try {
      const payload = { name: name.trim(), code: code.trim(), description: description.trim() }
      const res = isEdit
        ? await api.put(`/dictionaries/${editDict!.id}`, payload)
        : await api.post('/dictionaries', payload)
      if (res.code === 0) {
        toast.success(isEdit ? '字典已更新' : '字典已创建')
        onOpenChange(false)
        onSuccess()
      } else {
        toast.error(res.message || '操作失败')
      }
    } catch {
      toast.error('操作失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[460px]">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{isEdit ? '编辑字典' : '新建字典'}</DialogTitle>
            <DialogDescription>
              {isEdit ? '修改字典信息' : '创建一个新的数据字典'}
            </DialogDescription>
          </DialogHeader>
          <DialogBody className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="dict-name">名称 *</Label>
              <Input
                id="dict-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="例如：项目标签"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="dict-code">编码 *</Label>
              <Input
                id="dict-code"
                value={code}
                onChange={(e) => setCode(e.target.value)}
                placeholder="例如：project_tags"
                disabled={isEdit}
                className="font-mono"
              />
              {!isEdit && (
                <p className="text-xs text-muted-foreground">唯一标识，创建后不可修改</p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="dict-desc">描述</Label>
              <Input
                id="dict-desc"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="简要描述字典用途"
              />
            </div>
          </DialogBody>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              取消
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting ? '提交中...' : isEdit ? '保存' : '创建'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

function DictItemFormDialog({
  open,
  onOpenChange,
  dictId,
  editItem,
  onSuccess,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  dictId: number
  editItem: DictItem | null
  onSuccess: () => void
}) {
  const isEdit = !!editItem
  const [label, setLabel] = useState('')
  const [value, setValue] = useState('')
  const [sortOrder, setSortOrder] = useState(0)
  const [enabled, setEnabled] = useState(true)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (open && editItem) {
      setLabel(editItem.label)
      setValue(editItem.value)
      setSortOrder(editItem.sort_order)
      setEnabled(editItem.enabled)
    } else if (open) {
      setLabel('')
      setValue('')
      setSortOrder(0)
      setEnabled(true)
    }
  }, [open, editItem])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!label.trim() || !value.trim()) {
      toast.error('请填写显示文本和存储值')
      return
    }
    setSubmitting(true)
    try {
      const payload = {
        label: label.trim(),
        value: value.trim(),
        sort_order: sortOrder,
        enabled,
      }
      const res = isEdit
        ? await api.put(`/dictionaries/${dictId}/items/${editItem!.id}`, payload)
        : await api.post(`/dictionaries/${dictId}/items`, payload)
      if (res.code === 0) {
        toast.success(isEdit ? '字典项已更新' : '字典项已创建')
        onOpenChange(false)
        onSuccess()
      } else {
        toast.error(res.message || '操作失败')
      }
    } catch {
      toast.error('操作失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[460px]">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{isEdit ? '编辑字典项' : '新增字典项'}</DialogTitle>
            <DialogDescription>
              {isEdit ? '修改字典项信息' : '向当前字典添加一个选项'}
            </DialogDescription>
          </DialogHeader>
          <DialogBody className="space-y-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="item-label">显示文本 *</Label>
                <Input
                  id="item-label"
                  value={label}
                  onChange={(e) => setLabel(e.target.value)}
                  placeholder="例如：生产环境"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="item-value">存储值 *</Label>
                <Input
                  id="item-value"
                  value={value}
                  onChange={(e) => setValue(e.target.value)}
                  placeholder="例如：prod"
                  className="font-mono"
                />
              </div>
            </div>
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="item-sort">排序</Label>
                <Input
                  id="item-sort"
                  type="number"
                  value={sortOrder}
                  onChange={(e) => setSortOrder(Number(e.target.value) || 0)}
                />
              </div>
              <div className="flex items-center gap-3 pt-6">
                <Switch checked={enabled} onCheckedChange={setEnabled} />
                <Label>启用</Label>
              </div>
            </div>
          </DialogBody>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              取消
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting ? '提交中...' : isEdit ? '保存' : '创建'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
