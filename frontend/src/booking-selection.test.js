import assert from 'node:assert/strict'
import test from 'node:test'
import { bookingSummary, validateBookingSelection } from './booking-selection.js'

const valid = {
  resource:{id:'11111111-1111-4111-8111-111111111111',name:'Miza A',type:'DESK',floor:0},
  filters:{date:'2026-11-12',start:'09:00',end:'11:00'},
}
test('booking confirmation requires resource and a chosen interval',()=>{
 assert.equal(validateBookingSelection(valid),true)
 assert.equal(bookingSummary(valid).floor,0)
 assert.equal(validateBookingSelection({...valid,filters:{...valid.filters,start:'12:00'}}),false)
 assert.equal(validateBookingSelection({}),false)
})
