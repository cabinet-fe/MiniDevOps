import type { Hono } from 'hono'
import { repo } from './repos'
import { task } from './tasks'
import { remote } from './remotes'
export function registerRoutes(app: Hono) {
  app.route('/repos', repo)
  app.route('/tasks', task)
  app.route('/remotes', remote)
}
