import { Hono } from 'hono'
import { registerRoutes } from './routes'
import { logger } from 'hono/logger'
import { date } from 'cat-kit/be'

const app = new Hono()

app.use(
  logger((msg, ...rest: string[]) => {
    console.log(msg, ...rest, date().format('yyyy-MM-dd HH:mm:ss'))
  })
)

registerRoutes(app)

export default app
