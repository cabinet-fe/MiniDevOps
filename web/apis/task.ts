import { defineService } from '@/utils/define-service'
import type { Task } from '@/types'

export const taskService = defineService('/task', client => {
  return {
    getTasks: (repoId?: number) => {
      return client.get<Task[]>('/list', { params: { repoId } })
    },
    getTaskPage: (params: { page: number; pageSize: number }) => {
      return client.get<{
        list: Task[]
        total: number
      }>('/page', { params })
    },
    createTask: (data: Partial<Task>) => {
      return client.post('/create', data)
    },
    updateTask: (data: Partial<Task>) => {
      return client.post('/update', data)
    },
    runTask: (id: number) => {
      return client.post('/run', { id })
    },
    deleteTask: (id: number) => {
      return client.post('/delete', { id })
    }
  }
})
