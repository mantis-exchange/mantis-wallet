package config

import "os"

type Config struct {
	Port               string
	DBURL              string
	ETHNode            string
	BTCNode            string
	AccountServiceAddr string
}

func Load() *Config {
	return &Config{
		Port:               getEnv("PORT", "50054"),
		DBURL:              getEnv("DB_URL", "postgres://mantis:mantis@localhost:5432/mantis?sslmode=disable"),
		ETHNode:            getEnv("ETH_NODE", "http://localhost:8545"),
		BTCNode:            getEnv("BTC_NODE", "http://localhost:18332"),
		AccountServiceAddr: getEnv("ACCOUNT_SERVICE_ADDR", "http://localhost:50053"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
