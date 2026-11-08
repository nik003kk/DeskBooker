import { Navigate, useLocation } from 'react-router-dom'
import { useAuth } from '../AuthContext.jsx'
import { paths } from '../routes.js'

export default function ProtectedRoute({ children }) {
  const { checking, isAuthenticated } = useAuth()
  const location = useLocation()
  if (checking) return <p role="status">Preverjam prijavo ...</p>
  if (!isAuthenticated) return <Navigate to={paths.login} state={{ from: location.pathname }} replace />
  return children
}
