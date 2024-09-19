package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/BurntSushi/toml"
)

const (
	serverAddrEnvVar  = "SERVER_ADDR"
	serverPortEnvVar  = "SERVER_PORT"
	frontendURLEnvVar = "FRONTEND_URL"
	configPathEnvVar  = "GROROTI_CONFIG"
	voteStepEnvVar    = "VOTE_STEP"
	qrCodeSizeEnvVar  = "QR_CODE_SIZE"
	cleanOverTime     = "CLEAN_OVER_TIME"
	enableTracing     = "ENABLE_TRACING"
	OTLPEndpoint      = "OTLP_ENDPOINT"
)

func parse(path string) (Config, error) {
	var config Config

	// check if file exists
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		log.Warn().Msgf("Unable to parse, configuration file not found: %v. Skipping", err)
		return (Config{}), nil
	} else {
		// try to parse file
		if _, err := toml.DecodeFile(path, &config); err != nil {
			return (Config{}), err
		}
	}
	log.Info().Msgf("Configuration extracted from %s file", path)
	return config, nil
}

func LoadConfig() (config Config, err error) {
	configPathFromEnv := os.Getenv(configPathEnvVar)
	if configPathFromEnv != "" {
		config, err = parse(configPathFromEnv)
	} else {
		config, err = parse("config.toml")
	}
	if err != nil {
		return (Config{}), err
	}
	return config, err
}

func (c *Config) SetDefaults() {

	if c.ServerAddr == "" {
		c.ServerAddr = "0.0.0.0"
	}

	if c.ServerPort == 0 {
		c.ServerPort = 3000
	}

	if c.FrontendURL == "" {
		c.FrontendURL = "http://localhost:3000"
	}

	if c.VoteStep == 0.0 {
		c.VoteStep = 0.5
	}

	if c.QrCodeSize == 0 {
		c.QrCodeSize = 384
	}

	if c.CleanOverTime == 0 {
		c.CleanOverTime = 30
	}

	if c.OTLPEndpoint == "" {
		c.OTLPEndpoint = "localhost:4318"
	}
}

func (c *Config) SetConfigFromEnv() (err error) {
	serverAddrFromEnv := os.Getenv(serverAddrEnvVar)
	if serverAddrFromEnv != "" {
		c.ServerAddr = serverAddrFromEnv
	}

	portFromEnv := os.Getenv(serverPortEnvVar)
	if portFromEnv != "" {
		serverPort, err := strconv.Atoi(portFromEnv)
		if err != nil {
			err = fmt.Errorf("%w %s", ErrInvalidVar, serverPortEnvVar)
			return err
		}
		c.ServerPort = serverPort
	}

	frontendURLFromEnv := os.Getenv(frontendURLEnvVar)
	if frontendURLFromEnv != "" {
		c.FrontendURL = frontendURLFromEnv
	}

	qrCodeSizeFromEnv := os.Getenv(qrCodeSizeEnvVar)
	if qrCodeSizeFromEnv != "" {
		size, err := strconv.Atoi(qrCodeSizeFromEnv)
		if err != nil {
			err = fmt.Errorf("%w %s", ErrInvalidVar, qrCodeSizeFromEnv)
			return err
		}
		c.QrCodeSize = size
	}

	cleanOverTimeFromEnv := os.Getenv(cleanOverTime)
	if cleanOverTimeFromEnv != "" {
		cot, err := strconv.Atoi(cleanOverTimeFromEnv)
		if err != nil {
			err = fmt.Errorf("%w %s", ErrInvalidVar, cleanOverTimeFromEnv)
			return err
		}
		c.CleanOverTime = cot
	}

	enableTracingFromEnv := os.Getenv(enableTracing)
	if enableTracingFromEnv == "true" {
		c.EnableTracing = true
	}

	OTLPEndpointFromEnv := os.Getenv(OTLPEndpoint)
	if OTLPEndpointFromEnv != "" {
		c.OTLPEndpoint = OTLPEndpointFromEnv
	}

	return nil
}
