package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/kodra-pay/payout-service/internal/dto"
	"github.com/kodra-pay/payout-service/internal/models"
	"github.com/kodra-pay/payout-service/internal/repositories"
)

type PayoutService struct {
	repo *repositories.PayoutRepository
}

func NewPayoutService(repo *repositories.PayoutRepository) *PayoutService { return &PayoutService{repo: repo} }

func (s *PayoutService) Create(ctx context.Context, req dto.PayoutRequest) dto.PayoutResponse {
	p := &models.Payout{
		MerchantID:      req.MerchantID,
		Reference:       req.Reference,
		Amount:          req.Amount,
		Currency:        req.Currency,
		RecipientName:   req.RecipientName,
		RecipientAccount: req.RecipientAccount,
		RecipientBank:   req.RecipientBank,
		Status:          "pending",
		Narration:       req.Narration,
	}
	if p.Reference == "" {
		p.Reference = "pyt_" + uuid.NewString()
	}
	_ = s.repo.Create(ctx, p)
	return dto.PayoutResponse{
		ID:       p.ID,
		Status:   p.Status,
		Amount:   p.Amount,
		Currency: p.Currency,
		Reference: p.Reference,
	}
}

func (s *PayoutService) Get(ctx context.Context, id string) dto.PayoutResponse {
	p, _ := s.repo.GetByID(ctx, id)
	if p == nil {
		return dto.PayoutResponse{}
	}
	return dto.PayoutResponse{
		ID:       p.ID,
		Status:   p.Status,
		Amount:   p.Amount,
		Currency: p.Currency,
		Reference: p.Reference,
	}
}

func (s *PayoutService) List(ctx context.Context, merchantID string) []dto.PayoutResponse {
	list, _ := s.repo.ListByMerchant(ctx, merchantID, 50)
	var resp []dto.PayoutResponse
	for _, p := range list {
		resp = append(resp, dto.PayoutResponse{
			ID:       p.ID,
			Status:   p.Status,
			Amount:   p.Amount,
			Currency: p.Currency,
			Reference: p.Reference,
		})
	}
	return resp
}

func (s *PayoutService) Cancel(_ context.Context, id string) map[string]string {
	return map[string]string{"id": id, "status": "cancelled"}
}
