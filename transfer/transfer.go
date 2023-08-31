package transfer

type TransferRequest struct {
	ToAccount int `json:"to-account"`
	Amount    int `json:"amount"`
}
