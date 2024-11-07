import { Hono } from 'hono'
import { repo } from './repos'
import { task } from './tasks'
import { remote } from './remotes'
import { taskWS } from './tasks-ws'

export const http = new Hono()
export const ws = new Hono()

http.route('/', repo)
http.route('/', task)
http.route('/', remote)

ws.route('/', taskWS)
