import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Button, Card, Field, SelectField } from './ui.jsx'
import { useAuth } from '../AuthContext.jsx'
import { paths } from '../routes.js'
import { createResourceWithFeatures, validateResource } from '../resource-admin-api.js'

const defaults = { name:'', type:'DESK', floor:'', capacity:'1', features:'', is_active:true }
const types = [
  {value:'DESK',label:'Delovna miza'},
  {value:'MEETING_ROOM',label:'Sejna soba'},
  {value:'PARKING_SPOT',label:'Parkirno mesto'},
]

export default function CreateResourcePage() {
  const {token,logout} = useAuth()
  const [values,setValues] = useState(defaults)
  const [errors,setErrors] = useState({})
  const [notice,setNotice] = useState('')
  const [busy,setBusy] = useState(false)
  const [created,setCreated] = useState(null)
  const set = (key,value) => {
    setValues(v => ({...v,[key]:value}))
    setErrors(e => ({...e,[key]:undefined}))
    setNotice('')
  }
  async function submit(event) {
    event.preventDefault()
    if (busy) return
    const validation = validateResource(values)
    setErrors(validation)
    if (Object.keys(validation).length) return
    setBusy(true)
    setNotice('')
    try {
      const resource = await createResourceWithFeatures(values,{token,logout,apiBaseUrl:import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'})
      setCreated(resource)
      setValues(defaults)
    } catch (error) {
      if (error.resource) setCreated(error.resource)
      setNotice(error.message || 'Vira ni bilo mogoče shraniti.')
    } finally {
      setBusy(false)
    }
  }
  return <Card className="auth-card">
    <span className="eyebrow">Upravljanje virov</span>
    <h1>Dodaj nov vir</h1>
    {created && <p role="status" className="demo-notice">Ustvarjen vir: {created.name}. <Link to={paths.adminResources}>Na seznam virov</Link></p>}
    {notice && <p role="alert" className="auth-feedback auth-feedback--error">{notice}</p>}
    <form className="auth-form" onSubmit={submit} noValidate>
      <Field label="Naziv vira" required maxLength={150} value={values.name} onChange={e=>set('name',e.target.value)} aria-invalid={!!errors.name}/>
      {errors.name && <span className="form-error" role="alert">{errors.name}</span>}
      <SelectField label="Vrsta vira" options={types} value={values.type} onChange={e=>set('type',e.target.value)}/>
      <Field label="Nadstropje (neobvezno)" type="number" min={0} step={1} value={values.floor} onChange={e=>set('floor',e.target.value)} aria-invalid={!!errors.floor}/>
      {errors.floor && <span className="form-error" role="alert">{errors.floor}</span>}
      <Field label="Kapaciteta" type="number" required min={1} step={1} value={values.capacity} onChange={e=>set('capacity',e.target.value)} aria-invalid={!!errors.capacity}/>
      {errors.capacity && <span className="form-error" role="alert">{errors.capacity}</span>}
      <Field label="Oprema (ločena z vejicami)" value={values.features} onChange={e=>set('features',e.target.value)} hint="Primer: Monitor, USB-C, Projektor" aria-invalid={!!errors.features}/>
      {errors.features && <span className="form-error" role="alert">{errors.features}</span>}
      <label className="form-field"><span>Aktivno <input type="checkbox" checked={values.is_active} onChange={e=>set('is_active',e.target.checked)}/></span></label>
      <Button type="submit" disabled={busy}>{busy?'Shranjujem ...':'Dodaj vir'}</Button>
    </form>
  </Card>
}
