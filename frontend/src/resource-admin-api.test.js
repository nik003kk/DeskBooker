import assert from 'node:assert/strict'
import test from 'node:test'
import { validateResource,createResourceWithFeatures } from './resource-admin-api.js'
const input={name:'Miza A',type:'DESK',floor:'1',capacity:'1',features:'Monitor, USB-C',is_active:true}
test('resource validation checks required fields and equipment',()=>{
  assert.deepEqual(validateResource(input),{})
  for(const values of [
    {...input,name:' '},{...input,floor:'-1'},{...input,capacity:'0'},
    {...input,capacity:'1.5'},{...input,type:'OTHER'},{...input,features:'Monitor, monitor'},
  ]) assert.ok(Object.keys(validateResource(values)).length)
})
test('creates a resource then attaches equipment',async()=>{
  const calls=[]
  const result=await createResourceWithFeatures(input,{token:'jwt',fetchImpl:async(url,opts)=>{
    calls.push({url,opts})
    return {ok:true,status:201,json:async()=>url.endsWith('/features')?{features:[]}:{resource:{id:'res-1',name:'Miza A'}}}
  }})
  assert.equal(result.id,'res-1')
  assert.equal(calls.length,2)
  assert.deepEqual(JSON.parse(calls[1].opts.body),{features:['Monitor','USB-C']})
  assert.equal(calls[0].opts.headers.Authorization,'Bearer jwt')
})
test('equipment failure is not falsely reported as full success',async()=>{
  await assert.rejects(()=>createResourceWithFeatures(input,{token:'jwt',fetchImpl:async(url)=>url.endsWith('/features')?
    {ok:false,status:500,json:async()=>({})}:{ok:true,status:201,json:async()=>({resource:{id:'res-1',name:'Miza A'}})}
  }),error=>error.resource?.id==='res-1')
})
