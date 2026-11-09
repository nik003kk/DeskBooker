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
      reservation: { id }, options: { token: 'test-token' }, onClose: () => {}, onSuccess: value => { saved = value },
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
