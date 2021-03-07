package config

import (
	"context"
	"encoding/hex"

	"github.com/blend/go-sdk/configmeta"
	"github.com/blend/go-sdk/configutil"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/ex"
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
		configutil.SetString(&c.Secret, configutil.Env("SECRET"), configutil.String(c.Secret)),
		configutil.SetString(&c.ServiceName, configutil.Env("SERVICE_NAME"), configutil.Env("HEROKU_APP_NAME"), configutil.String(c.ServiceName), configutil.String("kana-server")),
		configutil.SetString(&c.Version, configutil.Env("VERSION"), configutil.Env("HEROKU_RELEASE_VERSION"), configutil.String(c.Version), configutil.String(configmeta.Version)),
		configutil.SetString(&c.GitRef, configutil.Env("GIT_REF"), configutil.Env("HEROKU_SLUG_COMMIT"), configutil.String(c.GitRef), configutil.String(configmeta.GitRef)),
	)
}

// GetSecret decodes the config secret.
func (c Config) GetSecret() ([]byte, error) {
	secret, err := hex.DecodeString(c.Secret)
	if err != nil {
		return nil, ex.New(err)
	}
	return secret, nil
}
