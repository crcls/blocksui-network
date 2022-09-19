package config

import (
	"os"
)

type Config struct {
	ChainName      string
	ContractsCID   string
	Env            string
	HomeDir        string
	NetworkName    string
	Port           string
	PrimitivesCID  string
	ProviderURL    string
	RecoveryPhrase string
	Web3Token      string
}

func (c *Config) WithPort(port string) *Config {
	c.Port = port
	return c
}

func New(env string) *Config {
	hd, err := os.UserHomeDir()
	if err != nil {
		hd = "/"
	}

	return &Config{
		ChainName:      os.Getenv("CHAIN_NAME"),
		ContractsCID:   os.Getenv("CONTRACTS_CID"),
		Env:            env,
		HomeDir:        hd,
		NetworkName:    os.Getenv("NETWORK_NAME"),
		PrimitivesCID:  os.Getenv("PRIMITIVES_CID"),
		ProviderURL:    os.Getenv("PROVIDER_URL"),
		RecoveryPhrase: os.Getenv("RECOVERY_PHRASE"),
		Web3Token:      os.Getenv("WEB3STORAGE_TOKEN"),
	}
}
