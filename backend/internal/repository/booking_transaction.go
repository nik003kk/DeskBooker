package repository

import (
 "context"
 "errors"
 "fmt"

 "github.com/jackc/pgx/v5"
 "github.com/jackc/pgx/v5/pgconn"
 "github.com/nik003kk/DeskBooker/backend/internal/models"
)

var ErrResourceUnavailable=errors.New("resource is not active")

// TxStarter is implemented by *pgxpool.Pool.
type TxStarter interface { Begin(context.Context)(pgx.Tx,error) }
type BookingStore struct { db TxStarter }
func NewBookingStore(db TxStarter)*BookingStore{return &BookingStore{db:db}}

// Each attempt is its own transaction. PostgreSQL exclusion prevents
// overlapping bookings and a unique index prevents duplicate check-in PINs.
func(s *BookingStore)CreateConfirmedReservation(ctx context.Context,in models.CreateReservationInput)(models.Reservation,error){
 if in.UserID==""||in.ResourceID==""||in.StartsAt.IsZero()||in.EndsAt.IsZero()||!in.EndsAt.After(in.StartsAt){
  return models.Reservation{},errors.New("invalid confirmed reservation")
 }
 for attempt:=0;attempt<12;attempt++{
  pin,err:=newCheckInPIN()
  if err!=nil{return models.Reservation{},err}
  booked,err:=s.tryConfirmedReservation(ctx,in,pin)
  if err==nil{return booked,nil}
  var dbErr *pgconn.PgError
  if errors.As(err,&dbErr)&&dbErr.Code=="23505"&&dbErr.ConstraintName=="reservations_check_in_code_lower_unique_idx"{
   continue
  }
  return models.Reservation{},err
 }
 return models.Reservation{},errors.New("could not allocate a unique check-in PIN")
}

func(s *BookingStore)tryConfirmedReservation(ctx context.Context,in models.CreateReservationInput,pin string)(models.Reservation,error){
 tx,err:=s.db.Begin(ctx)
 if err!=nil{return models.Reservation{},fmt.Errorf("begin booking: %w",err)}
 defer tx.Rollback(ctx)
 var active bool
 if err=tx.QueryRow(ctx,"SELECT is_active FROM public.resources WHERE id=$1 FOR UPDATE",in.ResourceID).Scan(&active);err!=nil{
  return models.Reservation{},fmt.Errorf("lock resource: %w",err)
 }
 if !active{return models.Reservation{},ErrResourceUnavailable}
 booked,err:=scanReservation(tx.QueryRow(ctx,
  "INSERT INTO public.reservations(user_id,resource_id,starts_at,ends_at,status,confirmed_at,check_in_code) "+
   "VALUES($1,$2,$3,$4,'CONFIRMED',NOW(),$5) RETURNING "+reservationColumns,
  in.UserID,in.ResourceID,in.StartsAt,in.EndsAt,pin))
 if err!=nil{return models.Reservation{},fmt.Errorf("insert confirmed reservation: %w",err)}
 if err=tx.Commit(ctx);err!=nil{return models.Reservation{},fmt.Errorf("commit reservation: %w",err)}
 return booked,nil
}
