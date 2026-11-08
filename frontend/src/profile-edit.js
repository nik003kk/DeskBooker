import { authorizedFetch } from './auth-api.js'
export function validateProfileEdit(values) {
  const errors = {}
  for (const [field,label] of [['first_name','Ime'],['last_name','Priimek']]) {
    const value = values[field] || ''
    if (!value.trim() || value !== value.trim() || [...value].length > 100 ||
        /[\x00-\x1f\x7f]/.test(value)) errors[field] = label + ' mora biti izpolnjeno (največ 100 znakov).'
  }
  const department = values.department || ''
  if (department && (department !== department.trim() || [...department].length > 100)) {
    errors.department = 'Oddelek naj ima največ 100 znakov brez odvečnih presledkov.'
  }
  return errors
}
export async function saveProfileEdit(values, { token, logout, fetchImpl = fetch, apiBaseUrl } = {}) {
  const payload = { first_name: values.first_name, last_name: values.last_name,
    department: values.department ? values.department : null }
  const response = await authorizedFetch('/users/me', {
    token, logout, fetchImpl, apiBaseUrl, method:'PATCH',
    headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload),
  })
  const data = await response.json()
  if (!data?.user?.id) throw new Error('Strežnik je vrnil neveljaven odgovor.')
  return data.user
}
