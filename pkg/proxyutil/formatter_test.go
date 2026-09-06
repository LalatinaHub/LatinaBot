package proxyutil

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/LalatinaHub/common/model"
	"github.com/stretchr/testify/assert"
)

func TestConvertProxyToURL_VMess(t *testing.T) {
	node := &model.ProxyNode{
		VPN:        "vmess",
		Remark:     "SG Node",
		Server:     "sg1.example.com",
		ServerPort: 443,
		UUID:       "12345678-1234-1234-1234-123456789abc",
		Transport:  "ws",
		Path:       "/vmess",
		Host:       "sg1.example.com",
		TLS:        true,
	}

	res := ConvertProxyToURL(node)
	assert.True(t, strings.HasPrefix(res, "vmess://"))

	encoded := strings.TrimPrefix(res, "vmess://")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	assert.NoError(t, err)

	var vmessOpts map[string]any
	err = json.Unmarshal(decoded, &vmessOpts)
	assert.NoError(t, err)
	assert.Equal(t, "SG Node", vmessOpts["ps"])
	assert.Equal(t, "sg1.example.com", vmessOpts["add"])
	assert.Equal(t, "tls", vmessOpts["tls"])
}

func TestConvertProxyToURL_VLESS(t *testing.T) {
	node := &model.ProxyNode{
		VPN:        "vless",
		Remark:     "VLESS SG",
		Server:     "sg1.example.com",
		ServerPort: 443,
		UUID:       "12345678-1234-1234-1234-123456789abc",
		Transport:  "ws",
		Path:       "/vless",
		SNI:        "sg1.example.com",
		TLS:        true,
	}

	res := ConvertProxyToURL(node)
	assert.True(t, strings.HasPrefix(res, "vless://"))
	assert.Contains(t, res, "12345678-1234-1234-1234-123456789abc@sg1.example.com:443")
	assert.Contains(t, res, "security=tls")
	assert.Contains(t, res, "#VLESS%20SG")
}

func TestConvertProxyToURL_Trojan(t *testing.T) {
	node := &model.ProxyNode{
		VPN:        "trojan",
		Remark:     "Trojan SG",
		Server:     "sg1.example.com",
		ServerPort: 443,
		Password:   "trojanpass",
		Transport:  "ws",
		Path:       "/trojan",
		SNI:        "sg1.example.com",
		TLS:        true,
	}

	res := ConvertProxyToURL(node)
	assert.True(t, strings.HasPrefix(res, "trojan://"))
	assert.Contains(t, res, "trojanpass@sg1.example.com:443")
	assert.Contains(t, res, "security=tls")
}
