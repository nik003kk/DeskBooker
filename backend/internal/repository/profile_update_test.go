package repository

import (
    "context"
    "strings"
    "testing"

    "github.com/jackc/pgx/v5"
    "github.com/nik003kk/DeskBooker/backend/internal/models"
)
func TestUpdateProfileUsesBoundParamsAndCannotUpdateEmail(t *testing.T) {
    db:=&stubDB{err:pgx.ErrNoRows}
    first:="Ana"
    _,err:=New(db).UpdateUserProfile(context.Background(),"user-id",models.UpdateUserProfileInput{FirstName:&first})
    if err==nil {t.Fatal("expected no rows")}
    if strings.Contains(db.sql,"SET email") || !strings.Contains(db.sql,"UPDATE public.users SET") {
        t.Fatalf("unexpected SQL: %s",db.sql)
    }
    if len(db.args)!=7 || db.args[0]!="user-id" {t.Fatalf("bound arguments: %#v",db.args)}
}
