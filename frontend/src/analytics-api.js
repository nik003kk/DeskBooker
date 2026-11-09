import { authorizedFetch } from './auth-api.js'

export const analyticsStatuses = ['PENDING', 'CONFIRMED', 'CHECKED_IN', 'COMPLETED', 'CANCELLED', 'EXPIRED']
export const statusLabels = { PENDING: 'V čakanju', CONFIRMED: 'Potrjeno', CHECKED_IN: 'Prisotnost potrjena', COMPLETED: 'Zaključeno', CANCELLED: 'Preklicano', EXPIRED: 'Poteklo' }
export const weekdayLabels = ['Pon', 'Tor', 'Sre', 'Čet', 'Pet']

export function defaultAnalyticsPeriod(days = 7, now = new Date()) {
  const parts = Object.fromEntries(new Intl.DateTimeFormat('en-GB', { timeZone: 'Europe/Ljubljana', year: 'numeric', month: '2-digit', day: '2-digit' }).formatToParts(now).filter(p => p.type !== 'literal').map(p => [p.type, p.value]))
  const end = new Date(`${parts.year}-${parts.month}-${parts.day}T00:00:00Z`)
  const start = new Date(end); start.setUTCDate(start.getUTCDate() - days + 1)
  return { from: start.toISOString().slice(0, 10), to: end.toISOString().slice(0, 10) }
}

function validDate(value) {
  return typeof value === 'string' && /^\d{4}-\d{2}-\d{2}$/.test(value) && Number.isFinite(Date.parse(value)) && new Date(value).toISOString().slice(0, 10) === value
}
export function analyticsQuery(period) {
  if (!validDate(period?.from) || !validDate(period?.to) || period.to < period.from || (Date.parse(period.to) - Date.parse(period.from)) / 86400000 > 365) {
    throw new Error('Izberi veljavno obdobje, dolgo največ 366 dni.')
  }
  return new URLSearchParams(period).toString()
}

const count = value => Number.isSafeInteger(value) && value >= 0
const percent = value => value === null || Number.isFinite(value) && value >= 0 && value <= 100
export function validateAnalytics(report, period) {
  return report?.from === period.from && report?.to === period.to && report.timezone === 'Europe/Ljubljana' &&
    Number.isFinite(Date.parse(report.generated_at)) && count(report.total_reservations) && count(report.active_reservations) && report.active_reservations <= report.total_reservations &&
    analyticsStatuses.every(status => count(report.status_counts?.[status])) && analyticsStatuses.reduce((sum, s) => sum + report.status_counts[s], 0) === report.total_reservations &&
    percent(report.occupancy_rate) && count(report.no_show?.eligible) && count(report.no_show.expired) && count(report.no_show.attended) && report.no_show.eligible === report.no_show.expired + report.no_show.attended && percent(report.no_show.rate) &&
    Array.isArray(report.daily) && Array.isArray(report.weekdays) && report.weekdays.length === 5 &&
    [...report.daily, ...report.weekdays].every(p => Number.isInteger(p.weekday) && p.weekday >= 1 && p.weekday <= 5 && Number.isFinite(p.booked_minutes) && p.booked_minutes >= 0 && Number.isFinite(p.capacity_minutes) && p.capacity_minutes >= p.booked_minutes && percent(p.occupancy_rate))
}

export async function loadAnalytics(period, options = {}) {
  const response = await authorizedFetch('/analytics?' + analyticsQuery(period), options)
  const body = await response.json()
  if (!validateAnalytics(body?.analytics, period)) throw new Error('Strežnik ni vrnil veljavnih statističnih podatkov.')
  return body.analytics
}
export function formatPercent(value) { return value === null ? '—' : new Intl.NumberFormat('sl-SI', { maximumFractionDigits: 2 }).format(value) + ' %' }
