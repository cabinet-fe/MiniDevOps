import { WebCache, type CacheKey } from 'cat-kit/fe'

export const session = WebCache.create('session')
export const local = WebCache.create('local')

export const TOKEN: CacheKey<string> = 'token'
