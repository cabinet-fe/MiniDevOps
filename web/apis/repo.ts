import { defineService } from '@/utils/define-service'
import type { Repo } from '@/types'

export const repoService = defineService('/repo', client => {
  return {
    getRepos: () => {
      return client.get<Repo[]>('/list')
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
