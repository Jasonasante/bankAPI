package misc

import (
	"fmt"
	"math/rand"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

func Generate() string {
	u2 := uuid.NewV4()
	return fmt.Sprintf("%x", u2)
}

func RangeIn(low, hi int64) int64 {
	return low + rand.Int63n(hi-low)
}

// this receives a password and encrypts it, protect a user's password in the database.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// this checks whether the inputted string when trying to login matches the encrypted password
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
