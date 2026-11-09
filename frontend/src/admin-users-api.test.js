import assert from 'node:assert/strict'
import test from 'node:test'
import { canAdministerUsers, changeUserRole, listAdminUsers } from './admin-users-api.js'
const id = 'bbbbbbbb-1111-4111-8111-111111111111'
const account = { id, email: 'employee@example.test', first_name: 'Test', last_name: 'User', role: 'employee', status: 'active' }

test('user list requires a valid paginated response and JWT', async () => {
  const result = await listAdminUsers({ token: 'test', fetchImpl: async (url, options) => {
    assert.ok(url.endsWith('/admin/users?limit=25&offset=0')); assert.equal(options.headers.Authorization, 'Bearer test')
    return Response.json({ users: [account], total: 1, limit: 25, offset: 0 })
  } })
  assert.equal(result.users[0].role, 'employee')
  await assert.rejects(listAdminUsers({ token: 'test', fetchImpl: async () => Response.json({ users: [], total: 1, limit: 25, offset: 20 }) }), /veljavnega seznama/)
  assert.equal(canAdministerUsers({ role: 'admin', status: 'active' }), true)
  assert.equal(canAdministerUsers({ role: 'manager', status: 'active' }), false)
})
test('role change maps product roles and verifies the server result', async () => {
  const result = await changeUserRole(id, 'manager', { token: 'test', fetchImpl: async (_, options) => {
    assert.equal(options.method, 'PATCH'); assert.deepEqual(JSON.parse(options.body), { role: 'OFFICE_MANAGER' })
    return Response.json({ user: { ...account, role: 'manager' } })
  } })
  assert.equal(result.role, 'manager')
  await assert.rejects(changeUserRole(id, 'admin', { currentUserId: id.toUpperCase(), fetchImpl: () => assert.fail('self role must not fetch') }), /Svoje vloge/)
  await assert.rejects(changeUserRole(id, 'root', { fetchImpl: () => assert.fail('invalid role must not fetch') }), /veljavno vlogo/)
  await assert.rejects(changeUserRole(id, 'manager', { token: 'test', fetchImpl: async () => Response.json({ user: account }) }), /ni potrdil/)
})
test('expired JWT logs out and forbidden requests cannot report success', async () => {
  let loggedOut = false
  await assert.rejects(listAdminUsers({ token: 'test', logout: () => { loggedOut = true }, fetchImpl: async () => new Response('{}', { status: 401 }) }), /Seja je potekla/)
  assert.equal(loggedOut, true)
  await assert.rejects(changeUserRole(id, 'manager', { token: 'test', fetchImpl: async () => new Response('{}', { status: 403 }) }))
})
