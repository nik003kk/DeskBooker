package repository

import (
 "context"
 "fmt"
 "time"
)

// IsResourceAvailable excludes inactive resources and overlapping
// PENDING/CONFIRMED/CHECKED_IN reservations. Boundaries are half-open [start,end).
// Cancelled and completed history is retained but does not block.
func (s *Store) IsResourceAvailable(ctx context.Context,id string,start,end time.Time)(bool,error){
 var available bool
 err:=s.db.QueryRow(ctx,
  "SELECT r.is_active AND NOT EXISTS ("+
   "SELECT 1 FROM public.reservations b "+
   "WHERE b.resource_id=r.id AND b.status IN "+blockingReservationStatusesSQL+" "+
   "AND b.starts_at < $3 AND b.ends_at > $2) "+
   "FROM public.resources r WHERE r.id=$1",
  id,start,end).Scan(&available)
 if err!=nil{return false,fmt.Errorf("check resource availability: %w",err)}
 return available,nil
}

