package server

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

const defaultFrontendOrigin = "http://localhost:5173"

// AllowedOrigins parses a comma-separated allowlist of complete web origins.
// Wildcards, paths and credentials are rejected to avoid permissive CORS settings.
func AllowedOrigins(value string) (map[string]struct{}, error) {
	if strings.TrimSpace(value) == "" {
		value = defaultFrontendOrigin
	}

	allowed := make(map[string]struct{})
	for _, raw := range strings.Split(value, ",") {
		origin := strings.TrimSpace(raw)
		parsed, err := url.Parse(origin)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") ||
			parsed.Host == "" || parsed.User != nil || parsed.Path != "" ||
			parsed.RawQuery != "" || parsed.Fragment != "" ||
			parsed.Opaque != "" || parsed.Host != parsed.Hostname() && parsed.Port() == "" {
			return nil, errors.New("CORS_ALLOWED_ORIGINS must contain valid HTTP(S) origins without paths")
		}
		if parsed.String() != origin || origin == "*" {
			return nil, errors.New("CORS_ALLOWED_ORIGINS must contain valid HTTP(S) origins without paths")
		}
		allowed[origin] = struct{}{}
	}
	return allowed, nil
}

// CORS allows browser requests only from configured frontend origins.
// No Access-Control-Allow-Credentials is sent; authentication is a later issue.
func CORS(allowed map[string]struct{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}

		c.Header("Vary", "Origin")
		_, ok := allowed[origin]
		if !ok {
			if c.Request.Method == http.MethodOptions &&
				c.GetHeader("Access-Control-Request-Method") != "" {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": gin.H{"code": "cors_origin_denied", "message": "Origin not allowed"},
				})
				return
			}
			c.Next()
			return
		}

		c.Header("Access-Control-Allow-Origin", origin)
		if c.Request.Method == http.MethodOptions &&
			c.GetHeader("Access-Control-Request-Method") != "" {
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			c.Header("Access-Control-Max-Age", "600")
			c.Status(http.StatusNoContent)
			c.Abort()
			return
		}
		c.Next()
	}
}
