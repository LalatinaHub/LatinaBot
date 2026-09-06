package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/LalatinaHub/common/model"
)

type serverRepo struct {
	db *sql.DB
}

// NewServerRepository creates a new ServerRepository instance.
func NewServerRepository(db *sql.DB) ServerRepository {
	return &serverRepo{db: db}
}

func (r *serverRepo) GetAll(ctx context.Context) ([]model.Server, error) {
	query := "SELECT id, code, domain, ip, country, users_count, users_max FROM servers;"
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query servers: %w", err)
	}
	defer rows.Close()

	var servers []model.Server
	for rows.Next() {
		var s model.Server
		err := rows.Scan(
			&s.ID,
			&s.Code,
			&s.Domain,
			&s.IP,
			&s.Country,
			&s.UsersCount,
			&s.UsersMax,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan server: %w", err)
		}
		servers = append(servers, s)
	}

	return servers, rows.Err()
}

func (r *serverRepo) Update(ctx context.Context, s *model.Server) error {
	query := `UPDATE servers SET code = ?, domain = ?, ip = ?, country = ?, users_count = ?, users_max = ? WHERE id = ?;`
	_, err := r.db.ExecContext(ctx, query, s.Code, s.Domain, s.IP, s.Country, s.UsersCount, s.UsersMax, s.ID)
	if err != nil {
		return fmt.Errorf("failed to update server %d: %w", s.ID, err)
	}
	return nil
}

func (r *serverRepo) UpdateUserCount(ctx context.Context, serverID int64, count int64) error {
	query := "UPDATE servers SET users_count = ? WHERE id = ?;"
	_, err := r.db.ExecContext(ctx, query, count, serverID)
	if err != nil {
		return fmt.Errorf("failed to update server %d user count: %w", serverID, err)
	}
	return nil
}
