import { useEffect, useRef, useCallback } from 'react'

interface UseWebSocketOptions {
  url: string
  onMessage: (data: string) => void
  onOpen?: () => void
  onClose?: () => void
  enabled?: boolean
}

export function useWebSocket({ url, onMessage, onOpen, onClose, enabled = true }: UseWebSocketOptions) {
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined)

  const connect = useCallback(() => {
    if (!enabled) return
    const token = localStorage.getItem('access_token')
    const wsUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}${url}${url.includes('?') ? '&' : '?'}token=${token}`
    
    const ws = new WebSocket(wsUrl)
    wsRef.current = ws
    
    ws.onopen = () => { onOpen?.() }
    ws.onmessage = (e) => { onMessage(e.data) }
    ws.onclose = () => {
      onClose?.()
      // Auto-reconnect after 3s
      const tid = setTimeout(() => connect(), 3000)
      reconnectTimer.current = tid
    }
    ws.onerror = () => { ws.close() }
  }, [url, onMessage, onOpen, onClose, enabled])

  useEffect(() => {
    connect()
    return () => {
      if (reconnectTimer.current != null) clearTimeout(reconnectTimer.current)
      wsRef.current?.close()
    }
  }, [connect])

  const send = useCallback((data: string) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(data)
    }
  }, [])

  return { send }
}
