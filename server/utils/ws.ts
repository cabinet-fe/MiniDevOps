import type { WSContext } from 'hono/ws'

export class WSMessage<Data = any> {
  private data: Data

  private msgGetter = (data: Data) => JSON.stringify(data)

  constructor(data: Data, msgGetter?: (data: Data) => string) {
    this.data = data
    if (msgGetter) this.msgGetter = msgGetter
  }

  private subscribers: Set<WSPool> = new Set()

  private publish() {
    const data = this.msgGetter(this.data)

    this.subscribers.forEach(pool => {
      pool.getConnections().forEach(ws => ws.send(data))
    })
  }

  update(updater: (data: Data) => Data | void | undefined) {
    const returnedData = updater(this.data)
    if (returnedData !== undefined) {
      this.data = returnedData
    }

    this.publish()
  }

  addSubscriber(wsPool: WSPool) {
    this.subscribers.add(wsPool)
  }
}

export class WSPool {
  private connections = new Set<WSContext>()

  add(ws: WSContext) {
    this.connections.add(ws)
  }

  delete(ws: WSContext) {
    this.connections.delete(ws)
  }

  /** 订阅消息 */
  subscribe(message: WSMessage) {
    message.addSubscriber(this)
  }

  getConnections() {
    return this.connections
  }
}
