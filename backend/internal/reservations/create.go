// Package reservations handles authenticated reservation lifecycle requests.
package reservations

import (
 "context"
 "encoding/json"
 "errors"
 "io"
 "mime"
 "net/http"
 "time"

 "github.com/gin-gonic/gin"
 "github.com/jackc/pgx/v5"
 "github.com/jackc/pgx/v5/pgconn"
 "github.com/nik003kk/DeskBooker/backend/internal/models"
 "github.com/nik003kk/DeskBooker/backend/internal/repository"
)

type Creator interface {
 CreateConfirmedReservation(context.Context,models.CreateReservationInput)(models.Reservation,error)
}
type createRequest struct {
 ResourceID string `json:"resource_id"`
 StartsAt time.Time `json:"starts_at"`
 EndsAt time.Time `json:"ends_at"`
}
func fail(c *gin.Context,status int,code,message string){
 c.JSON(status,gin.H{"error":gin.H{"code":code,"message":message}})
}

// CreateHandler derives the user from a verified JWT, never from JSON.
// Timestamps are RFC3339 with explicit offsets; database enforces nonoverlap.
func CreateHandler(store Creator)gin.HandlerFunc{
 return func(c *gin.Context){
  c.Header("Cache-Control","no-store")
  userID:=c.GetString("auth_user_id")
  if userID=="" {fail(c,401,"unauthorized","Authentication required");return}
  media,_,err:=mime.ParseMediaType(c.GetHeader("Content-Type"))
  if err!=nil||media!="application/json" {fail(c,415,"unsupported_media_type","Expected application/json");return}
  dec:=json.NewDecoder(io.LimitReader(c.Request.Body,8193))
  dec.DisallowUnknownFields()
  var req createRequest
  if err=dec.Decode(&req);err!=nil {fail(c,400,"invalid_json","Invalid reservation JSON");return}
  var extra any
  if err=dec.Decode(&extra);!errors.Is(err,io.EOF){fail(c,400,"invalid_json","Only one JSON object allowed");return}
  if !validUUID(req.ResourceID)||ValidateWindow(time.Now(),req.StartsAt,req.EndsAt)!=nil{
   fail(c,400,"invalid_reservation","Invalid resource or time interval");return
  }
  saved,err:=store.CreateConfirmedReservation(c.Request.Context(),models.CreateReservationInput{
   UserID:userID,ResourceID:req.ResourceID,StartsAt:req.StartsAt,EndsAt:req.EndsAt,
  })
  if err!=nil{
   var pgErr *pgconn.PgError
   if errors.As(err,&pgErr) {
    if pgErr.Code=="23P01" {fail(c,409,"reservation_conflict","The selected time is already booked");return}
    if pgErr.Code=="23503" {fail(c,404,"resource_not_found","Resource or user no longer exists");return}
   }
   if errors.Is(err,pgx.ErrNoRows){fail(c,404,"resource_not_found","Resource not found");return}
   if errors.Is(err,repository.ErrResourceUnavailable){fail(c,409,"resource_unavailable","Resource is not active");return}
   fail(c,500,"reservation_failed","Could not create reservation");return
  }
  // PIN is returned only on initial creation, never in list DTOs.
  response:=gin.H{"reservation":saved.Public()}
  if saved.CheckInCode!=nil {response["check_in_pin"]=*saved.CheckInCode}
  c.JSON(http.StatusCreated,response)
 }
}
func validUUID(id string)bool {
 if len(id)!=36{return false}
 for i,r:=range id{
  switch i {case 8,13,18,23: if r!='-'{return false}
  default: if !(r>='0'&&r<='9'||r>='a'&&r<='f'||r>='A'&&r<='F'){return false}}
 }
 return true
}
