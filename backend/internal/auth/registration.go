// Package auth contains HTTP handlers for registration and future authentication.
package auth

import (
    "context"
    "encoding/json"
    "errors"
    "io"
    "net/http"
    "net/mail"
    "mime"
    "strings"
    "unicode"
    "unicode/utf8"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5/pgconn"
    "golang.org/x/crypto/bcrypt"

    "github.com/nik003kk/DeskBooker/backend/internal/models"
)

const maxRegistrationBody = 8 << 10 // 8 KiB

// UserCreator is satisfied by repository.Store; tests can supply a fake.
type UserCreator interface {
    CreateUser(context.Context, models.CreateUserInput) (models.User, error)
}

type registrationRequest struct {
    FirstName string `json:"first_name"`
    LastName string `json:"last_name"`
    Email string `json:"email"`
    Password string `json:"password"`
}

func responseError(c *gin.Context, status int, code, message string) {
    c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}

func validName(value string) bool {
    if value == "" || strings.TrimSpace(value) != value ||
        utf8.RuneCountInString(value) > 100 {
        return false
    }
    for _, r := range value {
        if unicode.IsControl(r) { return false }
    }
    return true
}

func normalizedCompanyEmail(value, domain string) (string, bool) {
    if value == "" || value != strings.TrimSpace(value) ||
        len(value) > 254 || strings.ContainsAny(value, "<>\t\r\n ") {
        return "", false
    }
    addr, err := mail.ParseAddress(value)
    if err != nil || addr.Address != value || strings.Count(value, "@") != 1 {
        return "", false
    }
    parts := strings.SplitN(value, "@", 2)
    if parts[0] == "" || parts[1] == "" || !strings.EqualFold(parts[1], domain) {
        return "", false
    }
    return strings.ToLower(value), true
}

// RegisterHandler requires a configured corporate mail domain.
// When unconfigured, registration fails closed with HTTP 503.
func RegisterHandler(store UserCreator, companyDomain string) gin.HandlerFunc {
    domain := strings.ToLower(strings.TrimSpace(companyDomain))
    return func(c *gin.Context) {
        if domain == "" {
            responseError(c, http.StatusServiceUnavailable, "registration_unavailable", "Registration is not configured")
            return
        }
        mediaType, _, mediaErr := mime.ParseMediaType(c.GetHeader("Content-Type"))
        if mediaErr != nil || mediaType != "application/json" {
            responseError(c, http.StatusUnsupportedMediaType, "unsupported_media_type", "Expected application/json")
            return
        }
        decoder := json.NewDecoder(io.LimitReader(c.Request.Body, maxRegistrationBody+1))
        decoder.DisallowUnknownFields()
        var request registrationRequest
        if err := decoder.Decode(&request); err != nil {
            responseError(c, http.StatusBadRequest, "invalid_json", "Invalid registration JSON")
            return
        }
        var extra any
        if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
            responseError(c, http.StatusBadRequest, "invalid_json", "Only one JSON object is allowed")
            return
        }
        if !validName(request.FirstName) || !validName(request.LastName) {
            responseError(c, http.StatusBadRequest, "invalid_name", "First and last names are required (max 100 characters)")
            return
        }
        email, ok := normalizedCompanyEmail(request.Email, domain)
        if !ok {
            responseError(c, http.StatusBadRequest, "invalid_email", "A valid company email address is required")
            return
        }
        if len(request.Password) < 12 || len(request.Password) > 72 ||
            strings.TrimSpace(request.Password) == "" {
            responseError(c, http.StatusBadRequest, "invalid_password", "Password must be 12 to 72 bytes long")
            return
        }

        hashed, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
        if err != nil {
            responseError(c, http.StatusInternalServerError, "registration_failed", "Could not register user")
            return
        }
        user, err := store.CreateUser(c.Request.Context(), models.CreateUserInput{
            FirstName: request.FirstName,
            LastName: request.LastName,
            Email: email,
            PasswordHash: string(hashed),
        })
        if err != nil {
            var pgErr *pgconn.PgError
            if errors.As(err, &pgErr) &&
                pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_lower_unique_idx" {
                responseError(c, http.StatusConflict, "email_exists", "Email is already registered")
                return
            }
            responseError(c, http.StatusInternalServerError, "registration_failed", "Could not register user")
            return
        }
        c.JSON(http.StatusCreated, gin.H{"user": user.Public()})
    }
}
