import { Link, useLocation } from 'react-router-dom'
import { Card, StatusBadge, Button } from './ui.jsx'
import { paths } from '../routes.js'
import { bookingSummary } from '../booking-selection.js'
import { resourceTypeNames } from '../resource-card.js'

export default function BookingConfirmPage() {
  const { state } = useLocation()
  const summary = bookingSummary(state)
  if (!summary) return (
    <Card className="auth-card" aria-labelledby="booking-title">
      <h1 id="booking-title">Izberi vir in termin</h1>
      <p>Za rezervacijo najprej poišči prost vir in določi datum ter uro.</p>
      <Link to={paths.search}>Nazaj na iskanje virov</Link>
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
      <p className="demo-notice">Pregled podatkov. Potrditev rezervacije bo povezana z API-jem v DB-055.</p>
      <Button disabled>Potrdi (kmalu)</Button>
      <p className="auth-bottom"><Link to={paths.search}>Nazaj na iskanje</Link></p>
    </Card>
  )
}
