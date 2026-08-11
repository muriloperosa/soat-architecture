package wiring

import (
	"gorm.io/gorm"

	"github.com/muriloperosa/soat-architecture/internal/infrastructure/config"
)

// Container compõe as dependências compartilhadas da aplicação
// (config, conexão de banco e, futuramente, repositórios e use cases por domínio).
type Container struct {
	Config *config.Config
	DB     *gorm.DB
}

// NewContainer monta o grafo de dependências da aplicação.
func NewContainer(cfg *config.Config, db *gorm.DB) *Container {
	return &Container{
		Config: cfg,
		DB:     db,
	}
}
