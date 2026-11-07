package main

import (
	"agniscrm/internal/config"
	"agniscrm/internal/handler"
	"agniscrm/internal/storage/memory"
	"log"
)

func main() {
	cfg := config.Must()
	h := handler.New(memory.NewWithSeed())
	if err := h.Run(cfg.Addr); err != nil {
		log.Fatal(err)
	}
}
