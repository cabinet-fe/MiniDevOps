// 仓库路由
import { Hono } from 'hono'
import { db } from '../db'
import { getPageParams } from '../utils/page'

export const repo = new Hono()

repo.get('/page', async c => {
  const rows = await db.repo.findMany(getPageParams(c))

  return c.json({
    msg: '成功',
    data: {
      rows,
      total: await db.repo.count()
    }
  })
})

repo.post('/', async c => {
  const data = await c.req.json()
  const ret = await db.repo.create({
    data
  })
  return c.json({
    data: ret,
    msg: '新增成功'
  })
})

/**  */
repo.get('/:id', async c => {
  const { id } = c.req.param()
  const data = await db.repo.findUnique({
    where: { id: +id }
  })
  return c.json({
    msg: '获取成功',
    data
  })
})
