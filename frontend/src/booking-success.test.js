import assert from 'node:assert/strict'
import test from 'node:test'
import { confirmedBookingDetails } from './booking-success.js'

const id = '11111111-1111-4111-8111-111111111111'
const state = {
  selection:{resource:{id,name:'Miza A',type:'DESK',floor:1},filters:{date:'2026-11-10',start:'09:00',end:'10:00'}},
  booking:{reservation:{id:'reservation-1',resource_id:id,status:'CONFIRMED'},check_in_pin:'0042'},
}
test('confirmation shows correct booking details and leading-zero PIN',()=>{
  const d=confirmedBookingDetails(state)
  assert.equal(d.id,'reservation-1')
  assert.equal(d.pin,'0042')
  assert.equal(d.resourceName,'Miza A')
  assert.equal(d.start,'09:00')
})
test('missing PIN or mismatched resource never renders fake confirmation',()=>{
  assert.equal(confirmedBookingDetails(null),null)
  assert.equal(confirmedBookingDetails({...state,booking:{...state.booking,check_in_pin:'42'}}),null)
  assert.equal(confirmedBookingDetails({...state,booking:{...state.booking,reservation:{...state.booking.reservation,resource_id:'other'}}}),null)
})
