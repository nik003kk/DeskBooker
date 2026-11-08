import { bookingSummary } from './booking-selection.js'

export function confirmedBookingDetails(state) {
  const reservation = state?.booking?.reservation
  const pin = state?.booking?.check_in_pin
  const summary = bookingSummary(state?.selection)
  if (!summary || !reservation || reservation.status !== 'CONFIRMED' ||
      typeof reservation.id !== 'string' || !reservation.id ||
      reservation.resource_id !== state.selection.resource.id ||
      !/^\d{4}$/.test(pin || '')) return null
  return { ...summary, id: reservation.id, pin, status: reservation.status }
}
