package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(*Account) error
	UpdateAccount(*Account) error
	GetAccountByID(int) (*Account, error)
}

type SQLiteStore struct {
	db *sql.DB
}

func NewDB() (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", "./db/bankApi.db")
	if err != nil {
		log.Fatal(err)
	}

	return &SQLiteStore{
		db: db,
	}, nil
}

func (s *SQLiteStore) Init() error {
	return s.CreateAccountTable()
}

func (s *SQLiteStore) CreateAccountTable() error {
	_, acctTblErr := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS "account" (
			"id" SERIAL PRIMARY KEY,
			"first_name" VARCHAR(64),
			"last_name" VARCHAR(64),
			"username" VARCHAR(64) NOT NULL UNIQUE,
			"password" VARCHAR(64) NOT NULL,
			"bank_number" NUMBER,
			"balance" NUMBER,
			"created_at" TIMESTAMP
		)`,
	)
	if acctTblErr != nil {
		return acctTblErr
	}
	return nil
}

func (s *SQLiteStore) CreateAccount(acc *Account) error {
	stmt, err := s.db.Prepare(`
	INSERT INTO "account" (
	"id", 
	"first_name",
	"last_name",
	"username",
	"password",
	"bank_number" ,
	"balance" ,
	"created_at") values (?, ?, ?, ?, ?, ?, ?,?)
	`)
	if err != nil {
		fmt.Println("error preparing account table:", err)
		return err
	}
	_, errorWithTable := stmt.Exec(
		acc.ID,
		acc.FirstName,
		acc.LastName,
		acc.Username,
		acc.Password,
		acc.BankNumber,
		acc.Balance,
		acc.CreatedAt,
	)
	if errorWithTable != nil {
		fmt.Println("error adding to accout table:", errorWithTable)
		return errorWithTable
	}
	s.displayInfo("account")
	return nil
}

func (s *SQLiteStore) DeleteAccount(*Account) error {
	return nil
}

func (s *SQLiteStore) UpdateAccount(*Account) error {
	return nil
}

func (s *SQLiteStore) GetAccountByID(int) (*Account, error) {
	return nil, nil
}

func (s *SQLiteStore) displayInfo(table string) {
	switch table {
	case "account":
		row, err := s.db.Query(`SELECT * FROM "account"`)
		if err != nil {
			log.Fatal(err)
		}

		defer row.Close()
		for row.Next() { // Iterate and fetch the records from result cursor
			// account := new(Account)
			var (
				id, bankNumber, balance int
				firstName, lastName     string
				username, password      string
				createdAt               time.Time
			)
			err = row.Scan(&id, &firstName, &lastName, &username, &password, &bankNumber, &balance, &createdAt)
			if err != nil {
				fmt.Println("error with scanning rows in", err)
				// return err
			}
			fmt.Println("id:=", id, "first name:=", firstName, "last name:=", lastName, "username:=", username, "password:=", password, "bank number:=", bankNumber, "balance:=", balance, "created at:=", createdAt)
		}
	}
}
