import { Http, type HttpResponse, type MergedConfig, str } from 'cat-kit/fe'
import { message } from 'ultra-ui'
import { session, TOKEN } from './cache'

const BASE_URL = '/api/v1'

const after = (res: HttpResponse) => {
  if (res.code >= 400) message.error(res.data.msg || res.message)
  if (res.data?.data !== undefined) {
    res.data = res.data.data
  }
  return res
}

const before = (config: MergedConfig) => {
  config.headers['Authorization'] = `Bearer ${session.get(TOKEN)}`
  return config
}

export function defineService<
  T extends Record<string, (...args: any[]) => Promise<any>>
>(prefix: string, apis: (client: Http) => T) {
  const http = new Http({
    baseUrl: str.joinPath(BASE_URL, prefix),
    after,
    before
  })

  return apis(http)
}

export function defineNoAuthService<
  T extends Record<string, (...args: any[]) => Promise<any>>
>(prefix: string, apis: (client: Http) => T) {
  const http = new Http({
    baseUrl: str.joinPath(BASE_URL, prefix),
    after
  })

  return apis(http)
}
