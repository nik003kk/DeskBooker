import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Card, ButtonLink, StatusBadge } from './ui.jsx'
import { paths } from '../routes.js'
import { useAuth } from '../AuthContext.jsx'
import { listAdminResources } from '../resource-admin-api.js'

export default function AdminResourcesPage() {
  const { token, logout } = useAuth()
  const [items,setItems]=useState([])
  const [loading,setLoading]=useState(true)
  const [error,setError]=useState('')
  const reload=useCallback(async()=>{
    setLoading(true);setError('')
    try {setItems(await listAdminResources({token,logout,apiBaseUrl:import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'}))}
    catch(e){setError(e.message || 'Seznama ni mogoče naložiti.')}
    finally{setLoading(false)}
  },[token,logout])
  useEffect(()=>{reload()},[reload])
  return <Card className="admin-resource-page">
    <span className="eyebrow">Upravljanje</span>
    <div className="section-heading"><h1>Seznam virov</h1><ButtonLink to={paths.createResource}>Dodaj vir</ButtonLink></div>
    {loading && <p role="status">Nalagam vire ...</p>}
    {error && <p role="alert" className="auth-feedback auth-feedback--error">{error} <button type="button" onClick={reload}>Poskusi znova</button></p>}
    {!loading && !error && !items.length && <p role="status">Ni ustvarjenih virov.</p>}
    {!loading && !error && items.length>0 &&
      <div className="admin-table-wrap"><table className="admin-table">
        <thead><tr><th>Naziv</th><th>Vrsta</th><th>Nadstropje</th><th>Kapaciteta</th><th>Stanje</th></tr></thead>
        <tbody>{items.map(item=><tr key={item.id}><td>{item.name}</td><td>{item.type}</td>
          <td>{item.floor ?? '—'}</td><td>{item.capacity}</td>
          <td><StatusBadge tone={item.is_active?'blue':'neutral'}>{item.is_active?'Aktiven':'Neaktiven'}</StatusBadge></td></tr>)}</tbody>
      </table></div>
    }
  </Card>
}
