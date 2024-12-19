import { Http } from 'cat-kit/fe'
import { message } from 'ultra-ui'

export const http = new Http({
  baseUrl: '/api',
  after(res) {
    if (res.code >= 400) message.error(res.data.msg || res.message)
    if (res.data?.data !== undefined) {
      res.data = res.data.data
    }
    return res
  }
})
