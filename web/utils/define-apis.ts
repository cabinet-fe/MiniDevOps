import { Http } from 'cat-kit/fe'

export function defineAPI(
  prefix: string,
  apis: (client: Http) => Record<string, (...args: any[]) => Promise<any>>
) {
  const http = new Http({
    baseUrl: prefix
  })

  return apis
}
