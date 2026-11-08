import { Navigate } from 'react-router-dom'
import { useAuth } from '../AuthContext.jsx'
import { canManageResources } from '../resource-permissions.js'
import { paths } from '../routes.js'
import ProtectedRoute from './ProtectedRoute.jsx'

export default function ManagerRoute({ children }) {
  const { user, checking, isAuthenticated } = useAuth()
  return <ProtectedRoute>
    {checking ? <p role="status">Preverjam dovoljenja ...</p>
      : !isAuthenticated ? null
      : canManageResources(user) ? children
      : <section role="alert" className="auth-feedback auth-feedback--error">
          Za upravljanje virov potrebuješ aktivno vlogo vodje ali skrbnika.
          <p><NavigateFallback /></p>
        </section>}
  </ProtectedRoute>
}
function NavigateFallback() {
  return <a href={paths.home}>Na začetno stran</a>
}
