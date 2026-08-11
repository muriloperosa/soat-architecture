package main

import (
	"log"

	"github.com/muriloperosa/soat-architecture/internal/infrastructure/config"
	httpinfra "github.com/muriloperosa/soat-architecture/internal/infrastructure/http"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql"
)

// @title Sistema de Oficina Mecanica API
// @version 1.0
// @description API de gestão de ordens de serviço de uma oficina mecânica.
// @BasePath /v1
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	db, err := mysql.NewConnection(cfg)
	if err != nil {
		log.Fatalf("error connecting to MySQL: %v", err)
	}

	router := httpinfra.NewRouter(db)

	log.Printf("starting server on port %s\n", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}
