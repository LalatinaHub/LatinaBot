package payment

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPayment_MakePayment_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"qr_string": "00020101021226540014ID.LINKAJA.WWW01189360001100000000005204581253033605802ID5911Fool VPN6007JAKARTA61051234062070703A0163041234"}`))
	}))
	defer ts.Close()

	svc := NewService("dummy-key")
	svc.SetBaseURL(ts.URL)

	res := svc.MakePayment(context.Background(), 5000)
	assert.False(t, res.Error)
	assert.NotEmpty(t, res.QRString)
}

func TestPayment_CheckPayment_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/test-token-123", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "SUCCEEDED"}`))
	}))
	defer ts.Close()

	svc := NewService("dummy-key")
	svc.SetBaseURL(ts.URL)

	res := svc.CheckPayment(context.Background(), "test-token-123")
	assert.False(t, res.Error)
	assert.Equal(t, "success", res.Message)
}
