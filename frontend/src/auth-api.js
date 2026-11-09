export class AuthAPIError extends Error {
  constructor(status) {
    super(status === 401 ? 'Seja je potekla. Prijavi se znova.'
      : status === 403 ? 'Za to dejanje nimaš ustreznih pravic.'
      : 'Zahteva ni uspela.')
    this.status = status
  }
}
export async function authorizedFetch(path, { token, logout, fetchImpl = fetch, apiBaseUrl = import.meta.env?.VITE_API_BASE_URL || 'http://localhost:8080/api', ...options } = {}) {
  if (!token) { logout?.(); throw new AuthAPIError(401) }
  const response = await fetchImpl(apiBaseUrl.replace(/\/$/, '') + path, {
    ...options,
    headers: { ...(options.headers || {}), Authorization: 'Bearer ' + token },
  })
  if (response.status === 401) { logout?.(); throw new AuthAPIError(401) }
  if (response.status === 403) throw new AuthAPIError(403)
  if (!response.ok) throw new AuthAPIError(response.status)
  return response
}

