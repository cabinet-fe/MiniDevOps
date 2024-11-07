import type { WSContext } from 'hono/ws'

export const connections = new Set<WSContext>()

export const buildingTasks = new Set<number>()

export function publish() {
  connections.forEach(ws => {
    ws.send(JSON.stringify(Array.from(buildingTasks)))
  })
}

export function addBuildingTask(id: number) {
  buildingTasks.add(id)
  publish()
}

export function removeBuildingTask(id: number) {
  buildingTasks.delete(id)
  publish()
}
