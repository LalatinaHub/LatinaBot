package domain

import (
	"time"

	"github.com/LalatinaHub/common/model"
)

// Re-export common models for ease of use
type User = model.User
type Server = model.Server
type ProxyNode = model.ProxyNode
type KeyValue = model.KeyValue

// Wildcard represents a registered wildcard domain for CDN workers.
type Wildcard struct {
	ID     int64  `json:"id"`
	Domain string `json:"domain"`
}

// Donation represents a processed donation order ID.
type Donation struct {
	ID        int64     `json:"id"`
	OrderID   string    `json:"order_id"`
	CreatedAt time.Time `json:"created_at"`
}

// EdgeServerInfo represents ping and organization info returned by edge server /api/v1/info.
type EdgeServerInfo struct {
	Org  string `json:"org"`
	Ping int64  `json:"ping"` // in milliseconds
}

// ServerStatus represents hardware & system metrics from edge server /api/v1/status.
type ServerStatus struct {
	CPU  []float64 `json:"cpu"`
	Host HostInfo  `json:"host"`
	Mem  MemInfo   `json:"mem"`
	Disk DiskInfo  `json:"disk"`
	NIC  []NICInfo `json:"nic"`
}

type HostInfo struct {
	Hostname        string `json:"hostname"`
	Uptime          uint64 `json:"uptime"`
	OS              string `json:"os"`
	Platform        string `json:"platform"`
	PlatformVersion string `json:"platformVersion"`
}

type MemInfo struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
}

type DiskInfo struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
}

type NICInfo struct {
	Name      string `json:"name"`
	BytesSent uint64 `json:"bytesSent"`
	BytesRecv uint64 `json:"bytesRecv"`
}
