package transfer

import "time"

type TransferRequest struct {
	ToAccount int `json:"to-account"`
	Amount    int `json:"amount"`
}

type TransferResponse struct {
	Account MyBalance `json:"my-account"`
	Sent    bool      `json:"sent"`
}

type MyBalance struct {
	Username        string `json:"username"`
	MyAccountNumber int64  `json:"my-account"`
	Balance         int64  `json:"balance"`
}

type Transfer struct {
	From            int       `json:"from"`
	To              int       `json:"to"`
	Amount          int       `json:"amount"`
	Action          string    `json:"action"`
	PreviousBalance int64     `json:"previous-balance"`
	CurrentBalance  int64     `json:"current-balance"`
	CompletedAt     time.Time `json:"completed-at"`
}

func CreateTransfer(from, to, amount int, previous, new int64, action string, time time.Time) *Transfer {
	return &Transfer{
		From:            from,
		To:              to,
		Amount:          amount,
		Action:          action,
		PreviousBalance: previous,
		CurrentBalance:  new,
		CompletedAt:    time,
	}
}
