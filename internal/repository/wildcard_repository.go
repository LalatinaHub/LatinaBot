package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/LalatinaHub/LatinaBot/internal/domain"
)

type wildcardRepo struct {
	db *sql.DB
}

// NewWildcardRepository creates a new WildcardRepository instance.
func NewWildcardRepository(db *sql.DB) WildcardRepository {
	return &wildcardRepo{db: db}
}

func (r *wildcardRepo) GetAll(ctx context.Context) ([]domain.Wildcard, error) {
	query := "SELECT id, domain FROM wildcards;"
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query wildcards: %w", err)
	}
	defer rows.Close()

	var wildcards []domain.Wildcard
	for rows.Next() {
		var w domain.Wildcard
		if err := rows.Scan(&w.ID, &w.Domain); err != nil {
			return nil, fmt.Errorf("failed to scan wildcard: %w", err)
		}
		wildcards = append(wildcards, w)
	}

	return wildcards, rows.Err()
}

func (r *wildcardRepo) Create(ctx context.Context, domainName string) error {
	query := "INSERT INTO wildcards (domain) VALUES (?);"
	_, err := r.db.ExecContext(ctx, query, domainName)
	if err != nil {
		return fmt.Errorf("failed to insert wildcard %s: %w", domainName, err)
	}
	return nil
}

func (r *wildcardRepo) Exists(ctx context.Context, domainName string) (bool, error) {
	var exists int
	query := "SELECT 1 FROM wildcards WHERE domain = ? LIMIT 1;"
	err := r.db.QueryRowContext(ctx, query, domainName).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check wildcard existence %s: %w", domainName, err)
	}
	return exists == 1, nil
}
