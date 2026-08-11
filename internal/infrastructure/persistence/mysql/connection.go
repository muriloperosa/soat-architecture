package mysql

import (
	"fmt"

	"github.com/muriloperosa/soat-architecture/internal/infrastructure/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewConnection abre a conexão GORM com o MySQL a partir do Config.
func NewConnection(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	return db, nil
}
