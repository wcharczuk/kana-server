package config

import (
	"context"
	"encoding/base64"

	"github.com/blend/go-sdk/configmeta"
	"github.com/blend/go-sdk/configutil"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/oauth"
	"github.com/blend/go-sdk/web"
)

// Config is the application config type.
type Config struct {
	configmeta.Meta `yaml:",inline"`

	Secret string `yaml:"secret"`

	DB     db.Config     `yaml:"db"`
	OAuth  oauth.Config  `yaml:"oauth"`
	Logger logger.Config `yaml:"logger"`
	Web    web.Config    `yaml:"web"`
}

// Resolve resolves the config.
func (c *Config) Resolve(ctx context.Context) error {
	return configutil.Resolve(ctx,
		(&c.Meta).Resolve,
		(&c.DB).Resolve,
		(&c.OAuth).Resolve,
		(&c.Logger).Resolve,
		(&c.Web).Resolve,
	)
}

// GetSecret decodes the config secret.
func (c Config) GetSecret() ([]byte, error) {
	return base64.StdEncoding.DecodeString(c.Secret)
}
