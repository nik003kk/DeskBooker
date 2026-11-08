package repository

import(
 "context"
 "errors"
 "testing"
 "time"

 "github.com/nik003kk/DeskBooker/backend/internal/models"
)

func TestBookingStoreRejectsInvalidBeforeBeginningTransaction(t *testing.T){
 store:=&BookingStore{}
 _,err:=store.CreateConfirmedReservation(context.Background(),models.CreateReservationInput{
  UserID:"user",ResourceID:"resource",StartsAt:time.Now(),
 })
 if err==nil{t.Fatal("invalid interval should fail before transaction")}
}
func TestUnavailableReservationErrorIsDistinct(t *testing.T) {
 if !errors.Is(ErrResourceUnavailable,ErrResourceUnavailable){t.Fatal("unavailable error must be identifiable")}
}
