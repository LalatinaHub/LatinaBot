package trakteer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SupportItem represents an individual donation support record from Trakteer.
type SupportItem struct {
	OrderID        string `json:"order_id"`
	SupporterName  string `json:"supporter_name"`
	Quantity       int    `json:"quantity"`
	SupportMessage string `json:"support_message"`
}

// Response represents the Trakteer API response wrapper.
type Response struct {
	Result struct {
		Data []SupportItem `json:"data"`
	} `json:"result"`
}

// Client interacts with the Trakteer public API.
type Client struct {
	token      string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Trakteer API client.
func NewClient(token string) *Client {
	return &Client{
		token:   token,
		baseURL: "https://api.trakteer.id",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SetBaseURL allows overriding the base URL (useful for testing).
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

// GetDonations retrieves the latest 5 public donations with order IDs.
func (c *Client) GetDonations(ctx context.Context) (*Response, error) {
	reqURL := fmt.Sprintf("%s/v1/public/supports?limit=5&include=order_id", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create trakteer request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	if c.token != "" {
		req.Header.Set("key", c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute trakteer request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("trakteer returned non-200 status: %d", resp.StatusCode)
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode trakteer response: %w", err)
	}

	return &result, nil
}
