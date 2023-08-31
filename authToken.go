package main

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/Jasonasante/bankAPI.git/account"
	"github.com/Jasonasante/bankAPI.git/misc"
	"github.com/golang-jwt/jwt/v4"
)

func withJWTAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("calling JWT middleware...")
		tokenStr := r.Header.Get("x-jwt-token")
		token, err := validateJWT(tokenStr)
		if err != nil {
			WriteJSON(w, http.StatusUnauthorized, apiError{Error: "Invalid JWT"})
			return
		}
		if !token.Valid {
			WriteJSON(w, http.StatusUnauthorized, apiError{Error: "Invalid JWT"})
			return
		}
		userId, err := misc.GetID(r)
		if err != nil {
			WriteJSON(w, http.StatusBadRequest, apiError{Error: "Invalid Request"})
			return
		}
		account, err := s.GetAccountByID(userId)
		if err != nil {
			WriteJSON(w, http.StatusUnauthorized, apiError{Error: "Permission Denied"})
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		fmt.Println(claims)
		fmt.Printf("%T\n", int64(math.Round(claims["account-number"].(float64))))
		if account.BankNumber != int64(math.Round(claims["account-number"].(float64))) {
			WriteJSON(w, http.StatusUnauthorized, apiError{Error: "Permission Denied"})
			return
		}
		handlerFunc(w, r)
	}
}

func validateJWT(tokenStr string) (*jwt.Token, error) {
	secret := os.Getenv("jwtSecret")
	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(secret), nil
	})
}

func createJWT(account *account.Account) (string, error) {
	secret := os.Getenv("jwtSecret")
	// Create the Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"account-number": account.BankNumber,
		"expiresAt":      time.Now().AddDate(1, 0, 0).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	return token.SignedString([]byte(secret))
}
