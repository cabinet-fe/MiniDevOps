export const ROLES = [
  { value: 'admin', label: '管理员' },
  { value: 'ops', label: '运维' },
  { value: 'dev', label: '开发' },
]

export const BUILD_STATUSES = {
  pending: { label: '等待中', color: 'bg-yellow-500' },
  cloning: { label: '拉取代码', color: 'bg-blue-400' },
  building: { label: '构建中', color: 'bg-blue-500' },
  deploying: { label: '部署中', color: 'bg-purple-500' },
  success: { label: '成功', color: 'bg-green-500' },
  failed: { label: '失败', color: 'bg-red-500' },
  cancelled: { label: '已取消', color: 'bg-gray-500' },
}

export const DEPLOY_METHODS = [
  { value: 'rsync', label: 'Rsync' },
  { value: 'sftp', label: 'SFTP' },
  { value: 'scp', label: 'SCP' },
]

export const AUTH_TYPES = [
  { value: 'password', label: '密码' },
  { value: 'key', label: 'SSH密钥' },
]

export const REPO_AUTH_TYPES = [
  { value: 'none', label: '无需认证' },
  { value: 'password', label: '用户名/密码' },
  { value: 'token', label: 'Token' },
]

export const BUILD_SCRIPT_TYPES = [
  { value: 'bash', label: 'Bash', placeholder: 'npm install && npm run build' },
  { value: 'node', label: 'Node.js', placeholder: "const { execSync } = require('child_process');\nexecSync('npm install && npm run build', { stdio: 'inherit' });" },
  { value: 'python', label: 'Python', placeholder: "import subprocess\nsubprocess.run(['npm', 'install'], check=True)\nsubprocess.run(['npm', 'run', 'build'], check=True)" },
]
