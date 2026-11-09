import assert from 'node:assert/strict'
import test from 'node:test'
import { mkdtemp, writeFile, rm } from 'node:fs/promises'
import { pathToFileURL } from 'node:url'
import { build } from 'esbuild'
import { JSDOM } from 'jsdom'
import { analyticsFixture } from './analytics-fixture.mjs'

async function pageHarness(fetchImpl, check) {
  const dom = new JSDOM('<!doctype html><div id="root"></div>', { url: 'http://localhost/' })
  globalThis.window = dom.window; globalThis.document = dom.window.document; globalThis.IS_REACT_ACT_ENVIRONMENT = true
  const React = await import('react'); const { act } = React
  const { createRoot } = await import('react-dom/client'); const { MemoryRouter } = await import('react-router-dom')
  const directory = await mkdtemp(new URL('../node_modules/.analytics-page-', import.meta.url))
  const previousFetch = globalThis.fetch; globalThis.fetch = fetchImpl
  let root
  try {
    const bundle = await build({ entryPoints: ['src/components/AnalyticsPage.jsx'], bundle: true, packages: 'external', platform: 'node', format: 'esm', jsx: 'automatic', write: false,
      define: { 'import.meta.env.VITE_API_BASE_URL': 'undefined' }, plugins: [{ name: 'authenticated-manager', setup(builder) {
        builder.onResolve({ filter: /AuthContext\.jsx$/ }, () => ({ path: 'test-auth', namespace: 'test-auth' }))
        builder.onLoad({ filter: /.*/, namespace: 'test-auth' }, () => ({ contents: 'const logout=()=>{}; export function useAuth(){return {token:"test",logout,user:{id:"test",role:"admin",status:"active"},isAuthenticated:true,checking:false};}' }))
      } }] })
    const path = directory + '/page.mjs'; await writeFile(path, bundle.outputFiles[0].text)
    const { default: Page } = await import(pathToFileURL(path).href)
    root = createRoot(document.getElementById('root'))
    await act(async () => root.render(React.createElement(MemoryRouter, null, React.createElement(Page))))
    await check({ act, dom, root })
  } finally {
    if (root) await act(async () => root.unmount())
    globalThis.fetch = previousFetch; dom.window.close(); await rm(directory, { recursive: true, force: true })
  }
}
const responseFor = (url, empty = false) => Response.json({ analytics: analyticsFixture(Object.fromEntries(new URL(url).searchParams), empty) })

test('analytics page renders real KPIs and handles an empty period', async () => {
  await pageHarness(async url => responseFor(url), async () => {
    assert.equal(document.querySelectorAll('.analytics-kpi').length, 4)
    assert.equal(document.querySelectorAll('svg[role="img"]').length, 2)
    assert.equal(document.querySelectorAll('.analytics-bars rect').length, 5)
    assert.equal(document.querySelectorAll('.analytics-legend li').length, 6)
    assert.equal(document.querySelector('.analytics-donut-total').textContent, '6')
    assert.match(document.querySelector('.analytics-kpis').textContent, /25 %/)
    assert.match(document.querySelector('.analytics-kpis').textContent, /33,33 %/)
    assert.match(document.body.textContent, /09:00 in 17:00/)
  })
  await pageHarness(async url => responseFor(url, true), async () => {
    assert.match(document.body.textContent, /še ni rezervacij/)
    assert.match(document.querySelector('.analytics-kpis').textContent, /—/)
  })
})

test('analytics authorization errors show no data and allow retry', async () => {
  let calls = 0
  await pageHarness(async url => ++calls === 1 ? new Response('{}', { status: 403 }) : responseFor(url), async ({ act }) => {
    assert.match(document.querySelector('[role="alert"]').textContent, /aktivno vlogo/)
    assert.equal(document.querySelector('.analytics-kpis'), null)
    await act(async () => [...document.querySelectorAll('button')].find(button => button.textContent === 'Poskusi znova').click())
    assert.ok(document.querySelector('.analytics-kpis'))
  })
})
