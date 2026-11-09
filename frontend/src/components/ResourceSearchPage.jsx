import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { paths } from '../routes.js'
import { Card, StatusBadge, Button } from './ui.jsx'
import { useAuth } from '../AuthContext.jsx'
import ResourceSearchFilters from './ResourceSearchFilters.jsx'
import ResourceResultCard from './ResourceResultCard.jsx'
import { emptySearchFilters } from '../resource-search-filters.js'
import { searchAvailableResources } from '../resource-search-api.js'

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'

export default function ResourceSearchPage({ fixedType = '', title = 'Iskanje virov', description = 'Preveri proste delovne mize, sejne sobe in parkirišča glede na termin in opremo.' }) {
  const { token, logout } = useAuth()
  const navigate = useNavigate()
  const baseFilters = useMemo(() => ({ ...emptySearchFilters, type: fixedType }), [fixedType])
  const [selectedFilters, setSelectedFilters] = useState(baseFilters)
  const [dirty, setDirty] = useState(false)
  const [resources, setResources] = useState([])
  const [busy, setBusy] = useState(true)
  const [error, setError] = useState('')
  const [searched, setSearched] = useState(false)
  const pending = useRef(null)
  const applied = useRef({ filters: baseFilters, isFiltered: false })

  const load = useCallback(async (filters, isFiltered = false, preserveDraft = false) => {
    pending.current?.abort()
    const controller = new AbortController()
    pending.current = controller
    const scoped = { ...filters, type: fixedType || filters.type || '' }
    if (!preserveDraft) setDirty(false)
    applied.current = { filters: scoped, isFiltered }
    setBusy(true)
    setError('')
    setSearched(isFiltered)
    setSelectedFilters(scoped)
    try {
      const result = await searchAvailableResources(scoped, {
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
  }, [fixedType, token, logout])

  useEffect(() => {
    load(baseFilters)
    return () => { pending.current?.abort(); pending.current = null }
  }, [load, baseFilters])

  const refresh = useCallback(() => load(applied.current.filters, applied.current.isFiltered, true), [load])
  useEffect(() => {
    const resume = () => { if (document.visibilityState !== 'hidden') refresh() }
    window.addEventListener('focus', resume)
    document.addEventListener('visibilitychange', resume)
    return () => { window.removeEventListener('focus', resume); document.removeEventListener('visibilitychange', resume) }
  }, [refresh])

  return (
    <div className="page-stack resource-search-page">
      <header className="page-intro">
        <div>
          <span className="eyebrow">Rezervacijski viri</span>
          <h1>{title}</h1>
          <p>{description}</p>
        </div>
        <div className="resource-catalog-actions"><StatusBadge tone="blue">Dejanski podatki</StatusBadge><Button variant="secondary" disabled={busy} onClick={refresh}>Osveži vire</Button></div>
      </header>
      <div className="resource-search-layout">
        <ResourceSearchFilters key={fixedType || 'all'} fixedType={fixedType} busy={busy} onChange={() => setDirty(true)}
          onSearch={filters => load(filters, true)}
          onReset={() => load(baseFilters)} />
        <section className="resource-search-results" aria-labelledby="resource-search-results-heading">
          <div className="section-heading">
            <h2 id="resource-search-results-heading">{selectedFilters.date ? 'Razpoložljivi viri' : 'Aktivni viri'}</h2>
            {!busy && !error && <span className="result-count" aria-live="polite">
              {resources.length} {resources.length === 1 ? 'rezultat' : 'rezultatov'}
            </span>}
          </div>
          <p className="field-hint">{selectedFilters.date
            ? `Uporabljeni termin: ${selectedFilters.date}, ${selectedFilters.start}–${selectedFilters.end} · Europe/Ljubljana`
            : 'Prikazani so aktivni viri. Za preverjanje razpoložljivosti in rezervacijo izberi datum in čas ter klikni Poišči vire.'}</p>
          {dirty && <p role="status" className="resource-filter-notice">Filtri so spremenjeni. Za nove rezultate klikni Poišči vire.</p>}
          {busy && <Card className="empty-state" role="status"><p>Iščem razpoložljive vire ...</p></Card>}
          {!busy && error && (
            <Card className="empty-state" role="alert">
              <h3>Iskanje ni uspelo</h3>
              <p>{error}</p>
              <Button variant="secondary" onClick={refresh}>Poskusi znova</Button>
            </Card>
          )}
          {!busy && !error && resources.length === 0 && (
            <Card className="empty-state" role="status">
              <h3>{searched ? 'Ni ustreznih virov' : 'Ni aktivnih virov'}</h3>
              <p>{searched ? 'Za izbrani termin in lastnosti ni zadetkov. Spremeni ali ponastavi filtre.' : 'Aktivnih virov te vrste še ni. Ko jih administrator doda ali aktivira, osveži seznam.'}</p>
              {searched && <p className="field-hint">Za ponastavitev uporabi gumb »Ponastavi filtre«.</p>}
            </Card>
          )}
          {!busy && !error && resources.length > 0 && (
            <div className="resource-grid">
              {resources.map(resource => <ResourceResultCard key={resource.id} resource={resource}
                canBook={!dirty && Boolean(selectedFilters.date && selectedFilters.start && selectedFilters.end)}
                onBook={item => navigate(paths.bookingConfirm, {
                  state: { resource: item, filters: selectedFilters },
                })} />)}
            </div>
          )}
        </section>
      </div>
    </div>
  )
}

