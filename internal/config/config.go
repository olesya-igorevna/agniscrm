package config

import (
        "fmt"
        "os"
        "time"
)

type Config struct {
        Addr        string
        DatabaseURL string
        DBTimeout   time.Duration
}


func Load() (Config, error) {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
			addr = ":8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
			return Config{}, fmt.Errorf("DATABASE_URL не задана")
	}

	dbTimeout := 3 * time.Second
	if v := os.Getenv("DB_TIMEOUT"); v != "" {
			d, err := time.ParseDuration(v)
			if err != nil {
					return Config{}, fmt.Errorf("DB_TIMEOUT некорректен: %w", err)
			}
			dbTimeout = d
	}

	return Config{
			Addr:        addr,
			DatabaseURL: databaseURL,
			DBTimeout:   dbTimeout,
	}, nil
}
