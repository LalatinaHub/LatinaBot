package payment

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// PaymentResult contains the result of a QR payment creation.
type PaymentResult struct {
	Error    bool   `json:"error"`
	Message  string `json:"message"` // QR string or error message
	QRString string `json:"qr_string,omitempty"`
}

// CheckResult contains the result of a payment status verification.
type CheckResult struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

// Service manages Xendit QR payments.
type Service struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewService creates a new Payment service instance.
func NewService(apiKey string) *Service {
	return &Service{
		apiKey:  apiKey,
		baseURL: "https://api.xendit.co/qr_codes",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SetBaseURL allows overriding the base URL for testing.
func (s *Service) SetBaseURL(url string) {
	s.baseURL = url
}

// MakePayment creates a dynamic QR code payment for the given amount (in IDR).
func (s *Service) MakePayment(ctx context.Context, amount int) PaymentResult {
	if s.apiKey == "" {
		return PaymentResult{
			Error:   true,
			Message: "Payment gateway API key not configured",
		}
	}

	payload := map[string]any{
		"reference_id": fmt.Sprintf("order-id-%d", time.Now().UnixMilli()),
		"type":         "DYNAMIC",
		"currency":     "IDR",
		"amount":       amount,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return PaymentResult{Error: true, Message: err.Error()}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return PaymentResult{Error: true, Message: err.Error()}
	}

	s.setHeaders(req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return PaymentResult{Error: true, Message: err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		var resData struct {
			QRString string `json:"qr_string"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&resData); err == nil {
			return PaymentResult{
				Error:    false,
				Message:  resData.QRString,
				QRString: resData.QRString,
			}
		}
	}

	return PaymentResult{
		Error:   true,
		Message: fmt.Sprintf("Failed with status %d", resp.StatusCode),
	}
}

// CheckPayment checks whether a payment has succeeded.
func (s *Service) CheckPayment(ctx context.Context, token string) CheckResult {
	if s.apiKey == "" {
		return CheckResult{Error: true, Message: "Payment gateway API key not configured"}
	}

	checkURL := fmt.Sprintf("%s/%s", s.baseURL, token)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, checkURL, nil)
	if err != nil {
		return CheckResult{Error: true, Message: err.Error()}
	}

	s.setHeaders(req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return CheckResult{Error: true, Message: err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var resData struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&resData); err == nil && resData.Status == "SUCCEEDED" {
			return CheckResult{Error: false, Message: "success"}
		}
	}

	return CheckResult{Error: true, Message: "Payment not completed or failed"}
}

func (s *Service) setHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-version", "2022-07-31")
	auth := base64.StdEncoding.EncodeToString([]byte(s.apiKey + ":"))
	req.Header.Set("Authorization", "Basic "+auth)
}
