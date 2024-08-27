import type { Hono } from 'hono'
import { repo } from './repos'

export function registerRoutes(app: Hono) {
  app.route('/repos', repo)
}
