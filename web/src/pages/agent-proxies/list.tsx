import { useState, useEffect, useCallback } from 'react'
import { Cpu, Download, Loader2, RefreshCw } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { api } from '@/lib/api'
import { toast } from 'sonner'

interface AgentProxy {
  key: string
  name: string
  binary: string
  installed: boolean
  version: string
  path: string
  message: string
}

export function AgentProxyListPage() {
  const [proxies, setProxies] = useState<AgentProxy[]>([])
  const [loading, setLoading] = useState(true)
  const [busyKey, setBusyKey] = useState<string | null>(null)

  const fetchProxies = useCallback(async () => {
    try {
      const res = await api.get<AgentProxy[]>('/agent-proxies')
      if (res.code === 0 && res.data) {
        setProxies(Array.isArray(res.data) ? res.data : [])
      }
    } catch {
      toast.error('加载代理状态失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchProxies()
  }, [fetchProxies])

  const runAction = async (key: string, action: 'install' | 'upgrade') => {
    setBusyKey(key)
    try {
      const res = await api.post<{ proxy: AgentProxy; output: string }>(
        `/agent-proxies/${key}/${action}`,
      )
      if (res.code === 0) {
        toast.success(action === 'install' ? '安装完成' : '更新完成')
        await fetchProxies()
      } else {
        toast.error(res.message || '操作失败')
      }
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : '操作失败'
      toast.error(msg)
    } finally {
      setBusyKey(null)
    }
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0">
          <div>
            <CardTitle className="flex items-center gap-2">
              <Cpu className="size-5" />
              代理管理
            </CardTitle>
            <CardDescription>
              检测本机 CLI（opencode / Claude Code / reasonix），支持安装与更新
            </CardDescription>
          </div>
          <Button variant="outline" onClick={() => { setLoading(true); fetchProxies() }}>
            <RefreshCw className="mr-2 size-4" />
            刷新
          </Button>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex justify-center py-12">
              <Loader2 className="size-6 animate-spin text-muted-foreground" />
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>名称</TableHead>
                  <TableHead>二进制</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>版本</TableHead>
                  <TableHead>路径</TableHead>
                  <TableHead className="w-[180px]">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {proxies.map((p) => (
                  <TableRow key={p.key}>
                    <TableCell className="font-medium">{p.name}</TableCell>
                    <TableCell className="font-mono text-sm">{p.binary}</TableCell>
                    <TableCell>
                      <Badge variant={p.installed ? 'default' : 'outline'}>
                        {p.installed ? '已安装' : '未安装'}
                      </Badge>
                    </TableCell>
                    <TableCell className="font-mono text-sm">{p.version || '—'}</TableCell>
                    <TableCell className="max-w-[240px] truncate font-mono text-xs text-muted-foreground">
                      {p.path || '—'}
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-2">
                        {!p.installed ? (
                          <Button
                            size="sm"
                            disabled={busyKey === p.key}
                            onClick={() => runAction(p.key, 'install')}
                          >
                            {busyKey === p.key ? (
                              <Loader2 className="mr-2 size-4 animate-spin" />
                            ) : (
                              <Download className="mr-2 size-4" />
                            )}
                            安装
                          </Button>
                        ) : (
                          <Button
                            size="sm"
                            variant="outline"
                            disabled={busyKey === p.key}
                            onClick={() => runAction(p.key, 'upgrade')}
                          >
                            {busyKey === p.key ? (
                              <Loader2 className="mr-2 size-4 animate-spin" />
                            ) : (
                              <RefreshCw className="mr-2 size-4" />
                            )}
                            更新
                          </Button>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
