package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// DatabasePinger allows the router to check database availability independently
// from the concrete PostgreSQL driver.
type DatabasePinger interface {
	Ping(context.Context) error
}

// NewRouter sets up REST JSON routes, CORS and middleware for DeskBooker.
// The caller must supply an explicit allowlist of frontend origins.
func NewRouter(db DatabasePinger, allowedOrigins map[string]struct{}) *gin.Engine {
    return NewRouterWithRegistration(db, allowedOrigins, nil)
}

// NewRouterWithRegistration optionally exposes the registration endpoint.
// Existing health-only tests can continue using NewRouter.
func NewRouterWithRegistration(db DatabasePinger, allowedOrigins map[string]struct{}, register gin.HandlerFunc) *gin.Engine {
    return NewRouterWithAuth(db, allowedOrigins, register, nil)
}

// NewRouterWithAuth exposes configured authentication handlers and keeps
// the legacy health/registration constructor behavior unchanged.
func NewRouterWithAuth(db DatabasePinger, allowedOrigins map[string]struct{}, register, login gin.HandlerFunc) *gin.Engine {
    return NewRouterWithJWT(db, allowedOrigins, register, login, nil, nil)
}

// NewRouterWithJWT additionally wires a JWT-protected current-user endpoint.
func NewRouterWithJWT(db DatabasePinger, allowedOrigins map[string]struct{},
    register, login, requireJWT, me gin.HandlerFunc) *gin.Engine {
    return NewRouterWithProfile(db, allowedOrigins, register, login, requireJWT, me, nil)
}

// NewRouterWithProfile adds authenticated profile retrieval.
func NewRouterWithProfile(db DatabasePinger, allowedOrigins map[string]struct{},
    register, login, requireJWT, me, profile gin.HandlerFunc) *gin.Engine {
    return NewRouterWithProfileUpdate(db, allowedOrigins, register, login, requireJWT, me, profile, nil)
}

// NewRouterWithProfileUpdate supports authenticated profile edits.
func NewRouterWithProfileUpdate(db DatabasePinger, allowedOrigins map[string]struct{},
    register, login, requireJWT, me, profile, updateProfile gin.HandlerFunc) *gin.Engine {
    return NewRouterWithResources(db, allowedOrigins, register, login, requireJWT, me, profile, updateProfile, nil, nil)
}

// NewRouterWithResources enables role-restricted resource creation.
func NewRouterWithResources(db DatabasePinger, allowedOrigins map[string]struct{},
    register, login, requireJWT, me, profile, updateProfile,
    requireManager, createResource gin.HandlerFunc) *gin.Engine {
    return NewRouterWithResourceFeatures(db, allowedOrigins, register, login, requireJWT, me,
        profile, updateProfile, requireManager, createResource, nil)
}

// NewRouterWithResourceFeatures adds batch resource-feature persistence.
func NewRouterWithResourceFeatures(db DatabasePinger, allowedOrigins map[string]struct{},
    register, login, requireJWT, me, profile, updateProfile,
    requireManager, createResource, addFeatures gin.HandlerFunc) *gin.Engine {
    return NewRouterWithAdminList(db,allowedOrigins,register,login,requireJWT,me,profile,updateProfile,requireManager,createResource,addFeatures,nil)
}

// NewRouterWithAdminList serves the manager resource inventory.
func NewRouterWithAdminList(db DatabasePinger,allowedOrigins map[string]struct{},
 register,login,requireJWT,me,profile,updateProfile,requireManager,createResource,addFeatures,adminList gin.HandlerFunc)*gin.Engine {
    return NewRouterWithResourceEdit(db,allowedOrigins,register,login,requireJWT,me,profile,updateProfile,requireManager,createResource,addFeatures,adminList,nil)
}

// NewRouterWithResourceEdit adds manager-only atomic resource editing.
func NewRouterWithResourceEdit(db DatabasePinger,allowedOrigins map[string]struct{},
 register,login,requireJWT,me,profile,updateProfile,requireManager,createResource,addFeatures,adminList,editResource gin.HandlerFunc)*gin.Engine {
    return NewRouterWithActivation(db,allowedOrigins,register,login,requireJWT,me,profile,updateProfile,requireManager,createResource,addFeatures,adminList,editResource,nil)
}

// NewRouterWithActivation enables role-restricted activation and deactivation.
func NewRouterWithActivation(db DatabasePinger,allowedOrigins map[string]struct{},
 register,login,requireJWT,me,profile,updateProfile,requireManager,createResource,addFeatures,adminList,editResource,activation gin.HandlerFunc)*gin.Engine {
    router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), CORS(allowedOrigins))

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"code": "not_found", "message": "Resource not found"},
		})
	})
	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": gin.H{"code": "method_not_allowed", "message": "Method not allowed"},
		})
	})
	router.HandleMethodNotAllowed = true

	api := router.Group("/api")
    if register != nil {
        api.POST("/auth/register", register)
    }
    if login != nil {
        api.POST("/auth/login", login)
    }
    if requireJWT != nil && me != nil {
        api.GET("/auth/me", requireJWT, me)
    }
    if requireJWT != nil && profile != nil {
        api.GET("/users/me", requireJWT, profile)
    }
    if requireJWT != nil && updateProfile != nil {
        api.PATCH("/users/me", requireJWT, updateProfile)
    }
    if requireJWT != nil && requireManager != nil && createResource != nil {
        api.POST("/resources", requireJWT, requireManager, createResource)
    }
    if requireJWT!=nil && requireManager!=nil && activation!=nil {api.PATCH("/resources/:id/activation",requireJWT,requireManager,activation)}
    if requireJWT!=nil && requireManager!=nil && editResource!=nil {api.PATCH("/resources/:id",requireJWT,requireManager,editResource)}
    if requireJWT!=nil && requireManager!=nil && adminList!=nil {api.GET("/admin/resources",requireJWT,requireManager,adminList)}
    if requireJWT != nil && requireManager != nil && addFeatures != nil {
        api.POST("/resources/:id/features", requireJWT, requireManager, addFeatures)
    }

	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "deskbooker-api",
		})
	})

	api.GET("/health/db", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			// Do not return internal DB connection errors or credentials to clients.
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "database": "connected"})
	})

	return router
}
