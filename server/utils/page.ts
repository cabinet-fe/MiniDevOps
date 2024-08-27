import type { Context } from 'hono'

export function getPageParams(c: Context) {
  const query = c.req.query() ?? {}
  const page = Number(query.page ?? 1)
  const size = Number(query.size ?? 10)
  return {
    skip: (page - 1) * size,
    take: size
  }
}
