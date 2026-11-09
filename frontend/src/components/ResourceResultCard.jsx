import { Button, Card, StatusBadge } from './ui.jsx'
import { resourceCardDetails } from '../resource-card.js'

export default function ResourceResultCard({ resource, onBook, canBook = false }) {
  const info = resourceCardDetails(resource)
  return (
    <Card as="article" className="listing-card resource-result-card">
      <div className="listing-top">
        <span className="listing-icon" aria-hidden="true">▤</span>
        <StatusBadge tone="blue">{info.typeName}</StatusBadge>
      </div>
      <h3>{info.title}</h3>
      <p>{info.floor} · {info.capacity}</p>
      <div aria-label="Oprema vira" className="resource-feature-tags">
        {info.features.length
          ? info.features.map(feature => <span key={feature} className="resource-feature-tag">{feature}</span>)
          : <span className="field-hint">Oprema ni navedena.</span>}
      </div>
      <Button variant="secondary" disabled={!canBook || typeof onBook !== 'function'}
        onClick={() => onBook?.(resource)}
        title={canBook ? 'Preglej in potrdi rezervacijo' : 'Za rezervacijo najprej izberi datum in čas ter klikni Poišči vire'}>
        Rezerviraj
      </Button>
    </Card>
  )
}
