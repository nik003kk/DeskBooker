import assert from 'node:assert/strict'
import test from 'node:test'
import { mkdtemp, writeFile, rm } from 'node:fs/promises'
import { pathToFileURL } from 'node:url'
import { build } from 'esbuild'
import { JSDOM } from 'jsdom'

test('returning to the app refreshes displayed rights and clears suspended sessions', async () => {
  const dom = new JSDOM('<!doctype html><div id="root"></div>', { url: 'http://localhost/', pretendToBeVisual: true })
  globalThis.window = dom.window; globalThis.document = dom.window.document; globalThis.IS_REACT_ACT_ENVIRONMENT = true
  let live = { id: 'current-user', email: 'current@example.test', first_name: 'Current', role: 'admin', status: 'active' }
  window.sessionStorage.setItem('deskbooker-auth-session', JSON.stringify({ access_token: 'test', expires_at: new Date(Date.now() + 600000).toISOString(), user: live }))
  const React = await import('react'); const { act } = React; const { createRoot } = await import('react-dom/client')
  const oldFetch = globalThis.fetch
  globalThis.fetch = async () => live.status === 'active' ? Response.json({ user: live }) : new Response('{}', { status: 403 })
  const directory = await mkdtemp(new URL('../node_modules/.current-rights-', import.meta.url))
  let root
  try {
    const output = await build({ entryPoints: ['src/AuthContext.jsx'], bundle: true, packages: 'external', platform: 'node', format: 'esm', jsx: 'automatic', write: false, define: { 'import.meta.env.VITE_API_BASE_URL': 'undefined' } })
    const path = directory + '/context.mjs'; await writeFile(path, output.outputFiles[0].text)
    const { AuthProvider, useAuth } = await import(pathToFileURL(path).href)
    function Probe() { const { user } = useAuth(); return React.createElement('p', { id: 'role' }, user?.role || 'anonymous') }
    root = createRoot(document.getElementById('root'))
    await act(async () => root.render(React.createElement(AuthProvider, null, React.createElement(Probe))))
    assert.equal(document.getElementById('role').textContent, 'admin')
    live = { ...live, role: 'employee' }
    await act(async () => window.dispatchEvent(new dom.window.Event('focus')))
    assert.equal(document.getElementById('role').textContent, 'employee')
    assert.equal(JSON.parse(window.sessionStorage.getItem('deskbooker-auth-session')).user.role, 'employee')
    live = { ...live, status: 'suspended' }
    await act(async () => window.dispatchEvent(new dom.window.Event('focus')))
    assert.equal(document.getElementById('role').textContent, 'anonymous')
    assert.equal(window.sessionStorage.getItem('deskbooker-auth-session'), null)
  } finally { if (root) await act(async () => root.unmount()); globalThis.fetch = oldFetch; dom.window.close(); await rm(directory, { recursive: true, force: true }) }
})
