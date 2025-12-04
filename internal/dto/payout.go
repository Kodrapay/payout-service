package dto

type PayoutRequest struct {
	MerchantID       string `json:"merchant_id"`
	Reference        string `json:"reference"`
	Amount           int64  `json:"amount"`
	Currency         string `json:"currency"`
	RecipientName    string `json:"recipient_name"`
	RecipientAccount string `json:"recipient_account"`
	RecipientBank    string `json:"recipient_bank"`
	Narration        string `json:"narration"`
}

type PayoutResponse struct {
	ID        string `json:"id"`
	Reference string `json:"reference,omitempty"`
	Status    string `json:"status"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
}
