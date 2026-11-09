import assert from 'node:assert/strict'
import test from 'node:test'
import { analyticsQuery, defaultAnalyticsPeriod, formatPercent, loadAnalytics } from './analytics-api.js'
import { analyticsFixture } from '../scripts/analytics-fixture.mjs'
const period = { from: '2026-11-03', to: '2026-11-09' }

test('analytics dates use the office timezone and civil days', () => {
  assert.deepEqual(defaultAnalyticsPeriod(7, new Date('2026-11-08T23:30:00Z')), period)
  assert.equal(analyticsQuery(period), 'from=2026-11-03&to=2026-11-09')
  for (const value of [{ from: '2026-02-30', to: '2026-03-01' }, { from: '2026-11-10', to: '2026-11-09' }, { from: '2020-01-01', to: '2026-01-01' }]) assert.throws(() => analyticsQuery(value), /veljavno obdobje/)
  assert.equal(formatPercent(null), '—')
})
test('analytics fetch uses JWT and verifies the selected period', async () => {
  const result = await loadAnalytics(period, { token: 'test', fetchImpl: async (url, options) => {
    assert.ok(url.endsWith('/analytics?from=2026-11-03&to=2026-11-09'))
    assert.equal(options.headers.Authorization, 'Bearer test')
    return Response.json({ analytics: analyticsFixture(period) })
  } })
  assert.equal(result.total_reservations, 6)
  for (const mutation of [r => { r.to = '2026-11-10' }, r => { r.total_reservations = 999 }, r => { r.occupancy_rate = 101 }, r => { r.status_counts.EXPIRED = -1 }]) {
    const report = analyticsFixture(period); mutation(report)
    await assert.rejects(loadAnalytics(period, { token: 'test', fetchImpl: async () => Response.json({ analytics: report }) }), /veljavnih statističnih/)
  }
})
test('empty analytics and expired authentication are handled explicitly', async () => {
  const result = await loadAnalytics(period, { token: 'test', fetchImpl: async () => Response.json({ analytics: analyticsFixture(period, true) }) })
  assert.equal(result.no_show.rate, null)
  let logout = false
  await assert.rejects(loadAnalytics(period, { token: 'test', logout: () => { logout = true }, fetchImpl: async () => new Response('{}', { status: 401 }) }), /Seja je potekla/)
  assert.equal(logout, true)
})
