import { authorizedFetch } from './auth-api.js'
import { validateBookingSelection } from './booking-selection.js'

// Match backend's Europe/Ljubljana business timezone, independent of user's
// computer timezone. Detect nonexistent local hours at the spring DST jump.
function officeTimestamp(date, clock) {
  const wall = new Date(date + 'T' + clock + ':00Z')
  if (!Number.isFinite(wall.getTime())) throw new Error('Datum ali ura ni veljavna.')
  const formatter = new Intl.DateTimeFormat('en-GB', {
    timeZone: 'Europe/Ljubljana', year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit', hourCycle: 'h23',
  })
  function parts(time) {
    return Object.fromEntries(formatter.formatToParts(new Date(time))
      .filter(part => part.type !== 'literal').map(part => [part.type, part.value]))
  }
  const guess = parts(wall.getTime())
  const offset = Date.UTC(+guess.year, +guess.month - 1, +guess.day,
    +guess.hour, +guess.minute) - wall.getTime()
  const utc = wall.getTime() - offset
  const exact = parts(utc)
  if (exact.year + '-' + exact.month + '-' + exact.day !== date ||
      exact.hour + ':' + exact.minute !== clock) {
    throw new Error('Ta lokalni čas ne obstaja zaradi premika ure.')
  }
  return new Date(utc).toISOString()
}

export function bookingRequest(selection) {
  if (!validateBookingSelection(selection)) throw new Error('Najprej izberi vir in veljaven termin.')
  const { resource, filters } = selection
  const starts_at = officeTimestamp(filters.date, filters.start)
  const ends_at = officeTimestamp(filters.date, filters.end)
  if (Date.parse(ends_at) <= Date.parse(starts_at)) throw new Error('Končni čas mora biti po začetnem.')
  return { resource_id: resource.id, starts_at, ends_at }
}

export async function createBooking(selection, {
  token, logout, apiBaseUrl, fetchImpl = fetch,
} = {}) {
  const payload = bookingRequest(selection)
  let response
  try {
    response = await authorizedFetch('/reservations', {
      token, logout, apiBaseUrl, fetchImpl, method: 'POST',
      headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload),
    })
  } catch (error) {
    if (error.status === 409) throw new Error('Izbrani termin je že zaseden. Izberi drugega.')
    if (error.status === 400) throw new Error('Termin ni veljaven. Preveri datum in čas.')
    if (error.status === 404) throw new Error('Izbrani vir ni več na voljo.')
    throw error
  }
  const body = await response.json()
  if (body?.reservation?.status !== 'CONFIRMED' ||
      typeof body.reservation.id !== 'string' ||
      !/^\d{4}$/.test(body?.check_in_pin || '')) {
    throw new Error('Strežnik ni vrnil veljavne potrditve rezervacije.')
  }
  return { ...body, resource: selection.resource }
}
