import { authorizedFetch } from './auth-api.js'
export async function fetchProfile({ token, logout, fetchImpl = fetch, apiBaseUrl } = {}) {
  const response = await authorizedFetch('/users/me', { token, logout, fetchImpl, apiBaseUrl })
  const data = await response.json()
  if (!data?.user?.id || !data.user.email) throw new Error('Strežnik je vrnil neveljavne podatke.')
  return data.user
}
export function accountStatusLabel(status) {
  return ({ active:'Aktiven', inactive:'Neaktiven', suspended:'Začasno onemogočen' })[status] || 'Neznano'
}
