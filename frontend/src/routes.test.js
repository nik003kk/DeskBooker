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

test('primary navigation covers home, resources, reservations and login', () => {
  assert.equal(navigation[0].path, paths.home)
  assert.equal(navigation[0].end, true)
  for (const path of [paths.desks, paths.rooms, paths.parking, paths.equipment, paths.reservations, paths.login]) {
    assert.ok(navigation.some((route) => route.path === path), `missing route ${path}`)
  }
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
