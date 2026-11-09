import { Link } from 'react-router-dom'
import { useAuth } from '../AuthContext.jsx'
import { canAdministerUsers } from '../admin-users-api.js'
import ProtectedRoute from './ProtectedRoute.jsx'
import { Card } from './ui.jsx'
import { paths } from '../routes.js'

export default function AdminUsersRoute({ children }) {
  const { user, checking, isAuthenticated } = useAuth()
  return <ProtectedRoute>{checking ? <p role="status">Preverjam administratorske pravice ...</p> : !isAuthenticated ? null : canAdministerUsers(user) ? children
    : <Card><h1>Uporabniki niso na voljo</h1><p role="alert">Za upravljanje uporabnikov potrebuješ aktivno vlogo administratorja.</p><Link to={paths.home}>Na začetno stran</Link></Card>}</ProtectedRoute>
}
