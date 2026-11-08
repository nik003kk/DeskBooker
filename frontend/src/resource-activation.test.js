import assert from 'node:assert/strict'
import test from 'node:test'
import {setResourceActive} from './resource-activation.js'
test('deactivation sends boolean status and verifies server response',async()=>{
 let request
 const item=await setResourceActive('uuid',false,{token:'jwt',fetchImpl:async(url,opts)=>{
   request={url,opts}
   return {ok:true,status:200,json:async()=>({resource:{id:'uuid',is_active:false}})}
 }})
 assert.equal(item.is_active,false)
 assert.equal(request.url,'http://localhost:8080/api/resources/uuid/activation')
 assert.equal(request.opts.headers.Authorization,'Bearer jwt')
 assert.deepEqual(JSON.parse(request.opts.body),{is_active:false})
})
test('activation rejects malformed success',async()=>{
 await assert.rejects(()=>setResourceActive('uuid',true,{token:'jwt',fetchImpl:async()=>({
  ok:true,status:200,json:async()=>({resource:{id:'uuid',is_active:false}})
 })}),/ni potrdil/)
})
test('invalid request does not fetch',async()=>{
 await assert.rejects(()=>setResourceActive('',false,{}),/Neveljavni/)
})
