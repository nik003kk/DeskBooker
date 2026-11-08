import { authorizedFetch } from './auth-api.js'
import { parseSearchFeatures } from './resource-search-filters.js'

export function buildResourceSearchQuery(filters = {}) {
  const params = new URLSearchParams()
  if (filters.type) params.set('type', filters.type)
  if (filters.date && filters.start && filters.end) {
    params.set('date', filters.date)
    params.set('start', filters.start)
    params.set('end', filters.end)
  }
  const features = parseSearchFeatures(filters.features)
  if (features.length) params.set('features', features.join(','))
  const query = params.toString()
  return '/resources' + (query ? '?' + query : '')
}

export async function searchAvailableResources(filters, {
  token, logout, apiBaseUrl, fetchImpl = fetch, signal,
} = {}) {
  const response = await authorizedFetch(buildResourceSearchQuery(filters), {
    token, logout, apiBaseUrl, fetchImpl, signal,
  })
  const body = await response.json()
  if (!Array.isArray(body?.resources) ||
    body.resources.some(item => !item || typeof item.id !== 'string' ||
      typeof item.name !== 'string' || typeof item.type !== 'string')) {
    throw new Error('Strežnik je vrnil neveljavne podatke o virih.')
  }
  return body.resources
}
