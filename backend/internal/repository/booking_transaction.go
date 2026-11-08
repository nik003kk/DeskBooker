package repository

import (
 "context"
 "errors"
 "fmt"

 "github.com/jackc/pgx/v5"
 "github.com/nik003kk/DeskBooker/backend/internal/models"
)

// ErrResourceUnavailable means a resource was deactivated before booking.
var ErrResourceUnavailable=errors.New("resource is not active")

// TxStarter is implemented by *pgxpool.Pool, permitting transaction tests.
type TxStarter interface { Begin(context.Context)(pgx.Tx,error) }
type BookingStore struct { db TxStarter }
func NewBookingStore(db TxStarter)*BookingStore{return &BookingStore{db:db}}

// CreateConfirmedReservation serializes writes against updates/deactivation
// of the resource row. The GiST exclusion constraint is the final safeguard
// against concurrent overlapping bookings (SQLSTATE 23P01).
func(s *BookingStore)CreateConfirmedReservation(ctx context.Context,in models.CreateReservationInput)(models.Reservation,error){
 if in.UserID==""||in.ResourceID==""||in.StartsAt.IsZero()||in.EndsAt.IsZero()||!in.EndsAt.After(in.StartsAt){
  return models.Reservation{},errors.New("invalid confirmed reservation")
 }
 tx,err:=s.db.Begin(ctx)
 if err!=nil{return models.Reservation{},fmt.Errorf("begin booking: %w",err)}
 defer tx.Rollback(ctx)
 var active bool
 if err=tx.QueryRow(ctx,"SELECT is_active FROM public.resources WHERE id=$1 FOR UPDATE",in.ResourceID).Scan(&active);err!=nil{
  return models.Reservation{},fmt.Errorf("lock resource: %w",err)
 }
 if !active{return models.Reservation{},ErrResourceUnavailable}
 booked,err:=scanReservation(tx.QueryRow(ctx,
  "INSERT INTO public.reservations(user_id,resource_id,starts_at,ends_at,status,confirmed_at) "+
   "VALUES($1,$2,$3,$4,'CONFIRMED',NOW()) RETURNING "+reservationColumns,
  in.UserID,in.ResourceID,in.StartsAt,in.EndsAt))
 if err!=nil{return models.Reservation{},fmt.Errorf("insert confirmed reservation: %w",err)}
 if err=tx.Commit(ctx);err!=nil{return models.Reservation{},fmt.Errorf("commit reservation: %w",err)}
 return booked,nil
}
