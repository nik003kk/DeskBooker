import { authorizedFetch } from './auth-api.js'
import { analyticsQuery } from './analytics-api.js'

export async function fetchAnalyticsCSV(period, options = {}) {
  const response = await authorizedFetch('/analytics/export.csv?' + analyticsQuery(period), options)
  if (!response.headers.get('Content-Type')?.toLowerCase().startsWith('text/csv')) throw new Error('Strežnik ni vrnil CSV poročila.')
  const text = (await response.text()).replace(/^\uFEFF/, '')
  if (!text.startsWith('section,metric,date,weekday,value,unit')) throw new Error('CSV poročilo ni veljavno.')
  return { blob: new Blob(['\uFEFF', text], { type: 'text/csv;charset=utf-8' }), filename: `deskbooker-analytics-${period.from}-${period.to}.csv` }
}

export function saveAnalyticsCSV(file) {
  const url = URL.createObjectURL(file.blob)
  const link = document.createElement('a')
  try { link.href = url; link.download = file.filename; document.body.append(link); link.click() }
  finally { link.remove(); URL.revokeObjectURL(url) }
}
