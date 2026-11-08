import assert from 'node:assert/strict'
import test from 'node:test'
import { resourceCardDetails, resourceTypeNames } from './resource-card.js'

test('resource card displays name, floor, capacity and equipment', () => {
  const info = resourceCardDetails({
    name:'Miza A-101',type:'DESK',floor:0,capacity:1,features:['Monitor','USB-C'],
  })
  assert.equal(info.title,'Miza A-101')
  assert.equal(info.floor,'0. nadstropje')
  assert.equal(info.capacity,'1 oseb')
  assert.deepEqual(info.features,['Monitor','USB-C'])
  assert.equal(info.typeName,resourceTypeNames.DESK)
})

test('resource card handles missing floor and equipment', () => {
  const info = resourceCardDetails({name:'Soba',type:'MEETING_ROOM',floor:null,capacity:8})
  assert.equal(info.floor,'Nadstropje ni navedeno')
  assert.deepEqual(info.features,[])
  assert.equal(info.capacity,'8 oseb')
})
