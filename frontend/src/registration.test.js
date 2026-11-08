import assert from 'node:assert/strict'
import test from 'node:test'
import { registrationErrorMessage, submitRegistration, validateRegistration } from './registration.js'

const valid = {
  first_name: 'Ana',
  last_name: 'Novak',
  email: 'Ana@Gmail.COM',
  password: 'VeryStrongPass123!',
  confirm_password: 'VeryStrongPass123!',
}

test('registration validates names, domain, password and confirmation', () => {
  assert.deepEqual(validateRegistration(valid, 'gmail.com'), {})
  assert.ok(validateRegistration({ ...valid, first_name: ' ' }, 'gmail.com').first_name)
  assert.ok(validateRegistration({ ...valid, last_name: 'X'.repeat(101) }, 'gmail.com').last_name)
  assert.ok(validateRegistration({ ...valid, email: 'ana@example.com' }, 'gmail.com').email)
  assert.ok(validateRegistration({ ...valid, email: 'not-an-email' }, 'gmail.com').email)
  assert.ok(validateRegistration({ ...valid, password: 'short' }, 'gmail.com').password)
  assert.ok(validateRegistration({ ...valid, password: 'ž'.repeat(37) }, 'gmail.com').password)
  assert.ok(validateRegistration({ ...valid, confirm_password: 'WrongPassword123' }, 'gmail.com').confirm_password)
})

test('request sends only supported backend fields', async () => {
  let recorded
  const user = await submitRegistration(valid, {
    apiBaseUrl: 'http://localhost:8080/api/',
    fetchImpl: async (url, options) => {
      recorded = { url, options }
      return { ok: true, json: async () => ({ user: { id: 'user-1', email: 'ana@gmail.com' } }) }
    },
  })
  assert.deepEqual(user, { id: 'user-1', email: 'ana@gmail.com' })
  assert.equal(recorded.url, 'http://localhost:8080/api/auth/register')
  assert.equal(recorded.options.method, 'POST')
  assert.equal(recorded.options.headers['Content-Type'], 'application/json')
  assert.deepEqual(JSON.parse(recorded.options.body), {
    first_name: valid.first_name,
    last_name: valid.last_name,
    email: 'ana@gmail.com',
    password: valid.password,
  })
  assert.ok(!recorded.options.body.includes('confirm_password'))
})

test('duplicate email and server errors become readable messages', async () => {
  await assert.rejects(
    () => submitRegistration(valid, {
      fetchImpl: async () => ({ ok: false, status: 409, json: async () => ({ error: { code: 'email_exists' } }) }),
    }),
    /že registriran/,
  )
  assert.match(registrationErrorMessage(400, 'invalid_email'), /domeno/)
  assert.match(registrationErrorMessage(503), /ni na voljo/)
  assert.match(registrationErrorMessage(500), /ni uspela/)
})

test('unexpected success response cannot create false success', async () => {
  await assert.rejects(
    () => submitRegistration(valid, { fetchImpl: async () => ({ ok: true, json: async () => ({}) }) }),
    /neveljaven odgovor/,
  )
})
