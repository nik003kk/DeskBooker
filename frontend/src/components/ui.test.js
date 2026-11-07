import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

const source = readFileSync(fileURLToPath(new URL('./ui.jsx', import.meta.url)), 'utf8')

test('shared UI primitives are exported', () => {
  for (const name of ['Button', 'ButtonLink', 'Card', 'Field', 'SelectField', 'StatusBadge']) {
    assert.match(source, new RegExp(`export function ${name}\\(`))
  }
})

test('form controls have associated accessible labels', () => {
  assert.match(source, /<label htmlFor=\{inputId\}>/)
  assert.match(source, /<label htmlFor=\{id \|\| generatedId\}>/)
  assert.match(source, /aria-describedby/)
})

test('buttons default to non-submitting type', () => {
  assert.match(source, /type = 'button'/)
})
