package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type kvRepo struct {
	db *sql.DB
}

// NewKVRepository creates a new KVRepository instance.
func NewKVRepository(db *sql.DB) KVRepository {
	return &kvRepo{db: db}
}

func (r *kvRepo) Get(ctx context.Context, key string) (string, error) {
	var val string
	query := "SELECT value FROM kv WHERE key = ? LIMIT 1;"
	err := r.db.QueryRowContext(ctx, query, key).Scan(&val)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("failed to get kv key %s: %w", key, err)
	}
	return val, nil
}
