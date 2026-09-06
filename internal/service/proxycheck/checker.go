package proxycheck

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

const (
	ipResolverDomain = "myip.shylook.workers.dev"
	ipResolverPath   = "/"
)

// CheckResult represents the outcome of checking a proxy IP.
type CheckResult struct {
	Proxy   string `json:"proxy,omitempty"`
	Port    string `json:"port,omitempty"`
	ProxyIP bool   `json:"proxyip"`
	Delay   int64  `json:"delay,omitempty"`
	IP      string `json:"ip,omitempty"`
	Message string `json:"message,omitempty"`
	Raw     map[string]any `json:"-"`
}

func (r CheckResult) MarshalJSON() ([]byte, error) {
	m := make(map[string]any)
	for k, v := range r.Raw {
		m[k] = v
	}
	m["proxyip"] = r.ProxyIP
	if r.Proxy != "" {
		m["proxy"] = r.Proxy
	}
	if r.Port != "" {
		m["port"] = r.Port
	}
	if r.Delay > 0 {
		m["delay"] = r.Delay
	}
	if r.Message != "" {
		m["message"] = r.Message
	}
	return json.Marshal(m)
}

// Checker tests proxy IP against Cloudflare worker resolver.
type Checker struct {
	resolverDomain string
	timeout        time.Duration
}

// NewChecker creates a new Checker instance.
func NewChecker() *Checker {
	return &Checker{
		resolverDomain: ipResolverDomain,
		timeout:        5 * time.Second,
	}
}

// CheckIP verifies if proxyIP (host:port) acts as a valid proxy.
func (c *Checker) CheckIP(ctx context.Context, proxyIP string) CheckResult {
	if proxyIP == "" {
		return CheckResult{ProxyIP: false}
	}

	parts := strings.Split(proxyIP, ":")
	host := parts[0]
	port := "443"
	if len(parts) > 1 {
		port = parts[1]
	}

	start := time.Now()

	// 1. Fetch via proxy target
	proxyResp, err := c.sendRequest(net.JoinHostPort(host, port), c.resolverDomain)
	if err != nil {
		return CheckResult{ProxyIP: false, Message: err.Error()}
	}

	// 2. Fetch directly
	directResp, err := c.sendRequest(net.JoinHostPort(c.resolverDomain, "443"), c.resolverDomain)
	if err != nil {
		return CheckResult{ProxyIP: false, Message: err.Error()}
	}

	delay := time.Since(start).Milliseconds()

	var ipInfo map[string]any
	if err := json.Unmarshal([]byte(proxyResp), &ipInfo); err != nil {
		return CheckResult{ProxyIP: false, Message: "Invalid proxy response JSON"}
	}

	var myIPInfo map[string]any
	if err := json.Unmarshal([]byte(directResp), &myIPInfo); err != nil {
		return CheckResult{ProxyIP: false, Message: "Invalid direct response JSON"}
	}

	proxyIPStr, _ := ipInfo["ip"].(string)
	myIPStr, _ := myIPInfo["ip"].(string)

	if proxyIPStr != "" && proxyIPStr != myIPStr {
		return CheckResult{
			Proxy:   host,
			Port:    port,
			ProxyIP: true,
			Delay:   delay,
			IP:      proxyIPStr,
			Raw:     ipInfo,
		}
	}

	return CheckResult{ProxyIP: false}
}

func (c *Checker) sendRequest(targetAddr, sniHost string) (string, error) {
	dialer := &net.Dialer{Timeout: c.timeout}
	tlsConfig := &tls.Config{
		ServerName:         sniHost,
		InsecureSkipVerify: true,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", targetAddr, tlsConfig)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(c.timeout))

	req := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nUser-Agent: Mozilla/5.0\r\nConnection: close\r\n\r\n", ipResolverPath, sniHost)
	if _, err := conn.Write([]byte(req)); err != nil {
		return "", err
	}

	data, err := io.ReadAll(conn)
	if err != nil && len(data) == 0 {
		return "", err
	}

	raw := string(data)
	idx := strings.Index(raw, "\r\n\r\n")
	if idx != -1 {
		return raw[idx+4:], nil
	}
	return raw, nil
}
