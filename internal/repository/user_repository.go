package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/LalatinaHub/common/model"
	"github.com/google/uuid"
)

type userRepo struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository instance.
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) GetUser(ctx context.Context, id int64) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, token, password, expired, server_code, quota, relay, adblock, vpn FROM users WHERE id = ?;", id)

	var (
		u          model.User
		expiredStr string
		adblockVal any
	)

	err := row.Scan(
		&u.ID,
		&u.Token,
		&u.Password,
		&expiredStr,
		&u.ServerCode,
		&u.Quota,
		&u.Relay,
		&adblockVal,
		&u.VPN,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query user %d: %w", id, err)
	}

	u.Adblock = parseBool(adblockVal)
	u.Expired = parseDate(expiredStr)

	return &u, nil
}

func (r *userRepo) CreateUser(ctx context.Context, id int64) (*model.User, error) {
	token := generateRandomToken(8)
	password := uuid.New().String()
	expired := time.Now().AddDate(0, 0, 7)
	expiredStr := expired.Format("2006-01-02")

	query := `INSERT INTO users (id, token, password, expired, server_code, quota, relay, adblock, vpn)
	          VALUES (?, ?, ?, ?, '', 10000, '', 1, 'vless');`

	_, err := r.db.ExecContext(ctx, query, id, token, password, expiredStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create user %d: %w", id, err)
	}

	return &model.User{
		ID:         id,
		Token:      token,
		Password:   password,
		Expired:    expired,
		ServerCode: "",
		Quota:      10000,
		Relay:      "",
		Adblock:    true,
		VPN:        "vless",
	}, nil
}

func (r *userRepo) UpdateUser(ctx context.Context, u *model.User) error {
	query := `UPDATE users SET 
	          token = ?, password = ?, expired = ?, server_code = ?, 
	          quota = ?, relay = ?, adblock = ?, vpn = ?
	          WHERE id = ?;`

	adblockInt := 0
	if u.Adblock {
		adblockInt = 1
	}

	_, err := r.db.ExecContext(ctx, query,
		u.Token,
		u.Password,
		u.Expired.Format("2006-01-02"),
		u.ServerCode,
		u.Quota,
		u.Relay,
		adblockInt,
		u.VPN,
		u.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user %d: %w", u.ID, err)
	}
	return nil
}

func (r *userRepo) GetAllUsers(ctx context.Context) ([]model.User, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, token, password, expired, server_code, quota, relay, adblock, vpn FROM users;")
	if err != nil {
		return nil, fmt.Errorf("failed to query all users: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var (
			u          model.User
			expiredStr string
			adblockVal any
		)

		err := rows.Scan(
			&u.ID,
			&u.Token,
			&u.Password,
			&expiredStr,
			&u.ServerCode,
			&u.Quota,
			&u.Relay,
			&adblockVal,
			&u.VPN,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		u.Adblock = parseBool(adblockVal)
		u.Expired = parseDate(expiredStr)
		users = append(users, u)
	}

	return users, rows.Err()
}

func (r *userRepo) DeleteUsers(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("DELETE FROM users WHERE id IN (%s);", strings.Join(placeholders, ","))
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete users: %w", err)
	}
	return nil
}

func (r *userRepo) CleanExpiredUsers(ctx context.Context, olderThanDays int) ([]int64, error) {
	users, err := r.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	cutoffDuration := time.Duration(olderThanDays) * 24 * time.Hour

	var expiredIDs []int64
	for _, u := range users {
		if now.Sub(u.Expired) > cutoffDuration {
			expiredIDs = append(expiredIDs, u.ID)
		}
	}

	if len(expiredIDs) > 0 {
		if err := r.DeleteUsers(ctx, expiredIDs); err != nil {
			return nil, err
		}
	}

	return expiredIDs, nil
}

func (r *userRepo) CleanExceededQuota(ctx context.Context, thresholdMB int64) error {
	query := "UPDATE users SET server_code = '' WHERE quota <= ? AND server_code != '';"
	_, err := r.db.ExecContext(ctx, query, thresholdMB)
	if err != nil {
		return fmt.Errorf("failed to reset server_code for quota-exceeded users: %w", err)
	}
	return nil
}

func parseBool(val any) bool {
	if b, ok := val.(bool); ok {
		return b
	}
	if i, ok := val.(int64); ok {
		return i > 0
	}
	if s, ok := val.(string); ok {
		return s == "1" || strings.ToLower(s) == "true"
	}
	return false
}

func parseDate(s string) time.Time {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	return time.Time{}
}

func generateRandomToken(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	randBytes := make([]byte, n)
	_, _ = rand.Read(randBytes)
	for i := range b {
		b[i] = charset[int(randBytes[i])%len(charset)]
	}
	return string(b)
}
