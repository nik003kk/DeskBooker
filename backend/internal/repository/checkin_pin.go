package repository

import (
 "crypto/rand"
 "fmt"
 "math/big"
)

// newCheckInPIN uses cryptographically secure randomness and zero padding.
// The database's unique index and a bounded retry loop handle collisions.
func newCheckInPIN()(string,error){
 n,err:=rand.Int(rand.Reader,big.NewInt(10000))
 if err!=nil{return "",fmt.Errorf("generate PIN: %w",err)}
 return fmt.Sprintf("%04d",n.Int64()),nil
}
