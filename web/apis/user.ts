import { http } from '@/utils/http'

export const userService = {
  async login(data: { username: string; password: string }) {
    const res = await http.post('/post', data)
    return res.data
  }
}
