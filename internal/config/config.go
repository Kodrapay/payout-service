package config

import (
	"os"
	"strings"
)

type Config struct {
	ServiceName        string
	Port               string
	PostgresDSN        string
	MerchantServiceURL string
	TransactionServiceURL string
}

func Load(serviceName, defaultPort string) Config {
	dsn := getEnv("POSTGRES_URL", "postgres://kodrapay:kodrapay_password@postgres:5432/kodrapay?sslmode=disable")
	if !strings.Contains(strings.ToLower(dsn), "sslmode=") {
		if strings.Contains(dsn, "?") {
			dsn += "&sslmode=disable"
		} else {
			dsn += "?sslmode=disable"
		}
	}

	return Config{
		ServiceName:            serviceName,
		Port:                   getEnv("PORT", defaultPort),
		PostgresDSN:            dsn,
		MerchantServiceURL:     getEnv("MERCHANT_SERVICE_URL", "http://merchant-service:7002"),
		TransactionServiceURL:  getEnv("TRANSACTION_SERVICE_URL", "http://transaction-service:7004"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
