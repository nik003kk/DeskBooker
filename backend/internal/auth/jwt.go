package auth

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "errors"
    "strings"
    "time"
)

const (
    tokenIssuer = "deskbooker-api"
    tokenAudience = "deskbooker-api"
    tokenTTL = 15 * time.Minute
    maxJWTLength = 4096
)

var (
    ErrInvalidToken = errors.New("invalid or expired token")
    ErrJWTSecret = errors.New("JWT_SECRET must contain at least 32 bytes")
)

var jwtEncoding = base64.RawURLEncoding

// JWTService signs short-lived bearer tokens using an application-specific
// HS256 secret. Secrets must be generated randomly and never committed.
type JWTService struct {
    key []byte
    now func() time.Time
}

type JWTClaims struct {
    Subject string `json:"sub"`
    Issuer string `json:"iss"`
    Audience string `json:"aud"`
    IssuedAt int64 `json:"iat"`
    NotBefore int64 `json:"nbf"`
    ExpiresAt int64 `json:"exp"`
}

type jwtHeader struct {
    Algorithm string `json:"alg"`
    Type string `json:"typ"`
}

func NewJWTService(secret string) (*JWTService, error) {
    if len(secret) < 32 {
        return nil, ErrJWTSecret
    }
    return &JWTService{key: []byte(secret), now: time.Now}, nil
}

// Issue creates an HS256 JWT bound to a user ID and valid for 15 minutes.
func (s *JWTService) Issue(userID string) (string, time.Time, error) {
    if s == nil || userID == "" {
        return "", time.Time{}, ErrInvalidToken
    }
    now := s.now().UTC().Truncate(time.Second)
    exp := now.Add(tokenTTL)
    header, _ := json.Marshal(jwtHeader{Algorithm: "HS256", Type: "JWT"})
    payload, err := json.Marshal(JWTClaims{
        Subject: userID, Issuer: tokenIssuer, Audience: tokenAudience,
        IssuedAt: now.Unix(), NotBefore: now.Unix(), ExpiresAt: exp.Unix(),
    })
    if err != nil {
        return "", time.Time{}, err
    }
    message := jwtEncoding.EncodeToString(header)+"."+jwtEncoding.EncodeToString(payload)
    mac := hmac.New(sha256.New, s.key)
    _, _ = mac.Write([]byte(message))
    return message+"."+jwtEncoding.EncodeToString(mac.Sum(nil)), exp, nil
}

// Verify validates the exact algorithm, HMAC, issuer, audience, subject and
// temporal claims. It never trusts unverified token contents.
func (s *JWTService) Verify(token string) (JWTClaims, error) {
    invalid := JWTClaims{}
    if s == nil || token == "" || len(token) > maxJWTLength {
        return invalid, ErrInvalidToken
    }
    parts := strings.Split(token, ".")
    if len(parts) != 3 {
        return invalid, ErrInvalidToken
    }
    headerBytes, err := jwtEncoding.DecodeString(parts[0])
    if err != nil || len(headerBytes) > 256 {
        return invalid, ErrInvalidToken
    }
    var header jwtHeader
    if json.Unmarshal(headerBytes, &header) != nil ||
        header.Algorithm != "HS256" || header.Type != "JWT" {
        return invalid, ErrInvalidToken
    }
    providedMAC, err := jwtEncoding.DecodeString(parts[2])
    if err != nil || len(providedMAC) != sha256.Size {
        return invalid, ErrInvalidToken
    }
    mac := hmac.New(sha256.New, s.key)
    _, _ = mac.Write([]byte(parts[0]+"."+parts[1]))
    if !hmac.Equal(providedMAC, mac.Sum(nil)) {
        return invalid, ErrInvalidToken
    }
    payload, err := jwtEncoding.DecodeString(parts[1])
    if err != nil || len(payload) > 2048 {
        return invalid, ErrInvalidToken
    }
    var claims JWTClaims
    if json.Unmarshal(payload, &claims) != nil {
        return invalid, ErrInvalidToken
    }
    now := s.now().UTC().Unix()
    if claims.Issuer != tokenIssuer || claims.Audience != tokenAudience ||
        claims.Subject == "" || len(claims.Subject) > 128 ||
        claims.IssuedAt <= 0 || claims.NotBefore <= 0 || claims.ExpiresAt <= 0 ||
        claims.IssuedAt > now || claims.NotBefore > now ||
        claims.ExpiresAt <= now || claims.ExpiresAt <= claims.IssuedAt ||
        claims.ExpiresAt-claims.IssuedAt > int64(tokenTTL.Seconds()) {
        return invalid, ErrInvalidToken
    }
    return claims, nil
}
