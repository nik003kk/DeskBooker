import { useEffect, useRef, useState } from 'react'
import { useAuth } from '../AuthContext.jsx'
import { listAdminUsers, changeUserRole, userRoleLabels, userStatusLabels } from '../admin-users-api.js'
import { Button, Card, StatusBadge } from './ui.jsx'

export default function AdminUsersPage() {
  const { token, logout, user } = useAuth()
  const [items, setItems] = useState([])
  const [total, setTotal] = useState(0)
  const [offset, setOffset] = useState(0)
  const [refresh, setRefresh] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [drafts, setDrafts] = useState({})
  const [pending, setPending] = useState(null)
  const [saving, setSaving] = useState(false)
  const inflight = useRef(false)
  const alive = useRef(true)
  const mutation = useRef(null)
  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'
  const limit = 25
  useEffect(() => { alive.current = true; return () => { alive.current = false; mutation.current?.abort() } }, [])
  useEffect(() => {
    const controller = new AbortController()
    setLoading(true); setError(''); setItems([]); setDrafts({}); setPending(null)
    listAdminUsers({ token, logout, apiBaseUrl, limit, offset, signal: controller.signal }).then(page => {
      if (!controller.signal.aborted) { setItems(page.users); setTotal(page.total) }
    }).catch(err => { if (!controller.signal.aborted) setError(err.status === 403 ? 'Nimaš več pravic za upravljanje uporabnikov.' : 'Seznama uporabnikov ni mogoče naložiti. Poskusi znova.') })
      .finally(() => { if (!controller.signal.aborted) setLoading(false) })
    return () => controller.abort()
  }, [token, logout, apiBaseUrl, offset, refresh])
  async function save() {
    if (inflight.current || !pending) return
    inflight.current = true; setSaving(true); setError(''); setMessage('')
    const controller = new AbortController(); mutation.current = controller
    try {
      const saved = await changeUserRole(pending.user.id, pending.role, { token, logout, apiBaseUrl, currentUserId: user.id, signal: controller.signal })
      if (alive.current) {
        setItems(current => current.map(row => row.id === saved.id ? saved : row)); setDrafts(current => { const next = { ...current }; delete next[saved.id]; return next })
        setPending(null); setMessage(`Vloga za ${saved.first_name} ${saved.last_name} je shranjena.`)
      }
    } catch (err) {
      if (alive.current && !controller.signal.aborted) {
        if (err.status === 403) { setItems([]); setPending(null); setError('Sprememba ni dovoljena. Osveži seznam in preveri svoje pravice.') }
        else setError(err.status === 404 ? 'Uporabnik ni več na voljo. Osveži seznam.' : 'Sprememba vloge ni uspela. Poskusi znova.')
      }
    } finally { inflight.current = false; if (alive.current) setSaving(false) }
  }
  return <Card className="admin-users-page">
    <span className="eyebrow">Administracija</span>
    <div className="section-heading"><div><h1>Uporabniki</h1><p>Upravljaj vloge zaposlenih in vodij pisarne.</p></div><Button variant="secondary" disabled={loading || saving} onClick={() => { setMessage(''); setRefresh(value => value + 1) }}>Osveži seznam</Button></div>
    <p className="field-hint">Svoja vloga je zaščitena. Sprememba drugega uporabnika začne veljati ob naslednji zaščiteni API zahtevi.</p>
    {loading && <p role="status">Nalagam uporabnike ...</p>}
    {error && !pending && <p role="alert" className="auth-feedback auth-feedback--error">{error}</p>}
    {message && <p role="status" className="admin-users-success">{message}</p>}
    {!loading && !error && items.length === 0 && <p role="status">Na tej strani ni uporabnikov.</p>}
    {!loading && items.length > 0 && <div className="admin-table-wrap"><table className="admin-table"><caption className="sr-only">Seznam uporabnikov in njihovih vlog</caption>
      <thead><tr><th>Uporabnik</th><th>E-pošta</th><th>Oddelek</th><th>Vloga</th><th>Stanje</th><th>Dejanje</th></tr></thead>
      <tbody>{items.map(item => {
        const own = item.id.toLowerCase() === user.id.toLowerCase()
        const role = drafts[item.id] ?? item.role
        return <tr key={item.id}><th scope="row">{item.first_name} {item.last_name}{own && <small className="admin-users-own">Tvoj račun</small>}</th><td>{item.email}</td><td>{item.department || '—'}</td>
          <td><select aria-label={'Vloga za ' + item.email} value={role} disabled={own || saving} onChange={event => setDrafts(current => ({ ...current, [item.id]: event.target.value }))}>{Object.entries(userRoleLabels).map(([value, label]) => <option key={value} value={value}>{label}</option>)}</select></td>
          <td><StatusBadge tone={item.status === 'active' ? 'green' : 'neutral'}>{userStatusLabels[item.status]}</StatusBadge></td>
          <td>{own ? <span className="field-hint">Zaščiten račun</span> : <Button variant="secondary" disabled={saving || role === item.role} onClick={() => { setError(''); setMessage(''); setPending({ user: item, role }) }}>Shrani vlogo</Button>}</td></tr>
      })}</tbody></table></div>}
    {!loading && !error && <div className="admin-users-pagination"><p>Prikaz {items.length ? offset + 1 : 0}–{offset + items.length} od {total}</p><Button variant="secondary" disabled={saving || offset === 0} onClick={() => setOffset(value => Math.max(0, value - limit))}>Prejšnja stran</Button><Button variant="secondary" disabled={saving || offset + limit >= total} onClick={() => setOffset(value => value + limit)}>Naslednja stran</Button></div>}
    {pending && <RoleDialog error={error} pending={pending} saving={saving} onClose={() => setPending(null)} onSave={save} />}
  </Card>
}
function RoleDialog({ pending, saving, onClose, onSave, error }) {
  const dialog = useRef(null)
  useEffect(() => { const previous = document.activeElement; const element = dialog.current; element.showModal(); return () => { element.close(); previous?.focus() } }, [])
  return <dialog className="check-in-dialog" ref={dialog} aria-labelledby="role-dialog-title" onCancel={event => { if (saving) event.preventDefault(); else onClose() }}>
    <h2 id="role-dialog-title">Sprememba vloge</h2><p>{pending.user.first_name} {pending.user.last_name} · {pending.user.email}</p>
    <p>{userRoleLabels[pending.user.role]} → <strong>{userRoleLabels[pending.role]}</strong></p><p>Nove pravice bodo veljale pri naslednji zaščiteni zahtevi uporabnika.</p>
    {error && <p role="alert" className="auth-feedback auth-feedback--error">{error}</p>}
    <div className="check-in-actions"><Button disabled={saving} onClick={onSave}>{saving ? 'Shranjujem ...' : 'Potrdi spremembo'}</Button><Button variant="secondary" disabled={saving} onClick={onClose}>Prekliči</Button></div>
  </dialog>
}
