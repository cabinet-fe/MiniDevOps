import { useEffect, useRef, useState, useCallback } from 'react'
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
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { SearchAddon } from '@xterm/addon-search'
import { WebglAddon } from '@xterm/addon-webgl'
import '@xterm/xterm/css/xterm.css'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { useWebSocket } from '@/hooks/use-websocket'
import { api } from '@/lib/api'

export type BuildStatus =
  | 'pending'
  | 'cloning'
  | 'building'
  | 'deploying'
  | 'distributing'
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
  distributing: { label: '分发中', className: 'bg-purple-500 text-white' },
  success: { label: '成功', className: 'bg-green-600 text-white' },
  failed: { label: '失败', className: 'bg-red-600 text-white' },
  cancelled: { label: '已取消', className: 'bg-zinc-500 text-white' },
}

export function BuildLogViewer({
  buildId,
  projectId,
  status = 'pending',
  initialLogs = '',
  streamUrl,
  className,
}: BuildLogViewerProps) {
  const [copied, setCopied] = useState(false)
  const [autoScroll, setAutoScroll] = useState(true)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [searchOpen, setSearchOpen] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [matchCount, setMatchCount] = useState(0)
  const [currentMatch, setCurrentMatch] = useState(0)
  const [lineCount, setLineCount] = useState(0)

  const terminalRef = useRef<HTMLDivElement>(null)
  const xtermRef = useRef<Terminal | null>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)
  const searchAddonRef = useRef<SearchAddon | null>(null)
  const userScrolling = useRef(false)
  const logsBufferRef = useRef<string[]>([])
  const initializedRef = useRef(false)

  const wsUrl =
    streamUrl ??
    (projectId
      ? `/ws/projects/${projectId}/builds/${buildId}/logs`
      : `/ws/builds/${buildId}/logs`)

  const isRunning = ['pending', 'cloning', 'building', 'deploying', 'distributing'].includes(status)

  // Initialize xterm terminal
  useEffect(() => {
    if (!terminalRef.current || initializedRef.current) return

    const term = new Terminal({
      allowProposedApi: true,
      disableStdin: true,
      cursorBlink: false,
      cursorStyle: 'bar',
      cursorInactiveStyle: 'none',
      fontSize: 13,
      fontFamily: "'JetBrains Mono', 'Fira Code', 'Cascadia Code', Menlo, Monaco, 'Courier New', monospace",
      lineHeight: 1.4,
      scrollback: 100000,
      convertEol: true,
      theme: {
        background: '#09090b',
        foreground: '#d4d4d8',
        cursor: '#d4d4d8',
        selectionBackground: '#3b82f680',
        selectionForeground: '#ffffff',
        black: '#27272a',
        red: '#ef4444',
        green: '#22c55e',
        yellow: '#eab308',
        blue: '#3b82f6',
        magenta: '#a855f7',
        cyan: '#06b6d4',
        white: '#d4d4d8',
        brightBlack: '#52525b',
        brightRed: '#f87171',
        brightGreen: '#4ade80',
        brightYellow: '#facc15',
        brightBlue: '#60a5fa',
        brightMagenta: '#c084fc',
        brightCyan: '#22d3ee',
        brightWhite: '#fafafa',
      },
    })

    const fitAddon = new FitAddon()
    const searchAddon = new SearchAddon()

    term.loadAddon(fitAddon)
    term.loadAddon(searchAddon)

    // Listen for search result changes to get accurate match count
    searchAddon.onDidChangeResults((e) => {
      if (e) {
        setMatchCount(e.resultCount)
        setCurrentMatch(e.resultIndex === -1 ? 0 : e.resultIndex + 1)
      } else {
        setMatchCount(0)
        setCurrentMatch(0)
      }
    })

    term.open(terminalRef.current)

    // Try to load WebGL addon for better performance
    try {
      const webglAddon = new WebglAddon()
      webglAddon.onContextLoss(() => {
        webglAddon.dispose()
      })
      term.loadAddon(webglAddon)
    } catch {
      // WebGL not supported, fall back to canvas renderer
    }

    fitAddon.fit()

    // Track scrolling for auto-scroll behavior
    term.onScroll(() => {
      const viewport = term.buffer.active
      const isAtBottom =
        viewport.baseY <= viewport.viewportY
      if (!isAtBottom && !userScrolling.current) {
        userScrolling.current = true
        setAutoScroll(false)
      }
      if (isAtBottom && userScrolling.current) {
        userScrolling.current = false
      }
    })

    xtermRef.current = term
    fitAddonRef.current = fitAddon
    searchAddonRef.current = searchAddon
    initializedRef.current = true

    // Write any buffered logs
    if (logsBufferRef.current.length > 0) {
      const content = logsBufferRef.current.join('\n')
      term.write(content)
      setLineCount(logsBufferRef.current.length)
      logsBufferRef.current = []
    }

    return () => {
      initializedRef.current = false
      term.dispose()
      xtermRef.current = null
      fitAddonRef.current = null
      searchAddonRef.current = null
    }
  }, [])

  // Handle terminal resize when fullscreen changes or window resizes
  useEffect(() => {
    const handleResize = () => {
      if (fitAddonRef.current && xtermRef.current) {
        try {
          fitAddonRef.current.fit()
        } catch {
          // ignore fit errors during transitions
        }
      }
    }

    // Fit after a short delay to account for CSS transitions
    const timer = setTimeout(handleResize, 100)
    window.addEventListener('resize', handleResize)

    return () => {
      clearTimeout(timer)
      window.removeEventListener('resize', handleResize)
    }
  }, [isFullscreen])

  // Load initial/historical logs
  useEffect(() => {
    const term = xtermRef.current

    if (initialLogs) {
      const lines = initialLogs.split('\n').filter(Boolean)
      if (term) {
        term.clear()
        term.write(lines.join('\r\n'))
        setLineCount(lines.length)
      } else {
        logsBufferRef.current = lines
      }
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
        if (term) {
          term.clear()
          if (lines.length > 0) {
            term.write(lines.join('\r\n'))
          }
          setLineCount(lines.length)
        } else {
          logsBufferRef.current = lines
        }
      } catch {
        // ignore load errors
      }
    }

    void loadLogs()

    return () => {
      cancelled = true
    }
  }, [buildId, initialLogs])

  // WebSocket for streaming logs
  useWebSocket({
    url: wsUrl,
    onMessage: (data) => {
      const term = xtermRef.current
      if (term) {
        term.write('\r\n' + data)
        setLineCount((prev) => prev + 1)
        if (autoScroll) {
          term.scrollToBottom()
        }
      } else {
        logsBufferRef.current.push(data)
      }
    },
    enabled: !!streamUrl || isRunning,
  })

  // Auto-scroll when autoScroll state changes
  useEffect(() => {
    if (autoScroll && xtermRef.current) {
      xtermRef.current.scrollToBottom()
    }
  }, [autoScroll, lineCount])

  // Search functionality
  const searchOptions = {
    caseSensitive: false,
    regex: false,
    wholeWord: false,
    decorations: {
      matchBackground: '#854d0e',
      matchBorder: '#a16207',
      matchOverviewRuler: '#eab308',
      activeMatchBackground: '#1d4ed8',
      activeMatchBorder: '#3b82f6',
      activeMatchColorOverviewRuler: '#3b82f6',
    },
  }

  const performSearch = useCallback(
    (query: string, direction: 'next' | 'prev' = 'next') => {
      const searchAddon = searchAddonRef.current
      if (!searchAddon || !query) {
        setMatchCount(0)
        setCurrentMatch(0)
        return
      }

      if (direction === 'next') {
        searchAddon.findNext(query, searchOptions)
      } else {
        searchAddon.findPrevious(query, searchOptions)
      }
      // matchCount and currentMatch are updated via onDidChangeResults callback
    },
    []
  )

  const handleSearchChange = useCallback(
    (query: string) => {
      setSearchQuery(query)
      setCurrentMatch(0)
      setMatchCount(0)
      if (query) {
        performSearch(query, 'next')
      } else {
        searchAddonRef.current?.clearDecorations()
      }
    },
    [performSearch]
  )

  const navigateMatch = useCallback(
    (direction: 'prev' | 'next') => {
      if (!searchQuery) return
      performSearch(searchQuery, direction)
    },
    [searchQuery, performSearch]
  )

  const handleCopy = async () => {
    const term = xtermRef.current
    if (!term) return
    // Select all and get content
    term.selectAll()
    const text = term.getSelection()
    term.clearSelection()
    await navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const enableAutoScroll = () => {
    setAutoScroll(true)
    userScrolling.current = false
    xtermRef.current?.scrollToBottom()
  }

  const handleCloseSearch = useCallback(() => {
    setSearchOpen(false)
    setSearchQuery('')
    setMatchCount(0)
    setCurrentMatch(0)
    searchAddonRef.current?.clearDecorations()
  }, [])

  const statusInfo = STATUS_VARIANTS[status] ?? STATUS_VARIANTS.pending

  return (
    <div
      className={cn(
        'flex flex-col overflow-hidden rounded-lg border border-border bg-zinc-950 dark:bg-zinc-950',
        isFullscreen && 'fixed inset-0 z-50 rounded-none border-0',
        className
      )}
    >
      <div className="flex items-center justify-between border-b border-zinc-800/60 bg-zinc-900/80 px-4 py-2 gap-2">
        <div className="flex items-center gap-2">
          <Badge className={statusInfo.className}>{statusInfo.label}</Badge>
          <span className="text-xs text-zinc-500">{lineCount} 行</span>
        </div>
        <div className="flex items-center gap-1">
          {searchOpen ? (
            <div className="flex items-center gap-1 rounded-md border border-zinc-700 bg-zinc-800 px-2">
              <Search className="size-3.5 text-zinc-500" />
              <Input
                value={searchQuery}
                onChange={(e) => handleSearchChange(e.target.value)}
                placeholder="搜索日志..."
                className="h-7 w-40 border-0 bg-transparent text-sm text-zinc-200 placeholder:text-zinc-500 focus-visible:ring-0 px-1"
                autoFocus
                onKeyDown={(e) => {
                  if (e.key === 'Enter') navigateMatch(e.shiftKey ? 'prev' : 'next')
                  if (e.key === 'Escape') handleCloseSearch()
                }}
              />
              {searchQuery && (
                <span className="text-xs text-zinc-500 whitespace-nowrap">
                  {matchCount > 0
                    ? `${currentMatch}/${matchCount}`
                    : '无匹配'}
                </span>
              )}
              <Button
                variant="ghost"
                size="icon"
                className="size-6 text-zinc-300 hover:bg-zinc-700 hover:text-white"
                onClick={() => navigateMatch('prev')}
                disabled={matchCount === 0}
              >
                <ChevronUp className="size-3.5" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                className="size-6 text-zinc-300 hover:bg-zinc-700 hover:text-white"
                onClick={() => navigateMatch('next')}
                disabled={matchCount === 0}
              >
                <ChevronDown className="size-3.5" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                className="size-6 text-zinc-300 hover:bg-zinc-700 hover:text-white"
                onClick={handleCloseSearch}
              >
                <X className="size-3.5" />
              </Button>
            </div>
          ) : (
            <Button
              variant="ghost"
              size="icon"
              className="size-8 text-zinc-300 hover:bg-zinc-700 hover:text-white"
              onClick={() => setSearchOpen(true)}
            >
              <Search className="size-4" />
            </Button>
          )}
          {!autoScroll && (
            <Button
              variant="ghost"
              size="sm"
              className="h-8 gap-1 text-zinc-300 hover:bg-zinc-700 hover:text-white"
              onClick={enableAutoScroll}
            >
              <ArrowDownToLine className="size-3.5" />
              跟随
            </Button>
          )}
          <Button
            variant="ghost"
            size="icon"
            className="size-8 text-zinc-300 hover:bg-zinc-700 hover:text-white"
            onClick={handleCopy}
          >
            {copied ? <Check className="size-4" /> : <Copy className="size-4" />}
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="size-8 text-zinc-300 hover:bg-zinc-700 hover:text-white"
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
        ref={terminalRef}
        className="xterm-container"
        style={{
          height: isFullscreen ? 'calc(100vh - 52px)' : '480px',
          padding: '8px',
        }}
      />
    </div>
  )
}
