package models

import "time"

type Payout struct {
	ID             int       `json:"id"`
	MerchantID     int       `json:"merchant_id"`
	Reference      int       `json:"reference"`
	Amount         int64     `json:"amount"`
	Currency       string    `json:"currency"`
	RecipientName  string    `json:"recipient_name"`
	RecipientAccount string  `json:"recipient_account"`
	RecipientBank  string    `json:"recipient_bank"`
	Status         string    `json:"status"`
	Narration      string    `json:"narration,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
