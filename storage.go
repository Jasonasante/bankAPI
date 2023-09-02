package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Jasonasante/bankAPI.git/account"
	"github.com/Jasonasante/bankAPI.git/misc"
	"github.com/Jasonasante/bankAPI.git/transfer"
	_ "github.com/mattn/go-sqlite3"
)

type Storage interface {
	CreateAccount(*account.Account) error
	DeleteAccount(int) error
	UpdateAccount(id int, update *account.UpdateAccountRequest) error
	GetAccountByID(int) (*account.Account, error)
	GetAllAccounts() ([]*account.Account, error)
	VerifyLogin(account.LoginRequest) (*account.Account, error)
	CreateTransfer(trans *transfer.Transfer) error
	GetAccountBalance(id int) (*transfer.MyBalance, error)
	DepositWithdrawIntoMyAccount(id int, deposit *transfer.TransferRequest) (*transfer.MyBalance, error)
	Transfer(id int, request *transfer.TransferRequest) (*transfer.TransferResponse, error)
	GetAllTransfers() ([]*transfer.Transfer, error)
	GetMyTransfers(id int) ([]*transfer.Transfer, error)
}

type QueryResult interface {
	Scan(dest ...interface{}) error
}

type SQLiteStore struct {
	db *sql.DB
}

//
// Initialisation
//

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
	// if err:=s.DropTable("transfer"); err!=nil {return err}
	if err := s.CreateAccountTable(); err != nil {
		return err
	}
	if err := s.CreateTransferTable(); err != nil {
		return err
	}
	return nil
}

func (s *SQLiteStore) DropTable(name string) error {
	_, tblErr := s.db.Exec("DROP TABLE " + name)

	if tblErr != nil {
		return tblErr
	}

	return nil
}

//
// Account
//

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

//
// Login
//

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

//
// Transfer
//

func (s *SQLiteStore) CreateTransferTable() error {
	_, transferTblErr := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS "transfer" (
			"id" INTEGER PRIMARY KEY AUTOINCREMENT,
			"from" INTEGER NOT NULL,
			"to" INTEGER,
			"amount" INTEGER NOT NULL,
			"action" TEXT NOT NULL,
			"previous_balance" INTERGER NOT NULL,
			"current_balance" INTEGER NOT NULL,
			"completed_at" TIMESTAMP
		)`,
	)
	if transferTblErr != nil {
		return transferTblErr
	}
	return nil
}

func (s *SQLiteStore) CreateTransfer(trans *transfer.Transfer) error {
	stmt, err := s.db.Prepare(`
	INSERT INTO "transfer" (
		"from",
		"to",
		"amount",
		"action",
		"previous_balance",
		"current_balance",
		"completed_at") values ( ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		fmt.Println("error preparing account table:", err)
		return err
	}
	_, errorWithTable := stmt.Exec(
		trans.From,
		trans.To,
		trans.Amount,
		trans.Action,
		trans.PreviousBalance,
		trans.CurrentBalance,
		trans.CompletedAt,
	)
	if errorWithTable != nil {
		fmt.Println("error adding to account table:", errorWithTable)
		return errorWithTable
	}
	return nil
}

func (s *SQLiteStore) GetAllTransfers() ([]*transfer.Transfer, error) {
	transferArray := []*transfer.Transfer{}
	row, err := s.db.Query(`SELECT * FROM "transfer"`)
	if err != nil {
		log.Fatal(err)
	}

	defer row.Close()
	for row.Next() {
		transfer, err := ScanIntoTransfer(row)
		if err != nil {
			fmt.Println("error with scanning rows in transfer table", err)
			return nil, err
		}
		transferArray = append(transferArray, transfer)
	}
	return transferArray, nil
}

func (s *SQLiteStore) GetMyTransfers(id int) ([]*transfer.Transfer, error) {
	transferArray := []*transfer.Transfer{}
	row, err := s.db.Query(`SELECT * FROM "transfer" WHERE ("from" = ? AND "action" = 'withdrawal') OR ("to" = ? AND "action" = 'deposit') OR ("from" = ? AND "to"= ?)`, id, id, id, id)
	if err != nil {
		log.Fatal(err)
	}

	defer row.Close()
	for row.Next() {
		transfer, err := ScanIntoTransfer(row)
		if err != nil {
			fmt.Println("error with scanning rows in transfer table", err)
			return nil, err
		}
		transferArray = append(transferArray, transfer)
	}
	return transferArray, nil
}

func (s *SQLiteStore) GetAccountBalance(id int) (*transfer.MyBalance, error) {
	account, err := ScanIntoAccount(s.db.QueryRow(`SELECT * FROM "account" WHERE id = ?`, id))
	if err != nil {
		fmt.Println("error retrieving account by ID from accounts table")
		return nil, err
	}
	myAccount := &transfer.MyBalance{
		Username:        account.Username,
		MyAccountNumber: account.BankNumber,
		Balance:         account.Balance,
	}
	return myAccount, nil
}

func (s *SQLiteStore) DepositWithdrawIntoMyAccount(id int, deposit *transfer.TransferRequest) (*transfer.MyBalance, error) {
	account, err := ScanIntoAccount(s.db.QueryRow(`SELECT * FROM "account" WHERE id = ?`, id))
	if err != nil {
		fmt.Println("error retrieving account by ID from accounts table")
		return nil, err
	}
	account.Balance += int64(deposit.Amount)
	if account.Balance < 0 {
		return nil, fmt.Errorf("insufficent funds")
	}
	if err := updateBalance(s, account.Balance, id); err != nil {
		fmt.Printf("Could Not Withdraw from %v Account %v", id, err)
		return nil, err
	}
	myAccount := &transfer.MyBalance{
		Username:        account.Username,
		MyAccountNumber: account.BankNumber,
		Balance:         account.Balance,
	}
	return myAccount, nil
}

func (s *SQLiteStore) Transfer(id int, request *transfer.TransferRequest) (*transfer.TransferResponse, error) {
	if request.Amount < 0 {
		return nil, fmt.Errorf("invalid amount - > cannot complete transaction")
	}
	account, err := ScanIntoAccount(s.db.QueryRow(`SELECT * FROM "account" WHERE id = ?`, id))
	if err != nil {
		fmt.Println("error retrieving account by ID from accounts table")
		return nil, err
	}
	toAccount, err := ScanIntoAccount(s.db.QueryRow(`SELECT * FROM "account" WHERE id = ?`, request.ToAccount))
	if err != nil {
		fmt.Println("error retrieving account by ID from accounts table")
		return nil, err
	}
	account.Balance -= int64(request.Amount)
	toAccount.Balance += int64(request.Amount)
	if account.Balance < 0 {
		return nil, fmt.Errorf("insufficent funds - > cannot complete transaction")
	}
	if err := updateBalance(s, account.Balance, id); err != nil {
		fmt.Printf("Could Not Withdraw from %v Account %v", id, err)
		return nil, err
	}

	if err := updateBalance(s, toAccount.Balance, request.ToAccount); err != nil {
		fmt.Printf("Could Not Deposit into %v Account %v", request.ToAccount, err)
		account.Balance += int64(request.Amount)
		if err := updateBalance(s, account.Balance, id); err != nil {
			fmt.Printf("Could Not Return Amount Back to %v Account %v", id, err)
			return nil, err
		}
		return nil, err
	}
	myAccount := &transfer.TransferResponse{
		Account: transfer.MyBalance{
			Username:        account.Username,
			MyAccountNumber: account.BankNumber,
			Balance:         account.Balance,
		},
		Sent: true,
	}
	return myAccount, nil
}

func updateBalance(s *SQLiteStore, balance int64, id int) error {
	_, err := s.db.Exec(`UPDATE account SET "balance" = ? WHERE "id" = ?`, balance, id)
	if err != nil {
		return fmt.Errorf("Transaction Error Please Try Again Later")
	}
	return nil
}

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

func ScanIntoTransfer(row QueryResult) (*transfer.Transfer, error) {
	transfer := new(transfer.Transfer)
	err := row.Scan(
		&transfer.ID,
		&transfer.From,
		&transfer.To,
		&transfer.Amount,
		&transfer.Action,
		&transfer.PreviousBalance,
		&transfer.CurrentBalance,
		&transfer.CompletedAt)
	// PrintAccount(account)
	return transfer, err
}
