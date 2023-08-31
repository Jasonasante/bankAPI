package account

import (
	"time"

	"github.com/Jasonasante/bankAPI.git/misc"
)

type LoginResponse struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateAccountRequest struct {
	FirstName string    `json:"first-name"`
	LastName  string    `json:"last-name"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created-at"`
}

type Account struct {
	ID         int       `json:"id"`
	FirstName  string    `json:"first-name"`
	LastName   string    `json:"last-name"`
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	BankNumber int64     `json:"bank-number"`
	Balance    int64     `json:"balance"`
	CreatedAt  time.Time `json:"created-at"`
}

func CreateAccount(firstName, lastName, username, password string) *Account {
	return &Account{
		FirstName:  firstName,
		LastName:   lastName,
		Username:   username,
		Password:   password,
		BankNumber: misc.RangeIn(10000000, 99999999),
		Balance:    0,
		CreatedAt:  time.Now().UTC(),
	}
}
