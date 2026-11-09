export function checkInWindow(reservation, now = Date.now()) {
  if (reservation?.status === 'CHECKED_IN') return { state: 'checked_in', canSubmit: false, seconds: 0 }
  if (reservation?.status !== 'CONFIRMED') return { state: 'unavailable', canSubmit: false, seconds: 0 }
  const start = Date.parse(reservation.starts_at)
  const end = Date.parse(reservation.ends_at)
  if (!Number.isFinite(start) || !Number.isFinite(end) || end <= start || !Number.isFinite(now)) {
    return { state: 'unavailable', canSubmit: false, seconds: 0 }
  }
  const opensAt = start - 10 * 60_000
  const closesAt = Math.min(start + 15 * 60_000, end)
  const state = now < opensAt ? 'waiting' : now > start + 15 * 60_000 || now >= end ? 'closed' : 'open'
  return { state, canSubmit: state === 'open', opensAt, closesAt,
    seconds: Math.max(0, Math.ceil(((state === 'waiting' ? opensAt : closesAt) - now) / 1000)) }
}

export function formatRemaining(seconds) {
  const value = Math.max(0, Math.ceil(seconds))
  if (value >= 86400) return `${Math.floor(value / 86400)} d ${Math.floor(value % 86400 / 3600)} h`
  if (value >= 3600) return `${Math.floor(value / 3600)} h ${Math.floor(value % 3600 / 60)} min`
  return `${Math.floor(value / 60)}:${String(value % 60).padStart(2, '0')}`
}

export function checkInWindowLabel(window) {
  if (window.state === 'waiting') return 'Check-in bo na voljo čez ' + formatRemaining(window.seconds) + '.'
  if (window.state === 'open') return 'Preostali čas za check-in: ' + formatRemaining(window.seconds)
  if (window.state === 'closed') return 'Čas za check-in je potekel.'
  if (window.state === 'checked_in') return 'Prisotnost je že potrjena.'
  return 'Check-in za to rezervacijo ni na voljo.'
}
