import assert from 'node:assert/strict'
import test from 'node:test'
import { buildResourceSearchQuery, searchAvailableResources } from './resource-search-api.js'

test('empty filters request every active resource', () => {
  assert.equal(buildResourceSearchQuery({}), '/resources')
})
test('search query encodes type date time and all equipment', () => {
  const path = buildResourceSearchQuery({
    type:'DESK',date:'2026-11-09',start:'09:00',end:'17:00',features:'Monitor, USB-C',
  })
  const params = new URLSearchParams(path.split('?')[1])
  assert.equal(params.get('type'),'DESK')
  assert.equal(params.get('date'),'2026-11-09')
  assert.equal(params.get('start'),'09:00')
  assert.equal(params.get('end'),'17:00')
  assert.equal(params.get('features'),'Monitor,USB-C')
})
test('searches with JWT and returns real resources with equipment', async () => {
  let request
  const found=await searchAvailableResources({type:'DESK'},{
    token:'jwt',fetchImpl:async(url,opts)=>{
      request={url,opts}
      return {status:200,ok:true,json:async()=>({
        resources:[{id:'uuid',name:'Desk',type:'DESK',features:['Monitor']}],
      })}
    },
  })
  assert.equal(request.opts.headers.Authorization,'Bearer jwt')
  assert.match(request.url,/type=DESK/)
  assert.deepEqual(found[0].features,['Monitor'])
})
test('valid empty results and malformed responses are distinguished', async () => {
  assert.deepEqual(await searchAvailableResources({},{
    token:'jwt',fetchImpl:async()=>({ok:true,status:200,json:async()=>({resources:[]})}),
  }),[])
  await assert.rejects(()=>searchAvailableResources({},{
    token:'jwt',fetchImpl:async()=>({ok:true,status:200,json:async()=>({not_resources:[]})}),
  }),/neveljavne podatke/)
})
test('401 removes session; 403 preserves it', async () => {
  for(const status of [401,403]){
    let loggedOut=false
    await assert.rejects(()=>searchAvailableResources({},{
      token:'jwt',logout:()=>{loggedOut=true},
      fetchImpl:async()=>({ok:false,status}),
    }),error=>error.status===status)
    assert.equal(loggedOut,status===401)
  }
})
