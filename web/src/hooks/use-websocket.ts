import { useEffect, useEffectEvent, useRef, useCallback } from 'react'

interface UseWebSocketOptions {
  url: string
  onMessage: (data: string) => void
  onOpen?: () => void
  onClose?: () => void
  enabled?: boolean
}

export function useWebSocket({ url, onMessage, onOpen, onClose, enabled = true }: UseWebSocketOptions) {
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const connectTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const handleMessage = useEffectEvent((data: string) => {
    onMessage(data)
  })
  const handleOpen = useEffectEvent(() => {
    onOpen?.()
  })
  const handleClose = useEffectEvent(() => {
    onClose?.()
  })

  useEffect(() => {
    if (!enabled) return

    let cancelled = false

    const clearTimers = () => {
      if (connectTimer.current != null) {
        clearTimeout(connectTimer.current)
        connectTimer.current = null
      }
      if (reconnectTimer.current != null) {
        clearTimeout(reconnectTimer.current)
        reconnectTimer.current = null
      }
    }

    const connect = () => {
      if (cancelled) return

      const token = localStorage.getItem('access_token')
      const wsUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}${url}${url.includes('?') ? '&' : '?'}token=${token}`
      const ws = new WebSocket(wsUrl)
      wsRef.current = ws

      ws.onopen = () => {
        if (cancelled || wsRef.current !== ws) return
        handleOpen()
      }

      ws.onmessage = (e) => {
        if (cancelled || wsRef.current !== ws) return
        handleMessage(e.data)
      }

      ws.onclose = () => {
        if (wsRef.current === ws) {
          wsRef.current = null
        }
        if (cancelled) return
        handleClose()
        reconnectTimer.current = setTimeout(connect, 3000)
      }

      ws.onerror = () => {
        ws.close()
      }
    }

    // Delay the initial connection so React StrictMode can tear down the
    // first dev-only effect pass before any socket is created.
    connectTimer.current = setTimeout(connect, 0)

    return () => {
      cancelled = true
      clearTimers()
      const ws = wsRef.current
      wsRef.current = null
      ws?.close()
    }
  }, [enabled, url])

  const send = useCallback((data: string) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(data)
    }
  }, [])

  return { send }
}
