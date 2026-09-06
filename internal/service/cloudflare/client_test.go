package cloudflare

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloudflare_GetDNSRecords(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/zones/test-zone/dns_records", r.URL.Path)
		assert.Equal(t, "test@example.com", r.Header.Get("X-Auth-Email"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result": [{"id": "rec-1", "name": "vpn.example.com", "type": "A", "content": "1.2.3.4"}]}`))
	}))
	defer ts.Close()

	client := NewClient("test@example.com", "api-key", "acc-id", "test-zone", "svc", "env")
	client.SetBaseURL(ts.URL)

	records, err := client.GetDNSRecords(context.Background(), "test-zone")
	assert.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, "rec-1", records[0].ID)
	assert.Equal(t, "vpn.example.com", records[0].Name)
}
