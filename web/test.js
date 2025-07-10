import { createRequire } from 'node:module'
import { pathToFileURL } from 'node:url'

const cwd = process.cwd()
const url = pathToFileURL(cwd)
const require1 = createRequire(url)
const require2 = createRequire(import.meta.url)

console.log(url.href, import.meta.url)

// console.log(require1.resolve('ultra-ui'))
console.log(require2.resolve('ultra-ui'))
