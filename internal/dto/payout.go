package dto

type PayoutRequest struct {
	MerchantID       int    `json:"merchant_id"`
	Reference        int    `json:"reference"`
	Amount           int64  `json:"amount"`
	Currency         string `json:"currency"`
	RecipientName    string `json:"recipient_name"`
	RecipientAccount string `json:"recipient_account"`
	RecipientBank    string `json:"recipient_bank"`
	Narration        string `json:"narration"`
}

type PayoutResponse struct {
	ID        int    `json:"id"`
	Reference int    `json:"reference,omitempty"`
	Status    string `json:"status"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
}

type PayoutStatusUpdateRequest struct {
	Status string `json:"status"`
}
