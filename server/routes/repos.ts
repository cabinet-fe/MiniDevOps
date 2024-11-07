// 仓库路由
import { Hono } from 'hono'
import { db } from '../db'
import { getPageParams } from '../utils/page'
import { gitClone } from '../utils/git'
import { $ } from 'bun'
import path from 'path'
import { exists } from 'fs/promises'
import { runCmd } from '../utils/cmd'

export const repo = new Hono().basePath('/repos')

repo.get('/page', async c => {
  const rows = await db.repo.findMany({
    ...getPageParams(c),
    select: {
      address: true,
      name: true,
      id: true,
      codePath: true
    }
  })

  return c.json({
    msg: '成功',
    data: {
      rows,
      total: await db.repo.count()
    }
  })
})

repo.get('/list', async c => {
  const rows = await db.repo.findMany({
    select: {
      name: true,
      id: true
    }
  })
  return c.json({ msg: '成功', data: rows })
})

repo.put('/:id', async c => {
  const id = c.req.param('id')
  const data = await c.req.json()
  await db.repo.update({ where: { id: +id }, data })
  return c.json({ msg: '成功' })
})

repo.post('/', async c => {
  const data = await c.req.json()

  let { address, username, pwd, codePath } = data

  address = address.replace(/^https?:\/\//, '')
  username = encodeURIComponent(username)

  const codeExist = await exists(codePath)
  if (!codeExist) {
    await gitClone({ address, username, pwd, destination: codePath })
  }

  const ret = await db.repo.create({
    data
  })
  return c.json({
    data: ret,
    msg: '新增成功'
  })
})

/** 获取 */
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

/** 删除 */
repo.delete('/:id', async c => {
  const { id } = c.req.param()

  const task = await db.task.findFirst({
    where: { repoId: +id }
  })
  if (task)
    return c.json(
      {
        msg: `任务${task.name}引用了这个仓库，请先删除对应的仓库`
      },
      400
    )

  await db.repo.delete({
    where: { id: +id }
  })
  return c.json({
    msg: '删除成功'
  })
})

repo.get('/:id/branch', async c => {
  const { id } = c.req.param()
  const data = await db.repo.findUnique({ where: { id: +id } })
  if (!data) throw new Error('仓库不存在')

  const repoName = data.address
    .slice(data.address.lastIndexOf('/') + 1)
    .replace(/\.git$/, '')

  const so = await runCmd(
    $`git branch -lr`.cwd(path.resolve(data.codePath, repoName))
  )

  const branches = so
    .trim()
    .split('\n')
    .map(v => v.trim())
    .filter(v => !!v && !v.startsWith('origin/HEAD'))
  return c.json({ msg: '获取成功', data: branches })
})
