package auth

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/json"
    "errors"
    "strings"
    "testing"
    "time"
)

const testJWTSecret = "very-long-local-test-secret-32-bytes-minimum"

func TestJWTIssueVerifyAndExpiry(t *testing.T) {
    service, err := NewJWTService(testJWTSecret)
    if err != nil { t.Fatal(err) }
    fixed := time.Date(2030, 1, 2, 10, 0, 0, 0, time.UTC)
    service.now = func() time.Time { return fixed }
    token, exp, err := service.Issue("user-123")
    if err != nil { t.Fatal(err) }
    if exp.Sub(fixed) != tokenTTL { t.Fatalf("unexpected TTL: %v", exp.Sub(fixed)) }
    claims, err := service.Verify(token)
    if err != nil || claims.Subject != "user-123" ||
        claims.Issuer != tokenIssuer || claims.Audience != tokenAudience {
        t.Fatalf("invalid claims: %#v %v", claims, err)
    }
    service.now = func() time.Time { return fixed.Add(tokenTTL) }
    if _, err := service.Verify(token); !errors.Is(err, ErrInvalidToken) {
        t.Fatalf("expired token was accepted: %v", err)
    }
    service.now = func() time.Time { return fixed.Add(-time.Second) }
    if _, err := service.Verify(token); !errors.Is(err, ErrInvalidToken) {
        t.Fatalf("future token was accepted: %v", err)
    }
}

func signTestToken(t *testing.T, secret string, header any, claims any) string {
    t.Helper()
    h, err := json.Marshal(header)
    if err != nil { t.Fatal(err) }
    p, err := json.Marshal(claims)
    if err != nil { t.Fatal(err) }
    message := jwtEncoding.EncodeToString(h)+"."+jwtEncoding.EncodeToString(p)
    mac := hmac.New(sha256.New, []byte(secret))
    _, _ = mac.Write([]byte(message))
    return message+"."+jwtEncoding.EncodeToString(mac.Sum(nil))
}

func TestJWTRejectsTamperingWrongAlgorithmAndInvalidClaims(t *testing.T) {
    service, err := NewJWTService(testJWTSecret)
    if err != nil { t.Fatal(err) }
    fixed := time.Date(2030, 1, 2, 10, 0, 0, 0, time.UTC)
    service.now = func() time.Time { return fixed }
    good, _, err := service.Issue("user-123")
    if err != nil { t.Fatal(err) }
    parts := strings.Split(good, ".")
    claims := JWTClaims{
        Subject: "user-123", Issuer: tokenIssuer, Audience: tokenAudience,
        IssuedAt: fixed.Unix(), NotBefore: fixed.Unix(),
        ExpiresAt: fixed.Add(tokenTTL).Unix(),
    }
    cases := map[string]string{
        "altered signature": parts[0]+"."+parts[1]+"."+strings.Repeat("A", 43),
        "missing signature": parts[0]+"."+parts[1]+".",
        "unsigned alg none": signTestToken(t, testJWTSecret, jwtHeader{Algorithm:"none", Type:"JWT"}, claims),
        "wrong alg": signTestToken(t, testJWTSecret, jwtHeader{Algorithm:"HS512", Type:"JWT"}, claims),
        "wrong audience": signTestToken(t, testJWTSecret, jwtHeader{Algorithm:"HS256", Type:"JWT"}, func() JWTClaims { c:=claims; c.Audience="other"; return c }()),
        "wrong issuer": signTestToken(t, testJWTSecret, jwtHeader{Algorithm:"HS256", Type:"JWT"}, func() JWTClaims { c:=claims; c.Issuer="other"; return c }()),
        "no expiry": signTestToken(t, testJWTSecret, jwtHeader{Algorithm:"HS256", Type:"JWT"}, func() JWTClaims { c:=claims; c.ExpiresAt=0; return c }()),
        "long expiry": signTestToken(t, testJWTSecret, jwtHeader{Algorithm:"HS256", Type:"JWT"}, func() JWTClaims { c:=claims; c.ExpiresAt+=3600; return c }()),
        "missing subject": signTestToken(t, testJWTSecret, jwtHeader{Algorithm:"HS256", Type:"JWT"}, func() JWTClaims { c:=claims; c.Subject=""; return c }()),
        "garbage": "not-a-jwt",
    }
    for name, value := range cases {
        t.Run(name, func(t *testing.T) {
            if _, err := service.Verify(value); !errors.Is(err, ErrInvalidToken) {
                t.Fatalf("accepted invalid token: %v", err)
            }
        })
    }
    other, _ := NewJWTService("another-secret-key-with-more-than-32-bytes")
    other.now = func() time.Time { return fixed }
    if _, err := other.Verify(good); !errors.Is(err, ErrInvalidToken) {
        t.Fatal("JWT signed by another secret was accepted")
    }
}

func TestJWTRejectsWeakSecretAndEmptySubject(t *testing.T) {
    if _, err := NewJWTService("too-short"); !errors.Is(err, ErrJWTSecret) {
        t.Fatalf("weak key accepted: %v", err)
    }
    service, _ := NewJWTService(testJWTSecret)
    if _, _, err := service.Issue(""); !errors.Is(err, ErrInvalidToken) {
        t.Fatalf("empty subject accepted: %v", err)
    }
}
