import { analyticsQuery, formatPercent } from './analytics-api.js'

export function previousAnalyticsPeriod(period) {
  analyticsQuery(period)
  const days = (Date.parse(period.to) - Date.parse(period.from)) / 86400000 + 1
  const to = new Date(Date.parse(period.from) - 86400000)
  const from = new Date(to); from.setUTCDate(from.getUTCDate() - days + 1)
  return { from: from.toISOString().slice(0, 10), to: to.toISOString().slice(0, 10) }
}
export function percentagePointChange(current, previous) {
  if (current === null || previous === null) return 'Ni primerljivih podatkov'
  const delta = Math.round((current - previous) * 100) / 100
  return (delta > 0 ? '+' : '') + new Intl.NumberFormat('sl-SI', { maximumFractionDigits: 2 }).format(delta) + ' odstotne točke'
}
export function reservationChange(current, previous) {
  if (previous === 0) return current === 0 ? 'Brez spremembe' : 'Novo glede na obdobje brez rezervacij'
  const delta = (current - previous) / previous * 100
  return (delta > 0 ? '+' : '') + formatPercent(delta)
}
