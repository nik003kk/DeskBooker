import { useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { Card, StatusBadge, Button } from './ui.jsx'
import { paths } from '../routes.js'
import { bookingSummary } from '../booking-selection.js'
import { resourceTypeNames } from '../resource-card.js'
import { createBooking } from '../booking-api.js'
import { useAuth } from '../AuthContext.jsx'

export default function BookingConfirmPage() {
  const { state } = useLocation()
  const { token, logout } = useAuth()
  const summary = bookingSummary(state)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [confirmed, setConfirmed] = useState(null)

  async function submit() {
    if (busy || confirmed || !summary) return
    setBusy(true)
    setError('')
    try {
      const result = await createBooking(state, {
        token, logout, apiBaseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api',
      })
      setConfirmed(result)
    } catch (err) {
      setError(err instanceof TypeError
        ? 'Povezava s strežnikom ni uspela.'
        : err.message || 'Rezervacija ni uspela.')
    } finally {
      setBusy(false)
    }
  }
  if (!summary) return (
    <Card className="auth-card" aria-labelledby="booking-title">
      <h1 id="booking-title">Izberi vir in termin</h1>
      <p>Za rezervacijo najprej poišči prost vir in določi datum ter uro.</p>
      <Link to={paths.search}>Nazaj na iskanje virov</Link>
    </Card>
  )
  if (confirmed) return (
    <Card className="auth-card" aria-labelledby="booking-success-title">
      <StatusBadge tone="blue">Potrjeno</StatusBadge>
      <h1 id="booking-success-title">Rezervacija je uspešna!</h1>
      <p><strong>{summary.resourceName}</strong>, {summary.date} od {summary.start} do {summary.end}.</p>
      <p>Check-in PIN: <strong className="booking-pin">{confirmed.check_in_pin}</strong></p>
      <p className="field-hint">PIN shrani zase. Ob ponovnem odpiranju ta odgovor ni na voljo.</p>
      <Link to={paths.search}>Nazaj na iskanje</Link>
    </Card>
  )
  return (
    <Card className="auth-card booking-confirm-card" aria-labelledby="booking-title">
      <span className="eyebrow">Pred zaključkom</span>
      <h1 id="booking-title">Potrdi rezervacijo</h1>
      <StatusBadge tone="blue">{resourceTypeNames[summary.resourceType] || 'Vir'}</StatusBadge>
      <dl className="profile-details">
        <div><dt>Vir</dt><dd>{summary.resourceName}</dd></div>
        <div><dt>Nadstropje</dt><dd>{summary.floor ?? 'Ni navedeno'}</dd></div>
        <div><dt>Datum</dt><dd>{summary.date}</dd></div>
        <div><dt>Začetek</dt><dd>{summary.start}</dd></div>
        <div><dt>Konec</dt><dd>{summary.end}</dd></div>
      </dl>
      {error && <p role="alert" className="auth-feedback auth-feedback--error">{error}</p>}
      <Button disabled={busy} onClick={submit}>{busy ? 'Potrjujem ...' : 'Potrdi rezervacijo'}</Button>
      <p className="auth-bottom"><Link to={paths.search}>Nazaj na iskanje</Link></p>
    </Card>
  )
}
