import assert from 'node:assert/strict'
import test from 'node:test'
import { previousAnalyticsPeriod, percentagePointChange, reservationChange } from './analytics-periods.js'

test('comparison uses equal civil-day ranges across DST, leap day and year boundary', () => {
  assert.deepEqual(previousAnalyticsPeriod({ from: '2030-04-01', to: '2030-04-07' }), { from: '2030-03-25', to: '2030-03-31' })
  assert.deepEqual(previousAnalyticsPeriod({ from: '2028-03-01', to: '2028-03-01' }), { from: '2028-02-29', to: '2028-02-29' })
  assert.deepEqual(previousAnalyticsPeriod({ from: '2030-01-01', to: '2030-01-07' }), { from: '2029-12-25', to: '2029-12-31' })
  assert.equal(percentagePointChange(null, 0), 'Ni primerljivih podatkov')
  assert.equal(percentagePointChange(25, 20), '+5 odstotne točke')
  assert.equal(reservationChange(0, 0), 'Brez spremembe')
  assert.match(reservationChange(5, 0), /Novo/)
  assert.equal(reservationChange(15, 10), '+50 %')
})
