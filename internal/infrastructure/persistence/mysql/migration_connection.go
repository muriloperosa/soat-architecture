package mysql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/muriloperosa/soat-architecture/internal/infrastructure/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewMigrationConnection cria uma conexão com o banco de dados MySQL para ser usada em migrações.
func NewMigrationConnection(cfg *config.Config) (*gorm.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuração é obrigatória")
	}

	dsn := buildMigrationDSN(cfg)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	configureMigrationPool(sqlDB)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping MySQL: %w", err)
	}

	return db, nil
}

func buildMigrationDSN(cfg *config.Config) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=30s&writeTimeout=30s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)
}

func configureMigrationPool(sqlDB *sql.DB) {
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(0)
	sqlDB.SetConnMaxLifetime(time.Minute)
}
