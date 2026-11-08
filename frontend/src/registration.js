// Validation and API boundary for registration. Backend is authoritative.
const emailPattern = /^[^\s@<>]+@[^\s@<>]+\.[^\s@<>]+$/
const encoder = new TextEncoder()

export function validateRegistration(values, companyDomain = '') {
  const errors = {}
  for (const [key, label] of [['first_name', 'Ime'], ['last_name', 'Priimek']]) {
    const value = values[key] || ''
    if (!value.trim()) errors[key] = `${label} je obvezno polje.`
    else if (value !== value.trim() || [...value].length > 100 || /[\x00-\x1f\x7f]/.test(value)) {
      errors[key] = `${label} mora imeti največ 100 znakov in brez odvečnih presledkov.`
    }
  }
  const email = values.email || ''
  if (email.length > 254 || !emailPattern.test(email) || email !== email.trim()) {
    errors.email = 'Vnesi veljaven e-poštni naslov.'
  } else if (companyDomain.trim() && email.split('@')[1].toLowerCase() !== companyDomain.trim().toLowerCase()) {
    errors.email = `Uporabi e-naslov z domeno @${companyDomain.trim().toLowerCase()}.`
  }
  const password = values.password || ''
  const bytes = encoder.encode(password).length
  if (bytes < 12 || bytes > 72 || !password.trim()) {
    errors.password = 'Geslo mora vsebovati od 12 do 72 bajtov.'
  }
  if (!values.confirm_password || values.confirm_password !== password) {
    errors.confirm_password = 'Gesli se ne ujemata.'
  }
  return errors
}

export function registrationErrorMessage(status, code) {
  if (status === 409 || code === 'email_exists') return 'Ta e-poštni naslov je že registriran.'
  if (code === 'invalid_email') return 'E-poštni naslov ni dovoljen. Preveri službeno domeno.'
  if (code === 'invalid_password') return 'Geslo mora vsebovati od 12 do 72 bajtov.'
  if (code === 'invalid_name') return 'Preveri ime in priimek.'
  if (status === 503) return 'Registracija trenutno ni na voljo. Obrni se na skrbnika.'
  if (status >= 500) return 'Registracija trenutno ni uspela. Poskusi znova.'
  return 'Preveri vnesene podatke in poskusi znova.'
}

export async function submitRegistration(values, { apiBaseUrl = 'http://localhost:8080/api', signal, fetchImpl = fetch } = {}) {
  const payload = {
    first_name: values.first_name,
    last_name: values.last_name,
    email: values.email.toLowerCase(),
    password: values.password,
  }
  const response = await fetchImpl(`${apiBaseUrl.replace(/\/$/, '')}/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
    signal,
  })
  let body
  try { body = await response.json() } catch { body = null }
  if (!response.ok) {
    const error = new Error(registrationErrorMessage(response.status, body?.error?.code))
    error.status = response.status
    throw error
  }
  if (!body?.user?.id) throw new Error('Strežnik je vrnil neveljaven odgovor.')
  return body.user
}
