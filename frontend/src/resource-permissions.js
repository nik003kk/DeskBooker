// Backend persists roles as 'manager' / 'admin'; product labels are OFFICE_MANAGER / ADMIN.
export function canManageResources(user) {
  return Boolean(user && (user.role==='manager' || user.role==='admin') && user.status==='active')
}
