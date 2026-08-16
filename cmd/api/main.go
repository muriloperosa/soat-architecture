package main

import (
	"log"

	"github.com/muriloperosa/soat-architecture/internal/infrastructure/config"
	httpinfra "github.com/muriloperosa/soat-architecture/internal/infrastructure/http"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
)

// @title Sistema de Oficina Mecânica API
// @version 1.0
// @description API de gestão de ordens de serviço de uma oficina mecânica. Tech Challenge Fase 1 - Soat.
// @BasePath /
// @Host http://localhost:8080
// @Schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	db, err := mysql.NewConnection(cfg)
	if err != nil {
		log.Fatalf("error connecting to MySQL: %v", err)
	}

	container := wiring.NewContainer(cfg, db)
	router := httpinfra.NewRouter(container)

	log.Printf("starting server on port %s\n", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}
