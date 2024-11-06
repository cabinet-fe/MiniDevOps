import { Hono } from 'hono'
import { db } from '../db'
import { getPageParams } from '../utils/page'

export const task = new Hono()

task.get('/page', async c => {
  const rows = await db.task.findMany({
    ...getPageParams(c),
    include: {
      repo: {
        select: {
          name: true
        }
      }
    }
  })

  return c.json({
    msg: '成功',
    data: {
      rows,
      total: await db.task.count()
    }
  })
})

task.delete('/:id', async c => {
  const id = c.req.param('id')
  await db.task.delete({ where: { id: +id } })
  return c.json({ msg: '成功' })
})

task.put('/:id', async c => {
  const id = c.req.param('id')
  const data = await c.req.json()
  await db.task.update({ where: { id: +id }, data })
  return c.json({ msg: '成功' })
})

task.post('/', async c => {
  const data = await c.req.json()
  await db.task.create({ data })
  return c.json({ msg: '成功' })
})

task.post('/:id/build', async c => {
  const id = c.req.param('id')

  return c.json({ msg: '成功' })
})

task.post('/:id/stop-build', async c => {
  const id = c.req.param('id')
  return c.json({ msg: '成功' })
})