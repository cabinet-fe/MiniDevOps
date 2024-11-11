import { Hono } from 'hono'
import { db } from '../db'
import { getPageParams } from '../utils/page'
import { runCmds } from '../utils/cmd'
import path from 'path'
import { getRepoName, gitCheckout } from '../utils/git'
import fs from 'fs/promises'
import { taskProgressMessage, taskResultMessage } from '../service/tasks'

export const task = new Hono().basePath('/tasks')

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

// 构建
const abortMap = new Map<number, () => void>()
task.post('/:id/build', async c => {
  const id = +c.req.param('id')

  const task = await db.task.findUnique({
    where: { id },
    include: { repo: true }
  })
  if (!task) {
    return c.json({ msg: '任务不存在' }, 404)
  }

  const repoDir = path.resolve(
    task.repo.codePath,
    getRepoName(task.repo.address)
  )

  if (!(await fs.exists(repoDir))) {
    return c.json({ msg: '仓库不存在' }, 404)
  }

  const scripts = task.script?.split('\n') ?? []

  taskProgressMessage.update(s => {
    s.add(id)
  })

  ~(async function () {
    try {
      await gitCheckout(repoDir, task.branch)

      await runCmds(scripts, repoDir)

      taskResultMessage.update(() => {
        return {
          taskName: task.name,
          status: 'success'
        }
      })
    } catch (err) {
      taskResultMessage.update(() => {
        return {
          taskName: task.name,
          status: 'error',
          error: err
        }
      })
    } finally {
      taskProgressMessage.update(s => {
        s.delete(id)
      })

      abortMap.delete(id)
    }
  })()
  return c.json({ msg: '成功' })
})

// 停止构建
task.post('/:id/stop-build', async c => {
  const id = +c.req.param('id')
  taskProgressMessage.update(s => {
    s.delete(id)
  })

  const task = await db.task.findUnique({ where: { id } })

  const abort = abortMap.get(id)

  abort?.()
  abortMap.delete(id)

  taskResultMessage.update(() => {
    return {
      taskName: task!.name,
      status: 'error',
      error: '构建已停止'
    }
  })
  return c.json({ msg: '成功' })
})
