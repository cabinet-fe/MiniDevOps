import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router'
import { Rocket, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useAuthStore } from '@/stores/auth-store'
import { cn } from '@/lib/utils'

export function LoginPage() {
  const navigate = useNavigate()
  const { login, isAuthenticated, fetchMe, token } = useAuthStore()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (isAuthenticated) {
      navigate('/', { replace: true })
      return
    }

    if (!token) {
      return
    }

    let active = true
    fetchMe().then(() => {
      if (active && useAuthStore.getState().isAuthenticated) {
        navigate('/', { replace: true })
      }
    })
    return () => {
      active = false
    }
  }, [isAuthenticated, navigate, fetchMe, token])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await login(username, password)
      navigate('/', { replace: true })
    } catch (err) {
      setError(err instanceof Error ? err.message : '登录失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="relative flex min-h-screen items-center justify-center overflow-hidden bg-gradient-to-br from-zinc-100 via-zinc-50 to-zinc-100 dark:from-zinc-950 dark:via-zinc-900 dark:to-zinc-950">
      <div className="absolute inset-0 overflow-hidden">
        <div className="absolute -left-40 -top-40 h-80 w-80 rounded-full bg-emerald-500/10 blur-3xl dark:bg-emerald-500/20" />
        <div className="absolute -right-40 -bottom-40 h-80 w-80 rounded-full bg-blue-500/10 blur-3xl dark:bg-blue-500/20" />
        <div className="absolute left-1/2 top-1/2 h-60 w-60 -translate-x-1/2 -translate-y-1/2 rounded-full bg-violet-500/5 blur-3xl dark:bg-violet-500/10" />
      </div>

      <Card className="relative z-10 w-full max-w-md border-border bg-card/80 shadow-2xl backdrop-blur-xl">
        <CardHeader className="space-y-1 text-center">
          <div className="mx-auto mb-4 flex size-14 items-center justify-center rounded-xl bg-gradient-to-br from-emerald-500 to-teal-600">
            <Rocket className="size-8 text-white" />
          </div>
          <CardTitle className="text-2xl font-bold tracking-tight text-foreground">
            BuildFlow
          </CardTitle>
          <CardDescription>
            持续集成与部署平台
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <div className="rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400">
                {error}
              </div>
            )}
            <div className="space-y-2">
              <Label htmlFor="username">用户名</Label>
              <Input
                id="username"
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="请输入用户名"
                required
                autoComplete="username"
                className="focus-visible:ring-emerald-500/50"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">密码</Label>
              <Input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="请输入密码"
                required
                autoComplete="current-password"
                className="focus-visible:ring-emerald-500/50"
              />
            </div>
            <Button
              type="submit"
              disabled={loading}
              className={cn(
                'w-full bg-gradient-to-r from-emerald-500 to-emerald-600 text-white hover:from-emerald-600 hover:to-emerald-700',
                loading && 'opacity-80'
              )}
            >
              {loading ? (
                <>
                  <Loader2 className="size-4 animate-spin" />
                  登录中...
                </>
              ) : (
                '登录'
              )}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
