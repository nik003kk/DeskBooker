package reservations

import (
 "errors"
 "time"
)

var ErrInvalidReservationWindow=errors.New("invalid reservation time window")

const maxReservationDuration=24*time.Hour

// ValidateWindow requires a future, positive, minute-precision booking
// no longer than 24 hours. UTC instants are compared, not local wall clocks.
func ValidateWindow(now,start,end time.Time)error{
 if start.IsZero()||end.IsZero()||!end.After(start)||
  start.Before(now)||end.Sub(start)>maxReservationDuration||
  start.Second()!=0||start.Nanosecond()!=0||end.Second()!=0||end.Nanosecond()!=0 {
  return ErrInvalidReservationWindow
 }
 return nil
}
