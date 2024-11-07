import { Hono } from 'hono'
import { http, ws } from './routes'
import { logger } from 'hono/logger'
import { date } from 'cat-kit/be'
import { websocket } from './ws'

const app = new Hono()

app.use(
  logger((msg, ...rest: string[]) => {
    console.log(msg, ...rest, date().format('yyyy-MM-dd HH:mm:ss'))
  })
)

app.onError((err, c) => {
  return c.json({ msg: err.message }, 500)
})

app.route('/api', http)
app.route('/ws', ws)

export default {
  fetch: app.fetch,
  port: 3000,
  websocket
}
