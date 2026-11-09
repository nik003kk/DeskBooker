export const emptySearchFilters = Object.freeze({
  type: '', date: '', start: '', end: '', features: '',
})

const validTypes = new Set(['', 'DESK', 'MEETING_ROOM', 'PARKING_SPOT'])
const datePattern = /^\d{4}-\d{2}-\d{2}$/
const clockPattern = /^([01]\d|2[0-3]):[0-5]\d$/

export function parseSearchFeatures(raw) {
  if (!raw?.trim()) return []
  return raw.split(',').map(s => s.trim()).filter(Boolean)
}

export function validateSearchFilters(values) {
  const errors = {}
  if (!validTypes.has(values.type || '')) errors.type = 'Neveljavna vrsta vira.'
  const date = values.date || ''
  const start = values.start || ''
  const end = values.end || ''
  if (date || start || end) {
    const validDate = datePattern.test(date) && !Number.isNaN(Date.parse(date)) &&
      new Date(date + 'T12:00:00Z').toISOString().slice(0, 10) === date
    if (!validDate || !clockPattern.test(start) || !clockPattern.test(end) || start >= end) {
      errors.date = 'Vnesi veljaven datum in začetni čas pred končnim časom.'
    }
  }
  const names = parseSearchFeatures(values.features)
  if (names.length > 20 || names.some(name => [...name].length > 100 ||
      /[\x00-\x1f\x7f]/.test(name)) ||
      new Set(names.map(name => name.toLocaleLowerCase())).size !== names.length) {
    errors.features = 'Dovoljenih je največ 20 različnih lastnosti do 100 znakov.'
  }
  if (values.features?.trim() && (values.features.startsWith(',') ||
      values.features.endsWith(',') || values.features.includes(',,'))) {
    errors.features = 'Oprema naj bo ločena z vejicami, brez praznih vrednosti.'
  }
  return errors
}
