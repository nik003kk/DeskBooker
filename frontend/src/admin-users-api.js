import { authorizedFetch } from './auth-api.js'

export const userRoleLabels = { employee: 'Zaposleni', manager: 'Vodja pisarne', admin: 'Administrator' }
export const userStatusLabels = { active: 'Aktiven', inactive: 'Neaktiven', suspended: 'Začasno onemogočen' }
const publicRoles = { employee: 'EMPLOYEE', manager: 'OFFICE_MANAGER', admin: 'ADMIN' }
export function canAdministerUsers(user) { return user?.role === 'admin' && user.status === 'active' }
function validUser(user) {
  return typeof user?.id === 'string' && typeof user.email === 'string' && typeof user.first_name === 'string' && typeof user.last_name === 'string' && Object.hasOwn(userRoleLabels, user.role) && Object.hasOwn(userStatusLabels, user.status)
}
export async function listAdminUsers({ limit = 25, offset = 0, ...options } = {}) {
  if (!Number.isSafeInteger(limit) || limit < 1 || limit > 100 || !Number.isSafeInteger(offset) || offset < 0 || offset > 1000000) throw new Error('Neveljavna stran uporabnikov.')
  const response = await authorizedFetch('/admin/users?' + new URLSearchParams({ limit, offset }), options)
  const body = await response.json()
  if (!Array.isArray(body?.users) || body.users.length > limit || !body.users.every(validUser) || new Set(body.users.map(u => u.id)).size !== body.users.length || !Number.isSafeInteger(body.total) || body.total < body.users.length || body.limit !== limit || body.offset !== offset) throw new Error('Strežnik ni vrnil veljavnega seznama uporabnikov.')
  return body
}
export async function changeUserRole(id, role, { currentUserId, ...options } = {}) {
  if (typeof id !== 'string' || !/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(id)) throw new Error('Neveljaven uporabnik.')
  if (!Object.hasOwn(publicRoles, role)) throw new Error('Izberi veljavno vlogo.')
  if (id.toLowerCase() === currentUserId?.toLowerCase()) throw new Error('Svoje vloge ne moreš spremeniti.')
  const response = await authorizedFetch('/admin/users/' + encodeURIComponent(id) + '/role', {
    ...options, method: 'PATCH', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ role: publicRoles[role] }),
  })
  const body = await response.json()
  if (!validUser(body?.user) || body.user.id.toLowerCase() !== id.toLowerCase() || body.user.role !== role) throw new Error('Strežnik ni potrdil spremembe vloge. Osveži seznam.')
  return body.user
}
