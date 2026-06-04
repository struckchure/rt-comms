package internal

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Port        string `mapstructure:"PORT"`
	DatabaseURL string `mapstructure:"DATABASE_URL"`
	STUNURL     string `mapstructure:"STUN_URL"`
}

func LoadConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.SetDefault("PORT", ":8080")
	v.SetDefault("STUN_URL", "stun:stun.l.google.com:19302")
	v.AutomaticEnv()
	_ = v.ReadInConfig()
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.Port == "" {
		c.Port = ":8080"
	}
	return nil
}
