export function validateLogin(values) {
  const errors = {}
  if (!/^[^\s@<>]+@[^\s@<>]+\.[^\s@<>]+$/.test(values.email || '')) {
    errors.email = 'Vnesi veljaven e-poštni naslov.'
  }
  if (!values.password) errors.password = 'Vnesi geslo.'
  return errors
}
export async function submitLogin(values, { fetchImpl = fetch, apiBaseUrl = 'http://localhost:8080/api', signal } = {}) {
  const response = await fetchImpl(apiBaseUrl.replace(/\/$/, '') + '/auth/login', {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email: values.email.toLowerCase(), password: values.password }), signal,
  })
  let data
  try { data = await response.json() } catch { data = null }
  if (!response.ok) {
    if (response.status === 401) throw new Error('Napačen e-poštni naslov ali geslo.')
    if (response.status >= 500) throw new Error('Prijava trenutno ni na voljo. Poskusi pozneje.')
    throw new Error('Preveri vnesene podatke.')
  }
  if (!data?.access_token || !data?.user?.id || data?.token_type !== 'Bearer') {
    throw new Error('Strežnik je vrnil neveljaven odgovor.')
  }
  return data
}
