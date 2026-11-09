import { useState } from 'react'
import { Button, Card, Field, SelectField } from './ui.jsx'
import { emptySearchFilters, validateSearchFilters } from '../resource-search-filters.js'

const types = [
  { value: '', label: 'Vse vrste virov' },
  { value: 'DESK', label: 'Delovne mize' },
  { value: 'MEETING_ROOM', label: 'Sejne sobe' },
  { value: 'PARKING_SPOT', label: 'Parkirna mesta' },
]

export default function ResourceSearchFilters({ onSearch, onReset, busy = false, fixedType = '', onChange }) {
  const [values, setValues] = useState({ ...emptySearchFilters, type: fixedType })
  const [errors, setErrors] = useState({})
  function update(field, value) {
    if (field === 'type' && fixedType) return
    onChange?.()
    setValues(current => ({ ...current, [field]: value }))
    setErrors({})
  }
  function submit(event) {
    event.preventDefault()
    const problems = validateSearchFilters(values)
    setErrors(problems)
    if (Object.keys(problems).length) return
    onSearch({ ...values, type: fixedType || values.type })
  }
  function reset() {
    setValues({ ...emptySearchFilters, type: fixedType })
    setErrors({})
    onReset()
  }
  return (
    <Card className="resource-search-panel" aria-labelledby="resource-search-filter-heading">
      <h2 id="resource-search-filter-heading">Filtri</h2>
      <form className="auth-form" role="search" onSubmit={submit} noValidate>
        <SelectField label="Vrsta vira" options={types} value={values.type} disabled={Boolean(fixedType)}
          onChange={e => update('type', e.target.value)} />
        <Field label="Datum" type="date" value={values.date} onChange={e => update('date', e.target.value)} />
        <Field label="Začetni čas" type="time" step={60} value={values.start}
          onChange={e => update('start', e.target.value)} />
        <Field label="Končni čas" type="time" step={60} value={values.end}
          onChange={e => update('end', e.target.value)} />
        {errors.date && <p className="form-error" role="alert">{errors.date}</p>}
        <Field label="Oprema (ločena z vejicami)" value={values.features}
          placeholder="Monitor, USB-C" onChange={e => update('features', e.target.value)}
          hint="Vir mora vsebovati vse izbrane lastnosti." />
        {errors.features && <p className="form-error" role="alert">{errors.features}</p>}
        {errors.type && <p className="form-error" role="alert">{errors.type}</p>}
        <Button type="submit" disabled={busy}>{busy ? 'Iščem ...' : 'Poišči vire'}</Button>
        <Button variant="ghost" onClick={reset} disabled={busy}>Ponastavi filtre</Button>
      </form>
    </Card>
  )
}

