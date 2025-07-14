import { defineNoAuthService } from '@/utils/define-service'

export const userService = defineNoAuthService('/user', client => {
  return {
    login: (data: { username: string; password: string }) => {
      return client.post('/login', data)
    }
  }
})
