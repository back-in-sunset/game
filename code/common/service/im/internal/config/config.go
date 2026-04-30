package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ServiceName string        `yaml:"service_name"`
	NodeID      string        `yaml:"node_id"`
	Listen      Listen        `yaml:"listen"`
	Auth        Auth          `yaml:"auth"`
	Discovery   Discovery     `yaml:"discovery"`
	Mysql       Mysql         `yaml:"mysql"`
	Redis       Redis         `yaml:"redis"`
	Scope       ScopeDefaults `yaml:"scope"`
}

type Listen struct {
	WebSocket string `yaml:"websocket"`
	TCP       string `yaml:"tcp"`
	RPC       string `yaml:"rpc"`
}

type Auth struct {
	PublicKeyFile string `yaml:"public_key_file"`
}

type Discovery struct {
	Endpoints       []string `yaml:"endpoints"`
	ServicePrefix   string   `yaml:"service_prefix"`
	LeaseTTLSeconds int64    `yaml:"lease_ttl_seconds"`
}

type Mysql struct {
	DataSource string `yaml:"data_source"`
}

type Redis struct {
	Addr      string `yaml:"addr"`
	Password  string `yaml:"password"`
	DB        int    `yaml:"db"`
	KeyPrefix string `yaml:"key_prefix"`
}

type ScopeDefaults struct {
	DefaultEnvironment string `yaml:"default_environment"`
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
	if cfg.Mysql.DataSource == "" {
		return cfg, fmt.Errorf("mysql.data_source is required")
	}
	if cfg.Redis.Addr == "" {
		return cfg, fmt.Errorf("redis.addr is required")
	}
	if cfg.Redis.KeyPrefix == "" {
		cfg.Redis.KeyPrefix = "im"
	}
	return cfg, nil
}
