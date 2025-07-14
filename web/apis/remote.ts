import { defineService } from '@/utils/define-service'
import type { Remote } from '@/types'

export const remoteService = defineService('/remote', client => {
  return {
    getRemotes: () => {
      return client.get<Remote[]>('/list')
    },
    createRemote: (data: Partial<Remote>) => {
      return client.post('/create', data)
    },
    updateRemote: (data: Partial<Remote>) => {
      return client.post('/update', data)
    },
    deleteRemote: (id: number) => {
      return client.post('/delete', { id })
    }
  }
})
