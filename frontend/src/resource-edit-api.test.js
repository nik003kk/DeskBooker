import assert from 'node:assert/strict'
import test from 'node:test'
import {loadAdminResource,saveResourceEdit} from './resource-edit-api.js'
const values={name:'Desk',type:'DESK',floor:'',capacity:'1',features:'Monitor,USB-C',is_active:true}
test('loads resource and existing features',async()=>{
 const data=await loadAdminResource('uuid',{token:'jwt',fetchImpl:async()=>({
  ok:true,status:200,json:async()=>({resource:{id:'uuid'},features:['Monitor']})
 })})
 assert.deepEqual(data.features,['Monitor'])
})
test('atomically sends modified details and equipment',async()=>{
 let sent
 const result=await saveResourceEdit('uuid',values,{token:'jwt',fetchImpl:async(_,options)=>{
  sent=JSON.parse(options.body)
  assert.equal(options.method,'PATCH')
  return {ok:true,status:200,json:async()=>({resource:{id:'uuid'},features:['Monitor','USB-C']})}
 }})
 assert.equal(result.resource.id,'uuid')
 assert.equal(sent.floor,null)
 assert.deepEqual(sent.features,['Monitor','USB-C'])
})
