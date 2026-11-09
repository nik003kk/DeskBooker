import assert from 'node:assert/strict'
import test from 'node:test'
import { mkdtemp, writeFile, rm } from 'node:fs/promises'
import { pathToFileURL } from 'node:url'
import { build } from 'esbuild'
import { JSDOM } from 'jsdom'

const actor = 'aaaaaaaa-1111-4111-8111-111111111111'
const target = 'bbbbbbbb-1111-4111-8111-111111111111'
const accounts = [{ id: actor, email: 'admin@example.test', first_name: 'Nik', last_name: 'Admin', role: 'admin', status: 'active' }, { id: target, email: 'employee@example.test', first_name: 'Test', last_name: 'User', role: 'employee', status: 'active' }]
async function harness(allowed, fetchImpl, check) {
  const dom = new JSDOM('<!doctype html><div id="root"></div>', { url: 'http://localhost/' })
  globalThis.window = dom.window; globalThis.document = dom.window.document; globalThis.IS_REACT_ACT_ENVIRONMENT = true
  dom.window.HTMLDialogElement.prototype.showModal = function () { this.setAttribute('open', '') }
  dom.window.HTMLDialogElement.prototype.close = function () { this.removeAttribute('open') }
  const React = await import('react'); const { act } = React
  const { createRoot } = await import('react-dom/client'); const { MemoryRouter } = await import('react-router-dom')
  const directory = await mkdtemp(new URL('../node_modules/.admin-users-', import.meta.url))
  const oldFetch = globalThis.fetch; globalThis.fetch = fetchImpl
  let root
  try {
    const bundle = await build({ stdin: { contents: `import React from 'react'; import Page from './src/components/AdminUsersPage.jsx'; import Guard from './src/components/AdminUsersRoute.jsx'; export default function Test(){return <Guard><Page /></Guard>}`, resolveDir: process.cwd(), loader: 'jsx' }, bundle: true, packages: 'external', platform: 'node', format: 'esm', jsx: 'automatic', write: false,
      define: { 'import.meta.env.VITE_API_BASE_URL': 'undefined' }, plugins: [{ name: 'session', setup(builder) {
        builder.onResolve({ filter: /AuthContext\.jsx$/ }, () => ({ path: 'session', namespace: 'session' }))
        builder.onLoad({ filter: /.*/, namespace: 'session' }, () => ({ contents: `const logout=()=>{};const user=${JSON.stringify({ ...accounts[0], role: allowed ? 'admin' : 'manager' })};export function useAuth(){return {token:'test',logout,user,isAuthenticated:true,checking:false};}` }))
      } }] })
    const path = directory + '/page.mjs'; await writeFile(path, bundle.outputFiles[0].text)
    const { default: Page } = await import(pathToFileURL(path).href)
    root = createRoot(document.getElementById('root'))
    await act(async () => root.render(React.createElement(MemoryRouter, null, React.createElement(Page))))
    await check({ act, dom })
  } finally { if (root) await act(async () => root.unmount()); globalThis.fetch = oldFetch; dom.window.close(); await rm(directory, { recursive: true, force: true }) }
}

test('administrator can confirm another role once while own role stays disabled', async () => {
  let patches = 0
  await harness(true, async (_, options) => {
    if (options.method === 'PATCH') { patches++; assert.deepEqual(JSON.parse(options.body), { role: 'OFFICE_MANAGER' }); return Response.json({ user: { ...accounts[1], role: 'manager' } }) }
    return Response.json({ users: accounts, total: 2, limit: 25, offset: 0 })
  }, async ({ act, dom }) => {
    const selects = [...document.querySelectorAll('select')]
    assert.equal(selects[0].disabled, true)
    await act(async () => { selects[1].value = 'manager'; selects[1].dispatchEvent(new dom.window.Event('change', { bubbles: true })) })
    await act(async () => [...document.querySelectorAll('button')].find(b => b.textContent === 'Shrani vlogo').click())
    assert.ok(document.querySelector('dialog').hasAttribute('open')); assert.equal(patches, 0)
    await act(async () => { const confirm = [...document.querySelectorAll('button')].find(b => b.textContent === 'Potrdi spremembo'); confirm.click(); confirm.click() })
    assert.equal(patches, 1); assert.equal(document.querySelector('dialog'), null)
    assert.match(document.querySelector('.admin-users-success').textContent, /je shranjena/)
    assert.equal(document.querySelectorAll('select')[1].value, 'manager')
  })
})
test('manager cannot mount the user list or fetch any account', async () => {
  await harness(false, () => assert.fail('manager must not fetch users'), async () => {
    assert.match(document.querySelector('[role="alert"]').textContent, /aktivno vlogo administratorja/)
    assert.equal(document.querySelector('table'), null)
  })
})
test('failed role mutation keeps the original role and displays the modal error', async () => {
  await harness(true, async (_, options) => options.method === 'PATCH' ? new Response('{}', { status: 500 }) : Response.json({ users: accounts, total: 2, limit: 25, offset: 0 }), async ({ act, dom }) => {
    const select = document.querySelectorAll('select')[1]
    await act(async () => { select.value = 'admin'; select.dispatchEvent(new dom.window.Event('change', { bubbles: true })) })
    await act(async () => [...document.querySelectorAll('button')].find(b => b.textContent === 'Shrani vlogo').click())
    await act(async () => [...document.querySelectorAll('button')].find(b => b.textContent === 'Potrdi spremembo').click())
    assert.match(document.querySelector('dialog [role="alert"]').textContent, /ni uspela/)
    assert.equal(document.querySelector('.admin-users-success'), null)
    assert.match(document.querySelector('dialog').textContent, /Zaposleni → Administrator/)
  })
})
