import { useEffect, useRef, useState } from 'react'
import { Card, StatusBadge, Button } from './ui.jsx'
import { useAuth } from '../AuthContext.jsx'
import ResourceSearchFilters from './ResourceSearchFilters.jsx'
import ResourceResultCard from './ResourceResultCard.jsx'
import { emptySearchFilters } from '../resource-search-filters.js'
import { searchAvailableResources } from '../resource-search-api.js'

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'

export default function ResourceSearchPage() {
  const { token, logout } = useAuth()
  const [resources, setResources] = useState([])
  const [busy, setBusy] = useState(true)
  const [error, setError] = useState('')
  const [searched, setSearched] = useState(false)
  const pending = useRef(null)

  async function load(filters, isFiltered = false) {
    pending.current?.abort()
    const controller = new AbortController()
    pending.current = controller
    setBusy(true)
    setError('')
    setSearched(isFiltered)
    try {
      const result = await searchAvailableResources(filters, {
        token, logout, apiBaseUrl, signal: controller.signal,
      })
      if (pending.current === controller) setResources(result)
    } catch (err) {
      if (pending.current === controller && err.name !== 'AbortError') {
        setResources([])
        setError(err instanceof TypeError
          ? 'Povezava s strežnikom ni uspela.'
          : err.message || 'Virov ni bilo mogoče naložiti.')
      }
    } finally {
      if (pending.current === controller) {
        pending.current = null
        setBusy(false)
      }
    }
  }

  useEffect(() => {
    load(emptySearchFilters)
    return () => { pending.current?.abort(); pending.current = null }
  }, [token, logout])

  return (
    <div className="page-stack resource-search-page">
      <header className="page-intro">
        <div>
          <span className="eyebrow">Poišči pravo delovno mesto</span>
          <h1>Iskanje virov</h1>
          <p>Preveri proste delovne mize, sejne sobe in parkirišča glede na termin in opremo.</p>
        </div>
        <StatusBadge tone="blue">Dejanski podatki</StatusBadge>
      </header>
      <div className="resource-search-layout">
        <ResourceSearchFilters busy={busy}
          onSearch={filters => load(filters, true)}
          onReset={() => load(emptySearchFilters)} />
        <section className="resource-search-results" aria-labelledby="resource-search-results-heading">
          <div className="section-heading">
            <h2 id="resource-search-results-heading">Rezultati iskanja</h2>
            {!busy && !error && <span className="result-count" aria-live="polite">
              {resources.length} {resources.length === 1 ? 'rezultat' : 'rezultatov'}
            </span>}
          </div>
          {busy && <Card className="empty-state" role="status"><p>Iščem razpoložljive vire ...</p></Card>}
          {!busy && error && (
            <Card className="empty-state" role="alert">
              <h3>Iskanje ni uspelo</h3>
              <p>{error}</p>
              <Button variant="secondary" onClick={() => load(emptySearchFilters)}>Poskusi znova</Button>
            </Card>
          )}
          {!busy && !error && resources.length === 0 && (
            <Card className="empty-state" role="status">
              <h3>Ni ustreznih virov</h3>
              <p>Za izbrani termin in lastnosti ni zadetkov. Spremeni ali ponastavi filtre.</p>
              {searched && <p className="field-hint">Za ponastavitev uporabi gumb »Ponastavi filtre«.</p>}
            </Card>
          )}
          {!busy && !error && resources.length > 0 && (
            <div className="resource-grid">
              {resources.map(resource => <ResourceResultCard key={resource.id} resource={resource} />)}
            </div>
          )}
        </section>
      </div>
    </div>
  )
}
