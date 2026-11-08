import { Button, Card, StatusBadge } from './ui.jsx'
import { resourceCardDetails } from '../resource-card.js'

export default function ResourceResultCard({ resource }) {
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
      <Button variant="secondary" disabled
        title="Rezerviranje bo omogočeno v naslednjih nalogah">
        Rezerviraj (kmalu)
      </Button>
    </Card>
  )
}
