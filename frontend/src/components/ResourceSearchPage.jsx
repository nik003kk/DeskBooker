import { Card, StatusBadge } from './ui.jsx'

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
        <Card className="resource-search-panel" aria-labelledby="resource-search-filter-heading">
          <h2 id="resource-search-filter-heading">Filtri</h2>
          <p>Izberi termin, vrsto vira in zahtevano opremo.</p>
          <p className="demo-notice">Izbor filtrov bo dodan v naslednjem koraku.</p>
        </Card>
        <section className="resource-search-results" aria-labelledby="resource-search-results-heading">
          <div className="section-heading">
            <h2 id="resource-search-results-heading">Rezultati iskanja</h2>
          </div>
          <Card className="empty-state" role="status">
            <h3>Pripravljeno za iskanje</h3>
            <p>Ko bodo filtri povezani z API-jem, bodo na tem mestu prikazani prosti viri.</p>
          </Card>
        </section>
      </div>
    </div>
  )
}
