import { Link } from 'react-router-dom'
import { useAuth } from '../AuthContext.jsx'
import { canManageResources } from '../resource-permissions.js'
import { paths } from '../routes.js'
import ProtectedRoute from './ProtectedRoute.jsx'
import { Card } from './ui.jsx'

export default function AnalyticsRoute({ children }) {
  const { user, checking, isAuthenticated } = useAuth()
  return <ProtectedRoute>{checking ? <p role="status">Preverjam dostop do analitike ...</p>
    : !isAuthenticated ? null
    : canManageResources(user) ? children
    : <Card><h1>Analitika ni na voljo</h1><p role="alert">Za analitiko potrebuješ aktivno vlogo vodje ali skrbnika.</p><Link to={paths.home}>Na začetno stran</Link></Card>}
  </ProtectedRoute>
}
