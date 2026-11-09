import assert from 'node:assert/strict'
import test from 'node:test'
import { checkInPath, loadCheckIn, submitCheckIn } from './check-in-api.js'

const id = '11111111-1111-4111-8111-111111111111'
const reservation = { id, status: 'CONFIRMED', starts_at: '2030-01-01T09:00:00Z', ends_at: '2030-01-01T10:00:00Z' }
const options = { token: 'test-token' }

test('loads private detail with JWT and encodes reservation ID', async () => {
  assert.equal(checkInPath('a/b'), '/reservations/a%2Fb/check-in')
  const result = await loadCheckIn(id, { ...options, fetchImpl: async (url, request) => {
    assert.ok(url.endsWith('/reservations/' + id))
    assert.equal(request.headers.Authorization, 'Bearer test-token')
    return Response.json({ reservation, resource: { name: 'Miza 1' } })
  } })
  assert.equal(result.reservation.status, 'CONFIRMED')
})

test('preserves zero-prefixed PIN and accepts only server-confirmed check-in', async () => {
  const result = await submitCheckIn(id, '0042', { ...options, fetchImpl: async (url, request) => {
    assert.ok(url.endsWith('/check-in'))
    assert.equal(request.method, 'POST')
    assert.deepEqual(JSON.parse(request.body), { pin: '0042' })
    return Response.json({ reservation: { ...reservation, status: 'CHECKED_IN', checked_in_at: '2030-01-01T09:00:00Z' } })
  } })
  assert.equal(result.status, 'CHECKED_IN')
  await assert.rejects(submitCheckIn(id, '0042', { ...options, fetchImpl: async () => Response.json({ reservation }) }), /ni potrdil/)
})

test('invalid PIN never triggers a request', async () => {
  for (const pin of ['42', '12a4', '１２３４', '', '12345']) {
    await assert.rejects(submitCheckIn(id, pin, { ...options, fetchImpl: () => assert.fail('must not fetch') }), /štirimestno/)
  }
})

test('shows wrong PIN, eligibility and unavailable reservation errors', async () => {
  for (const [status, message] of [[422, /Napačna PIN koda/], [409, /trenutno ni dovoljen/], [404, /ni na voljo/]]) {
    await assert.rejects(submitCheckIn(id, '0042', { ...options, fetchImpl: async () => new Response('{}', { status }) }), message)
  }
})

test('expired JWT logs out and does not show confirmation', async () => {
  let loggedOut = false
  await assert.rejects(submitCheckIn(id, '0042', { ...options, logout: () => { loggedOut = true }, fetchImpl: async () => new Response('{}', { status: 401 }) }), /Seja je potekla/)
  assert.equal(loggedOut, true)
})

test('rejects mismatched reservation data and propagates network failures', async () => {
  await assert.rejects(loadCheckIn(id, { ...options, fetchImpl: async () => Response.json({ reservation: { ...reservation, id: 'other' }, resource: { name: 'Desk' } }) }), /veljavnih podatkov/)
  await assert.rejects(submitCheckIn(id, '0042', { ...options, fetchImpl: async () => { throw new TypeError('offline') } }), /offline/)
})
