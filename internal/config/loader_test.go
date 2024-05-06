package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {

	// When there isn't config file, it should return an empty config
	config, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if config.ServerAddr != "" {
		t.Error("Expected an error here")
	}

	// When file isn't a toml, it should raise an error
	b1 := []byte("{'test': 'tests'}")
	err = os.WriteFile("config.json", b1, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("config.json")

	err = os.Setenv(configPathEnvVar, "config.json")
	if err != nil {
		t.Fatal(err)
	}

	_, err = LoadConfig()
	if err == nil {
		t.Error("Expected an error here")
	}

	err = os.Unsetenv(configPathEnvVar)
	if err != nil {
		t.Fatal(err)
	}

	// When file is good, it sould not raise any error
	b2 := []byte("server_addr = '0.0.0.0'")
	err = os.WriteFile("config.toml", b2, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("config.toml")
	config, err = LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if config.ServerAddr != "0.0.0.0" {
		t.Errorf("Expected %s, got %s", "0.0.0.0", config.ServerAddr)
	}
}

func TestSetDefaults(t *testing.T) {
	c, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}

	c.SetDefaults()

	if c.ServerAddr != "0.0.0.0" {
		t.Errorf("Expected %s, got %s", "0.0.0.0", c.ServerAddr)
	}
	if c.ServerPort != 3000 {
		t.Errorf("Expected %d, got %d", 3000, c.ServerPort)
	}
	if c.FrontendURL != "http://localhost:3000" {
		t.Errorf("Expected %s, got %s", "http://localhost:3000", c.FrontendURL)
	}
	if c.VoteStep != 0.5 {
		t.Errorf("Expected %f, got %f", 0.5, c.VoteStep)
	}
	if c.QrCodeSize != 384 {
		t.Errorf("Expected %d, got %d", 384, c.QrCodeSize)
	}
}

func TestSetConfigFromEnv(t *testing.T) {
	c, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}

	_ = os.Setenv(serverAddrEnvVar, "0.0.0.0")
	_ = os.Setenv(serverPortEnvVar, "3000")
	_ = os.Setenv(frontendURLEnvVar, "https://groroti.domain.tld")
	_ = os.Setenv(voteStepEnvVar, "0.5")
	err = os.Setenv(qrCodeSizeEnvVar, "512")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Unsetenv(serverAddrEnvVar)
		_ = os.Unsetenv(serverPortEnvVar)
		_ = os.Unsetenv(frontendURLEnvVar)
		_ = os.Unsetenv(voteStepEnvVar)
		_ = os.Unsetenv(qrCodeSizeEnvVar)
	}()

	c.SetConfigFromEnv()

	if c.ServerAddr != "0.0.0.0" {
		t.Errorf("Expected %s, got %s", "0.0.0.0", c.ServerAddr)
	}

}
