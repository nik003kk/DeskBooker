import assert from 'node:assert/strict'
import test from 'node:test'
import { fetchAnalyticsCSV } from './analytics-export.js'
const period = { from: '2026-11-01', to: '2026-11-09' }

test('CSV uses the applied period and authenticated fetch', async () => {
  const result = await fetchAnalyticsCSV(period, { token: 'test', fetchImpl: async (url, options) => {
    assert.ok(url.endsWith('/analytics/export.csv?from=2026-11-01&to=2026-11-09'))
    assert.equal(options.headers.Authorization, 'Bearer test')
    return new Response('section,metric,date,weekday,value,unit\r\nsummary,total_reservations,,,6,count\r\n', { headers: { 'Content-Type': 'text/csv; charset=utf-8' } })
  } })
  assert.equal(result.filename, 'deskbooker-analytics-2026-11-01-2026-11-09.csv')
  assert.match(await result.blob.text(), /total_reservations/)
})
test('CSV rejects error documents and expired sessions', async () => {
  await assert.rejects(fetchAnalyticsCSV(period, { token: 'test', fetchImpl: async () => Response.json({ error: 'failed' }) }), /ni vrnil CSV/)
  let logout = false
  await assert.rejects(fetchAnalyticsCSV(period, { token: 'test', logout: () => { logout = true }, fetchImpl: async () => new Response('{}', { status: 401 }) }), /Seja je potekla/)
  assert.equal(logout, true)
})
