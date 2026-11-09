import assert from 'node:assert/strict'
import test from 'node:test'
import { canManageResources } from './resource-permissions.js'
test('only active manager or admin can manage resources',()=>{
 for(const role of ['manager','admin']) assert.equal(canManageResources({role,status:'active'}),true)
 for(const role of ['employee','OFFICE_MANAGER','ADMIN']) assert.equal(canManageResources({role,status:'active'}),false)
 assert.equal(canManageResources({role:'admin',status:'suspended'}),false)
 assert.equal(canManageResources({role:'manager',status:'inactive'}),false)
 assert.equal(canManageResources(null),false)
})
