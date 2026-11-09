import { authorizedFetch } from './auth-api.js'

export async function setResourceActive(id,isActive,{token,logout,apiBaseUrl,fetchImpl=fetch}={}) {
  if (!id || typeof isActive!=='boolean') throw new Error('Neveljavni podatki za aktivacijo vira.')
  const response=await authorizedFetch('/resources/'+encodeURIComponent(id)+'/activation',{
    token,logout,apiBaseUrl,fetchImpl,
    method:'PATCH',headers:{'Content-Type':'application/json'},
    body:JSON.stringify({is_active:isActive}),
  })
  const data=await response.json()
  if (!data?.resource?.id || data.resource.is_active!==isActive)
    throw new Error('Strežnik ni potrdil spremembe stanja vira.')
  return data.resource
}
