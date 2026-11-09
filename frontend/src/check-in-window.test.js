import assert from 'node:assert/strict'
import test from 'node:test'
import { checkInWindow, checkInWindowLabel, formatRemaining } from './check-in-window.js'
const start = Date.parse('2030-01-01T09:00:00Z')
const reservation = { status: 'CONFIRMED', starts_at: new Date(start).toISOString(), ends_at: new Date(start + 3600_000).toISOString() }

test('opening and closing boundaries match backend and use instants', () => {
  for (const [offset, state] of [[-600001, 'waiting'], [-600000, 'open'], [0, 'open'], [900000, 'open'], [900001, 'closed']]) {
    const result = checkInWindow(reservation, start + offset)
    assert.equal(result.state, state)
    assert.equal(result.canSubmit, state === 'open')
  }
  assert.equal(checkInWindow({ ...reservation, starts_at: '2030-01-01T10:00:00+01:00' }, start).state, 'open')
})

test('short reservation closes at its end and terminal statuses stay disabled', () => {
  assert.equal(checkInWindow({ ...reservation, ends_at: new Date(start + 60000).toISOString() }, start + 60000).state, 'closed')
  for (const status of ['CANCELLED', 'COMPLETED', 'EXPIRED', 'PENDING', 'CHECKED_IN']) {
    assert.equal(checkInWindow({ ...reservation, status }, start).canSubmit, false)
  }
  assert.equal(checkInWindow(null).canSubmit, false)
  assert.equal(checkInWindow({ ...reservation, starts_at: 'bad' }, start).canSubmit, false)
})

test('countdown rounds remaining seconds up and never becomes negative', () => {
  assert.equal(checkInWindow(reservation, start - 600001).seconds, 1)
  assert.equal(checkInWindow(reservation, start + 900001).seconds, 0)
  assert.equal(formatRemaining(65), '1:05')
  assert.equal(formatRemaining(3661), '1 h 1 min')
  assert.equal(formatRemaining(90000), '1 d 1 h')
  assert.match(checkInWindowLabel(checkInWindow(reservation, start)), /15:00/)
  assert.match(checkInWindowLabel(checkInWindow(reservation, start + 900001)), /potekel/)
})
