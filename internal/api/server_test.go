package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LalatinaHub/LatinaBot/internal/service/proxycheck"
	"github.com/stretchr/testify/assert"
)

func TestHTTPServer_Root(t *testing.T) {
	srv := NewServer(":8080", proxycheck.NewChecker())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Welcome to LatinaBot")
}

func TestHTTPServer_ProxyCheck_EmptyIP(t *testing.T) {
	srv := NewServer(":8080", proxycheck.NewChecker())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/proxy/check", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "{}", w.Body.String())
}
