import assert from 'node:assert/strict'
import test from 'node:test'
import { bookingRequest, createBooking } from './booking-api.js'

const selection = {
  resource: {id:'11111111-1111-4111-8111-111111111111',name:'Miza 1',type:'DESK'},
  filters: {date:'2026-11-10',start:'09:00',end:'10:00'},
}
test('Slovenian winter time is correctly converted to UTC',()=>{
  const payload = bookingRequest(selection)
  assert.equal(payload.starts_at,'2026-11-10T08:00:00.000Z')
  assert.equal(payload.ends_at,'2026-11-10T09:00:00.000Z')
})
test('rejects nonexistent daylight-saving hour',()=>{
  assert.throws(()=>bookingRequest({...selection,filters:{
    date:'2026-03-29',start:'02:30',end:'04:00',
  }}),/ne obstaja/)
})
test('creates confirmed reservation using bearer token and returns PIN',async()=>{
  let sent
  const result=await createBooking(selection,{token:'JWT',fetchImpl:async(url,opts)=>{
    sent={url,opts}
    return {ok:true,status:201,json:async()=>({
      reservation:{id:'reservation-id',status:'CONFIRMED'},check_in_pin:'0042',
    })}
  }})
  assert.equal(sent.url,'http://localhost:8080/api/reservations')
  assert.equal(sent.opts.headers.Authorization,'Bearer JWT')
  assert.equal(JSON.parse(sent.opts.body).resource_id,selection.resource.id)
  assert.equal(result.check_in_pin,'0042')
})
test('maps conflict to usable message',async()=>{
  await assert.rejects(()=>createBooking(selection,{token:'jwt',fetchImpl:async()=>({
    ok:false,status:409,
  })}),/že zaseden/)
})
