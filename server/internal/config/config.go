package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port        string
	DatabaseURL string
	BscRpcHTTP  string
	BscRpcWS    string
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigType("toml")
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("config")
		v.AddConfigPath(".")
		v.AddConfigPath("./server")
		v.AddConfigPath(filepath.Join(".", "config"))
	}

	v.SetDefault("port", "8080")
	v.SetDefault("database_url", "")
	v.SetDefault("bsc_rpc_http", "https://bsc-dataseed.binance.org")
	v.SetDefault("bsc_rpc_ws", "")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("load config: %w", err)
		}
	}

	return &Config{
		Port:        v.GetString("port"),
		DatabaseURL: v.GetString("database_url"),
		BscRpcHTTP:  v.GetString("bsc_rpc_http"),
		BscRpcWS:    v.GetString("bsc_rpc_ws"),
	}, nil
}
