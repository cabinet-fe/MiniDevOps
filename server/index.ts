import { Hono } from 'hono'
import { registerRoutes } from './routes'
import { logger } from 'hono/logger'
import { date } from 'cat-kit/be'

const app = new Hono().basePath('/api')

app.use(
  logger((msg, ...rest: string[]) => {
    console.log(msg, ...rest, date().format('yyyy-MM-dd HH:mm:ss'))
  })
)

app.onError((err, c) => {
  return c.json({ msg: err.stack }, 500)
})

registerRoutes(app)

export default {
  fetch: app.fetch
  // websocket: {}
}
