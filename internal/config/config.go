package config

import (
	"github.com/gin-gonic/gin"
	"os"
)

type Config struct{ Addr string }

func Must() Config {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		mode = gin.DebugMode
	}
	gin.SetMode(mode)
	return Config{Addr: addr}
}
