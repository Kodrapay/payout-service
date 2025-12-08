package dto

type DeductBalanceRequest struct {
	MerchantID int    `json:"merchant_id"`
	Amount     int64  `json:"amount"`
	Currency   string `json:"currency"`
}

type DeductBalanceResponse struct {
	Success bool `json:"success"`
}
