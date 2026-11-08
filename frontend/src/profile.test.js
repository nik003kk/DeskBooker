import assert from 'node:assert/strict'
import test from 'node:test'
import { accountStatusLabel, fetchProfile } from './profile.js'
test('fetches private profile with bearer token', async () => {
  let header
  const user = await fetchProfile({
    token:'jwt', fetchImpl: async (_, opts) => {
      header = opts.headers.Authorization
      return { ok:true, status:200, json:async()=>({user:{id:'u1',email:'ana@gmail.com',status:'active'}}) }
    },
  })
  assert.equal(header,'Bearer jwt')
  assert.equal(user.status,'active')
})
test('maps status and rejects malformed profile', async () => {
  assert.equal(accountStatusLabel('suspended'),'Začasno onemogočen')
  await assert.rejects(() => fetchProfile({token:'jwt',fetchImpl:async()=>({
    ok:true,status:200,json:async()=>({user:{}}),
  })}), /neveljavne/)
})
