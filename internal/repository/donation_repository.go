package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type donationRepo struct {
	db *sql.DB
}

// NewDonationRepository creates a new DonationRepository instance.
func NewDonationRepository(db *sql.DB) DonationRepository {
	return &donationRepo{db: db}
}

func (r *donationRepo) Exists(ctx context.Context, orderID string) (bool, error) {
	var exists int
	query := "SELECT 1 FROM donations WHERE order_id = ? LIMIT 1;"
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check donation %s: %w", orderID, err)
	}
	return exists == 1, nil
}

func (r *donationRepo) Create(ctx context.Context, orderID string) error {
	query := "INSERT INTO donations (order_id) VALUES (?);"
	_, err := r.db.ExecContext(ctx, query, orderID)
	if err != nil {
		return fmt.Errorf("failed to insert donation %s: %w", orderID, err)
	}
	return nil
}
