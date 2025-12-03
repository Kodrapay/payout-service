package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/kodra-pay/payout-service/internal/dto"
)

type PayoutService struct{}

func NewPayoutService() *PayoutService { return &PayoutService{} }

func (s *PayoutService) Create(_ context.Context, req dto.PayoutRequest) dto.PayoutResponse {
	return dto.PayoutResponse{
		ID:       "payout_" + uuid.NewString(),
		Status:   "processing",
		Amount:   req.Amount,
		Currency: req.Currency,
	}
}

func (s *PayoutService) Get(_ context.Context, id string) dto.PayoutResponse {
	return dto.PayoutResponse{
		ID:       id,
		Status:   "processing",
		Amount:   0,
		Currency: "NGN",
	}
}

func (s *PayoutService) Cancel(_ context.Context, id string) map[string]string {
	return map[string]string{"id": id, "status": "cancelled"}
}
