import assert from 'node:assert/strict'
import test from 'node:test'
import { saveProfileEdit, validateProfileEdit } from './profile-edit.js'
const valid = {first_name:'Ana',last_name:'Novak',department:''}
test('validates profile changes', () => {
  assert.deepEqual(validateProfileEdit(valid),{})
  assert.ok(validateProfileEdit({...valid,first_name:''}).first_name)
  assert.ok(validateProfileEdit({...valid,last_name:' '.repeat(2)}).last_name)
  assert.ok(validateProfileEdit({...valid,department:'x'.repeat(101)}).department)
})
test('saves only authorized fields and permits clearing department', async () => {
  let request
  const result = await saveProfileEdit(valid,{token:'jwt',fetchImpl:async(url,opts)=>{
    request={url,opts}
    return {ok:true,status:200,json:async()=>({user:{id:'u1',first_name:'Ana'}})}
  }})
  assert.equal(result.id,'u1')
  assert.equal(request.url,'http://localhost:8080/api/users/me')
  assert.equal(request.opts.method,'PATCH')
  assert.equal(request.opts.headers.Authorization,'Bearer jwt')
  assert.deepEqual(JSON.parse(request.opts.body),{first_name:'Ana',last_name:'Novak',department:null})
  assert.ok(!request.opts.body.includes('email'))
  assert.ok(!request.opts.body.includes('role'))
})
