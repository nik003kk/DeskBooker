import { useEffect, useState } from 'react'
import { Card, StatusBadge } from './ui.jsx'
import { useAuth } from '../AuthContext.jsx'
import { accountStatusLabel, fetchProfile } from '../profile.js'

export default function ProfilePage() {
  const { token, logout } = useAuth()
  const [profile, setProfile] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  useEffect(() => {
    const controller = new AbortController()
    let active = true
    setLoading(true)
    fetchProfile({ token, logout, apiBaseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api',
      fetchImpl: (url, opts) => fetch(url, { ...opts, signal: controller.signal }) })
      .then(user => { if (active) setProfile(user) })
      .catch(err => { if (active && err.name !== 'AbortError') setError(err.message || 'Profila ni mogoče naložiti.') })
      .finally(() => { if (active) setLoading(false) })
    return () => { active = false; controller.abort() }
  }, [token, logout])
  return (
    <Card className="auth-card" aria-labelledby="profile-title">
      <span className="eyebrow">Uporabniški račun</span>
      <h1 id="profile-title">Moj profil</h1>
      {loading && <p role="status">Nalagam profil ...</p>}
      {error && <p role="alert" className="auth-feedback auth-feedback--error">{error}</p>}
      {!loading && profile && (
        <>
          <StatusBadge tone="blue">{accountStatusLabel(profile.status)}</StatusBadge>
          <dl className="profile-details">
            <div><dt>Ime</dt><dd>{profile.first_name}</dd></div>
            <div><dt>Priimek</dt><dd>{profile.last_name}</dd></div>
            <div><dt>E-poštni naslov</dt><dd>{profile.email}</dd></div>
            <div><dt>Oddelek</dt><dd>{profile.department || 'Ni naveden'}</dd></div>
            <div><dt>Vloga</dt><dd>{profile.role}</dd></div>
            <div><dt>Stanje računa</dt><dd>{accountStatusLabel(profile.status)}</dd></div>
          </dl>
        </>
      )}
    </Card>
  )
}
