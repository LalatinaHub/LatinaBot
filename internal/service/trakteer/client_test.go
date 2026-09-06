package trakteer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_GetDonations_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "secret-token", r.Header.Get("key"))
		assert.Equal(t, "/v1/public/supports?limit=5&include=order_id", r.URL.RequestURI())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"result": {
				"data": [
					{
						"order_id": "12345678-1234-1234-1234-123456789abc",
						"supporter_name": "Test User",
						"quantity": 2,
						"support_message": "Semangat!"
					}
				]
			}
		}`))
	}))
	defer ts.Close()

	client := NewClient("secret-token")
	client.SetBaseURL(ts.URL)

	resp, err := client.GetDonations(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Result.Data, 1)
	assert.Equal(t, "12345678-1234-1234-1234-123456789abc", resp.Result.Data[0].OrderID)
	assert.Equal(t, "Test User", resp.Result.Data[0].SupporterName)
	assert.Equal(t, 2, resp.Result.Data[0].Quantity)
}
