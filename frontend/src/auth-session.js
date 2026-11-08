const key = 'deskbooker-auth-session'

export function readAuthSession(storage, now = Date.now()) {
  try {
    const raw = storage.getItem(key)
    if (!raw) return null
    const parsed = JSON.parse(raw)
    if (typeof parsed?.access_token !== 'string' || !parsed.access_token ||
        typeof parsed?.expires_at !== 'string' ||
        !Number.isFinite(Date.parse(parsed.expires_at)) ||
        Date.parse(parsed.expires_at) <= now ||
        typeof parsed?.user?.id !== 'string') {
      storage.removeItem(key)
      return null
    }
    return parsed
  } catch {
    try { storage.removeItem(key) } catch {}
    return null
  }
}
export function writeAuthSession(storage, data) {
  const session = { access_token: data.access_token, expires_at: data.expires_at, user: data.user }
  if (!readableSession(session)) throw new Error('Neveljavni prijavni podatki.')
  storage.setItem(key, JSON.stringify(session))
  return session
}
function readableSession(session) {
  return Boolean(session?.access_token && session?.user?.id &&
    Date.parse(session?.expires_at) > Date.now())
}
export function clearAuthSession(storage) { storage.removeItem(key) }
