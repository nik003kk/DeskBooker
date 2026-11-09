import assert from 'node:assert/strict'
import test from 'node:test'
import { mkdtemp, writeFile, rm } from 'node:fs/promises'
import { pathToFileURL } from 'node:url'
import { build } from 'esbuild'
import { JSDOM } from 'jsdom'

test('check-in modal preserves PIN, handles failures, prevents duplicate submissions and restores focus', async () => {
  const dom = new JSDOM('<!doctype html><button id="opener">Odpri check-in</button><div id="root"></div>', { url: 'http://localhost/' })
  globalThis.window = dom.window
  globalThis.document = dom.window.document
  globalThis.IS_REACT_ACT_ENVIRONMENT = true
  dom.window.HTMLDialogElement.prototype.showModal = function () { this.setAttribute('open', '') }
  dom.window.HTMLDialogElement.prototype.close = function () { this.removeAttribute('open') }
  const React = await import('react')
  const { act } = React
  const { createRoot } = await import('react-dom/client')
  const directory = await mkdtemp(new URL('../node_modules/.check-in-ui-', import.meta.url))
  const previousFetch = globalThis.fetch
  let root
  try {
    const bundle = await build({ entryPoints: ['src/components/CheckInModal.jsx'], bundle: true, packages: 'external', platform: 'node', format: 'esm', jsx: 'automatic', write: false })
    const modulePath = directory + '/modal.mjs'
    await writeFile(modulePath, bundle.outputFiles[0].text)
    const { default: CheckInModal } = await import(pathToFileURL(modulePath).href)
    const opener = document.getElementById('opener')
    opener.focus()
    let saved = null
    const id = '11111111-1111-4111-8111-111111111111'
    root = createRoot(document.getElementById('root'))
    await act(async () => root.render(React.createElement(CheckInModal, {
      reservation: { id, status: 'CONFIRMED', starts_at: new Date(Date.now() + 300000).toISOString(), ends_at: new Date(Date.now() + 3600000).toISOString() }, options: { token: 'test-token' }, onClose: () => {}, onSuccess: value => { saved = value },
    })))
    const inputs = [...document.querySelectorAll('input')]
    assert.equal(inputs.length, 4)
    assert.equal(document.activeElement, inputs[0])
    assert.ok(document.querySelector('dialog').hasAttribute('open'))
    assert.equal(document.querySelector('button[type="submit"]').disabled, true)
    async function paste(value) {
      await act(async () => {
        const event = new dom.window.Event('paste', { bubbles: true, cancelable: true })
        Object.defineProperty(event, 'clipboardData', { value: { getData: () => value } })
        inputs[0].dispatchEvent(event)
      })
    }
    await paste('0042')
    assert.equal(inputs.map(input => input.value).join(''), '0042')
    assert.equal(document.querySelector('button[type="submit"]').disabled, false)
    let requests = 0
    globalThis.fetch = async (_, request) => {
      requests++
      assert.equal(request.headers.Authorization, 'Bearer test-token')
      assert.deepEqual(JSON.parse(request.body), { pin: '0042' })
      return new Response('{}', { status: 422 })
    }
    await act(async () => document.querySelector('form').dispatchEvent(new dom.window.Event('submit', { bubbles: true, cancelable: true })))
    assert.match(document.querySelector('[role="alert"]').textContent, /Napačna PIN koda/)
    assert.equal(inputs.map(input => input.value).join(''), '')
    assert.equal(document.activeElement, inputs[0])
    assert.equal(saved, null)
    await paste('0042')
    globalThis.fetch = async () => {
      requests++
      return Response.json({ reservation: { id, status: 'CHECKED_IN', starts_at: '2030-01-01T09:00:00Z', ends_at: '2030-01-01T10:00:00Z', checked_in_at: '2030-01-01T09:00:00Z' } })
    }
    await act(async () => {
      const form = document.querySelector('form')
      form.dispatchEvent(new dom.window.Event('submit', { bubbles: true, cancelable: true }))
      form.dispatchEvent(new dom.window.Event('submit', { bubbles: true, cancelable: true }))
    })
    assert.equal(requests, 2)
    assert.equal(saved.status, 'CHECKED_IN')
    await act(async () => root.render(React.createElement(CheckInModal, {
      reservation: { id, status: 'CONFIRMED', starts_at: new Date(Date.now() - 3600000).toISOString(), ends_at: new Date(Date.now() + 3600000).toISOString() },
      options: { token: 'test-token' }, onClose: () => {}, onSuccess: () => assert.fail('closed window accepted'),
    })))
    assert.match(document.querySelector('.check-in-countdown').textContent, /potekel/)
    assert.equal(document.querySelector('button[type="submit"]').disabled, true)
    await act(async () => document.querySelector('form').dispatchEvent(new dom.window.Event('submit', { bubbles: true, cancelable: true })))
    assert.equal(requests, 2)
    await act(async () => root.unmount())
    root = null
    assert.equal(document.activeElement, opener)
  } finally {
    if (root) await act(async () => root.unmount())
    globalThis.fetch = previousFetch
    dom.window.close()
    await rm(directory, { recursive: true, force: true })
  }
})

test('confirmed presence comes from server status and timestamp', async () => {
  const React = await import('react')
  const { renderToStaticMarkup } = await import('react-dom/server')
  const directory = await mkdtemp(new URL('../node_modules/.check-in-status-', import.meta.url))
  try {
    const bundle = await build({ entryPoints: ['src/components/CheckInStatus.jsx'], bundle: true, packages: 'external', platform: 'node', format: 'esm', jsx: 'automatic', write: false })
    const modulePath = directory + '/status.mjs'
    await writeFile(modulePath, bundle.outputFiles[0].text)
    const { default: CheckInStatus } = await import(pathToFileURL(modulePath).href)
    const reservation = { status: 'CHECKED_IN', checked_in_at: '2030-01-01T09:00:00Z' }
    const result = renderToStaticMarkup(React.createElement(CheckInStatus, { reservation }))
    assert.match(result, /Prisotnost potrjena/)
    assert.match(result, /status-badge--green/)
    assert.match(result, /CHECKED_IN/)
    assert.match(result, /dateTime="2030-01-01T09:00:00Z"/)
    assert.match(result, /10:00/)
    for (const value of [{ status: 'CONFIRMED' }, { status: 'CHECKED_IN' }, { status: 'CANCELLED', checked_in_at: reservation.checked_in_at }]) {
      const html = renderToStaticMarkup(React.createElement(CheckInStatus, { reservation: value }))
      assert.doesNotMatch(html, /Prisotnost potrjena/)
      assert.doesNotMatch(html, /status-badge--green/)
    }
  } finally {
    await rm(directory, { recursive: true, force: true })
  }
})

test('check-in page confirms presence and reloads persisted confirmation without PIN', async () => {
  const dom = new JSDOM('<!doctype html><div id="root"></div>', { url: 'http://localhost/' })
  globalThis.window = dom.window
  globalThis.document = dom.window.document
  globalThis.IS_REACT_ACT_ENVIRONMENT = true
  dom.window.HTMLDialogElement.prototype.showModal = function () { this.setAttribute('open', '') }
  dom.window.HTMLDialogElement.prototype.close = function () { this.removeAttribute('open') }
  const React = await import('react')
  const { act } = React
  const { createRoot } = await import('react-dom/client')
  const { MemoryRouter, Route, Routes } = await import('react-router-dom')
  const directory = await mkdtemp(new URL('../node_modules/.check-in-page-', import.meta.url))
  const previousFetch = globalThis.fetch
  let root
  try {
    const bundle = await build({
      entryPoints: ['src/components/CheckInPage.jsx'], bundle: true, packages: 'external', platform: 'node', format: 'esm', jsx: 'automatic', write: false,
      define: { 'import.meta.env.VITE_API_BASE_URL': 'undefined' },
      plugins: [{ name: 'authenticated-user', setup(builder) {
        builder.onResolve({ filter: /AuthContext\.jsx$/ }, () => ({ path: 'test-auth', namespace: 'test-auth' }))
        builder.onLoad({ filter: /.*/, namespace: 'test-auth' }, () => ({ contents: 'const logout = () => {}; export function useAuth() { return { token: "test-token", logout }; }' }))
      } }],
    })
    const modulePath = directory + '/page.mjs'
    await writeFile(modulePath, bundle.outputFiles[0].text)
    const { default: CheckInPage } = await import(pathToFileURL(modulePath).href)
    const id = '11111111-1111-4111-8111-111111111111'
    let reservation = { id, status: 'CONFIRMED', starts_at: new Date(Date.now() + 300000).toISOString(), ends_at: new Date(Date.now() + 3600000).toISOString() }
    let posts = 0
    globalThis.fetch = async (url, request) => {
      if (request.method === 'POST') {
        posts++
        assert.ok(url.endsWith('/check-in'))
        assert.deepEqual(JSON.parse(request.body), { pin: '0042' })
        reservation = { ...reservation, status: 'CHECKED_IN', checked_in_at: new Date().toISOString() }
        return Response.json({ reservation })
      }
      return Response.json({ reservation, resource: { name: 'Miza za check-in' } })
    }
    const view = () => React.createElement(MemoryRouter, { initialEntries: ['/reservations/' + id + '/check-in'] },
      React.createElement(Routes, null, React.createElement(Route, { path: '/reservations/:id/check-in', element: React.createElement(CheckInPage) })))
    root = createRoot(document.getElementById('root'))
    await act(async () => root.render(view()))
    assert.match(document.body.textContent, /Miza za check-in/)
    const button = [...document.querySelectorAll('button')].find(element => element.textContent === 'Potrdi prihod s PIN-om')
    assert.equal(button.disabled, false)
    await act(async () => button.click())
    await act(async () => {
      const event = new dom.window.Event('paste', { bubbles: true, cancelable: true })
      Object.defineProperty(event, 'clipboardData', { value: { getData: () => '0042' } })
      document.querySelector('input').dispatchEvent(event)
    })
    await act(async () => document.querySelector('form').dispatchEvent(new dom.window.Event('submit', { bubbles: true, cancelable: true })))
    assert.equal(posts, 1)
    assert.equal(document.querySelector('dialog'), null)
    assert.match(document.querySelector('.status-badge--green').textContent, /Prisotnost potrjena/)
    await act(async () => root.unmount())
    root = createRoot(document.getElementById('root'))
    await act(async () => root.render(view()))
    assert.match(document.querySelector('.status-badge--green').textContent, /Prisotnost potrjena/)
    assert.ok(document.querySelector('time'))
    assert.doesNotMatch(document.body.textContent, /0042|Potrdi prihod s PIN-om/)
    assert.equal(posts, 1)
  } finally {
    if (root) await act(async () => root.unmount())
    globalThis.fetch = previousFetch
    dom.window.close()
    await rm(directory, { recursive: true, force: true })
  }
})
