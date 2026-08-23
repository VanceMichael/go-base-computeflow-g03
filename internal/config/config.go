package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port          int
	DatabasePath  string
	BusinessZone  *time.Location
	SessionTTL    time.Duration
	ShutdownGrace time.Duration
}

func Load() (Config, error) {
	port := 8080
	if v := os.Getenv("PORT"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, err
		}
		port = parsed
	}
	path := os.Getenv("DATABASE_PATH")
	if path == "" {
		path = "./computeflow.db"
	}
	zoneName := os.Getenv("BUSINESS_TIMEZONE")
	if zoneName == "" {
		zoneName = "Asia/Shanghai"
	}
	zone, err := time.LoadLocation(zoneName)
	if err != nil {
		return Config{}, err
	}
	ttl := 8 * time.Hour
	if v := os.Getenv("SESSION_TTL"); v != "" {
		ttl, err = time.ParseDuration(v)
		if err != nil {
			return Config{}, err
		}
	}
	return Config{Port: port, DatabasePath: path, BusinessZone: zone, SessionTTL: ttl, ShutdownGrace: 10 * time.Second}, nil
}
