package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"agnis-crm/internal/config"
	"agnis-crm/internal/handler"
	"agnis-crm/internal/storage"
	"agnis-crm/internal/usecase"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()

	// Слои снизу вверх: storage → usecase → handler.
	st := storage.NewWithDB(db, cfg.DBTimeout)
	uc := usecase.New(st)
	h := handler.New(uc)

	// gin.New() без встроенных Logger/Recovery: логирование и восстановление
	router := gin.New()
	h.Register(router)

	log.Printf("serving HTTP on %s", cfg.Addr)
	if err := router.Run(cfg.Addr); err != nil {
		log.Fatal(err)
	}
}
