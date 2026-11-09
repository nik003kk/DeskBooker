import assert from 'node:assert/strict'
import test from 'node:test'
import { navigation, paths, sections } from './routes.js'

test('all navigation paths are unique and absolute', () => {
  const routePaths = navigation.map(({ path }) => path)
  assert.equal(new Set(routePaths).size, routePaths.length)
  for (const path of routePaths) {
    assert.ok(path.startsWith('/'), `non-absolute path: ${path}`)
  }
})

test('primary navigation covers home, resources, reservations and login and registration', () => {
  assert.equal(navigation[0].path, paths.home)
  assert.equal(navigation[0].end, true)
  for (const path of [paths.desks, paths.rooms, paths.parking, paths.equipment, paths.reservations, paths.login, paths.register, paths.profile, paths.search]) {
    assert.ok(navigation.some((route) => route.path === path), `missing route ${path}`)
  }
})

test('booking confirmation is a private workflow route, not a navigation item', () => {
  assert.equal(paths.bookingConfirm, '/reservations/confirm')
  assert.ok(!navigation.some(route => route.path === paths.bookingConfirm))
})

test('booking success is a private workflow route, not in the main menu', () => {
  assert.equal(paths.bookingSuccess, '/reservations/success')
  assert.ok(!navigation.some(route => route.path === paths.bookingSuccess))
})

test('resource cards link to existing navigation paths', () => {
  const registered = new Set(navigation.map(({ path }) => path))
  assert.equal(sections.length, 4)
  for (const section of sections) {
    assert.ok(registered.has(section.path))
    assert.ok(section.label.trim().length > 0)
    assert.ok(section.description.trim().length > 0)
  }
})


test('check-in is a private reservation route outside the main menu', () => {
  assert.equal(paths.checkIn, '/reservations/:id/check-in')
  assert.ok(!navigation.some(route => route.path === paths.checkIn))
})

test('analytics navigation points to its manager workflow', () => {
  assert.equal(paths.analytics, '/analytics')
  assert.ok(navigation.some(route => route.path === paths.analytics))
})

test('live resource categories are bound to their API types', () => {
  for (const [path, type] of [[paths.desks, 'DESK'], [paths.rooms, 'MEETING_ROOM'], [paths.parking, 'PARKING_SPOT']]) {
    assert.equal(sections.find(section => section.path === path)?.resourceType, type)
  }
})
