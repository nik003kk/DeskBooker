import assert from 'node:assert/strict'
import test from 'node:test'
import { listAdminResources } from './resource-admin-api.js'
test('lists both active and inactive resources',async()=>{
 let url
 const result=await listAdminResources({token:'jwt',fetchImpl:async(u,opts)=>{
   url=u;assert.equal(opts.headers.Authorization,'Bearer jwt')
   return {ok:true,status:200,json:async()=>({resources:[{id:'one',is_active:true},{id:'two',is_active:false}]})}
 }})
 assert.equal(url,'http://localhost:8080/api/admin/resources')
 assert.equal(result.length,2)
 assert.equal(result[1].is_active,false)
})
