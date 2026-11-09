package auth

import (
    "errors"
    "strings"

    "golang.org/x/crypto/bcrypt"
)

// ErrInvalidPassword means the supplied password does not meet the registration
// policy. Count bytes, not runes: bcrypt uses a maximum of 72 bytes.
var ErrInvalidPassword = errors.New("password must be 12 to 72 bytes long")

// ErrInvalidCredentials is intentionally identical for a wrong password,
// malformed stored hash and unknown account at the authentication boundary.
var ErrInvalidCredentials = errors.New("invalid email or password")

const (
    minPasswordBytes = 12
    maxPasswordBytes = 72
)

// HashPassword validates a raw password and returns a salted bcrypt hash.
// Never persist the raw password or include the hash in public DTOs.
func HashPassword(password string) (string, error) {
    if len(password) < minPasswordBytes || len(password) > maxPasswordBytes || strings.TrimSpace(password) == "" {
        return "", ErrInvalidPassword
    }
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    return string(hash), nil
}

// VerifyPassword compares the supplied password with the stored bcrypt hash.
// Failure is deliberately generic to avoid exposing credential details.
func VerifyPassword(storedHash, password string) error {
    if storedHash == "" || len(password) > maxPasswordBytes {
        return ErrInvalidCredentials
    }
    if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err != nil {
        return ErrInvalidCredentials
    }
    return nil
}
