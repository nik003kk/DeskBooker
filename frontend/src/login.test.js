import assert from 'node:assert/strict'
import test from 'node:test'
import { validateLogin, submitLogin } from './login.js'
test('validates email and password', () => {
  assert.ok(validateLogin({ email: '', password: '' }).email)
  assert.ok(validateLogin({ email: 'x@gmail.com', password: '' }).password)
  assert.deepEqual(validateLogin({ email: 'x@gmail.com', password: 'correct' }), {})
})
test('calls existing login endpoint and parses token', async () => {
  let request
  const result = await submitLogin({ email: 'Ana@Gmail.COM', password: 'test-password' }, {
    fetchImpl: async (url, options) => {
      request = { url, options }
      return { ok: true, json: async () => ({ user: { id: 'one' }, access_token: 'token', token_type: 'Bearer' }) }
    },
  })
  assert.equal(request.url, 'http://localhost:8080/api/auth/login')
  assert.deepEqual(JSON.parse(request.options.body), { email: 'ana@gmail.com', password: 'test-password' })
  assert.equal(result.access_token, 'token')
})
test('handles bad credentials and malformed successful response', async () => {
  await assert.rejects(() => submitLogin({email:'x@gmail.com', password:'wrong'}, {
    fetchImpl: async () => ({ ok: false, status: 401, json: async () => ({}) }),
  }), /Napačen/)
  await assert.rejects(() => submitLogin({email:'x@gmail.com', password:'wrong'}, {
    fetchImpl: async () => ({ ok: true, json: async () => ({}) }),
  }), /neveljaven odgovor/)
})
