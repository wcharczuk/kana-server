package main

import (
	"context"

	"github.com/blend/go-sdk/configutil"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/web"
)

// Config is the application config type.
type Config struct {
	Logger logger.Config `yaml:"logger"`
	Web    web.Config    `yaml:"web"`
}

// Resolve resolves the config.
func (c *Config) Resolve(ctx context.Context) error {
	return configutil.Resolve(ctx,
		(&c.Logger).Resolve,
		(&c.Web).Resolve,
	)
}

func main() {
	var cfg Config
	configutil.MustRead(&cfg)
}
