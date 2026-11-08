package auth

import (
    "context"
    "encoding/json"
    "errors"
    "io"
    "mime"
    "net/http"
    "net/mail"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5"

    "github.com/nik003kk/DeskBooker/backend/internal/models"
)

// UserFinder is implemented by repository.Store.
type UserFinder interface {
    GetUserByEmail(context.Context, string) (models.User, error)
}

type loginRequest struct {
    Email string `json:"email"`
    Password string `json:"password"`
}

// LoginHandler checks credentials but deliberately does not issue a session,
// cookie or token. Session-based authentication is a separate follow-up.
func LoginHandler(store UserFinder) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Cache-Control", "no-store")
        mediaType, _, mediaErr := mime.ParseMediaType(c.GetHeader("Content-Type"))
        if mediaErr != nil || mediaType != "application/json" {
            responseError(c, http.StatusUnsupportedMediaType, "unsupported_media_type", "Expected application/json")
            return
        }

        decoder := json.NewDecoder(io.LimitReader(c.Request.Body, maxRegistrationBody+1))
        decoder.DisallowUnknownFields()
        var request loginRequest
        if err := decoder.Decode(&request); err != nil {
            responseError(c, http.StatusBadRequest, "invalid_json", "Invalid login JSON")
            return
        }
        var extra any
        if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
            responseError(c, http.StatusBadRequest, "invalid_json", "Only one JSON object is allowed")
            return
        }

        // Accept existing registered users even if the configured registration
        // domain changes. A login is not a new registration.
        if !validLoginEmail(request.Email) || request.Password == "" ||
            len(request.Password) > maxPasswordBytes {
            responseError(c, http.StatusBadRequest, "invalid_credentials", "Invalid email or password")
            return
        }
        user, err := store.GetUserByEmail(c.Request.Context(), strings.ToLower(request.Email))
        if err != nil {
            if errors.Is(err, pgx.ErrNoRows) {
                responseError(c, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password")
                return
            }
            responseError(c, http.StatusInternalServerError, "login_failed", "Could not sign in")
            return
        }
        if VerifyPassword(user.PasswordHash, request.Password) != nil {
            responseError(c, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password")
            return
        }

        c.JSON(http.StatusOK, gin.H{"user": user.Public()})
    }
}

func validLoginEmail(value string) bool {
    if value == "" || value != strings.TrimSpace(value) ||
        len(value) > 254 || strings.ContainsAny(value, "<>\t\r\n ") ||
        strings.Count(value, "@") != 1 {
        return false
    }
    parsed, err := mail.ParseAddress(value)
    if err != nil || parsed.Address != value {
        return false
    }
    parts := strings.SplitN(value, "@", 2)
    return parts[0] != "" && parts[1] != ""
}
