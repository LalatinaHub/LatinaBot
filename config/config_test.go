package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_Defaults(t *testing.T) {
	_ = os.Unsetenv("BOT_TOKEN")
	_ = os.Unsetenv("APP_ENV")
	_ = os.Unsetenv("PORT")

	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "development", cfg.AppEnv)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, ":8080", cfg.HTTPAddress())
	assert.False(t, cfg.IsProduction())
	assert.Error(t, cfg.Validate()) // missing bot token
}

func TestLoadConfig_WithValues(t *testing.T) {
	t.Setenv("BOT_TOKEN", "123456:ABC-DEF")
	t.Setenv("ADMIN_ID", "12345678")
	t.Setenv("GROUP_ID", "-10012345678")
	t.Setenv("PORT", "9090")
	t.Setenv("APP_ENV", "production")

	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "123456:ABC-DEF", cfg.BotToken)
	assert.Equal(t, int64(12345678), cfg.AdminID)
	assert.Equal(t, int64(-10012345678), cfg.GroupID)
	assert.Equal(t, "9090", cfg.Port)
	assert.Equal(t, ":9090", cfg.HTTPAddress())
	assert.True(t, cfg.IsProduction())
	assert.NoError(t, cfg.Validate())
}
