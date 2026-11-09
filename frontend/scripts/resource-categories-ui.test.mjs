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
