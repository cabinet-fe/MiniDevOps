import { Hono } from 'hono'
import { upgradeWebSocket } from '../ws'
import { connections, publish } from '../state'

export const taskWS = new Hono().basePath('/tasks')

taskWS.get(
  '/progress',
  upgradeWebSocket(c => {
    return {
      onMessage(ev, ws) {
        if (ev.data === 'connect') {
          connections.add(ws)
          publish()
        }
      },
      onClose: (_, ws) => {
        connections.delete(ws)
      }
    }
  })
)
