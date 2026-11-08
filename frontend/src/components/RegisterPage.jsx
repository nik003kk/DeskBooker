import { useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import { Button, Card, Field, StatusBadge } from './ui.jsx'
import { paths } from '../routes.js'
import { submitRegistration, validateRegistration } from '../registration.js'

const initialValues = { first_name: '', last_name: '', email: '', password: '', confirm_password: '' }
const domain = import.meta.env.VITE_REGISTRATION_EMAIL_DOMAIN || ''
const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'

export default function RegisterPage() {
  const [values, setValues] = useState(initialValues)
  const [errors, setErrors] = useState({})
  const [serverError, setServerError] = useState('')
  const [createdUser, setCreatedUser] = useState(null)
  const [submitting, setSubmitting] = useState(false)
  const pending = useRef(null)

  useEffect(() => () => pending.current?.abort(), [])

  function update(field, value) {
    setValues((current) => ({ ...current, [field]: value }))
    setErrors((current) => ({ ...current, [field]: undefined }))
    setServerError('')
  }

  async function handleSubmit(event) {
    event.preventDefault()
    if (submitting) return
    const validation = validateRegistration(values, domain)
    setErrors(validation)
    setServerError('')
    if (Object.keys(validation).length > 0) return

    const controller = new AbortController()
    pending.current = controller
    setSubmitting(true)
    try {
      const user = await submitRegistration(values, { apiBaseUrl, signal: controller.signal })
      setCreatedUser(user)
      setValues(initialValues)
    } catch (error) {
      if (error.name !== 'AbortError') {
        setServerError(error instanceof TypeError
          ? 'Povezava s strežnikom ni uspela. Preveri, ali backend deluje.'
          : error.message || 'Registracija ni uspela.')
      }
    } finally {
      if (pending.current === controller) pending.current = null
      setSubmitting(false)
    }
  }

  if (createdUser) {
    return (
      <Card className="auth-card" aria-labelledby="register-success-title">
        <StatusBadge tone="blue">Račun ustvarjen</StatusBadge>
        <h1 id="register-success-title">Registracija je uspela!</h1>
        <p>Račun za <strong>{createdUser.email}</strong> je ustvarjen. Zdaj se lahko prijaviš.</p>
        <Link className="ui-button ui-button--primary" to={paths.login}>Na prijavo</Link>
      </Card>
    )
  }

  return (
    <div className="auth-layout">
      <Card className="auth-card" aria-labelledby="register-title">
        <span className="eyebrow">Dobrodošli v DeskBooker</span>
        <h1 id="register-title">Ustvari račun</h1>
        <p>Registriraj se s svojim službenim e-poštnim naslovom.</p>
        <form onSubmit={handleSubmit} noValidate className="auth-form">
          <div className="auth-name-grid">
            <Field label="Ime" autoComplete="given-name" required maxLength={100}
              value={values.first_name} onChange={(e) => update('first_name', e.target.value)}
              aria-invalid={Boolean(errors.first_name)} aria-errormessage={errors.first_name ? 'register-first-error' : undefined} />
            <Field label="Priimek" autoComplete="family-name" required maxLength={100}
              value={values.last_name} onChange={(e) => update('last_name', e.target.value)}
              aria-invalid={Boolean(errors.last_name)} aria-errormessage={errors.last_name ? 'register-last-error' : undefined} />
          </div>
          <div className="auth-name-grid auth-error-grid">
            <span id="register-first-error" className="form-error" role={errors.first_name ? 'alert' : undefined}>{errors.first_name}</span>
            <span id="register-last-error" className="form-error" role={errors.last_name ? 'alert' : undefined}>{errors.last_name}</span>
          </div>
          <Field label="Službeni e-poštni naslov" type="email" autoComplete="email" required
            placeholder={domain ? 'ime@' + domain : 'ime@podjetje.si'}
            value={values.email} onChange={(e) => update('email', e.target.value)}
            aria-invalid={Boolean(errors.email)} aria-errormessage={errors.email ? 'register-email-error' : undefined}
            hint={domain ? 'Dovoljena domena: @' + domain : 'Uporabi e-poštni naslov podjetja.'} />
          {errors.email && <span id="register-email-error" role="alert" className="form-error">{errors.email}</span>}
          <Field label="Geslo" type="password" autoComplete="new-password" required
            value={values.password} onChange={(e) => update('password', e.target.value)}
            aria-invalid={Boolean(errors.password)} aria-errormessage={errors.password ? 'register-password-error' : undefined}
            hint="Najmanj 12 in največ 72 bajtov; pri šumnikih se štejejo bajti." />
          {errors.password && <span id="register-password-error" role="alert" className="form-error">{errors.password}</span>}
          <Field label="Ponovi geslo" type="password" autoComplete="new-password" required
            value={values.confirm_password} onChange={(e) => update('confirm_password', e.target.value)}
            aria-invalid={Boolean(errors.confirm_password)} aria-errormessage={errors.confirm_password ? 'register-confirm-error' : undefined} />
          {errors.confirm_password && <span id="register-confirm-error" role="alert" className="form-error">{errors.confirm_password}</span>}
          {serverError && <div className="auth-feedback auth-feedback--error" role="alert">{serverError}</div>}
          <Button type="submit" disabled={submitting} className="auth-submit">
            {submitting ? 'Ustvarjanje računa ...' : 'Ustvari račun'}
          </Button>
        </form>
        <p className="auth-bottom">Že imaš račun? <Link to={paths.login}>Prijava</Link></p>
      </Card>
      <aside className="auth-aside">
        <span className="eyebrow">En račun, več možnosti</span>
        <h2>Tvoj delovni dan, po tvoje.</h2>
        <p>DeskBooker povezuje delovna mesta, sejne sobe, parkirišča in opremo.</p>
        <p className="demo-notice">Po registraciji lahko račun uporabiš za prijavo. Rezervacije bodo omogočene v naslednjih korakih.</p>
      </aside>
    </div>
  )
}
