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
