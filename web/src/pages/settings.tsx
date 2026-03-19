import { useState, useRef } from 'react'
import { Download, Upload, FileJson } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { toast } from 'sonner'

export function SettingsPage() {
  const [restoreConfirm, setRestoreConfirm] = useState(false)
  const [restoring, setRestoring] = useState(false)
  const restoreInputRef = useRef<HTMLInputElement>(null)
  const importInputRef = useRef<HTMLInputElement>(null)

  const handleBackup = async () => {
    try {
      const token = localStorage.getItem('access_token')
      const res = await fetch('/api/v1/system/backup', {
        method: 'POST',
        headers: token ? { Authorization: `Bearer ${token}` } : {},
        credentials: 'include',
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
        credentials: 'include',
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
        credentials: 'include',
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

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">系统设置</h1>
        <p className="mt-1 text-sm text-zinc-500">备份、恢复与导入</p>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <Card className="border-zinc-200 dark:border-zinc-800">
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

        <Card className="border-zinc-200 dark:border-zinc-800">
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

        <Card className="border-zinc-200 dark:border-zinc-800 md:col-span-2">
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
    </div>
  )
}
