import { $ } from 'bun'
import { runCmd } from './cmd'

export function getRepoName(gitAddress: string) {
  return gitAddress.slice(gitAddress.lastIndexOf('/') + 1).replace(/\.git$/, '')
}

export async function gitClone(options: {
  /** 仓库地址 */
  address: string
  /** 仓库用户名 */
  username: string
  /** 密码 */
  pwd: string
  /** 拉去到目标目录 */
  destination: string
}) {
  const { address, username, pwd, destination } = options
  // 执行Git命令克隆远程仓库到指定目录

  await runCmd(
    $`git clone https://${username}:${pwd}@${address}`.cwd(destination)
  )
}

/**
 * 切换分支并拉取
 */
export async function gitCheckout(cwd: string, branch: string) {
  await runCmd($`git checkout .`.cwd(cwd))
  await runCmd($`git checkout ${branch.replace('origin/', '')}`.cwd(cwd))
  await runCmd($`git pull --no-rebase`.cwd(cwd))
}

/**
 * 拉取
 */
export async function gitPull(cwd: string, branch?: string) {
  if (branch) {
    return gitCheckout(cwd, branch)
  }

  return runCmd($`git pull`.cwd(cwd))
}
