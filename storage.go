package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Jasonasante/bankAPI.git/account"
	"github.com/Jasonasante/bankAPI.git/misc"
	_ "github.com/mattn/go-sqlite3"
)

type Storage interface {
	CreateAccount(*account.Account) error
	DeleteAccount(int) error
	UpdateAccount(id int, update *account.UpdateAccountRequest) error
	GetAccountByID(int) (*account.Account, error)
	GetAllAccounts() ([]*account.Account, error)
	VerifyLogin(account.LoginRequest) (*account.Account, error)
}

type QueryResult interface {
	Scan(dest ...interface{}) error
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
			"id" INTEGER PRIMARY KEY AUTOINCREMENT,
			"first_name" VARCHAR(64),
			"last_name" VARCHAR(64),
			"username" VARCHAR(64) NOT NULL UNIQUE,
			"password" VARCHAR(64) NOT NULL,
			"bank_number" NUMBER NOT NULL UNIQUE,
			"balance" NUMBER,
			"created_at" TIMESTAMP
		)`,
	)
	if acctTblErr != nil {
		return acctTblErr
	}
	return nil
}

func (s *SQLiteStore) CreateAccount(acc *account.Account) error {
	stmt, err := s.db.Prepare(`
	INSERT INTO "account" (
	"first_name",
	"last_name",
	"username",
	"password",
	"bank_number" ,
	"balance" ,
	"created_at") values ( ?, ?, ?, ?, ?, ?,?)
	`)
	if err != nil {
		fmt.Println("error preparing account table:", err)
		return err
	}
	_, errorWithTable := stmt.Exec(
		acc.FirstName,
		acc.LastName,
		acc.Username,
		acc.Password,
		acc.BankNumber,
		acc.Balance,
		acc.CreatedAt,
	)
	if errorWithTable != nil {
		fmt.Println("error adding to account table:", errorWithTable)
		return errorWithTable
	}
	return nil
}

func (s *SQLiteStore) DeleteAccount(id int) error {
	_, err := s.db.Exec(`DELETE FROM "account" WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *SQLiteStore) UpdateAccount(id int, update *account.UpdateAccountRequest) error {
	_, err := s.db.Exec(`UPDATE account SET "first_name" = ?, "last_name" = ?, "username" = ?, "password" = ? WHERE "id" = ?`, update.FirstName, update.LastName, update.Username, update.Password, id)
	if err != nil {
		fmt.Printf("Could Not Update Account %v", err)
		return err
	}
	return nil
}

func (s *SQLiteStore) GetAccountByID(id int) (*account.Account, error) {
	account, err := ScanIntoAccount(s.db.QueryRow(`SELECT * FROM "account" WHERE id = ?`, id))
	if err != nil {
		fmt.Println("error retrieving account by ID from accounts table")
		return nil, err
	}

	return account, nil
}

func (s *SQLiteStore) GetAllAccounts() ([]*account.Account, error) {
	accountArray := []*account.Account{}
	row, err := s.db.Query(`SELECT * FROM "account"`)
	if err != nil {
		log.Fatal(err)
	}

	defer row.Close()
	for row.Next() {
		account, err := ScanIntoAccount(row)
		if err != nil {
			fmt.Println("error with scanning rows in account table", err)
			return nil, err
		}
		accountArray = append(accountArray, account)
	}
	return accountArray, nil
}

func (s *SQLiteStore) VerifyLogin(login account.LoginRequest) (*account.Account, error) {
	account, err := ScanIntoAccount(s.db.QueryRow(`SELECT * FROM "account" WHERE username= ?`, login.Username))
	if err != nil {
		return nil, fmt.Errorf("Account Does Not Exist")
	}
	if !misc.CheckPasswordHash(login.Password, account.Password) {
		return nil, fmt.Errorf("Access Denied")
	}
	return account, nil
}

// func (s *SQLiteStore) CheckIfUsernameAlreadyExists(username string) bool {
// 	_, err := ScanIntoAccount(s.db.QueryRow(`SELECT * FROM "account" WHERE username = ?`, username))
// 	return err == nil
// }

func ScanIntoAccount(row QueryResult) (*account.Account, error) {
	account := new(account.Account)
	err := row.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Username,
		&account.Password,
		&account.BankNumber,
		&account.Balance,
		&account.CreatedAt)
	// PrintAccount(account)
	return account, err
}

func PrintAccount(account *account.Account) {
	fmt.Println(
		"id:=", account.ID,
		"first name:=", account.FirstName,
		"last name:=", account.LastName,
		"username:=", account.Username,
		"password:=", account.Password,
		"bank number:=", account.BankNumber,
		"balance:=", account.Balance,
		"created at:=", account.CreatedAt)
}
