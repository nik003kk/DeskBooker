import { authorizedFetch } from './auth-api.js'
import { validateResource,resourcePayload } from './resource-admin-api.js'

export async function loadAdminResource(id,{token,logout,apiBaseUrl,fetchImpl=fetch}={}){
 const response=await authorizedFetch('/admin/resources/'+encodeURIComponent(id),{token,logout,apiBaseUrl,fetchImpl})
 const data=await response.json()
 if (!data?.resource?.id || !Array.isArray(data.features)) throw new Error('Neveljavni podatki vira.')
 return data
}
export async function saveResourceEdit(id,values,{token,logout,apiBaseUrl,fetchImpl=fetch}={}){
 const errors=validateResource(values)
 if (Object.keys(errors).length)throw new Error('Preveri vnesene podatke vira.')
 const payload={...resourcePayload(values),features:values.features.split(',').map(s=>s.trim()).filter(Boolean)}
 const response=await authorizedFetch('/resources/'+encodeURIComponent(id),{
  token,logout,apiBaseUrl,fetchImpl,method:'PATCH',
  headers:{'Content-Type':'application/json'},body:JSON.stringify(payload),
 })
 const data=await response.json()
 if (!data?.resource?.id)throw new Error('Neveljaven odgovor strežnika.')
 return data
}
