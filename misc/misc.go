package misc

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

func Generate() string {
	u2 := uuid.NewV4()
	return fmt.Sprintf("%x", u2)
}

func RangeIn(low, hi int64) int64 {
	max := big.NewInt(100000000)
	randomNumber, _ := rand.Int(rand.Reader, max)
	return 10000000 + randomNumber.Int64()
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

func GetID(r *http.Request) (int, error) {
	vars := mux.Vars(r) // Extract route variables and returns it as a map[string]string
	return strconv.Atoi(vars["id"])
}

func DefaultValue(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
