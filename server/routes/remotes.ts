import { Hono } from 'hono'
import { db } from '../db'
import { getPageParams } from '../utils/page'

export const remote = new Hono()

remote.get('/page', async c => {
  const rows = await db.remote.findMany(getPageParams(c))
  return c.json({
    msg: '成功',
    data: {
      rows,
      total: await db.remote.count()
    }
  })
})

remote.get('/list', async c => {
  const rows = await db.remote.findMany({
    select: {
      name: true,
      id: true
    }
  })
  return c.json({ msg: '成功', data: rows })
})

remote.post('/', async c => {
  const data = await c.req.json()
  const ret = await db.remote.create({
    data
  })
  return c.json({
    msg: '成功',
    data: ret
  })
})

remote.put('/:id', async c => {
  const id = c.req.param('id')
  const data = await c.req.json()
  const ret = await db.remote.update({
    data,
    where: { id: +id }
  })
  return c.json({ msg: '成功', data: ret })
})

remote.delete('/:id', async c => {
  const id = c.req.param('id')
  await db.remote.delete({ where: { id: +id } })
  return c.json({ msg: '成功' })
})
