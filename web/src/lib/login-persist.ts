const REMEMBER_KEY = 'buildflow_v1_login_remember'
const USERNAME_KEY = 'buildflow_v1_login_username'
const PASSWORD_KEY = 'buildflow_v1_login_password'

export function loadSavedCredentials(): {
  remember: boolean
  username: string
  password: string
} {
  try {
    const remember = localStorage.getItem(REMEMBER_KEY) === '1'
    const username = localStorage.getItem(USERNAME_KEY) ?? ''
    const password = remember ? (localStorage.getItem(PASSWORD_KEY) ?? '') : ''
    return { remember, username, password }
  } catch {
    return { remember: false, username: '', password: '' }
  }
}

export function persistCredentials(
  remember: boolean,
  username: string,
  password: string,
) {
  try {
    if (remember) {
      localStorage.setItem(REMEMBER_KEY, '1')
      localStorage.setItem(USERNAME_KEY, username)
      localStorage.setItem(PASSWORD_KEY, password)
    } else {
      localStorage.removeItem(REMEMBER_KEY)
      localStorage.removeItem(USERNAME_KEY)
      localStorage.removeItem(PASSWORD_KEY)
    }
  } catch {
    // private mode / quota
  }
}
