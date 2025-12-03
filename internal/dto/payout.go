package dto

type PayoutRequest struct {
	MerchantID  string  `json:"merchant_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	BankAccount string  `json:"bank_account"`
	Reason      string  `json:"reason"`
}

type PayoutResponse struct {
	ID       string  `json:"id"`
	Status   string  `json:"status"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}
