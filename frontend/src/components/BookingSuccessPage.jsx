import { Link, useLocation } from 'react-router-dom'
import { Card, StatusBadge } from './ui.jsx'
import { paths } from '../routes.js'
import { confirmedBookingDetails } from '../booking-success.js'
import { resourceTypeNames } from '../resource-card.js'

export default function BookingSuccessPage() {
  const { state } = useLocation()
  const details = confirmedBookingDetails(state)
  if (!details) return (
    <Card className="auth-card" aria-labelledby="booking-success-title">
      <h1 id="booking-success-title">Potrditev ni na voljo</h1>
      <p>Podatki in PIN so prikazani samo neposredno po uspešni rezervaciji. Če si osvežil stran, je ta prikaz lahko izgubljen.</p>
      <p>Ne ustvarjaj nove rezervacije samo zato, da bi znova prikazal PIN.</p>
      <Link to={paths.search}>Nazaj na iskanje virov</Link>
    </Card>
  )
  return (
    <Card className="auth-card booking-success-card" aria-labelledby="booking-success-title">
      <StatusBadge tone="blue">Rezervacija potrjena</StatusBadge>
      <h1 id="booking-success-title">Uspešno si rezerviral vir!</h1>
      <p>Rezervacija je potrjena. Shranjena je v podatkovni bazi DeskBooker.</p>
      <dl className="profile-details">
        <div><dt>Vir</dt><dd>{details.resourceName}</dd></div>
        <div><dt>Vrsta</dt><dd>{resourceTypeNames[details.resourceType] || details.resourceType}</dd></div>
        <div><dt>Nadstropje</dt><dd>{details.floor ?? 'Ni navedeno'}</dd></div>
        <div><dt>Datum</dt><dd>{details.date}</dd></div>
        <div><dt>Od</dt><dd>{details.start}</dd></div>
        <div><dt>Do</dt><dd>{details.end}</dd></div>
        <div><dt>Stanje</dt><dd>Potrjeno (CONFIRMED)</dd></div>
      </dl>
      <section aria-labelledby="booking-pin-title" className="booking-pin-section">
        <h2 id="booking-pin-title">Check-in PIN</h2>
        <p className="booking-pin" aria-label="Štirimestna PIN koda">{details.pin}</p>
        <p className="field-hint">PIN shrani na varno in ga ne deli z drugimi. Prikaže se samo pri uspešni potrditvi.</p>
      </section>
      <Link className="ui-button ui-button--primary" to={paths.search}>Poišči drug vir</Link>
      <p className="field-hint">Referenca rezervacije: {details.id}</p>
    </Card>
  )
}
