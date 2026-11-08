import { useEffect, useRef, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Button, Card, Field } from './ui.jsx'
import { paths } from '../routes.js'
import { submitLogin, validateLogin } from '../login.js'

export default function LoginPage({ onLogin = () => {} }) {
  const navigate = useNavigate()
  const [values, setValues] = useState({ email: '', password: '' })
  const [errors, setErrors] = useState({})
  const [message, setMessage] = useState('')
  const [busy, setBusy] = useState(false)
  const controllerRef = useRef(null)
  useEffect(() => () => controllerRef.current?.abort(), [])

  function change(name, value) {
    setValues(current => ({ ...current, [name]: value }))
    setErrors(current => ({ ...current, [name]: undefined }))
    setMessage('')
  }
  async function submit(event) {
    event.preventDefault()
    if (busy) return
    const validation = validateLogin(values)
    setErrors(validation)
    if (Object.keys(validation).length) return
    setBusy(true)
    setMessage('')
    const controller = new AbortController()
    controllerRef.current = controller
    try {
      const result = await submitLogin(values, {
        apiBaseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api',
        signal: controller.signal,
      })
      setValues(current => ({ ...current, password: '' }))
      onLogin(result)
      navigate(paths.home, { replace: true })
    } catch (error) {
      if (error.name !== 'AbortError') setMessage(error instanceof TypeError
        ? 'Strežnik ni dosegljiv. Preveri povezavo.'
        : error.message || 'Prijava ni uspela.')
    } finally {
      if (controllerRef.current === controller) controllerRef.current = null
      setBusy(false)
    }
  }
  return (
    <Card className="auth-card" aria-labelledby="login-title">
      <span className="eyebrow">Dobrodošel nazaj</span>
      <h1 id="login-title">Prijava</h1>
      <p>Prijavi se z registriranim e-poštnim naslovom in geslom.</p>
      <form className="auth-form" onSubmit={submit} noValidate>
        <Field label="E-poštni naslov" type="email" autoComplete="username" required
          value={values.email} onChange={e => change('email', e.target.value)}
          aria-invalid={Boolean(errors.email)} />
        {errors.email && <span role="alert" className="form-error">{errors.email}</span>}
        <Field label="Geslo" type="password" autoComplete="current-password" required
          value={values.password} onChange={e => change('password', e.target.value)}
          aria-invalid={Boolean(errors.password)} />
        {errors.password && <span role="alert" className="form-error">{errors.password}</span>}
        {message && <div role="alert" className="auth-feedback auth-feedback--error">{message}</div>}
        <Button type="submit" disabled={busy} className="auth-submit">
          {busy ? 'Prijavljanje ...' : 'Prijava'}
        </Button>
      </form>
      <p className="auth-bottom">Še nimaš računa? <Link to={paths.register}>Registracija</Link></p>
    </Card>
  )
}
