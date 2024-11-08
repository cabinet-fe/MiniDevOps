import { $, ShellPromise } from 'bun'

export async function runCmd(shellPromise: ShellPromise) {
  try {
    const text = await shellPromise.text('utf-8')
    return text
  } catch (err: any) {
    return Promise.reject(err.stderr.toString('utf-8'))
  }
}

export async function runCmds(scripts: string[], cwd?: string) {
  if (!cwd) {
    cwd = process.cwd()
  }

  try {
    for (const script of scripts) {
      await $`${{ raw: script }}`.cwd(cwd).text()
    }
  } catch (err: any) {
    return Promise.reject(err.stderr.toString())
  }
}
