import { createBunWebSocket } from 'hono/bun'
import type { ServerWebSocket, WebSocketHandler } from 'bun'

const bunSocket = createBunWebSocket<ServerWebSocket>()

export const websocket = bunSocket.websocket as unknown as WebSocketHandler

export const upgradeWebSocket = bunSocket.upgradeWebSocket
