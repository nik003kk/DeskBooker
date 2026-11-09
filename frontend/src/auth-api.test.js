import assert from 'node:assert/strict'
import test from 'node:test'
import { AuthAPIError, authorizedFetch } from './auth-api.js'

test('missing JWT is refused before fetching', async () => {
  let called = false
  let loggedOut = false
  await assert.rejects(() => authorizedFetch('/users/me', {
    token: '', logout: () => { loggedOut = true },
    fetchImpl: () => { called = true },
  }), error => error instanceof AuthAPIError && error.status === 401)
  assert.equal(called, false)
  assert.equal(loggedOut, true)
})
test('401 logs out but 403 preserves existing session', async () => {
  for (const status of [401,403]) {
    let loggedOut = false
    await assert.rejects(() => authorizedFetch('/users/me', {
      token: 'token', logout: () => { loggedOut = true },
      fetchImpl: async () => ({ ok:false, status }),
    }), error => error.status === status)
    assert.equal(loggedOut, status === 401)
  }
})
test('passes bearer token to protected request', async () => {
  let authorization
  const result = await authorizedFetch('/private', {
    token: 'token', fetchImpl: async (_, options) => {
      authorization = options.headers.Authorization
      return {ok:true,status:200}
    },
  })
  assert.equal(authorization, 'Bearer token')
  assert.equal(result.status, 200)
})
