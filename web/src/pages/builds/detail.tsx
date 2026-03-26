import { useState, useEffect, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router'
import {
  ArrowLeft,
  RotateCcw,
  XCircle,
  Download,
  Rocket,
  Clock,
  GitCommit,
  User,
  CheckCircle2,
  Circle,
  AlertCircle,
  Loader2,
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { BuildLogViewer, type BuildStatus } from '@/components/build-log-viewer'
import { api } from '@/lib/api'
import { BUILD_STATUSES, DISTRIBUTION_SUMMARY_LABELS } from '@/lib/constants'
import { cn } from '@/lib/utils'
import { toast } from 'sonner'
import { useWebSocket } from '@/hooks/use-websocket'

interface BuildDistributionRow {
  id: number
  build_id: number
  distribution_id: number
  status: string
  error_message: string
  method?: string
  remote_path?: string
}

interface BuildDetail {
  id: number
  project_id: number
  environment_id: number
  build_number: number
  status: BuildStatus
  current_stage: string
  trigger_type: string
  triggered_by: number
  commit_hash: string
  branch: string
  commit_message: string
  log_path: string
  artifact_path: string
  duration_ms: number
  error_message: string
  distribution_summary?: string
  started_at: string | null
  finished_at: string | null
  created_at: string
  project_name: string
  environment_name: string
  triggered_by_name: string
  distributions?: BuildDistributionRow[]
}

const TIMELINE_STAGES = ['pending', 'cloning', 'building', 'distributing', 'success'] as const

function normalizeTimelineStage(status: string, currentStage: string, distributionSummary?: string): string {
  if (status === 'failed' || status === 'cancelled') {
    return currentStage === 'deploying' ? 'distributing' : currentStage
  }
  if (status === 'success' && (distributionSummary === 'running' || distributionSummary === 'pending')) {
    return 'distributing'
  }
  if (currentStage === 'deploying') return 'distributing'
  return currentStage || status
}

function getStageState(
  currentStatus: string,
  currentStage: string,
  distributionSummary: string | undefined,
  stage: string
): 'completed' | 'active' | 'pending' | 'failed' {
  const order = TIMELINE_STAGES.indexOf(stage as (typeof TIMELINE_STAGES)[number])
  const effectiveStage = normalizeTimelineStage(currentStatus, currentStage, distributionSummary)
  const currentOrder = TIMELINE_STAGES.indexOf(
    effectiveStage as (typeof TIMELINE_STAGES)[number]
  )

  if (currentOrder === -1) return 'pending'

  if (currentStatus === 'failed' || currentStatus === 'cancelled') {
    if (order < currentOrder) return 'completed'
    if (order === currentOrder) return 'failed'
    return 'pending'
  }

  if (order < currentOrder) return 'completed'
  if (order === currentOrder) return currentStatus === 'success' && stage === 'success' ? 'completed' : 'active'
  return 'pending'
}

function formatDuration(ms: number): string {
  if (!ms) return '-'
  const seconds = Math.floor(ms / 1000)
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = seconds % 60
  return `${minutes}m ${remainingSeconds}s`
}

function formatTime(time: string | null): string {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

export function BuildDetailPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [build, setBuild] = useState<BuildDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [actionLoading, setActionLoading] = useState<string | null>(null)

  const fetchBuild = useCallback(async () => {
    if (!id) return
    try {
      const res = await api.get<BuildDetail>(`/builds/${id}`)
      if (res.code === 0 && res.data) {
        setBuild(res.data)
      }
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => {
    fetchBuild()
  }, [fetchBuild])

  const isRunning = build
    ? ['pending', 'cloning', 'building', 'deploying'].includes(build.status) ||
      (build.status === 'success' &&
        (build.distribution_summary === 'running' || build.distribution_summary === 'pending'))
    : false

  // WebSocket for real-time status updates
  useWebSocket({
    url: `/ws/builds/${id}/logs`,
    onMessage: () => {
      // Refresh build data periodically when receiving log messages
    },
    enabled: isRunning,
  })

  // Poll for status updates when running
  useEffect(() => {
    if (!isRunning) return
    const interval = setInterval(fetchBuild, 3000)
    return () => clearInterval(interval)
  }, [isRunning, fetchBuild])

  const handleAction = async (action: string) => {
    if (!build) return
    setActionLoading(action)
    try {
      let res
      switch (action) {
        case 'cancel':
          res = await api.post(`/builds/${build.id}/cancel`)
          break
        case 'deploy':
          res = await api.post(`/builds/${build.id}/deploy`)
          break
        case 'retry':
          res = await api.post(`/builds/${build.id}/retry`)
          break
        case 'download':
          {
            const blob = await api.download(`/builds/${build.id}/artifact`)
            const url = URL.createObjectURL(blob)
            const a = document.createElement('a')
            a.href = url
            a.download = `build-${build.build_number}-artifact`
            a.click()
            URL.revokeObjectURL(url)
            toast.success('下载开始')
          }
          return
      }
      if (res && res.code === 0) {
        toast.success('操作成功')
        fetchBuild()
      } else if (res) {
        toast.error(res.message || '操作失败')
      }
    } catch {
      toast.error('操作失败')
    } finally {
      setActionLoading(null)
    }
  }

  if (loading || !build) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-2 border-muted border-t-foreground" />
      </div>
    )
  }

  const statusInfo = BUILD_STATUSES[build.status as keyof typeof BUILD_STATUSES] ?? {
    label: build.status,
    color: 'bg-gray-500',
  }
  const distSummaryLabel =
    build.distribution_summary && DISTRIBUTION_SUMMARY_LABELS[build.distribution_summary]
      ? DISTRIBUTION_SUMMARY_LABELS[build.distribution_summary]
      : ''
  const currentStageInfo = build.current_stage
    ? BUILD_STATUSES[
        (build.current_stage === 'deploying' ? 'distributing' : build.current_stage) as keyof typeof BUILD_STATUSES
      ]
    : undefined

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="icon" onClick={() => navigate(-1)}>
            <ArrowLeft className="size-5" />
          </Button>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-2xl font-bold tracking-tight">
                Build #{build.build_number}
              </h1>
              <Badge className={cn(statusInfo.color, 'text-white')}>
                {statusInfo.label}
              </Badge>
            </div>
            <p className="mt-0.5 text-sm text-muted-foreground">
              {build.project_name} / {build.environment_name}
            </p>
            {distSummaryLabel && (
              <p className="mt-1 text-xs text-muted-foreground">分发：{distSummaryLabel}</p>
            )}
          </div>
        </div>
        <div className="flex gap-2">
          {isRunning && (
            <Button
              variant="destructive"
              size="sm"
              onClick={() => handleAction('cancel')}
              disabled={actionLoading === 'cancel'}
            >
              {actionLoading === 'cancel' ? (
                <Loader2 className="size-4 animate-spin" />
              ) : (
                <XCircle className="size-4" />
              )}
              取消构建
            </Button>
          )}
          {build.status === 'success' && (
            <>
              {build.artifact_path && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleAction('download')}
                  disabled={actionLoading === 'download'}
                >
                  <Download className="size-4" />
                  下载产物
                </Button>
              )}
              {build.artifact_path && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleAction('deploy')}
                  disabled={actionLoading === 'deploy'}
                >
                  <Rocket className="size-4" />
                  重新分发
                </Button>
              )}
            </>
          )}
          {(build.status === 'failed' || build.status === 'cancelled') && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => handleAction('retry')}
              disabled={actionLoading === 'retry'}
            >
              {actionLoading === 'retry' ? (
                <Loader2 className="size-4 animate-spin" />
              ) : (
                <RotateCcw className="size-4" />
              )}
              重新构建
            </Button>
          )}
        </div>
      </div>

      {/* Build Info Card */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base">构建信息</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-x-8 gap-y-3 sm:grid-cols-3 lg:grid-cols-4">
            <InfoItem label="项目" value={build.project_name} />
            <InfoItem label="环境" value={build.environment_name} />
            <InfoItem label="分支" value={build.branch || '-'} />
            <InfoItem label="当前阶段" value={currentStageInfo?.label ?? build.current_stage ?? '-'} />
            <InfoItem label="触发方式" value={build.trigger_type || '-'} />
            <InfoItem
              label="触发者"
              value={build.triggered_by_name || '-'}
              icon={<User className="size-3.5" />}
            />
            <InfoItem
              label="Commit"
              value={build.commit_hash ? build.commit_hash.slice(0, 7) : '-'}
              icon={<GitCommit className="size-3.5" />}
              mono
            />
            <InfoItem
              label="提交信息"
              value={build.commit_message || '-'}
              className="col-span-2 sm:col-span-1"
            />
            <InfoItem label="创建时间" value={formatTime(build.created_at)} icon={<Clock className="size-3.5" />} />
            <InfoItem label="开始时间" value={formatTime(build.started_at)} />
            <InfoItem label="结束时间" value={formatTime(build.finished_at)} />
            <InfoItem label="耗时" value={formatDuration(build.duration_ms)} />
          </div>
          {build.error_message && (
            <>
              <Separator className="my-3" />
              <div className="rounded-md bg-red-500/10 border border-red-500/20 p-3">
                <p className="text-sm font-medium text-red-400">错误信息</p>
                {currentStageInfo && (
                  <p className="mt-1 text-xs text-red-300/70">失败阶段：{currentStageInfo.label}</p>
                )}
                <p className="mt-1 text-sm text-red-300/80">{build.error_message}</p>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {build.distributions && build.distributions.length > 0 && (
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base">分发目标</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            {build.distributions.map((row) => (
              <div
                key={row.id}
                className="flex flex-wrap items-center justify-between gap-2 rounded-md border border-border/80 px-3 py-2 text-sm"
              >
                <div className="min-w-0">
                  <span className="font-mono text-xs text-muted-foreground">
                    #{row.distribution_id}
                  </span>{' '}
                  <span className="text-muted-foreground">
                    {row.method || '-'}/{row.remote_path || '-'}
                  </span>
                </div>
                <Badge
                  variant="outline"
                  className={
                    row.status === 'success'
                      ? 'border-green-500/50 text-green-600'
                      : row.status === 'failed'
                        ? 'border-red-500/50 text-red-600'
                        : ''
                  }
                >
                  {row.status === 'pending'
                    ? '待处理'
                    : row.status === 'running'
                      ? '进行中'
                      : row.status === 'success'
                        ? '成功'
                        : row.status === 'failed'
                          ? '失败'
                          : row.status === 'cancelled'
                            ? '已取消'
                            : row.status}
                </Badge>
                {row.error_message && (
                  <p className="w-full text-xs text-red-500/90">{row.error_message}</p>
                )}
              </div>
            ))}
          </CardContent>
        </Card>
      )}

      {/* Build Timeline */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-base">构建进度</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="mt-4 flex w-full items-center justify-between">
            {TIMELINE_STAGES.map((stage, i) => {
              const state = getStageState(build.status, build.current_stage, build.distribution_summary, stage)
              const stageLabel =
                BUILD_STATUSES[stage as keyof typeof BUILD_STATUSES]?.label ?? stage
              const isLast = i === TIMELINE_STAGES.length - 1

              return (
                <div key={stage} className={cn("relative flex items-center", isLast ? "" : "flex-1")}>
                  {/* Step point */}
                  <div className="relative flex flex-col items-center">
                    {/* Ring for active item */}
                    {state === 'active' && (
                      <span className="absolute -inset-1 animate-ping rounded-full bg-blue-400/50" />
                    )}

                    <div
                      className={cn(
                        'relative z-10 flex size-8 sm:size-10 items-center justify-center rounded-full border-2 transition-all duration-300 shadow-sm glassmorphism',
                        state === 'completed' && 'border-green-500 bg-green-500 text-white',
                        state === 'active' && 'border-blue-500 bg-background text-blue-500 ring-4 ring-blue-500/20',
                        state === 'pending' && 'border-muted bg-muted/30 text-muted-foreground',
                        state === 'failed' && 'border-red-500 bg-red-500 text-white'
                      )}
                    >
                      {state === 'completed' && <CheckCircle2 className="size-4 sm:size-5" />}
                      {state === 'active' && <Loader2 className="size-4 sm:size-5 animate-spin" />}
                      {state === 'pending' && <Circle className="size-2 sm:size-2.5 fill-current" />}
                      {state === 'failed' && <AlertCircle className="size-4 sm:size-5" />}
                    </div>

                    <span
                      className={cn(
                        'absolute -bottom-7 text-[10px] sm:text-xs font-medium whitespace-nowrap transition-colors duration-300',
                        state === 'completed' && 'text-green-600 dark:text-green-400',
                        state === 'active' && 'text-blue-600 dark:text-blue-400 font-bold',
                        state === 'pending' && 'text-muted-foreground/60',
                        state === 'failed' && 'text-red-600 dark:text-red-400'
                      )}
                    >
                      {stageLabel}
                    </span>
                  </div>

                  {/* Connecting Line */}
                  {!isLast && (
                    <div className="flex-1 px-2 sm:px-4">
                      <div
                        className={cn(
                          'h-1 w-full rounded-full transition-all duration-500',
                          state === 'completed' ? 'bg-green-500' : 'bg-muted shadow-inner'
                        )}
                      />
                    </div>
                  )}
                </div>
              )
            })}
          </div>
          {/* Spacer to account for absolute positioned labels */}
          <div className="h-7 w-full" />
        </CardContent>
      </Card>

      {/* Build Logs */}
      <div>
        <h2 className="mb-3 text-base font-semibold">构建日志</h2>
        <BuildLogViewer
          buildId={build.id}
          status={
            build.status === 'success' &&
            (build.distribution_summary === 'running' || build.distribution_summary === 'pending')
              ? 'distributing'
              : (build.status === 'deploying' ? 'distributing' : build.status)
          }
        />
      </div>
    </div>
  )
}

function InfoItem({
  label,
  value,
  icon,
  mono,
  className,
}: {
  label: string
  value: string
  icon?: React.ReactNode
  mono?: boolean
  className?: string
}) {
  return (
    <div className={className}>
      <p className="text-xs text-muted-foreground">{label}</p>
      <p
        className={cn(
          'mt-0.5 text-sm flex items-center gap-1',
          mono && 'font-mono'
        )}
      >
        {icon}
        <span className="truncate">{value}</span>
      </p>
    </div>
  )
}
