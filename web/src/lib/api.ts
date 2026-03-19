const API_BASE = '/api/v1'

interface ApiResponse<T> {
  code: number
  message: string
  data?: T
}

interface PaginatedData<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

let getToken: () => string | null = () => localStorage.getItem('access_token')

export function setTokenGetter(fn: () => string | null) {
  getToken = fn
}

async function request<T>(method: string, path: string, body?: unknown, isFormData = false): Promise<ApiResponse<T>> {
  const headers: Record<string, string> = {}
  const token = getToken()
  if (token) headers['Authorization'] = `Bearer ${token}`
  if (!isFormData) headers['Content-Type'] = 'application/json'
  
  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: isFormData ? body as FormData : body ? JSON.stringify(body) : undefined,
    credentials: 'include',
  })
  
  if (res.status === 401) {
    const shouldTryRefresh = path !== '/auth/login' && path !== '/auth/refresh'
    if (shouldTryRefresh) {
      const refreshRes = await fetch(`${API_BASE}/auth/refresh`, { method: 'POST', credentials: 'include' })
      if (refreshRes.ok) {
        const data = await refreshRes.json()
        if (data.data?.access_token) {
          localStorage.setItem('access_token', data.data.access_token)
          headers['Authorization'] = `Bearer ${data.data.access_token}`
          const retryRes = await fetch(`${API_BASE}${path}`, {
            method,
            headers,
            body: isFormData ? body as FormData : body ? JSON.stringify(body) : undefined,
            credentials: 'include',
          })
          return retryRes.json()
        }
      }
    }

    localStorage.removeItem('access_token')
    if (window.location.pathname !== '/login') {
      window.location.assign('/login')
    }
    throw new Error('Unauthorized')
  }
  
  return res.json()
}

export const api = {
  get: <T>(path: string) => request<T>('GET', path),
  post: <T>(path: string, body?: unknown) => request<T>('POST', path, body),
  put: <T>(path: string, body?: unknown) => request<T>('PUT', path, body),
  delete: <T>(path: string) => request<T>('DELETE', path),
  upload: <T>(path: string, formData: FormData) => request<T>('POST', path, formData, true),
  download: async (path: string) => {
    const token = getToken()
    const res = await fetch(`${API_BASE}${path}`, {
      headers: token ? { 'Authorization': `Bearer ${token}` } : {},
      credentials: 'include',
    })
    return res.blob()
  },
}

export type { ApiResponse, PaginatedData }
