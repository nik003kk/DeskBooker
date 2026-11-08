import { useState } from 'react'
import { Button, Field } from './ui.jsx'
import { useAuth } from '../AuthContext.jsx'
import { saveProfileEdit, validateProfileEdit } from '../profile-edit.js'

export default function EditProfileForm({ user, onSaved }) {
  const { token, logout } = useAuth()
  const [values, setValues] = useState({
    first_name: user.first_name || '', last_name: user.last_name || '', department: user.department || '',
  })
  const [errors, setErrors] = useState({})
  const [feedback, setFeedback] = useState('')
  const [busy, setBusy] = useState(false)
  const update = (name, value) => {
    setValues(current => ({ ...current, [name]: value }))
    setErrors(current => ({ ...current, [name]: undefined }))
    setFeedback('')
  }
  async function submit(event) {
    event.preventDefault()
    if (busy) return
    const issues = validateProfileEdit(values)
    setErrors(issues)
    if (Object.keys(issues).length) return
    setBusy(true)
    setFeedback('')
    try {
      const updated = await saveProfileEdit(values, {
        token, logout, apiBaseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api',
      })
      onSaved(updated)
      setFeedback('Profil je uspešno posodobljen.')
    } catch (error) {
      setFeedback(error.message || 'Sprememb ni bilo mogoče shraniti.')
    } finally { setBusy(false) }
  }
  return (
    <section className="profile-edit" aria-labelledby="profile-edit-title">
      <h2 id="profile-edit-title">Uredi profil</h2>
      <p>E-poštnega naslova in vloge ni mogoče spreminjati.</p>
      <form className="auth-form" onSubmit={submit} noValidate>
        <Field label="Ime" value={values.first_name} required maxLength={100}
          onChange={e => update('first_name',e.target.value)} aria-invalid={Boolean(errors.first_name)} />
        {errors.first_name && <span role="alert" className="form-error">{errors.first_name}</span>}
        <Field label="Priimek" value={values.last_name} required maxLength={100}
          onChange={e => update('last_name',e.target.value)} aria-invalid={Boolean(errors.last_name)} />
        {errors.last_name && <span role="alert" className="form-error">{errors.last_name}</span>}
        <Field label="Oddelek (neobvezno)" value={values.department} maxLength={100}
          onChange={e => update('department',e.target.value)} aria-invalid={Boolean(errors.department)} />
        {errors.department && <span role="alert" className="form-error">{errors.department}</span>}
        {feedback && <p role="status" className="auth-feedback">{feedback}</p>}
        <Button type="submit" disabled={busy}>{busy ? 'Shranjujem ...' : 'Shrani spremembe'}</Button>
      </form>
    </section>
  )
}
