import { $, ShellPromise } from 'bun'
import { ChildProcess, exec } from 'child_process'

export async function runCmd(shellPromise: ShellPromise) {
  try {
    const text = await shellPromise.text('utf-8')
    return text
  } catch (err: any) {
    return Promise.reject(err.stderr.toString('utf-8'))
  }
}

export async function runCommand(
  command: string,
  cwd?: string,
  cb?: (cp: ChildProcess) => void
) {
  return new Promise((resolve, reject) => {
    const cp = exec(command, { cwd }, (err, stdout, stderr) => {
      if (err) {
        reject(stderr)
      } else {
        resolve({
          stdout,
          stderr
        })
      }
    })
    cb?.(cp)
  })
}

export async function runCommands(
  commands: string[],
  cwd?: string,
  cb?: (abort: () => void) => void
) {
  for (const command of commands) {
    await runCommand(command, cwd, cp => cp.kill)
  }
}

export async function runCmds(scripts: string[], cwd?: string) {
  if (!cwd) {
    cwd = process.cwd()
  }

  try {
    await $`${{ raw: scripts.join('\n') }}`.cwd(cwd).text()
  } catch (err: any) {
    console.log(err)
    return Promise.reject(err.stderr.toString())
  }
}
