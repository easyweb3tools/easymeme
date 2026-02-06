package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port          string
	DatabaseURL   string
	BscRpcHTTP    string
	BscRpcWS      string
	BscScanAPIKey string
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
	v.SetDefault("bscscan_api_key", "")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	_ = v.BindEnv("port", "port", "PORT")
	_ = v.BindEnv("database_url", "database_url", "DATABASE_URL")
	_ = v.BindEnv("bsc_rpc_http", "bsc_rpc_http", "BSC_RPC_HTTP")
	_ = v.BindEnv("bsc_rpc_ws", "bsc_rpc_ws", "BSC_RPC_WS")
	_ = v.BindEnv("bscscan_api_key", "bscscan_api_key", "BSCSCAN_API_KEY")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("load config: %w", err)
		}
	}

	return &Config{
		Port:          v.GetString("port"),
		DatabaseURL:   v.GetString("database_url"),
		BscRpcHTTP:    v.GetString("bsc_rpc_http"),
		BscRpcWS:      v.GetString("bsc_rpc_ws"),
		BscScanAPIKey: v.GetString("bscscan_api_key"),
	}, nil
}
