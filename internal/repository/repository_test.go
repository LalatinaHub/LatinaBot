package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LalatinaHub/common/model"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_GetUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("User Found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "token", "password", "expired", "server_code", "quota", "relay", "adblock", "vpn",
		}).AddRow(123456, "tok12345", "pass-uuid", "2026-09-13", "sg1", 50000, "SG", 1, "vless")

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, token, password, expired, server_code, quota, relay, adblock, vpn FROM users WHERE id = ?;")).
			WithArgs(int64(123456)).
			WillReturnRows(rows)

		user, err := repo.GetUser(ctx, 123456)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, int64(123456), user.ID)
		assert.Equal(t, "tok12345", user.Token)
		assert.Equal(t, "sg1", user.ServerCode)
		assert.True(t, user.Adblock)
		assert.Equal(t, "vless", user.VPN)
	})

	t.Run("User Not Found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, token, password, expired, server_code, quota, relay, adblock, vpn FROM users WHERE id = ?;")).
			WithArgs(int64(999999)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		user, err := repo.GetUser(ctx, 999999)
		assert.NoError(t, err)
		assert.Nil(t, user)
	})
}

func TestUserRepository_CreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (id, token, password, expired, server_code, quota, relay, adblock, vpn)")).
		WithArgs(int64(123), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	user, err := repo.CreateUser(ctx, 123)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, int64(123), user.ID)
	assert.Equal(t, int64(10000), user.Quota)
	assert.Equal(t, "vless", user.VPN)
	assert.True(t, user.Adblock)
}

func TestUserRepository_UpdateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{
		ID:         123,
		Token:      "newtoken",
		Password:   "newpass",
		Expired:    time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		ServerCode: "id1",
		Quota:      20000,
		Relay:      "ID",
		Adblock:    false,
		VPN:        "trojan",
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET")).
		WithArgs("newtoken", "newpass", "2026-10-01", "id1", int64(20000), "ID", 0, "trojan", int64(123)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.UpdateUser(ctx, user)
	assert.NoError(t, err)
}

func TestUserRepository_DeleteUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users WHERE id IN (?,?);")).
		WithArgs(int64(1), int64(2)).
		WillReturnResult(sqlmock.NewResult(2, 2))

	err = repo.DeleteUsers(ctx, []int64{1, 2})
	assert.NoError(t, err)
}

func TestServerRepository_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewServerRepository(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"id", "code", "domain", "ip", "country", "users_count", "users_max",
	}).AddRow(1, "sg1", "sg1.example.com", "1.1.1.1", "SG", 10, 100)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, code, domain, ip, country, users_count, users_max FROM servers;")).
		WillReturnRows(rows)

	servers, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, servers, 1)
	assert.Equal(t, "sg1", servers[0].Code)
}

func TestWildcardRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewWildcardRepository(db)
	ctx := context.Background()

	// Exists check
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM wildcards WHERE domain = ? LIMIT 1;")).
		WithArgs("example.com").
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	exists, err := repo.Exists(ctx, "example.com")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Create check
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO wildcards (domain) VALUES (?);")).
		WithArgs("test.com").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(ctx, "test.com")
	assert.NoError(t, err)
}

func TestDonationRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewDonationRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM donations WHERE order_id = ? LIMIT 1;")).
		WithArgs("order-123").
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	exists, err := repo.Exists(ctx, "order-123")
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestKVRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewKVRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT value FROM kv WHERE key = ? LIMIT 1;")).
		WithArgs("apiToken").
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("secret-token"))

	val, err := repo.Get(ctx, "apiToken")
	assert.NoError(t, err)
	assert.Equal(t, "secret-token", val)
}
