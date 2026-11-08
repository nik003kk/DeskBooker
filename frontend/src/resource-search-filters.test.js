import assert from 'node:assert/strict'
import test from 'node:test'
import { emptySearchFilters, parseSearchFeatures, validateSearchFilters } from './resource-search-filters.js'

test('empty filters and valid time window are supported', () => {
  assert.deepEqual(validateSearchFilters(emptySearchFilters), {})
  assert.deepEqual(validateSearchFilters({
    type:'DESK',date:'2026-11-09',start:'09:00',end:'17:00',features:'Monitor, USB-C',
  }), {})
})
test('invalid date, partial time and reversed interval fail', () => {
  for (const values of [
    {date:'2026-02-30',start:'09:00',end:'17:00'},
    {date:'2026-11-09',start:'',end:'17:00'},
    {date:'2026-11-09',start:'18:00',end:'09:00'},
  ]) assert.ok(validateSearchFilters({...emptySearchFilters,...values}).date)
})
test('feature names are split and validated', () => {
  assert.deepEqual(parseSearchFeatures('Monitor, USB-C'),['Monitor','USB-C'])
  assert.ok(validateSearchFilters({...emptySearchFilters,features:'Monitor,monitor'}).features)
  assert.ok(validateSearchFilters({...emptySearchFilters,features:'Monitor,'}).features)
  assert.ok(validateSearchFilters({...emptySearchFilters,features:'x'.repeat(101)}).features)
  assert.ok(validateSearchFilters({...emptySearchFilters,type:'NOT_A_TYPE'}).type)
})
