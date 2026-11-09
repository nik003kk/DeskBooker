import assert from 'node:assert/strict'
import test from 'node:test'
import { mkdtemp, writeFile, rm } from 'node:fs/promises'
import { pathToFileURL } from 'node:url'
import { build } from 'esbuild'
import { JSDOM } from 'jsdom'

const fixtures = [
  { id: '11111111-1111-4111-8111-111111111111', name: 'Miza iz baze', type: 'DESK', floor: 0, capacity: 1, is_active: true, features: ['USB-C'] },
  { id: '22222222-2222-4222-8222-222222222222', name: 'Soba iz baze', type: 'MEETING_ROOM', floor: 2, capacity: 6, is_active: true, features: ['Projektor'] },
  { id: '33333333-3333-4333-8333-333333333333', name: 'Parkirišče iz baze', type: 'PARKING_SPOT', floor: 0, capacity: 1, is_active: true, features: [] },
]
async function harness(pathname, fetchImpl, check, authenticated = true) {
  const dom = new JSDOM('<!doctype html><div id="root"></div>', { url: 'http://localhost/', pretendToBeVisual: true })
  globalThis.window = dom.window; globalThis.document = dom.window.document; globalThis.IS_REACT_ACT_ENVIRONMENT = true
  const React = await import('react'); const { act } = React
  const { createRoot } = await import('react-dom/client'); const { MemoryRouter } = await import('react-router-dom')
  const directory = await mkdtemp(new URL('../node_modules/.resource-categories-', import.meta.url))
  const previousFetch = globalThis.fetch; globalThis.fetch = fetchImpl
  let root
  try {
    const bundled = await build({ entryPoints: ['src/App.jsx'], bundle: true, packages: 'external', platform: 'node', format: 'esm', jsx: 'automatic', write: false,
      define: { 'import.meta.env': '{}' }, plugins: [{ name: 'employee-session', setup(builder) {
        builder.onResolve({ filter: /AuthContext\.jsx$/ }, () => ({ path: 'auth', namespace: 'test-auth' }))
        builder.onLoad({ filter: /.*/, namespace: 'test-auth' }, () => ({ contents: `const logout=()=>{}; const user={id:'employee-user',first_name:'Test',role:'employee',status:'active'};export function useAuth(){return {token:'test',logout,user,isAuthenticated:${authenticated},checking:false};}` }))
      } }] })
    const path = directory + '/app.mjs'; await writeFile(path, bundled.outputFiles[0].text)
    const { default: App } = await import(pathToFileURL(path).href)
    root = createRoot(document.getElementById('root'))
    await act(async () => root.render(React.createElement(MemoryRouter, { initialEntries: [pathname] }, React.createElement(App))))
    await check({ act, dom })
  } finally { if (root) await act(async () => root.unmount()); globalThis.fetch = previousFetch; dom.window.close(); await rm(directory, { recursive: true, force: true }) }
}
function responseFor(url) {
  const type = new URL(url).searchParams.get('type')
  return Response.json({ resources: fixtures.filter(resource => !type || resource.type === type) })
}

test('all three resource routes render only their real API category and preserve it on reset', async () => {
  for (const [pathname, fixture] of [['/desks', fixtures[0]], ['/rooms', fixtures[1]], ['/parking', fixtures[2]]]) {
    const requested = []
    await harness(pathname, async (url, options) => { requested.push(url); assert.equal(options.headers.Authorization, 'Bearer test'); return responseFor(url) }, async ({ act }) => {
      assert.equal(document.querySelectorAll('.resource-result-card').length, 1)
      assert.equal(document.querySelector('.resource-result-card h3').textContent, fixture.name)
      assert.ok(new URL(requested[0]).searchParams.get('type') === fixture.type)
      assert.equal(document.querySelector('select').disabled, true)
      assert.equal(document.querySelector('.resource-result-card button').disabled, true)
      assert.doesNotMatch(document.querySelector('main').textContent, /Demo prikaz|Miza D-101|Rezerviraj \(kmalu\)/)
      await act(async () => [...document.querySelectorAll('button')].find(button => button.textContent === 'Ponastavi filtre').click())
      assert.equal(new URL(requested.at(-1)).searchParams.get('type'), fixture.type)
      assert.equal(document.querySelector('select').value, fixture.type)
    })
  }
})
test('general resource search still lists every category and allows changing type', async () => {
  await harness('/search', async url => responseFor(url), async () => {
    assert.equal(document.querySelectorAll('.resource-result-card').length, 3)
    assert.equal(document.querySelector('select').disabled, false)
  })
})

test('returning to a category refreshes edits and manual refresh removes deactivated resources', async () => {
  let name = 'Soba iz baze'; let active = true
  const requested = []
  await harness('/rooms', async url => { requested.push(url); return Response.json({ resources: active ? [{ ...fixtures[1], name }] : [] }) }, async ({ act, dom }) => {
    name = 'Soba po urejanju'
    await act(async () => window.dispatchEvent(new dom.window.Event('focus')))
    assert.equal(document.querySelector('.resource-result-card h3').textContent, name)
    active = false
    await act(async () => [...document.querySelectorAll('button')].find(button => button.textContent === 'Osveži vire').click())
    assert.equal(document.querySelectorAll('.resource-result-card').length, 0)
    assert.match(document.querySelector('.resource-search-results').textContent, /Ni aktivnih virov/)
    assert.ok(requested.every(url => new URL(url).searchParams.get('type') === 'MEETING_ROOM'))
  })
})
test('failed category loads have no demo fallback and retry the same type', async () => {
  let attempts = 0
  await harness('/parking', async url => ++attempts === 1 ? new Response('{}', { status: 500 }) : responseFor(url), async ({ act }) => {
    assert.equal(document.querySelectorAll('.resource-result-card').length, 0)
    assert.match(document.querySelector('[role="alert"]').textContent, /Iskanje ni uspelo/)
    await act(async () => [...document.querySelectorAll('button')].find(button => button.textContent === 'Poskusi znova').click())
    assert.equal(document.querySelector('.resource-result-card h3').textContent, fixtures[2].name)
    assert.equal(document.querySelector('select').value, 'PARKING_SPOT')
  })
})

async function fillInput(context, selector, value) {
  await context.act(async () => {
    const input = document.querySelector(selector)
    Object.getOwnPropertyDescriptor(context.dom.window.HTMLInputElement.prototype, 'value').set.call(input, value)
    input.dispatchEvent(new context.dom.window.Event('input', { bubbles: true }))
    input.dispatchEvent(new context.dom.window.Event('change', { bubbles: true }))
  })
}
async function applyTerm(context) {
  await fillInput(context, 'input[type="date"]', '2099-01-05')
  await fillInput(context, 'input[type="time"]', '09:00')
  // Time fields each have their own wrapper; use the second time input.
  await context.act(async () => {
    const input = document.querySelectorAll('input[type="time"]')[1]
    Object.getOwnPropertyDescriptor(context.dom.window.HTMLInputElement.prototype, 'value').set.call(input, '10:00')
    input.dispatchEvent(new context.dom.window.Event('input', { bubbles: true }))
    input.dispatchEvent(new context.dom.window.Event('change', { bubbles: true }))
  })
  await context.act(async () => document.querySelector('form[role="search"]').dispatchEvent(new context.dom.window.Event('submit', { bubbles: true, cancelable: true })))
}

test('each category reserves the actual resource with the latest applied term', async () => {
  for (const [pathname, fixture] of [['/desks', fixtures[0]], ['/rooms', fixtures[1]], ['/parking', fixtures[2]]]) {
    const requested = []; let booking = null
    await harness(pathname, async (url, options) => {
      if (options.method === 'POST') {
        booking = JSON.parse(options.body)
        return Response.json({ reservation: { id: 'booking-' + fixture.id, resource_id: booking.resource_id, starts_at: booking.starts_at, ends_at: booking.ends_at, status: 'CONFIRMED' }, check_in_pin: '0042' }, { status: 201 })
      }
      requested.push(new URL(url)); return responseFor(url)
    }, async context => {
      await applyTerm(context)
      assert.equal(document.querySelector('.resource-result-card button').disabled, false)
      assert.equal(requested.at(-1).searchParams.get('date'), '2099-01-05')
      assert.equal(requested.at(-1).searchParams.get('end'), '10:00')
      await context.act(async () => {
        const input = document.querySelectorAll('input[type="time"]')[1]
        Object.getOwnPropertyDescriptor(context.dom.window.HTMLInputElement.prototype, 'value').set.call(input, '11:00')
        input.dispatchEvent(new context.dom.window.Event('input', { bubbles: true }))
        input.dispatchEvent(new context.dom.window.Event('change', { bubbles: true }))
      })
      assert.equal(document.querySelector('.resource-result-card button').disabled, true)
      await context.act(async () => [...document.querySelectorAll('button')].find(button => button.textContent === 'Osveži vire').click())
      assert.equal(requested.at(-1).searchParams.get('end'), '10:00')
      assert.equal(document.querySelector('.resource-result-card button').disabled, true)
      await context.act(async () => document.querySelector('form[role="search"]').dispatchEvent(new context.dom.window.Event('submit', { bubbles: true, cancelable: true })))
      assert.equal(requested.at(-1).searchParams.get('end'), '11:00')
      await context.act(async () => document.querySelector('.resource-result-card button').click())
      assert.match(document.querySelector('.booking-confirm-card').textContent, new RegExp(fixture.name))
      assert.match(document.querySelector('.booking-confirm-card').textContent, /2099-01-05/)
      assert.match(document.querySelector('.booking-confirm-card').textContent, /11:00/)
      await context.act(async () => [...document.querySelectorAll('button')].find(button => button.textContent === 'Potrdi rezervacijo').click())
      assert.equal(booking.resource_id, fixture.id)
      assert.equal(booking.starts_at, '2099-01-05T08:00:00.000Z')
      assert.equal(booking.ends_at, '2099-01-05T10:00:00.000Z')
      assert.match(document.querySelector('.booking-success-card').textContent, new RegExp(fixture.name))
      assert.equal(document.querySelector('.booking-pin').textContent, '0042')
    })
  }
})

test('retry preserves the submitted term and equipment', async () => {
  const requested = []; let fail = false
  await harness('/desks', async url => { requested.push(url); if (fail) { fail = false; return new Response('{}', { status: 500 }) } return responseFor(url) }, async context => {
    await applyTerm(context)
    await fillInput(context, 'input[placeholder="Monitor, USB-C"]', 'USB-C')
    await context.act(async () => document.querySelector('form[role="search"]').dispatchEvent(new context.dom.window.Event('submit', { bubbles: true, cancelable: true })))
    const applied = requested.at(-1)
    fail = true
    await context.act(async () => [...document.querySelectorAll('button')].find(button => button.textContent === 'Osveži vire').click())
    assert.match(document.querySelector('[role="alert"]').textContent, /Iskanje ni uspelo/)
    await context.act(async () => [...document.querySelectorAll('button')].find(button => button.textContent === 'Poskusi znova').click())
    assert.equal(requested.at(-1), applied)
    assert.equal(document.querySelector('.resource-result-card button').disabled, false)
  })
})

test('late response from a previous category cannot replace the new category', async () => {
  let hold = false; let release
  await harness('/desks', async url => {
    if (hold && new URL(url).searchParams.get('type') === 'DESK') return new Promise(resolve => { release = () => resolve(responseFor(url)) })
    return responseFor(url)
  }, async context => {
    await applyTerm(context)
    hold = true
    await context.act(async () => [...document.querySelectorAll('button')].find(button => button.textContent === 'Osveži vire').click())
    await context.act(async () => document.querySelector('.site-nav a[href="/rooms"]').click())
    assert.equal(document.querySelector('.resource-result-card h3').textContent, fixtures[1].name)
    assert.equal(document.querySelector('input[type="date"]').value, '')
    assert.equal(document.querySelector('.resource-result-card button').disabled, true)
    await context.act(async () => release())
    assert.equal(document.querySelector('.resource-result-card h3').textContent, fixtures[1].name)
    assert.equal(document.querySelector('select').value, 'MEETING_ROOM')
  })
})

test('resource categories require a signed-in user before reading API data', async () => {
  await harness('/parking', () => assert.fail('anonymous category must not fetch resources'), async () => {
    assert.equal(document.querySelectorAll('.resource-result-card').length, 0)
    assert.ok(document.querySelector('.auth-card'))
    assert.match(document.querySelector('h1').textContent, /Prijav/)
  }, false)
})
