package config

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidVar = errors.New("invalid var")
)

type Config struct {
	ServerAddr    string  `toml:"server_addr"`
	ServerPort    int     `toml:"server_port"`
	FrontendURL   string  `toml:"frontend_url"`
	VoteStep      float64 `toml:"vote_step"`
	QrCodeSize    int     `toml:"qr_code_size"`
	CleanOverTime int     `toml:"clean_over_time"`
	EnableTracing bool    `toml:"enable_tracing"`
	OTLPEndpoint  string  `toml:"otlp_endpoint"`
}

func NewConfig(config Config) *Config {
	return &config
}

func (c *Config) BuildServerAddr() string {
	return fmt.Sprintf("%s:%d", c.ServerAddr, c.ServerPort)
}

func (c *Config) GetURL() string {
	return c.FrontendURL
}

func (c *Config) GetQrCodeSize() int {
	return c.QrCodeSize
}
