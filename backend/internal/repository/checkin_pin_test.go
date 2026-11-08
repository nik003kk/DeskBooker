package repository

import (
 "regexp"
 "testing"
)
func TestCheckInPINIsFourDigitsIncludingLeadingZeros(t *testing.T){
 pattern:=regexp.MustCompile(`^[0-9]{4}$`)
 seen:=map[string]bool{}
 for i:=0;i<400;i++{
  pin,err:=newCheckInPIN()
  if err!=nil||!pattern.MatchString(pin){t.Fatalf("bad PIN: %q %v",pin,err)}
  seen[pin]=true
 }
 if len(seen)<2{t.Fatal("unexpected identical 400 random PINs")}
}
