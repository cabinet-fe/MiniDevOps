export interface Repo {
  id: number
  name: string
  url: string
  createdAt: string
}

export interface Task {
  id: number
  repoId: number
  name: string
  status: string
  createdAt: string
  repo: Repo
}

export interface Remote {
  id: number
  name: string
  user: string
  addr: string
  createdAt: string
}
