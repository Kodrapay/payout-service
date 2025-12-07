package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/kodra-pay/payout-service/internal/models"
)

type PayoutRepository struct {
	db *sql.DB
}

func NewPayoutRepository(dsn string) (*PayoutRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	return &PayoutRepository{db: db}, nil
}

func (r *PayoutRepository) Create(ctx context.Context, p *models.Payout) error {
	query := `
		INSERT INTO payouts (merchant_id, reference, amount, currency, recipient_name, recipient_account, recipient_bank, status, narration, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		p.MerchantID, p.Reference, p.Amount, p.Currency,
		p.RecipientName, p.RecipientAccount, p.RecipientBank,
		p.Status, p.Narration,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *PayoutRepository) GetByID(ctx context.Context, id int) (*models.Payout, error) {
	query := `
		SELECT id, merchant_id, reference, amount, currency, recipient_name, recipient_account, recipient_bank, status, narration, created_at, updated_at
		FROM payouts
		WHERE id = $1
	`
	var p models.Payout
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.MerchantID, &p.Reference, &p.Amount, &p.Currency,
		&p.RecipientName, &p.RecipientAccount, &p.RecipientBank,
		&p.Status, &p.Narration, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func (r *PayoutRepository) ListByMerchant(ctx context.Context, merchantID int, limit int) ([]*models.Payout, error) {
	query := `
		SELECT id, merchant_id, reference, amount, currency, recipient_name, recipient_account, recipient_bank, status, narration, created_at, updated_at
		FROM payouts
		WHERE merchant_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, merchantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.Payout
	for rows.Next() {
		var p models.Payout
		if err := rows.Scan(
			&p.ID, &p.MerchantID, &p.Reference, &p.Amount, &p.Currency,
			&p.RecipientName, &p.RecipientAccount, &p.RecipientBank,
			&p.Status, &p.Narration, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}

// UpdateStatus updates the payout status and refreshed updated_at.
func (r *PayoutRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	query := `
		UPDATE payouts
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, query, id, status)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("payout not found")
	}
	return nil
}
