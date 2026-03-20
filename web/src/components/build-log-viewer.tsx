import { useEffect, useRef, useState, useCallback, useMemo } from 'react'
import {
  Copy,
  Check,
  Search,
  ChevronUp,
  ChevronDown,
  ArrowDownToLine,
  Maximize2,
  Minimize2,
  X,
} from 'lucide-react'
import Convert from 'ansi-to-html'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { useWebSocket } from '@/hooks/use-websocket'
import { api } from '@/lib/api'

type BuildStatus =
  | 'pending'
  | 'cloning'
  | 'building'
  | 'deploying'
  | 'success'
  | 'failed'
  | 'cancelled'

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
  pending: { label: '等待中', className: 'bg-yellow-500 text-white' },
  cloning: { label: '拉取代码', className: 'bg-blue-400 text-white' },
  building: { label: '构建中', className: 'bg-blue-600 text-white' },
  deploying: { label: '部署中', className: 'bg-purple-500 text-white' },
  success: { label: '成功', className: 'bg-green-600 text-white' },
  failed: { label: '失败', className: 'bg-red-600 text-white' },
  cancelled: { label: '已取消', className: 'bg-zinc-500 text-white' },
}

const VISIBLE_LINES = 200
const LINE_HEIGHT = 24

const ansiConverter = new Convert({
  fg: '#d4d4d8',
  bg: '#09090b',
  newline: false,
  escapeXML: true,
})

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
  const [autoScroll, setAutoScroll] = useState(true)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [searchOpen, setSearchOpen] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [currentMatch, setCurrentMatch] = useState(0)
  const containerRef = useRef<HTMLDivElement>(null)
  const userScrolling = useRef(false)
  const [scrollTop, setScrollTop] = useState(0)

  const wsUrl =
    streamUrl ??
    (projectId
      ? `/ws/projects/${projectId}/builds/${buildId}/logs`
      : `/ws/builds/${buildId}/logs`)

  const isRunning = ['pending', 'cloning', 'building', 'deploying'].includes(status)

  useEffect(() => {
    if (initialLogs) {
      setLogs(initialLogs.split('\n').filter(Boolean))
      return
    }

    let cancelled = false
    const loadLogs = async () => {
      try {
        const text = await api.getText(`/builds/${buildId}/log`)
        if (cancelled) return
        const lines = text.split('\n')
        if (lines.at(-1) === '') {
          lines.pop()
        }
        setLogs(lines)
      } catch {
        if (!cancelled) {
          setLogs([])
        }
      }
    }

    void loadLogs()

    return () => {
      cancelled = true
    }
  }, [buildId, initialLogs])

  useWebSocket({
    url: wsUrl,
    onMessage: (data) => setLogs((prev) => [...prev, data]),
    enabled: !!streamUrl || isRunning,
  })

  const matchedLines = useMemo(() => {
    if (!searchQuery) return []
    const lowerQ = searchQuery.toLowerCase()
    return logs.reduce<number[]>((acc, line, i) => {
      if (line.toLowerCase().includes(lowerQ)) acc.push(i)
      return acc
    }, [])
  }, [logs, searchQuery])

  useEffect(() => {
    if (autoScroll && containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight
    }
  }, [logs, autoScroll])

  const handleScroll = useCallback(() => {
    const el = containerRef.current
    if (!el) return
    setScrollTop(el.scrollTop)
    const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 50
    if (!atBottom && !userScrolling.current) {
      userScrolling.current = true
      setAutoScroll(false)
    }
    if (atBottom && userScrolling.current) {
      userScrolling.current = false
    }
  }, [])

  const handleCopy = async () => {
    await navigator.clipboard.writeText(logs.join('\n'))
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const scrollToLine = useCallback((lineIdx: number) => {
    if (containerRef.current) {
      containerRef.current.scrollTop = lineIdx * LINE_HEIGHT
      setAutoScroll(false)
    }
  }, [])

  const navigateMatch = useCallback(
    (direction: 'prev' | 'next') => {
      if (matchedLines.length === 0) return
      let next = currentMatch
      if (direction === 'next') {
        next = (currentMatch + 1) % matchedLines.length
      } else {
        next = (currentMatch - 1 + matchedLines.length) % matchedLines.length
      }
      setCurrentMatch(next)
      scrollToLine(matchedLines[next])
    },
    [currentMatch, matchedLines, scrollToLine]
  )

  const enableAutoScroll = () => {
    setAutoScroll(true)
    userScrolling.current = false
    if (containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight
    }
  }

  const statusInfo = STATUS_VARIANTS[status] ?? STATUS_VARIANTS.pending
  const totalLines = logs.length
  const useVirtual = totalLines > 5000
  const containerHeight = isFullscreen ? 'calc(100vh - 120px)' : '480px'

  const visibleRange = useMemo(() => {
    if (!useVirtual) return { start: 0, end: totalLines }
    const start = Math.max(0, Math.floor(scrollTop / LINE_HEIGHT) - 50)
    const end = Math.min(totalLines, start + VISIBLE_LINES + 100)
    return { start, end }
  }, [useVirtual, scrollTop, totalLines])

  const renderLine = useCallback(
    (line: string, idx: number) => {
      const isCurrentSearchMatch =
        searchQuery && matchedLines[currentMatch] === idx
      const isSearchMatch =
        searchQuery &&
        matchedLines.includes(idx) &&
        !isCurrentSearchMatch

      let html = ansiConverter.toHtml(line)

      if (searchQuery && (isSearchMatch || isCurrentSearchMatch)) {
        const escaped = searchQuery.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
        const re = new RegExp(`(${escaped})`, 'gi')
        html = html.replace(
          re,
          isCurrentSearchMatch
            ? '<mark class="bg-amber-500 text-black rounded px-0.5">$1</mark>'
            : '<mark class="bg-amber-500/30 rounded px-0.5">$1</mark>'
        )
      }

      return (
        <div
          key={idx}
          className={cn(
            'flex hover:bg-zinc-900/50',
            isCurrentSearchMatch && 'bg-amber-500/10',
            isSearchMatch && 'bg-amber-500/5'
          )}
          style={{ height: LINE_HEIGHT }}
        >
          <span className="w-12 shrink-0 select-none pr-4 text-right text-zinc-600 text-xs leading-6">
            {idx + 1}
          </span>
          <span
            className="whitespace-pre-wrap break-all text-zinc-300 leading-6"
            dangerouslySetInnerHTML={{ __html: html }}
          />
        </div>
      )
    },
    [searchQuery, matchedLines, currentMatch]
  )

  return (
    <div
      className={cn(
        'flex flex-col overflow-hidden rounded-lg border border-zinc-800 bg-zinc-950',
        isFullscreen && 'fixed inset-0 z-50 rounded-none border-0',
        className
      )}
    >
      <div className="flex items-center justify-between border-b border-zinc-800 px-4 py-2 gap-2">
        <div className="flex items-center gap-2">
          <Badge className={statusInfo.className}>{statusInfo.label}</Badge>
          <span className="text-xs text-zinc-500">{totalLines} 行</span>
        </div>
        <div className="flex items-center gap-1">
          {searchOpen ? (
            <div className="flex items-center gap-1 rounded-md border border-zinc-700 bg-zinc-900 px-2">
              <Search className="size-3.5 text-zinc-500" />
              <Input
                value={searchQuery}
                onChange={(e) => {
                  setSearchQuery(e.target.value)
                  setCurrentMatch(0)
                }}
                placeholder="搜索日志..."
                className="h-7 w-40 border-0 bg-transparent text-sm focus-visible:ring-0 px-1"
                autoFocus
                onKeyDown={(e) => {
                  if (e.key === 'Enter') navigateMatch(e.shiftKey ? 'prev' : 'next')
                  if (e.key === 'Escape') {
                    setSearchOpen(false)
                    setSearchQuery('')
                  }
                }}
              />
              {searchQuery && (
                <span className="text-xs text-zinc-500 whitespace-nowrap">
                  {matchedLines.length > 0
                    ? `${currentMatch + 1}/${matchedLines.length}`
                    : '无匹配'}
                </span>
              )}
              <Button
                variant="ghost"
                size="icon"
                className="size-6"
                onClick={() => navigateMatch('prev')}
                disabled={matchedLines.length === 0}
              >
                <ChevronUp className="size-3.5" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                className="size-6"
                onClick={() => navigateMatch('next')}
                disabled={matchedLines.length === 0}
              >
                <ChevronDown className="size-3.5" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                className="size-6"
                onClick={() => {
                  setSearchOpen(false)
                  setSearchQuery('')
                }}
              >
                <X className="size-3.5" />
              </Button>
            </div>
          ) : (
            <Button
              variant="ghost"
              size="icon"
              className="size-8 text-zinc-400 hover:text-white"
              onClick={() => setSearchOpen(true)}
            >
              <Search className="size-4" />
            </Button>
          )}
          {!autoScroll && (
            <Button
              variant="ghost"
              size="sm"
              className="h-8 gap-1 text-zinc-400 hover:text-white"
              onClick={enableAutoScroll}
            >
              <ArrowDownToLine className="size-3.5" />
              跟随
            </Button>
          )}
          <Button
            variant="ghost"
            size="icon"
            className="size-8 text-zinc-400 hover:text-white"
            onClick={handleCopy}
          >
            {copied ? <Check className="size-4" /> : <Copy className="size-4" />}
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="size-8 text-zinc-400 hover:text-white"
            onClick={() => setIsFullscreen((v) => !v)}
          >
            {isFullscreen ? (
              <Minimize2 className="size-4" />
            ) : (
              <Maximize2 className="size-4" />
            )}
          </Button>
        </div>
      </div>

      <div
        ref={containerRef}
        onScroll={handleScroll}
        className="overflow-y-auto p-4 font-mono text-sm"
        style={{ maxHeight: containerHeight }}
      >
        {logs.length === 0 ? (
          <p className="text-zinc-500">暂无日志...</p>
        ) : useVirtual ? (
          <div style={{ height: totalLines * LINE_HEIGHT, position: 'relative' }}>
            <div
              style={{
                position: 'absolute',
                top: visibleRange.start * LINE_HEIGHT,
                left: 0,
                right: 0,
              }}
            >
              {logs.slice(visibleRange.start, visibleRange.end).map((line, i) =>
                renderLine(line, visibleRange.start + i)
              )}
            </div>
          </div>
        ) : (
          <div className="select-text">
            {logs.map((line, i) => renderLine(line, i))}
          </div>
        )}
      </div>
    </div>
  )
}
