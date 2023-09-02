package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Jasonasante/bankAPI.git/account"
	"github.com/Jasonasante/bankAPI.git/misc"
	"github.com/Jasonasante/bankAPI.git/transfer"

	"github.com/gorilla/mux"
)

func WriteJSON(w http.ResponseWriter, status int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// apiFunc is a custom type representing a function signature.
// It takes a function that has a http.ResponseWriter and an http.Request as parameters and returns an error.
type apiFunc func(w http.ResponseWriter, r *http.Request) error

type apiError struct {
	Error string
}

func makeHttpHandler(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
		}
	}
}

type APIServer struct {
	listenAddr string
	store      Storage
}

// Storage is an interface type populated with methods. So any type/struct that contains these methods
// will be acceptable as an input parameter
func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr,
		store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()
	router.HandleFunc("/login", makeHttpHandler(s.handleLogin))
	router.HandleFunc("/account", makeHttpHandler(s.handleAccount))
	router.HandleFunc("/account/{id}", withJWTAuth(makeHttpHandler(s.handleGetAccountbyID), s.store))
	router.HandleFunc("/transfer", makeHttpHandler(s.handleTransfers))
	router.HandleFunc("/transfer/{id}", withJWTAuth(makeHttpHandler(s.handleTransferAccount), s.store))
	log.Println("server opened http://localhost" + s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

// CRUD

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetAccounts(w, r)
	case "POST":
		return s.handleCreateAccount(w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *APIServer) handleGetAccounts(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAllAccounts()
	if err != nil {
		fmt.Println("Failed to get all accounts")
		return err
	}
	return WriteJSON(w, http.StatusOK, accounts)
}

func (s *APIServer) handleGetAccountbyID(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		id, err := misc.GetID(r)
		if err != nil {
			fmt.Println("Invalid ID given!!!")
			return err
		}
		account, err := s.store.GetAccountByID(id)
		if err != nil {
			return fmt.Errorf("failed to retrieve account by id : %v", err)
		}
		return WriteJSON(w, http.StatusOK, account)

	case "DELETE":
		return s.handleDeleteAccount(w, r)

	case "PATCH":
		return s.handleUpdateAccount(w, r)
	}
	return fmt.Errorf("method not allowed : %v", r.Method)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	acctRequest := account.CreateAccountRequest{}
	if err := json.NewDecoder(r.Body).Decode(&acctRequest); err != nil {
		return err
	}
	defer r.Body.Close()
	password, err := misc.HashPassword(acctRequest.Password)
	if err != nil {
		fmt.Println("could not hash password")
		log.Fatal(err)
	}
	account := account.CreateAccount(acctRequest.FirstName, acctRequest.LastName, acctRequest.Username, password)
	if err := s.store.CreateAccount(account); err != nil {
		fmt.Println("error here", account)
		return err
	}
	tokenStr, err := createJWT(account)
	if err != nil {
		fmt.Println("jwt error")
		return err
	}
	fmt.Println("JWT token is", tokenStr)

	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed %v", r.Method)
	}
	loginReq := account.LoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		return err
	}
	defer r.Body.Close()
	account, err := s.store.VerifyLogin(loginReq)
	if err != nil {
		return err
	}

	tokenStr, err := createJWT(account)
	if err != nil {
		fmt.Println("jwt error")
		return err
	}
	fmt.Println("JWT token is", tokenStr)
	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleUpdateAccount(w http.ResponseWriter, r *http.Request) error {
	updateReq := account.UpdateAccountRequest{}
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		return err
	}
	id, err := misc.GetID(r)
	if err != nil {
		fmt.Println("Invalid ID given!!!")
		return err
	}

	currentUser, err := s.store.GetAccountByID(id)
	if err != nil {
		return fmt.Errorf("Account Does Not Exist")
	}
	if currentUser.Username != updateReq.CurrentUsername || !misc.CheckPasswordHash(updateReq.CurrentPassword, currentUser.Password) {
		return fmt.Errorf("Access Denied")
	}

	updateReq.Username = misc.DefaultValue(updateReq.Username, currentUser.Username)
	updateReq.Password, err = misc.HashPassword(misc.DefaultValue(updateReq.Password, currentUser.Password))
	if err != nil {
		return fmt.Errorf("Could Not Encrypt Password")
	}

	updateReq.FirstName = misc.DefaultValue(updateReq.FirstName, currentUser.FirstName)
	updateReq.LastName = misc.DefaultValue(updateReq.LastName, currentUser.LastName)
	if err := s.store.UpdateAccount(id, &updateReq); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, updateReq)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := misc.GetID(r)
	if err != nil {
		fmt.Println("invalid ID given!!!")
		return err
	}
	if err := s.store.DeleteAccount(id); err != nil {
		return fmt.Errorf("failed to delete account by id : %v", err)
	}
	return WriteJSON(w, http.StatusOK, map[string]int{"deleted": id})
}

//
// Transfers
//
func (s *APIServer) handleTransfers(w http.ResponseWriter, r *http.Request) error {
	allTransfers, err := s.store.GetAllTransfers()
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, allTransfers)
}

func (s *APIServer) handleMyBalance(w http.ResponseWriter, r *http.Request) error {
	id, err := misc.GetID(r)
	if err != nil {
		return fmt.Errorf("Permission Denied")
	}

	myBalance, err := s.store.GetAccountBalance(id)
	if err != nil {
		return err
	}
	transactions, err := s.store.GetMyTransfers(id)
	if err != nil {
		return err
	}
	myTransfers := transfer.MyTransfers{
		MyBalance:   *myBalance,
		MyTransfers: transactions,
	}
	return WriteJSON(w, http.StatusOK, myTransfers)
}

func (s *APIServer) handleDepositsAndWithdrawals(w http.ResponseWriter, r *http.Request) error {
	transferRequest := transfer.TransferRequest{}
	if err := json.NewDecoder(r.Body).Decode(&transferRequest); err != nil {
		return err
	}
	defer r.Body.Close()
	id, err := misc.GetID(r)
	if err != nil {
		return fmt.Errorf("Permission Denied")
	}
	currentUser, err := s.store.GetAccountByID(id)
	if err != nil {
		return fmt.Errorf("Account Does Not Exist")
	}
	myBalance, err := s.store.DepositWithdrawIntoMyAccount(id, &transferRequest)
	if err != nil {
		return err
	}
	time := time.Now().UTC()
	addTransfer := transfer.CreateTransfer(id, id, transferRequest.Amount, currentUser.Balance, myBalance.Balance, misc.DepositOrWithdrawal(transferRequest.Amount), time)
	s.store.CreateTransfer(addTransfer)
	return WriteJSON(w, http.StatusOK, myBalance)
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	transferRequest := transfer.TransferRequest{}
	if err := json.NewDecoder(r.Body).Decode(&transferRequest); err != nil {
		return err
	}
	defer r.Body.Close()
	id, err := misc.GetID(r)
	if err != nil {
		return fmt.Errorf("Permission Denied")
	}

	currentUser, err := s.store.GetAccountByID(id)
	if err != nil {
		return fmt.Errorf("Account Does Not Exist")
	}
	recipientUser, err := s.store.GetAccountByID(transferRequest.ToAccount)
	if err != nil {
		return fmt.Errorf("Account Does Not Exist")
	}

	transferResponse, err := s.store.Transfer(id, &transferRequest)
	if err != nil {
		return err
	}
	time := time.Now().UTC()
	addTransferFrom := transfer.CreateTransfer(id, transferRequest.ToAccount, transferRequest.Amount, currentUser.Balance, transferResponse.Account.Balance, "withdrawal", time)
	addTransferTo := transfer.CreateTransfer(id, transferRequest.ToAccount, transferRequest.Amount, recipientUser.Balance, recipientUser.Balance+int64(transferRequest.Amount), "deposit", time)
	if err := s.store.CreateTransfer(addTransferFrom); err != nil {
		return fmt.Errorf("Could Not Add Transaction To Table")
	}
	if err := s.store.CreateTransfer(addTransferTo); err != nil {
		return fmt.Errorf("Could Not Add Transaction To Table")
	}
	return WriteJSON(w, http.StatusOK, transferResponse)
}

//
func (s *APIServer) handleTransferAccount(w http.ResponseWriter, r *http.Request) error {

	switch r.Method {
	case "GET":
		// get current user balance
		return s.handleMyBalance(w, r)
	case "PATCH":
		// withdrawal/deposit into user account
		return s.handleDepositsAndWithdrawals(w, r)
	case "POST":
		// transfer from user to recipient
		return s.handleTransfer(w, r)
	}

	return fmt.Errorf("method not allowed : %v", r.Method)
}
