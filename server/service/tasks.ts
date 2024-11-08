import { WSMessage, WSPool } from '../utils/ws'

export const taskPool = new WSPool()

export const taskProgressMessage = new WSMessage(new Set<number>(), data => {
  return JSON.stringify({
    type: 'progress',
    data: Array.from(data)
  })
})

export const taskResultMessage = new WSMessage({}, data => {
  return JSON.stringify({
    type: 'result',
    data
  })
})

taskPool.subscribe(taskProgressMessage)

taskPool.subscribe(taskResultMessage)
