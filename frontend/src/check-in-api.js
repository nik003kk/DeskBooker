import { authorizedFetch } from './auth-api.js'

export function checkInPath(id) {
  return '/reservations/' + encodeURIComponent(id) + '/check-in'
}

function validReservation(value, id) {
  return value?.id === id && typeof value.status === 'string' &&
    Number.isFinite(Date.parse(value.starts_at)) && Number.isFinite(Date.parse(value.ends_at))
}

export async function loadCheckIn(id, options = {}) {
  const response = await authorizedFetch('/reservations/' + encodeURIComponent(id), options)
  const body = await response.json()
  if (!validReservation(body?.reservation, id) || typeof body?.resource?.name !== 'string') {
    throw new Error('Strežnik ni vrnil veljavnih podatkov rezervacije.')
  }
  return body
}

export async function submitCheckIn(id, pin, options = {}) {
  if (!/^\d{4}$/.test(pin)) throw new Error('Vnesi štirimestno PIN kodo.')
  let response
  try {
    response = await authorizedFetch(checkInPath(id), {
      ...options, method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ pin }),
    })
  } catch (error) {
    if (error.status === 422) throw new Error('Napačna PIN koda')
    if (error.status === 409) throw new Error('Check-in trenutno ni dovoljen. Preveri termin in stanje rezervacije.')
    if (error.status === 404) throw new Error('Rezervacija ni na voljo.')
    throw error
  }
  const body = await response.json()
  if (!validReservation(body?.reservation, id) || body.reservation.status !== 'CHECKED_IN' ||
      !Number.isFinite(Date.parse(body.reservation.checked_in_at))) {
    throw new Error('Strežnik ni potrdil prisotnosti. Osveži podatke rezervacije.')
  }
  return body.reservation
}
