package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/LalatinaHub/LatinaBot/internal/domain"
	"github.com/LalatinaHub/LatinaBot/pkg/logger"
	"github.com/LalatinaHub/common/model"
)

// DNSRecord represents a Cloudflare DNS record.
type DNSRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
}

// WorkerDomain represents a Cloudflare Worker custom domain binding.
type WorkerDomain struct {
	ID          string `json:"id"`
	ZoneID      string `json:"zone_id"`
	Hostname    string `json:"hostname"`
	Service     string `json:"service"`
	Environment string `json:"environment"`
}

// Client manages Cloudflare DNS and Worker domains.
type Client struct {
	email       string
	apiKey      string
	accountID   string
	zoneID      string
	serviceName string
	environment string
	baseURL     string
	httpClient  *http.Client
	cachedZone  string
	mu          sync.RWMutex
}

// NewClient creates a new Cloudflare client.
func NewClient(email, apiKey, accountID, zoneID, serviceName, environment string) *Client {
	return &Client{
		email:       email,
		apiKey:      apiKey,
		accountID:   accountID,
		zoneID:      zoneID,
		serviceName: serviceName,
		environment: environment,
		baseURL:     "https://api.cloudflare.com/client/v4",
		httpClient:  &http.Client{Timeout: 15 * time.Second},
	}
}

// SetBaseURL overrides base URL for testing.
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if c.email != "" {
		req.Header.Set("X-Auth-Email", c.email)
		req.Header.Set("X-Auth-Key", c.apiKey)
	} else if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
}

// GetZoneID retrieves zone ID by domain name or returns configured zoneID.
func (c *Client) GetZoneID(ctx context.Context, domainName string) (string, error) {
	c.mu.RLock()
	if c.zoneID != "" {
		defer c.mu.RUnlock()
		return c.zoneID, nil
	}
	if c.cachedZone != "" {
		defer c.mu.RUnlock()
		return c.cachedZone, nil
	}
	c.mu.RUnlock()

	reqURL := fmt.Sprintf("%s/zones?name=%s", c.baseURL, domainName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Result []struct {
			ID string `json:"id"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Result) == 0 {
		return "", fmt.Errorf("zone not found for domain %s", domainName)
	}

	c.mu.Lock()
	c.cachedZone = result.Result[0].ID
	c.mu.Unlock()

	return result.Result[0].ID, nil
}

// GetDNSRecords lists DNS records for a given zone.
func (c *Client) GetDNSRecords(ctx context.Context, zoneID string) ([]DNSRecord, error) {
	reqURL := fmt.Sprintf("%s/zones/%s/dns_records?per_page=100", c.baseURL, zoneID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Result []DNSRecord `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Result, nil
}

// PostDNSRecord creates an 'A' DNS record if it does not already exist.
func (c *Client) PostDNSRecord(ctx context.Context, zoneID, name, content string, proxied bool) error {
	existing, err := c.GetDNSRecords(ctx, zoneID)
	if err == nil {
		for _, rec := range existing {
			if rec.Name == name {
				return nil // already exists
			}
		}
	}

	reqURL := fmt.Sprintf("%s/zones/%s/dns_records", c.baseURL, zoneID)
	bodyData := map[string]any{
		"type":    "A",
		"name":    name,
		"content": content,
		"proxied": proxied,
	}
	bodyBytes, _ := json.Marshal(bodyData)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create DNS record %s, status: %d", name, resp.StatusCode)
	}

	return nil
}

// DeleteDNSRecord deletes a DNS record by ID.
func (c *Client) DeleteDNSRecord(ctx context.Context, zoneID, recordID string) error {
	reqURL := fmt.Sprintf("%s/zones/%s/dns_records/%s", c.baseURL, zoneID, recordID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetWorkerDomains returns all worker custom domain bindings.
func (c *Client) GetWorkerDomains(ctx context.Context) ([]WorkerDomain, error) {
	if c.accountID == "" {
		return nil, nil
	}

	reqURL := fmt.Sprintf("%s/accounts/%s/workers/domains", c.baseURL, c.accountID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Result []WorkerDomain `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Result, nil
}

// RegisterWorkersSubdomains binds wildcard domains to worker service.
func (c *Client) RegisterWorkersSubdomains(ctx context.Context, servers []model.Server, wildcards []domain.Wildcard) error {
	if c.accountID == "" || c.serviceName == "" {
		return nil
	}

	existing, err := c.GetWorkerDomains(ctx)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to list existing worker domains")
	}

	existingMap := make(map[string]bool)
	for _, ed := range existing {
		existingMap[ed.Hostname] = true
	}

	for _, srv := range servers {
		for _, wc := range wildcards {
			subdomain := fmt.Sprintf("%s.%s", strings.TrimSpace(wc.Domain), strings.TrimSpace(srv.Domain))
			if existingMap[subdomain] {
				continue
			}

			payload := map[string]any{
				"zone_id":     c.zoneID,
				"service":     c.serviceName,
				"environment": c.environment,
				"hostname":    subdomain,
			}
			bodyBytes, _ := json.Marshal(payload)

			reqURL := fmt.Sprintf("%s/accounts/%s/workers/domains", c.baseURL, c.accountID)
			req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, bytes.NewBuffer(bodyBytes))
			if err != nil {
				continue
			}
			c.setHeaders(req)

			resp, err := c.httpClient.Do(req)
			if err == nil {
				_ = resp.Body.Close()
			}
		}
	}

	return nil
}
