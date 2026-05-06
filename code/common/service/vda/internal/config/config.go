package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ServiceName string    `yaml:"service_name"`
	NodeID      string    `yaml:"node_id"`
	Listen      Listen    `yaml:"listen"`
	Redis       Redis     `yaml:"redis"`
	LiveKit     LiveKit   `yaml:"livekit"`
	IM          IM        `yaml:"im"`
	Discovery   Discovery `yaml:"discovery"`
	Log         Log       `yaml:"log"`
}

type Listen struct {
	RPC string `yaml:"rpc"`
}

type Redis struct {
	Addr      string `yaml:"addr"`
	Password  string `yaml:"password"`
	DB        int    `yaml:"db"`
	KeyPrefix string `yaml:"key_prefix"`
}

type LiveKit struct {
	Host      string `yaml:"host"`
	APIKey    string `yaml:"api_key"`
	APISecret string `yaml:"api_secret"`
}

type IM struct {
	Endpoint string `yaml:"endpoint"`
}

type Discovery struct {
	Endpoints       []string `yaml:"endpoints"`
	ServicePrefix   string   `yaml:"service_prefix"`
	LeaseTTLSeconds int64    `yaml:"lease_ttl_seconds"`
}

type Log struct {
	Level string `yaml:"level"`
}

func Load(path string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	if cfg.ServiceName == "" {
		return cfg, fmt.Errorf("service_name is required")
	}
	if cfg.NodeID == "" {
		return cfg, fmt.Errorf("node_id is required")
	}
	if cfg.Redis.Addr == "" {
		return cfg, fmt.Errorf("redis.addr is required")
	}
	if cfg.Redis.KeyPrefix == "" {
		cfg.Redis.KeyPrefix = "vda"
	}
	if cfg.LiveKit.Host == "" {
		return cfg, fmt.Errorf("livekit.host is required")
	}
	return cfg, nil
}
