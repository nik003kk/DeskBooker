import { authorizedFetch, AuthAPIError } from './auth-api.js'

export function validateResource(values) {
  const errors = {}
  const name = values.name ?? ''
  if (!name.trim() || name !== name.trim() || [...name].length > 150 || /[\x00-\x1f\x7f]/.test(name)) errors.name = 'Vnesi naziv brez odvečnih presledkov (največ 150 znakov).'
  if (!['DESK', 'MEETING_ROOM', 'PARKING_SPOT'].includes(values.type)) errors.type = 'Izberi vrsto vira.'
  if (!Number.isSafeInteger(Number(values.capacity)) || Number(values.capacity) <= 0) errors.capacity = 'Kapaciteta mora biti pozitivno celo število.'
  if (values.floor !== '' && values.floor != null && (!Number.isSafeInteger(Number(values.floor)) || Number(values.floor) < 0)) errors.floor = 'Nadstropje mora biti nenegativno celo število.'
  const features = values.features?.split(',').map(x => x.trim()).filter(Boolean) ?? []
  if (features.length > 20 || features.some(x => [...x].length > 100 || /[\x00-\x1f\x7f]/.test(x)) ||
      new Set(features.map(x => x.toLowerCase())).size !== features.length) errors.features = 'Dovoljeno je največ 20 različnih lastnosti (do 100 znakov).'
  return errors
}

export function resourcePayload(values) {
  return {
    name: values.name,
    type: values.type,
    floor: values.floor === '' || values.floor == null ? null : Number(values.floor),
    capacity: Number(values.capacity),
    is_active: values.is_active !== false,
  }
}

export async function createResourceWithFeatures(values, { token, logout, apiBaseUrl, fetchImpl = fetch } = {}) {
  const payload = resourcePayload(values)
  const response = await authorizedFetch('/resources', {
    token, logout, apiBaseUrl, fetchImpl, method: 'POST',
    headers: {'Content-Type':'application/json'}, body: JSON.stringify(payload),
  })
  const data = await response.json()
  if (!data?.resource?.id) throw new Error('Strežnik ni vrnil ID-ja vira.')
  const features = values.features?.split(',').map(s => s.trim()).filter(Boolean) ?? []
  if (features.length) {
    try {
      await authorizedFetch('/resources/' + encodeURIComponent(data.resource.id) + '/features', {
        token, logout, apiBaseUrl, fetchImpl, method: 'POST',
        headers: {'Content-Type':'application/json'}, body: JSON.stringify({features}),
      })
    } catch (error) {
      const partial = new Error('Vir je ustvarjen, dodajanje opreme pa ni uspelo. Preveri seznam in opremo dodaj ponovno.')
      partial.resource = data.resource
      partial.cause = error
      throw partial
    }
  }
  return data.resource
}
