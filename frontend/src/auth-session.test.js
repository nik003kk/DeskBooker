import assert from 'node:assert/strict'
import test from 'node:test'
import { readAuthSession, writeAuthSession, clearAuthSession } from './auth-session.js'
function memoryStorage() {
  const map = new Map()
  return { getItem: k => map.get(k) ?? null, setItem: (k,v) => map.set(k,v), removeItem: k => map.delete(k) }
}
test('saves, restores and clears valid session', () => {
  const store = memoryStorage()
  const item = { access_token:'jwt', user: { id:'user1' }, expires_at: new Date(Date.now()+60000).toISOString() }
  writeAuthSession(store,item)
  assert.deepEqual(readAuthSession(store),item)
  clearAuthSession(store)
  assert.equal(readAuthSession(store),null)
})
test('expired and malformed sessions are refused', () => {
  const store = memoryStorage()
  store.setItem('deskbooker-auth-session', '{oops')
  assert.equal(readAuthSession(store),null)
  assert.throws(() => writeAuthSession(store, { access_token:'bad',user:{id:'1'},expires_at:'yesterday' }))
  store.setItem('deskbooker-auth-session', JSON.stringify({
    access_token:'jwt', user:{id:'1'}, expires_at:new Date(Date.now()-20000).toISOString(),
  }))
  assert.equal(readAuthSession(store),null)
})
