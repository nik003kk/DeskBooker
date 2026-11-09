import { createContext, useCallback, useContext, useEffect, useState } from 'react'
import { clearAuthSession, readAuthSession, writeAuthSession } from './auth-session.js'

const AuthContext = createContext(null)
const baseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'

export function AuthProvider({ children }) {
  const [session, setSession] = useState(() => readAuthSession(window.sessionStorage))
  const [checking, setChecking] = useState(true)
  const logout = useCallback(() => {
    clearAuthSession(window.sessionStorage)
    setSession(null)
  }, [])
  const login = useCallback((data) => {
    setSession(writeAuthSession(window.sessionStorage, data))
  }, [])
  useEffect(() => {
    let active = true
    const current = readAuthSession(window.sessionStorage)
    if (!current) { setSession(null); setChecking(false); return }
    const controller = new AbortController()
    fetch(baseUrl.replace(/\/$/, '') + '/auth/me', {
      headers: { Authorization: 'Bearer ' + current.access_token },
      signal: controller.signal,
    }).then(async response => {
      if (!active) return
      if (!response.ok) { logout(); return }
      const data = await response.json()
      if (!active || !data?.user?.id || data.user.id !== current.user.id) { if (active) logout(); return }
      const refreshed = { ...current, user: data.user }
      setSession(refreshed)
      writeAuthSession(window.sessionStorage, refreshed)
    }).catch(() => {
      // Network outage should not silently grant access; clear the session.
      if (active) logout()
    }).finally(() => { if (active) setChecking(false) })
    return () => { active = false; controller.abort() }
  }, [logout])

  // Refresh the displayed role when returning from another tab/window.
  // Every protected API request remains independently authorized by the server.
  useEffect(() => {
    let active = true
    let controller = null
    function refreshCurrentUser() {
      if (document.visibilityState === 'hidden') return
      const current = readAuthSession(window.sessionStorage)
      if (!current) return
      controller?.abort()
      controller = new AbortController()
      const signal = controller.signal
      fetch(baseUrl.replace(/\/$/, '') + '/auth/me', {
        headers: { Authorization: 'Bearer ' + current.access_token }, signal,
      }).then(async response => {
        if (!active || signal.aborted || readAuthSession(window.sessionStorage)?.access_token !== current.access_token) return
        if (response.status === 401 || response.status === 403) { logout(); return }
        if (!response.ok) return
        const data = await response.json()
        if (!active || signal.aborted || readAuthSession(window.sessionStorage)?.access_token !== current.access_token) return
        if (!data?.user?.id || data.user.id !== current.user.id) { logout(); return }
        writeAuthSession(window.sessionStorage, { ...current, user: data.user })
        setSession(previous => previous?.access_token === current.access_token ? { ...previous, user: data.user } : previous)
      }).catch(() => {})
    }
    window.addEventListener('focus', refreshCurrentUser)
    document.addEventListener('visibilitychange', refreshCurrentUser)
    return () => { active = false; controller?.abort(); window.removeEventListener('focus', refreshCurrentUser); document.removeEventListener('visibilitychange', refreshCurrentUser) }
  }, [session?.access_token, logout])

  useEffect(() => {
    if (!session) return
    const delay = Date.parse(session.expires_at) - Date.now()
    if (delay <= 0) { logout(); return }
    const timer = window.setTimeout(logout, delay)
    return () => window.clearTimeout(timer)
  }, [session, logout])

  return (
    <AuthContext.Provider value={{ session, user: session?.user ?? null, token: session?.access_token ?? null,
      isAuthenticated: Boolean(session), checking, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const value = useContext(AuthContext)
  if (!value) throw new Error('useAuth requires AuthProvider')
  return value
}

