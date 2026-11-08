package auth
import (
 "net/http"
 "net/http/httptest"
 "testing"
 "github.com/gin-gonic/gin"
 "github.com/nik003kk/DeskBooker/backend/internal/models"
)
func TestResourceManagerOnlyActiveManagerOrAdmin(t *testing.T){
 gin.SetMode(gin.TestMode)
 tokens,_:=NewJWTService(testJWTSecret)
 token,_,err:=tokens.Issue("test-user")
 if err!=nil {t.Fatal(err)}
 db:=&jwtUserDB{user:models.User{ID:"test-user",Status:models.UserActive}}
 router:=gin.New()
 router.POST("/admin",RequireJWT(tokens),RequireResourceManager(db),func(c *gin.Context){c.Status(http.StatusNoContent)})
 check:=func(role models.UserRole,status models.UserStatus,want int){
  t.Helper();db.user.Role=role;db.user.Status=status
  req:=httptest.NewRequest("POST","/admin",nil)
  req.Header.Set("Authorization","Bearer "+token)
  w:=httptest.NewRecorder();router.ServeHTTP(w,req)
  if w.Code!=want{t.Fatalf("role=%s status=%s expected %d got %d",role,status,want,w.Code)}
 }
 check(models.RoleEmployee,models.UserActive,403)
 check(models.RoleManager,models.UserActive,204)
 check(models.RoleAdmin,models.UserActive,204)
 check(models.RoleManager,models.UserSuspended,403)
 check(models.RoleAdmin,models.UserInactive,403)
 check(models.UserRole("OFFICE_MANAGER"),models.UserActive,403)
 check(models.UserRole("ADMIN"),models.UserActive,403)
}
