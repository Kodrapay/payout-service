package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/kodra-pay/payout-service/internal/dto"
	"github.com/kodra-pay/payout-service/internal/models"
	"github.com/kodra-pay/payout-service/internal/repositories"
)

type PayoutService struct {
	repo               *repositories.PayoutRepository
	merchantServiceURL string
}

func NewPayoutService(repo *repositories.PayoutRepository, merchantServiceURL string) *PayoutService {
	return &PayoutService{
		repo:               repo,
		merchantServiceURL: merchantServiceURL,
	}
}

func (s *PayoutService) Create(ctx context.Context, req dto.PayoutRequest) (dto.PayoutResponse, error) {
	if req.Amount <= 0 || req.MerchantID == 0 { // int check
		return dto.PayoutResponse{}, fmt.Errorf("merchant_id and positive amount are required")
	}

	// Check available balance before creating payout
	available, err := s.getAvailableBalance(ctx, req.MerchantID, req.Currency)
	if err != nil {
		return dto.PayoutResponse{}, fmt.Errorf("failed to verify balance: %w", err)
	}
	if available < req.Amount {
		return dto.PayoutResponse{}, fmt.Errorf("insufficient available balance")
	}

	p := &models.Payout{
		MerchantID:       req.MerchantID, // int
		Reference:        req.Reference,  // int
		Amount:           req.Amount,
		Currency:         req.Currency,
		RecipientName:    req.RecipientName,
		RecipientAccount: req.RecipientAccount,
		RecipientBank:    req.RecipientBank,
		Status:           "pending",
		Narration:        req.Narration,
	}
	// Reference is an int. If req.Reference is 0, it means no reference was provided.
	// The DB will auto-generate p.ID.
	_ = s.repo.Create(ctx, p)

	// Simulate processing: move to processed after a short delay to mimic asynchronous payout handling.
	go func(payoutID int) { // int
		log.Printf("payout-service: starting simulated processing for payout %d", payoutID) // int
		time.Sleep(5 * time.Second)
		log.Printf("payout-service: attempting to update status for payout %d to 'processed'", payoutID) // int
		if err := s.repo.UpdateStatus(context.Background(), payoutID, "processed"); err != nil { // int
			log.Printf("payout-service: failed to auto-process payout %d: %v", payoutID, err) // int
		} else {
			log.Printf("payout-service: successfully auto-processed payout %d to 'processed'", payoutID) // int
		}
	}(p.ID) // int

	return dto.PayoutResponse{
		ID:        p.ID,        // int
		Status:    p.Status,
		Amount:    p.Amount,
		Currency:  p.Currency,
		Reference: p.Reference, // int
	}, nil
}

func (s *PayoutService) Get(ctx context.Context, id int) dto.PayoutResponse { // int
	p, _ := s.repo.GetByID(ctx, id) // int
	if p == nil {
		return dto.PayoutResponse{}
	}
	return dto.PayoutResponse{
		ID:        p.ID,        // int
		Status:    p.Status,
		Amount:    p.Amount,
		Currency:  p.Currency,
		Reference: p.Reference, // int
	}
}

// getAvailableBalance fetches merchant available balance from merchant-service
func (s *PayoutService) getAvailableBalance(ctx context.Context, merchantID int, currency string) (int64, error) { // int
	url := fmt.Sprintf("%s/merchants/%d/balance?currency=%s", strings.TrimRight(s.merchantServiceURL, "/"), merchantID, currency) // int
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("merchant service returned %d", resp.StatusCode)
	}
	var payload struct {
		AvailableBalance int64 `json:"available_balance"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, err
	}
	return payload.AvailableBalance, nil
}

func (s *PayoutService) List(ctx context.Context, merchantID int) []dto.PayoutResponse { // int
	list, _ := s.repo.ListByMerchant(ctx, merchantID, 50) // int
	var resp []dto.PayoutResponse
	for _, p := range list {
		resp = append(resp, dto.PayoutResponse{
			ID:        p.ID,        // int
			Status:    p.Status,
			Amount:    p.Amount,
			Currency:  p.Currency,
			Reference: p.Reference, // int
		})
	}
	return resp
}

func (s *PayoutService) Cancel(_ context.Context, id int) map[string]interface{} { // int, map[string]interface{}
	return map[string]interface{}{"id": id, "status": "cancelled"}
}

func (s *PayoutService) UpdateStatus(ctx context.Context, id int, status string) (dto.PayoutResponse, error) { // int
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

	if err := s.repo.UpdateStatus(ctx, id, normalized); err != nil { // int
		return dto.PayoutResponse{}, err
	}

	updated, err := s.repo.GetByID(ctx, id) // int
	if err != nil || updated == nil {
		return dto.PayoutResponse{}, fmt.Errorf("payout not found after update")
	}

	// Normalize "processed" to "completed" for display consistency
	displayStatus := updated.Status
	if displayStatus == "processed" {
		displayStatus = "completed"
	}

	return dto.PayoutResponse{
		ID:        updated.ID,        // int
		Reference: updated.Reference, // int
		Status:    displayStatus,
		Amount:    updated.Amount,
		Currency:  updated.Currency,
	}, nil
}
