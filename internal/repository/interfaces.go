package repository

import (
	"context"

	"github.com/LalatinaHub/LatinaBot/internal/domain"
	"github.com/LalatinaHub/common/model"
)

// UserRepository handles user persistence operations.
type UserRepository interface {
	GetUser(ctx context.Context, id int64) (*model.User, error)
	CreateUser(ctx context.Context, id int64) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	GetAllUsers(ctx context.Context) ([]model.User, error)
	DeleteUsers(ctx context.Context, ids []int64) error
	CleanExpiredUsers(ctx context.Context, olderThanDays int) ([]int64, error)
	CleanExceededQuota(ctx context.Context, thresholdMB int64) error
}

// ServerRepository handles cluster server queries and updates.
type ServerRepository interface {
	GetAll(ctx context.Context) ([]model.Server, error)
	Update(ctx context.Context, server *model.Server) error
	UpdateUserCount(ctx context.Context, serverID int64, count int64) error
}

// ProxyRepository handles proxy node queries.
type ProxyRepository interface {
	GetRandomCDNProxy(ctx context.Context) (*model.ProxyNode, error)
}

// WildcardRepository handles wildcard domain queries.
type WildcardRepository interface {
	GetAll(ctx context.Context) ([]domain.Wildcard, error)
	Create(ctx context.Context, domainName string) error
	Exists(ctx context.Context, domainName string) (bool, error)
}

// DonationRepository handles processed donation tracking.
type DonationRepository interface {
	Exists(ctx context.Context, orderID string) (bool, error)
	Create(ctx context.Context, orderID string) error
}

// KVRepository handles global key-value configuration queries.
type KVRepository interface {
	Get(ctx context.Context, key string) (string, error)
}

// Repositories holds all repository instances.
type Repositories struct {
	User     UserRepository
	Server   ServerRepository
	Proxy    ProxyRepository
	Wildcard WildcardRepository
	Donation DonationRepository
	KV       KVRepository
}
