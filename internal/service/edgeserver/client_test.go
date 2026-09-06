package edgeserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEdgeServer_GetInfo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/info", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"org": "Cloudflare, Inc."}`))
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	client := NewClient()

	info, err := client.GetInfo(context.Background(), u.Host)
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "Cloudflare, Inc.", info.Org)
}

func TestEdgeServer_GetStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/status", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"cpu": [12.5],
			"host": {"hostname": "srv-1", "uptime": 3600, "platform": "linux", "platformVersion": "Ubuntu 22.04"},
			"mem": {"total": 1000, "used": 500, "usedPercent": 50.0},
			"disk": {"total": 10000, "used": 2000, "usedPercent": 20.0},
			"nic": [{"name": "eth0", "bytesSent": 1024, "bytesRecv": 2048}]
		}`))
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	client := NewClient()

	status, err := client.GetStatus(context.Background(), u.Host)
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "Ubuntu 22.04", status.Host.PlatformVersion)
	assert.Equal(t, 50.0, status.Mem.UsedPercent)
}
