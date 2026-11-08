export const resourceTypeNames = Object.freeze({
  DESK: 'Delovna miza',
  MEETING_ROOM: 'Sejna soba',
  PARKING_SPOT: 'Parkirno mesto',
})

export function resourceCardDetails(resource) {
  return {
    title: resource.name || 'Neimenovan vir',
    typeName: resourceTypeNames[resource.type] || 'Rezervacijski vir',
    floor: Number.isInteger(resource.floor) ? resource.floor + '. nadstropje' : 'Nadstropje ni navedeno',
    capacity: Number.isInteger(resource.capacity) && resource.capacity > 0
      ? resource.capacity + ' oseb' : 'Kapaciteta ni navedena',
    features: Array.isArray(resource.features)
      ? resource.features.filter(name => typeof name === 'string' && name.trim())
      : [],
  }
}
