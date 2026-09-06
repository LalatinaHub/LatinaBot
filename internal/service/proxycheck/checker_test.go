package proxycheck

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChecker_EmptyIP(t *testing.T) {
	c := NewChecker()
	res := c.CheckIP(context.Background(), "")
	assert.False(t, res.ProxyIP)
}

func TestChecker_UnreachableIP(t *testing.T) {
	c := NewChecker()
	res := c.CheckIP(context.Background(), "127.0.0.1:54321")
	assert.False(t, res.ProxyIP)
	assert.NotEmpty(t, res.Message)
}
