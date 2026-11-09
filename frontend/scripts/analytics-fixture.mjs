export function analyticsFixture(period, empty = false) {
  const statuses = ['PENDING', 'CONFIRMED', 'CHECKED_IN', 'COMPLETED', 'CANCELLED', 'EXPIRED']
  const points = [1, 2, 3, 4, 5].map(weekday => ({ weekday, booked_minutes: empty ? 0 : 120, capacity_minutes: 480, occupancy_rate: empty ? 0 : 25 }))
  return { ...period, timezone: 'Europe/Ljubljana', generated_at: '2026-11-09T10:00:00Z', total_reservations: empty ? 0 : 6, active_reservations: empty ? 0 : 2,
    status_counts: Object.fromEntries(statuses.map(s => [s, empty ? 0 : 1])), occupancy_rate: empty ? 0 : 25,
    no_show: { expired: empty ? 0 : 1, attended: empty ? 0 : 2, eligible: empty ? 0 : 3, rate: empty ? null : 33.33 },
    daily: points.map((point, i) => ({ ...point, date: '2026-11-0' + (i + 2) })), weekdays: points }
}
