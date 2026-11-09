import assert from 'node:assert/strict'
import test from 'node:test'
import { mkdtemp, writeFile, rm } from 'node:fs/promises'
import { pathToFileURL } from 'node:url'
import { build } from 'esbuild'
import React from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { MemoryRouter } from 'react-router-dom'

test('direct analytics route blocks unauthorized users before rendering its data', async () => {
  const directory = await mkdtemp(new URL('../node_modules/.analytics-access-', import.meta.url))
  try {
    const result = await build({ entryPoints: ['src/components/AnalyticsRoute.jsx'], bundle: true, packages: 'external', platform: 'node', format: 'esm', jsx: 'automatic', write: false,
      plugins: [{ name: 'test-session', setup(builder) {
        builder.onResolve({ filter: /AuthContext\.jsx$/ }, () => ({ path: 'session', namespace: 'test-session' }))
        builder.onLoad({ filter: /.*/, namespace: 'test-session' }, () => ({ contents: 'export function useAuth(){return globalThis.__analyticsAccessSession;}' }))
      } }] })
    const path = directory + '/guard.mjs'; await writeFile(path, result.outputFiles[0].text)
    const { default: Guard } = await import(pathToFileURL(path).href)
    for (const [role, status, allowed] of [['admin', 'active', true], ['manager', 'active', true], ['employee', 'active', false], ['admin', 'suspended', false], ['manager', 'inactive', false]]) {
      globalThis.__analyticsAccessSession = { user: { role, status }, isAuthenticated: true, checking: false }
      const html = renderToStaticMarkup(React.createElement(MemoryRouter, null, React.createElement(Guard, null, React.createElement('div', null, 'Protected analytics payload'))))
      assert.equal(html.includes('Protected analytics payload'), allowed, role + '/' + status)
      if (!allowed) assert.match(html, /Za analitiko potrebuješ/)
    }
  } finally { delete globalThis.__analyticsAccessSession; await rm(directory, { recursive: true, force: true }) }
})
