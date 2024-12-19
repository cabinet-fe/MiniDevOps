import { Hono } from 'hono'

const app = new Hono()

app.get('/login', c => {
  return c.json({
    message: '登录成功'
  })
})
