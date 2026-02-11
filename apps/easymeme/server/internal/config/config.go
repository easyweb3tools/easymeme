package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port               string
	DatabaseURL        string
	BscRpcHTTP         string
	BscRpcWS           string
	BscScanAPIKey      string
	BaseServiceURL     string
	BaseServiceToken   string
	ApiKey             string
	ApiUserID          string
	ApiHmacSecret      string
	CorsAllowedOrigins []string
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
	v.SetDefault("base_service_url", "http://localhost:8081")
	v.SetDefault("base_service_token", "")
	v.SetDefault("api_key", "")
	v.SetDefault("api_user_id", "")
	v.SetDefault("api_hmac_secret", "")
	v.SetDefault("cors_allowed_origins", "http://localhost:3000")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	_ = v.BindEnv("port", "port", "PORT")
	_ = v.BindEnv("database_url", "database_url", "DATABASE_URL")
	_ = v.BindEnv("bsc_rpc_http", "bsc_rpc_http", "BSC_RPC_HTTP")
	_ = v.BindEnv("bsc_rpc_ws", "bsc_rpc_ws", "BSC_RPC_WS")
	_ = v.BindEnv("bscscan_api_key", "bscscan_api_key", "BSCSCAN_API_KEY")
	_ = v.BindEnv("base_service_url", "base_service_url", "BASE_SERVICE_URL")
	_ = v.BindEnv("base_service_token", "base_service_token", "BASE_SERVICE_TOKEN")
	_ = v.BindEnv("api_key", "api_key", "EASYMEME_API_KEY", "API_KEY")
	_ = v.BindEnv("api_user_id", "api_user_id", "EASYMEME_USER_ID")
	_ = v.BindEnv("api_hmac_secret", "api_hmac_secret", "EASYMEME_API_HMAC_SECRET")
	_ = v.BindEnv("cors_allowed_origins", "cors_allowed_origins", "CORS_ALLOWED_ORIGINS")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("load config: %w", err)
		}
	}

	return &Config{
		Port:               v.GetString("port"),
		DatabaseURL:        v.GetString("database_url"),
		BscRpcHTTP:         v.GetString("bsc_rpc_http"),
		BscRpcWS:           v.GetString("bsc_rpc_ws"),
		BscScanAPIKey:      v.GetString("bscscan_api_key"),
		BaseServiceURL:     v.GetString("base_service_url"),
		BaseServiceToken:   v.GetString("base_service_token"),
		ApiKey:             v.GetString("api_key"),
		ApiUserID:          v.GetString("api_user_id"),
		ApiHmacSecret:      v.GetString("api_hmac_secret"),
		CorsAllowedOrigins: splitOrigins(v.GetString("cors_allowed_origins")),
	}, nil
}

func splitOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}
