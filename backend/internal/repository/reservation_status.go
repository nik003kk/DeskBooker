package repository

// Keep search and direct availability aligned with the database exclusion
// constraint. Historical states, including EXPIRED, never block a new booking.
const blockingReservationStatusesSQL = "('PENDING','CONFIRMED','CHECKED_IN')"
