import { useState } from 'react'
import { analyticsQuery, defaultAnalyticsPeriod } from '../analytics-api.js'
import { Button, Field, SelectField } from './ui.jsx'

export default function AnalyticsFilters({ period, loading, onChange, onRefresh }) {
  const [preset, setPreset] = useState('week')
  const [draft, setDraft] = useState(period)
  const [error, setError] = useState('')
  function select(value) {
    setPreset(value); setError('')
    if (value !== 'custom') { const next = defaultAnalyticsPeriod(value === 'week' ? 7 : 30); setDraft(next); onChange(next) }
  }
  function apply(event) {
    event.preventDefault()
    try { analyticsQuery(draft); setError(''); onChange({ ...draft }) } catch (err) { setError(err.message) }
  }
  return <form className="analytics-filters" onSubmit={apply} aria-label="Obdobje analitike">
    <SelectField label="Obdobje" value={preset} onChange={event => select(event.target.value)} options={[{ value: 'week', label: 'Zadnjih 7 dni' }, { value: 'month', label: 'Zadnjih 30 dni' }, { value: 'custom', label: 'Izbrano obdobje' }]} />
    {preset === 'custom' && <><Field label="Od" type="date" required value={draft.from} onChange={event => setDraft({ ...draft, from: event.target.value })} /><Field label="Do" type="date" required value={draft.to} onChange={event => setDraft({ ...draft, to: event.target.value })} /><Button type="submit">Uporabi obdobje</Button></>}
    <Button variant="secondary" disabled={loading} onClick={onRefresh}>{loading ? 'Osvežujem ...' : 'Osveži podatke'}</Button>
    {error && <p role="alert" className="auth-feedback auth-feedback--error">{error}</p>}
  </form>
}
