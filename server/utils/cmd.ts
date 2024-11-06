import { ShellPromise } from 'bun'

export async function runCmd(shellPromise: ShellPromise) {
  try {
    const text = await shellPromise.text('utf-8')
    return text
  } catch (err: any) {
    return Promise.reject(err.stderr.toString('utf-8'))
  }
}
