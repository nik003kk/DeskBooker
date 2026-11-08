package models

import (
    "encoding/json"
    "strings"
    "testing"
    "time"
)

func TestUserPublicDTOHidesSensitiveData(t *testing.T) {
    user := User{
        ID: "uuid", Email: "someone@example.test", PasswordHash: "SECRET-HASH",
        FirstName: "Test", LastName: "User", Role: RoleEmployee,
        CreatedAt: time.Now().UTC(),
    }
    raw, err := json.Marshal(user.Public())
    if err != nil { t.Fatal(err) }
    text := string(raw)
    if strings.Contains(text, "SECRET-HASH") || strings.Contains(text, "password_hash") {
        t.Fatalf("public DTO leaks password hash: %s", text)
    }
    if !strings.Contains(text, "someone@example.test") || !strings.Contains(text, "employee") {
        t.Fatalf("public user fields missing: %s", text)
    }
}

func TestReservationDTOHidesCheckInCode(t *testing.T) {
    code := "PRIVATE-CHECK-IN"
    r := Reservation{ID: "uuid", CheckInCode: &code}
    raw, err := json.Marshal(r.Public())
    if err != nil { t.Fatal(err) }
    if strings.Contains(string(raw), code) || strings.Contains(string(raw), "check_in_code") {
        t.Fatalf("ordinary reservation DTO leaks check-in code: %s", raw)
    }
}

func TestCreateUserInputNeverSerializesHash(t *testing.T) {
    raw, err := json.Marshal(CreateUserInput{
        Email: "example@test.invalid", PasswordHash: "PRIVATE", FirstName: "A", LastName: "B",
    })
    if err != nil { t.Fatal(err) }
    if strings.Contains(string(raw), "PRIVATE") || strings.Contains(string(raw), "password") {
        t.Fatalf("input DTO leaked password hash: %s", raw)
    }
}
