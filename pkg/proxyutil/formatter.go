package proxyutil

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/LalatinaHub/common/model"
)

// ConvertProxyToURL serializes a model.ProxyNode into a standard proxy URI string (vmess, vless, trojan).
func ConvertProxyToURL(proxy *model.ProxyNode) string {
	if proxy == nil {
		return ""
	}

	vpn := strings.ToLower(strings.TrimSpace(proxy.VPN))

	switch vpn {
	case "vmess":
		tlsStr := ""
		if proxy.TLS {
			tlsStr = "tls"
		}

		vmessOpts := map[string]any{
			"v":    2,
			"ps":   proxy.Remark,
			"add":  proxy.Server,
			"port": proxy.ServerPort,
			"id":   proxy.UUID,
			"aid":  0,
			"net":  proxy.Transport,
			"path": proxy.Path,
			"type": "none",
			"host": proxy.Host,
			"tls":  tlsStr,
		}

		jsonBytes, err := json.Marshal(vmessOpts)
		if err != nil {
			return ""
		}
		return fmt.Sprintf("vmess://%s", base64.StdEncoding.EncodeToString(jsonBytes))

	case "trojan", "vless":
		auth := proxy.Password
		if auth == "" {
			auth = proxy.UUID
		}

		u := url.URL{
			Scheme: vpn,
			User:   url.User(auth),
			Host:   fmt.Sprintf("%s:%d", proxy.Server, proxy.ServerPort),
		}

		q := u.Query()
		if proxy.Path != "" {
			q.Set("path", proxy.Path)
		}
		if proxy.TLS {
			q.Set("security", "tls")
		} else {
			q.Set("security", "")
		}
		if proxy.Transport != "" {
			q.Set("type", proxy.Transport)
		}
		if proxy.SNI != "" {
			q.Set("sni", proxy.SNI)
		}
		u.RawQuery = q.Encode()

		if proxy.Remark != "" {
			u.Fragment = proxy.Remark
		}

		return u.String()

	case "shadowsocks":
		if proxy.Raw != "" {
			return proxy.Raw
		}
		auth := fmt.Sprintf("%s:%s", proxy.Method, proxy.Password)
		encodedAuth := base64.URLEncoding.EncodeToString([]byte(auth))
		tag := ""
		if proxy.Remark != "" {
			tag = "#" + url.QueryEscape(proxy.Remark)
		}
		return fmt.Sprintf("ss://%s@%s:%d%s", encodedAuth, proxy.Server, proxy.ServerPort, tag)

	default:
		if proxy.Raw != "" {
			return proxy.Raw
		}
		return ""
	}
}
