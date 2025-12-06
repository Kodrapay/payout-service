package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/kodra-pay/payout-service/internal/dto"
	"github.com/kodra-pay/payout-service/internal/models"
	"github.com/kodra-pay/payout-service/internal/repositories"
)

type PayoutService struct {
	repo *repositories.PayoutRepository
}

func NewPayoutService(repo *repositories.PayoutRepository) *PayoutService {
	return &PayoutService{repo: repo}
}

func (s *PayoutService) Create(ctx context.Context, req dto.PayoutRequest) dto.PayoutResponse {
	p := &models.Payout{
		MerchantID:       req.MerchantID,
		Reference:        req.Reference,
		Amount:           req.Amount,
		Currency:         req.Currency,
		RecipientName:    req.RecipientName,
		RecipientAccount: req.RecipientAccount,
		RecipientBank:    req.RecipientBank,
		Status:           "pending",
		Narration:        req.Narration,
	}
	if p.Reference == "" {
		p.Reference = "pyt_" + uuid.NewString()
	}
	_ = s.repo.Create(ctx, p)

	// Simulate processing: move to processed after a short delay to mimic asynchronous payout handling.
	go func(payoutID string) {
		time.Sleep(5 * time.Second)
		if err := s.repo.UpdateStatus(context.Background(), payoutID, "processed"); err != nil {
			log.Printf("payout-service: failed to auto-process payout %s: %v", payoutID, err)
		}
	}(p.ID)

	return dto.PayoutResponse{
		ID:        p.ID,
		Status:    p.Status,
		Amount:    p.Amount,
		Currency:  p.Currency,
		Reference: p.Reference,
	}
}

func (s *PayoutService) Get(ctx context.Context, id string) dto.PayoutResponse {
	p, _ := s.repo.GetByID(ctx, id)
	if p == nil {
		return dto.PayoutResponse{}
	}
	return dto.PayoutResponse{
		ID:        p.ID,
		Status:    p.Status,
		Amount:    p.Amount,
		Currency:  p.Currency,
		Reference: p.Reference,
	}
}

func (s *PayoutService) List(ctx context.Context, merchantID string) []dto.PayoutResponse {
	list, _ := s.repo.ListByMerchant(ctx, merchantID, 50)
	var resp []dto.PayoutResponse
	for _, p := range list {
		resp = append(resp, dto.PayoutResponse{
			ID:        p.ID,
			Status:    p.Status,
			Amount:    p.Amount,
			Currency:  p.Currency,
			Reference: p.Reference,
		})
	}
	return resp
}

func (s *PayoutService) Cancel(_ context.Context, id string) map[string]string {
	return map[string]string{"id": id, "status": "cancelled"}
}

func (s *PayoutService) UpdateStatus(ctx context.Context, id, status string) (dto.PayoutResponse, error) {
	normalized := status
	if normalized == "" {
		return dto.PayoutResponse{}, fmt.Errorf("status is required")
	}
	normalized = strings.ToLower(normalized)

	switch normalized {
	case "pending", "processing", "processed", "completed", "failed":
		// allowed
	default:
		return dto.PayoutResponse{}, fmt.Errorf("invalid status")
	}

	if err := s.repo.UpdateStatus(ctx, id, normalized); err != nil {
		return dto.PayoutResponse{}, err
	}

	updated, err := s.repo.GetByID(ctx, id)
	if err != nil || updated == nil {
		return dto.PayoutResponse{}, fmt.Errorf("payout not found after update")
	}

	// Normalize "processed" to "completed" for display consistency
	displayStatus := updated.Status
	if displayStatus == "processed" {
		displayStatus = "completed"
	}

	return dto.PayoutResponse{
		ID:        updated.ID,
		Reference: updated.Reference,
		Status:    displayStatus,
		Amount:    updated.Amount,
		Currency:  updated.Currency,
	}, nil
}
