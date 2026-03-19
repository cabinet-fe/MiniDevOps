import { useEffect, useRef, useState } from 'react'
import { Copy, Check } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useWebSocket } from '@/hooks/use-websocket'

type BuildStatus = 'pending' | 'running' | 'success' | 'failed'

interface BuildLogViewerProps {
  buildId: number
  projectId?: number
  status?: BuildStatus
  initialLogs?: string
  streamUrl?: string
  className?: string
}

const STATUS_VARIANTS: Record<
  BuildStatus,
  { label: string; className: string }
> = {
  pending: { label: 'Pending', className: 'bg-zinc-500 text-white' },
  running: { label: 'Running', className: 'bg-blue-600 text-white' },
  success: { label: 'Success', className: 'bg-green-600 text-white' },
  failed: { label: 'Failed', className: 'bg-red-600 text-white' },
}

export function BuildLogViewer({
  buildId,
  projectId,
  status = 'pending',
  initialLogs = '',
  streamUrl,
  className,
}: BuildLogViewerProps) {
  const [logs, setLogs] = useState<string[]>(
    initialLogs ? initialLogs.split('\n').filter(Boolean) : []
  )
  const [copied, setCopied] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  const wsUrl =
    streamUrl ??
    (projectId
      ? `/ws/projects/${projectId}/builds/${buildId}/logs`
      : `/ws/builds/${buildId}/logs`)

  useWebSocket({
    url: wsUrl,
    onMessage: (data) => setLogs((prev) => [...prev, data]),
    enabled: !!streamUrl || status === 'running',
  })

  useEffect(() => {
    if (containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight
    }
  }, [logs])

  const handleCopy = async () => {
    await navigator.clipboard.writeText(logs.join('\n'))
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const statusInfo = STATUS_VARIANTS[status] ?? STATUS_VARIANTS.pending

  return (
    <div
      className={cn(
        'flex flex-col overflow-hidden rounded-lg border border-zinc-800 bg-zinc-950',
        className
      )}
    >
      <div className="flex items-center justify-between border-b border-zinc-800 px-4 py-2">
        <Badge className={statusInfo.className}>{statusInfo.label}</Badge>
        <Button
          variant="ghost"
          size="sm"
          className="h-8 gap-1.5 text-zinc-400 hover:text-white"
          onClick={handleCopy}
        >
          {copied ? (
            <>
              <Check className="size-4" />
              Copied
            </>
          ) : (
            <>
              <Copy className="size-4" />
              Copy logs
            </>
          )}
        </Button>
      </div>
      <div
        ref={containerRef}
        className="max-h-[480px] overflow-y-auto p-4 font-mono text-sm"
      >
        {logs.length === 0 ? (
          <p className="text-zinc-500">No logs yet...</p>
        ) : (
          <div className="select-text">
            {logs.map((line, i) => (
              <div
                key={i}
                className="flex hover:bg-zinc-900/50"
                style={{ minHeight: '1.5em' }}
              >
                <span className="select-none pr-4 text-right text-zinc-600">
                  {i + 1}
                </span>
                <span className="whitespace-pre-wrap break-all text-zinc-300">
                  {line}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
