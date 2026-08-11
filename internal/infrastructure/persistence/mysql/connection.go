package mysql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/muriloperosa/soat-architecture/internal/infrastructure/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewConnection abre a conexão GORM com o MySQL a partir do Config,
// já com o connection pool configurado.
func NewConnection(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	configurePool(sqlDB, cfg)

	return db, nil
}

// configurePool aplica os limites de connection pool a partir do Config.
func configurePool(sqlDB *sql.DB, cfg *config.Config) {
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.DBConnMaxLifetimeMinutes) * time.Minute)
}
