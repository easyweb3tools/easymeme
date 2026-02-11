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
	RedisURL           string
	BscRpcHTTP         string
	BscRpcWS           string
	BscScanAPIKey      string
	TelegramBotToken   string
	WalletMasterKey    string
	ServiceTokens      map[string]string
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
		v.AddConfigPath(filepath.Join(".", "config"))
	}

	v.SetDefault("port", "8081")
	v.SetDefault("database_url", "")
	v.SetDefault("redis_url", "redis://localhost:6379")
	v.SetDefault("bsc_rpc_http", "https://bsc-dataseed.binance.org")
	v.SetDefault("bsc_rpc_ws", "")
	v.SetDefault("bscscan_api_key", "")
	v.SetDefault("telegram_bot_token", "")
	v.SetDefault("wallet_master_key", "")
	v.SetDefault("cors_allowed_origins", "")

	v.AutomaticEnv()
	_ = v.BindEnv("port", "PORT")
	_ = v.BindEnv("database_url", "DATABASE_URL")
	_ = v.BindEnv("redis_url", "REDIS_URL")
	_ = v.BindEnv("bsc_rpc_http", "BSC_RPC_HTTP")
	_ = v.BindEnv("bsc_rpc_ws", "BSC_RPC_WS")
	_ = v.BindEnv("bscscan_api_key", "BSCSCAN_API_KEY")
	_ = v.BindEnv("telegram_bot_token", "TELEGRAM_BOT_TOKEN")
	_ = v.BindEnv("wallet_master_key", "WALLET_MASTER_KEY")
	_ = v.BindEnv("cors_allowed_origins", "CORS_ALLOWED_ORIGINS")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("load config: %w", err)
		}
	}

	serviceTokens := make(map[string]string)
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "SERVICE_TOKEN_") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) != 2 {
				continue
			}
			serviceID := strings.ToLower(strings.TrimPrefix(parts[0], "SERVICE_TOKEN_"))
			serviceTokens[serviceID] = strings.TrimSpace(parts[1])
		}
	}

	return &Config{
		Port:               v.GetString("port"),
		DatabaseURL:        v.GetString("database_url"),
		RedisURL:           v.GetString("redis_url"),
		BscRpcHTTP:         v.GetString("bsc_rpc_http"),
		BscRpcWS:           v.GetString("bsc_rpc_ws"),
		BscScanAPIKey:      v.GetString("bscscan_api_key"),
		TelegramBotToken:   v.GetString("telegram_bot_token"),
		WalletMasterKey:    v.GetString("wallet_master_key"),
		ServiceTokens:      serviceTokens,
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
