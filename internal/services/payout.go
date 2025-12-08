package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/kodra-pay/payout-service/internal/dto"
	"github.com/kodra-pay/payout-service/internal/models"
	"github.com/kodra-pay/payout-service/internal/repositories"
)

type PayoutService struct {
	repo                  *repositories.PayoutRepository
	merchantServiceURL    string
	transactionServiceURL string
}

func NewPayoutService(repo *repositories.PayoutRepository, merchantServiceURL, transactionServiceURL string) *PayoutService {
	return &PayoutService{
		repo:                  repo,
		merchantServiceURL:    merchantServiceURL,
		transactionServiceURL: transactionServiceURL,
	}
}

func (s *PayoutService) Create(ctx context.Context, req dto.PayoutRequest) (dto.PayoutResponse, error) {
	if req.Amount <= 0 || req.MerchantID == 0 { // int check
		return dto.PayoutResponse{}, fmt.Errorf("merchant_id and positive amount are required")
	}

	amountKobo := int64(math.Round(req.Amount * 100))

	// Check available balance before creating payout
	available, err := s.getAvailableBalance(ctx, req.MerchantID, req.Currency)
	if err != nil {
		return dto.PayoutResponse{}, fmt.Errorf("failed to verify balance: %w", err)
	}
	if available < amountKobo {
		return dto.PayoutResponse{}, fmt.Errorf("insufficient available balance")
	}

	// Generate a reference if not provided to avoid duplicate zero values
	if req.Reference == 0 {
		req.Reference = int(time.Now().UnixNano() / 1e6) // ms timestamp
	}

	p := &models.Payout{
		MerchantID:       req.MerchantID, // int
		Reference:        req.Reference,  // int
		Amount:           amountKobo,
		Currency:         req.Currency,
		RecipientName:    req.RecipientName,
		RecipientAccount: req.RecipientAccount,
		RecipientBank:    req.RecipientBank,
		Status:           "pending",
		Narration:        req.Narration,
	}
	// Reference is an int. If req.Reference is 0, it means no reference was provided.
	// The DB will auto-generate p.ID.
	if err := s.repo.Create(ctx, p); err != nil {
		return dto.PayoutResponse{}, fmt.Errorf("failed to create payout: %w", err)
	}

	// Simulate processing: move to processed after a short delay to mimic asynchronous payout handling.
	go func(payoutID int) { // int
		log.Printf("payout-service: starting simulated processing for payout %d", payoutID) // int
		time.Sleep(5 * time.Second)
		// Only auto-process if the payout is still pending to avoid retrying failed/updated payouts.
		current, err := s.repo.GetByID(context.Background(), payoutID)
		if err != nil || current == nil {
			log.Printf("payout-service: skipping auto-process for payout %d: not found or error: %v", payoutID, err)
			return
		}
		if strings.ToLower(current.Status) != "pending" {
			log.Printf("payout-service: skipping auto-process for payout %d because status is %s", payoutID, current.Status)
			return
		}

		log.Printf("payout-service: attempting to update status for payout %d to 'processed'", payoutID) // int
		if _, err := s.UpdateStatus(context.Background(), payoutID, "processed"); err != nil {           // int
			log.Printf("payout-service: failed to auto-process payout %d: %v", payoutID, err) // int
		} else {
			log.Printf("payout-service: successfully auto-processed payout %d to 'processed'", payoutID) // int
		}
	}(p.ID) // int

	return dto.PayoutResponse{
		ID:        p.ID, // int
		Status:    p.Status,
		Amount:    float64(p.Amount) / 100,
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
		ID:        p.ID, // int
		Status:    p.Status,
		Amount:    float64(p.Amount) / 100,
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
		AvailableBalance float64 `json:"available_balance"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, err
	}
	return int64(math.Round(payload.AvailableBalance * 100)), nil
}

func (s *PayoutService) List(ctx context.Context, merchantID int) []dto.PayoutResponse { // int
	list, _ := s.repo.ListByMerchant(ctx, merchantID, 50) // int
	var resp []dto.PayoutResponse
	for _, p := range list {
		resp = append(resp, dto.PayoutResponse{
			ID:        p.ID, // int
			Status:    p.Status,
			Amount:    float64(p.Amount) / 100,
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

	// Fetch current state to avoid double-deducting on repeated calls
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return dto.PayoutResponse{}, err
	}
	if current == nil {
		return dto.PayoutResponse{}, fmt.Errorf("payout not found")
	}
	previousStatus := strings.ToLower(current.Status)

	// Avoid re-processing already finalized payouts
	if isFinalStatus(previousStatus) && isFinalStatus(normalized) {
		return dto.PayoutResponse{
			ID:        current.ID,
			Reference: current.Reference,
			Status:    current.Status,
			Amount:    float64(current.Amount) / 100,
			Currency:  current.Currency,
		}, nil
	}

	if err := s.repo.UpdateStatus(ctx, id, normalized); err != nil { // int
		return dto.PayoutResponse{}, err
	}

	updated, err := s.repo.GetByID(ctx, id) // int
	if err != nil || updated == nil {
		return dto.PayoutResponse{}, fmt.Errorf("payout not found after update")
	}

	// On completion, deduct available balance and record payout transaction
	if isFinalStatus(normalized) && !isFinalStatus(previousStatus) {
		if err := s.handlePayoutCompletion(ctx, updated); err != nil {
			_ = s.repo.UpdateStatus(context.Background(), id, "failed")
			return dto.PayoutResponse{}, fmt.Errorf("failed to finalize payout: %w", err)
		}
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
		Amount:    float64(updated.Amount) / 100,
		Currency:  updated.Currency,
	}, nil
}

func isFinalStatus(status string) bool {
	switch strings.ToLower(status) {
	case "processed", "completed":
		return true
	default:
		return false
	}
}

// handlePayoutCompletion deducts merchant available balance and logs a payout transaction
func (s *PayoutService) handlePayoutCompletion(ctx context.Context, p *models.Payout) error {
	if err := s.deductMerchantBalance(ctx, p.MerchantID, p.Currency, p.Amount); err != nil {
		return err
	}

	if err := s.recordPayoutTransaction(ctx, p); err != nil {
		// If the transaction log fails, we don't want to double-deduct on retry.
		// Log and continue so the payout remains completed.
		log.Printf("payout-service: failed to record payout transaction for payout %d: %v", p.ID, err)
	}

	return nil
}

func (s *PayoutService) deductMerchantBalance(ctx context.Context, merchantID int, currency string, amount int64) error {
	url := fmt.Sprintf("%s/internal/balance/payout", strings.TrimRight(s.merchantServiceURL, "/"))

	payload := map[string]interface{}{
		"merchant_id": merchantID,
		"currency":    currency,
		"amount":      float64(amount) / 100, // send in currency units
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("merchant service returned %d: %s", resp.StatusCode, string(b))
	}

	return nil
}

func (s *PayoutService) recordPayoutTransaction(ctx context.Context, p *models.Payout) error {
	if s.transactionServiceURL == "" {
		return fmt.Errorf("transaction service URL not configured")
	}

	reference := fmt.Sprintf("payout-%d", p.ID)
	if p.Reference != 0 {
		reference = fmt.Sprintf("payout-%d", p.Reference)
	}

	payload := map[string]interface{}{
		"reference":      reference,
		"merchant_id":    p.MerchantID,
		"amount":         float64(p.Amount) / 100, // send in currency units
		"currency":       p.Currency,
		"payment_method": "payout",
		"status":         "payout",
		"description":    fmt.Sprintf("Payout to %s (%s)", p.RecipientName, p.RecipientBank),
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/transactions", strings.TrimRight(s.transactionServiceURL, "/")), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("transaction service returned %d: %s", resp.StatusCode, string(b))
	}

	return nil
}
