package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/LalatinaHub/common/model"
)

type proxyRepo struct {
	db *sql.DB
}

// NewProxyRepository creates a new ProxyRepository instance.
func NewProxyRepository(db *sql.DB) ProxyRepository {
	return &proxyRepo{db: db}
}

func (r *proxyRepo) GetRandomCDNProxy(ctx context.Context) (*model.ProxyNode, error) {
	query := `SELECT id, server, ip, server_port, uuid, password, security, alter_id, method, 
	                 plugin, plugin_opts, host, tls, transport, path, service_name, insecure, 
	                 sni, remark, conn_mode, country_code, region, org, vpn, raw 
	          FROM proxies 
	          WHERE vpn != 'shadowsocks' AND conn_mode = 'cdn' 
	          ORDER BY RANDOM() LIMIT 1;`

	row := r.db.QueryRowContext(ctx, query)
	var p model.ProxyNode
	err := row.Scan(
		&p.ID,
		&p.Server,
		&p.IP,
		&p.ServerPort,
		&p.UUID,
		&p.Password,
		&p.Security,
		&p.AlterID,
		&p.Method,
		&p.Plugin,
		&p.PluginOpts,
		&p.Host,
		&p.TLS,
		&p.Transport,
		&p.Path,
		&p.ServiceName,
		&p.Insecure,
		&p.SNI,
		&p.Remark,
		&p.ConnMode,
		&p.CountryCode,
		&p.Region,
		&p.Org,
		&p.VPN,
		&p.Raw,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get random CDN proxy: %w", err)
	}

	return &p, nil
}
