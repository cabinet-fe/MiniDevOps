import { Hono } from 'hono'
import { upgradeWebSocket } from '../ws'
import { taskPool } from '../service/tasks'

export const taskWS = new Hono().basePath('/tasks')

taskWS.get(
  '/progress',
  upgradeWebSocket(c => {
    return {
      onMessage(ev, ws) {
        if (ev.data === 'connect') {
          taskPool.add(ws)
        }
      },
      onClose: (_, ws) => {
        taskPool.delete(ws)
      }
    }
  })
)
