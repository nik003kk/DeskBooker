import { Card, StatusBadge } from './ui.jsx'
import ResourceSearchFilters from './ResourceSearchFilters.jsx'
import ResourceResultCard from './ResourceResultCard.jsx'

// The page layout is introduced independently of its API wiring.
// DB-047 adds filters, DB-048 adds result cards, DB-049 connects the API.
export default function ResourceSearchPage() {
  return (
    <div className="page-stack resource-search-page">
      <header className="page-intro">
        <div>
          <span className="eyebrow">Poišči pravo delovno mesto</span>
          <h1>Iskanje virov</h1>
          <p>Filtriraj razpoložljive mize, sejne sobe in parkirna mesta glede na svoje potrebe.</p>
        </div>
        <StatusBadge tone="blue">Iskanje</StatusBadge>
      </header>
      <div className="resource-search-layout">
        <ResourceSearchFilters onSearch={() => {}} onReset={() => {}} />
        <section className="resource-search-results" aria-labelledby="resource-search-results-heading">
          <div className="section-heading">
            <h2 id="resource-search-results-heading">Rezultati iskanja</h2>
          </div>
          <p className="demo-notice">Primer prikaza rezultata. V DB-049 bodo podatki prišli iz PostgreSQL.</p>
          <div className="resource-grid">
            <ResourceResultCard resource={{
              id: 'demo-resource', name: 'Primer: Miza A-101', type: 'DESK',
              floor: 1, capacity: 1, features: ['Monitor', 'USB-C'],
            }} />
          </div>
        </section>
      </div>
    </div>
  )
}
