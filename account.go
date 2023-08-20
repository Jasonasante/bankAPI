package main

import "github.com/Jasonasante/bankAPI.git/misc"

type Account struct {
	ID         string `json:"id"`
	FirstName  string `json:"first-name"`
	LastName   string `json:"last-name"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	DOB        string `json:"date-of-birth"`
	BankNumber int64  `json:"bank-number"`
	Balance    int64  `json:"balance"`
}

func CreateAccount(firstName, lastName, username, password string) *Account {
	return &Account{
		ID:         misc.Generate(),
		FirstName:  firstName,
		LastName:   lastName,
		Username:   username,
		Password:   password,
		BankNumber: misc.RangeIn(10000000, 99999999),
		Balance:    0,
	}
}
