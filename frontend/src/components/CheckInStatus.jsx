import { StatusBadge } from './ui.jsx'

export default function CheckInStatus({ reservation }) {
  const timestamp = Date.parse(reservation?.checked_in_at)
  if (reservation?.status !== 'CHECKED_IN' || !Number.isFinite(timestamp)) {
    return <StatusBadge tone="blue">{reservation?.status || 'Stanje ni na voljo'}</StatusBadge>
  }
  const formatted = new Intl.DateTimeFormat('sl-SI', {
    timeZone: 'Europe/Ljubljana', dateStyle: 'medium', timeStyle: 'short',
  }).format(new Date(timestamp))
  return (
    <section className="check-in-confirmed" role="status" aria-live="polite" aria-label="Potrjena prisotnost">
      <StatusBadge tone="green"><span aria-hidden="true">✓ </span>Prisotnost potrjena</StatusBadge>
      <p>Stanje: <strong>CHECKED_IN</strong></p>
      <p>Čas potrditve: <time dateTime={reservation.checked_in_at}>{formatted}</time></p>
    </section>
  )
}
