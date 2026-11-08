package auth

import (
    "errors"
    "strings"
    "testing"

    "golang.org/x/crypto/bcrypt"
)

func TestPasswordHashAndVerification(t *testing.T) {
    const plain = "VeryStrongPass123!"
    first, err := HashPassword(plain)
    if err != nil { t.Fatal(err) }
    second, err := HashPassword(plain)
    if err != nil { t.Fatal(err) }
    if first == second || first == plain || !strings.HasPrefix(first, "$2") {
        t.Fatal("bcrypt hashes must be salted and must not match plaintext")
    }
    if err := VerifyPassword(first, plain); err != nil { t.Fatalf("correct password rejected: %v", err) }
    if err := VerifyPassword(first, "wrong-password"); !errors.Is(err, ErrInvalidCredentials) {
        t.Fatalf("wrong password returned %v", err)
    }
    if err := VerifyPassword("not-a-bcrypt-hash", plain); !errors.Is(err, ErrInvalidCredentials) {
        t.Fatalf("malformed stored hash returned %v", err)
    }
    if err := VerifyPassword("", plain); !errors.Is(err, ErrInvalidCredentials) {
        t.Fatalf("missing stored hash returned %v", err)
    }
}

func TestPasswordLimits(t *testing.T) {
    for _, tc := range []struct { name, password string; valid bool }{
        {"empty", "", false},
        {"too short", strings.Repeat("a", 11), false},
        {"whitespace only", strings.Repeat(" ", 12), false},
        {"minimum", strings.Repeat("a", 12), true},
        {"maximum", strings.Repeat("a", 72), true},
        {"too long", strings.Repeat("a", 73), false},
        {"multibyte over limit", strings.Repeat("ž", 37), false},
    } {
        t.Run(tc.name, func(t *testing.T) {
            hash, err := HashPassword(tc.password)
            if (err == nil) != tc.valid { t.Fatalf("valid=%v, hash=%q, err=%v", tc.valid, hash, err) }
            if !tc.valid && !errors.Is(err, ErrInvalidPassword) {
                t.Fatalf("expected ErrInvalidPassword, got %v", err)
            }
            if tc.valid && VerifyPassword(hash, tc.password) != nil {
                t.Fatal("valid password did not verify")
            }
        })
    }
}

func TestPasswordDoesNotAcceptBCryptTruncation(t *testing.T) {
    password := strings.Repeat("x", 72)
    hash, err := HashPassword(password)
    if err != nil { t.Fatal(err) }
    // bcrypt's comparison could silently ignore a suffix beyond 72 bytes.
    if err := VerifyPassword(hash, password+"extra"); !errors.Is(err, ErrInvalidCredentials) {
        t.Fatal("oversized password must not authenticate through bcrypt truncation")
    }
    if cost, err := bcrypt.Cost([]byte(hash)); err != nil || cost < bcrypt.DefaultCost {
        t.Fatalf("incorrect bcrypt cost %v: %v", cost, err)
    }
}
