import { defineService } from '@/utils/define-service'
import type { Repo } from '@/types'

export const repoService = defineService('/repositories', client => {
  return {
    getRepos: () => {
      return client.get<Repo[]>('/list')
    },

    getRepoPage: (params?: Record<string, any>) => {
      return client
        .get<{
          data: Repo[]
          total: number
        }>('/page', { params })
        .then(res => res.data)
    },

    createRepo: (data: Partial<Repo>) => {
      return client.post('/create', data)
    },
    updateRepo: (data: Partial<Repo>) => {
      return client.post('/update', data)
    },
    deleteRepo: (id: number) => {
      return client.post('/delete', { id })
    }
  }
})
