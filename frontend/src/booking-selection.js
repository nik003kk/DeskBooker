export function validateBookingSelection(selection) {
  const resource = selection?.resource
  const filters = selection?.filters
  if (!resource || typeof resource.id !== 'string' ||
      !/^[0-9a-f]{8}-[0-9a-f-]{27}$/i.test(resource.id) ||
      typeof resource.name !== 'string' || !resource.name ||
      !filters?.date || !filters?.start || !filters?.end || filters.start >= filters.end) {
    return false
  }
  return true
}

export function bookingSummary(selection) {
  if (!validateBookingSelection(selection)) return null
  return {
    resourceName: selection.resource.name,
    resourceType: selection.resource.type,
    floor: Number.isInteger(selection.resource.floor) ? selection.resource.floor : null,
    date: selection.filters.date,
    start: selection.filters.start,
    end: selection.filters.end,
  }
}
