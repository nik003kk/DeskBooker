import { useCallback, useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { useAuth } from '../AuthContext.jsx'
import { Button, Card, StatusBadge } from './ui.jsx'
import CheckInModal from './CheckInModal.jsx'
import { loadCheckIn } from '../check-in-api.js'
import { paths } from '../routes.js'

function officeTime(value) {
  return new Intl.DateTimeFormat('sl-SI', { timeZone: 'Europe/Ljubljana', dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
}

export default function CheckInPage() {
  const { id } = useParams()
  const { token, logout } = useAuth()
  const [data, setData] = useState(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [open, setOpen] = useState(false)
  const [reload, setReload] = useState(0)
  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'
  const refresh = useCallback(() => setReload(value => value + 1), [])
  useEffect(() => {
    const controller = new AbortController()
    setData(null); setError(''); setOpen(false); setLoading(true)
    loadCheckIn(id, { token, logout, apiBaseUrl, signal: controller.signal }).then(result => {
      if (!controller.signal.aborted) setData(result)
    }).catch(err => {
      if (!controller.signal.aborted) setError(err.status === 404 ? 'Rezervacija ni na voljo.' : 'Podatkov rezervacije ni mogoče naložiti. Poskusi znova.')
    }).finally(() => { if (!controller.signal.aborted) setLoading(false) })
    return () => controller.abort()
  }, [id, token, logout, apiBaseUrl, reload])

  if (loading) return <Card aria-busy="true"><p role="status">Nalagam rezervacijo ...</p></Card>
  if (!data) return <Card><h1>Check-in</h1><p role="alert">{error}</p><Button onClick={refresh}>Poskusi znova</Button></Card>
  const reservation = data.reservation
  return (
    <Card className="auth-card">
      <span className="eyebrow">Tvoja rezervacija</span>
      <h1>Check-in</h1>
      <StatusBadge tone="blue">{reservation.status}</StatusBadge>
      <h2>{data.resource.name}</h2>
      <dl className="profile-details">
        <div><dt>Začetek</dt><dd>{officeTime(reservation.starts_at)}</dd></div>
        <div><dt>Konec</dt><dd>{officeTime(reservation.ends_at)}</dd></div>
      </dl>
      {reservation.status === 'CONFIRMED' && <Button onClick={() => setOpen(true)}>Potrdi prihod s PIN-om</Button>}
      <p><Button variant="secondary" onClick={refresh}>Osveži podatke</Button></p>
      <p className="field-hint">Referenca rezervacije: {reservation.id}</p>
      <p><Link to={paths.search}>Nazaj na iskanje virov</Link></p>
      {open && <CheckInModal reservation={reservation} options={{ token, logout, apiBaseUrl }}
        onClose={() => setOpen(false)} onSuccess={saved => { setData(current => ({ ...current, reservation: saved })); setOpen(false) }} />}
    </Card>
  )
}
