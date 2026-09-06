package edgeserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/LalatinaHub/LatinaBot/internal/domain"
	"github.com/LalatinaHub/LatinaBot/internal/repository"
	"github.com/LalatinaHub/LatinaBot/pkg/logger"
	"github.com/LalatinaHub/common/model"
)

// Client interacts with FoolVPN edge servers.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new Edge Server client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetInfo fetches /api/v1/info from the edge server and measures response latency (ping).
func (c *Client) GetInfo(ctx context.Context, domainName string) (*domain.EdgeServerInfo, error) {
	reqURL := fmt.Sprintf("http://%s/api/v1/info", domainName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &domain.EdgeServerInfo{Org: "Unknown", Ping: 0}, nil
	}
	defer resp.Body.Close()

	ping := time.Since(start).Milliseconds()

	if resp.StatusCode == http.StatusOK {
		var data struct {
			Org string `json:"org"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
			return &domain.EdgeServerInfo{
				Org:  data.Org,
				Ping: ping,
			}, nil
		}
	}

	return &domain.EdgeServerInfo{Org: "Unknown", Ping: ping}, nil
}

// GetStatus retrieves system hardware & network metrics from /api/v1/status.
func (c *Client) GetStatus(ctx context.Context, domainName string) (*domain.ServerStatus, error) {
	reqURL := fmt.Sprintf("http://%s/api/v1/status", domainName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server %s returned status %d", domainName, resp.StatusCode)
	}

	var status domain.ServerStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode status: %w", err)
	}

	return &status, nil
}

// GetRelays retrieves available relay proxy configurations from /api/v1/relay.
func (c *Client) GetRelays(ctx context.Context, domainName string) ([]model.ProxyNode, error) {
	reqURL := fmt.Sprintf("http://%s/api/v1/relay", domainName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server %s returned status %d for relay", domainName, resp.StatusCode)
	}

	var relays []model.ProxyNode
	if err := json.NewDecoder(resp.Body).Decode(&relays); err != nil {
		return nil, fmt.Errorf("failed to decode relays: %w", err)
	}

	return relays, nil
}

// ReloadServers notifies all cluster servers to reload configurations concurrently.
func (c *Client) ReloadServers(ctx context.Context, servers []model.Server, apiToken string) {
	if apiToken == "" {
		logger.Warn().Msg("Cannot reload servers: apiToken is empty")
		return
	}

	var wg sync.WaitGroup
	for _, srv := range servers {
		wg.Add(1)
		go func(domainName string) {
			defer wg.Done()
			reqURL := fmt.Sprintf("http://%s/api/v1/%s", domainName, apiToken)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
			if err != nil {
				return
			}
			resp, err := c.httpClient.Do(req)
			if err == nil {
				_ = resp.Body.Close()
			}
		}(srv.Domain)
	}

	wg.Wait()
}

// AssignServerTenants recalculates active users count per server and updates the database.
func AssignServerTenants(ctx context.Context, serverRepo repository.ServerRepository, userRepo repository.UserRepository) error {
	servers, err := serverRepo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get servers: %w", err)
	}

	users, err := userRepo.GetAllUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get users: %w", err)
	}

	for _, srv := range servers {
		var count int64 = 0
		for _, u := range users {
			if u.ServerCode == srv.Code && u.Quota > 10 {
				count++
			}
		}
		if err := serverRepo.UpdateUserCount(ctx, srv.ID, count); err != nil {
			logger.Error().Err(err).Int64("server_id", srv.ID).Msg("Failed to update server tenant count")
		}
	}

	return nil
}
