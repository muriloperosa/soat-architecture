package mysql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewMigrationConnection cria uma conexão com o banco de dados MySQL para ser usada em migrações.
func NewMigrationConnection(cfg *config.Config) (*gorm.DB, error) {
	if cfg == nil {
		return nil, ErrConfigNotFound
	}

	return openMigrationConnection(mysql.Open(buildMigrationDSN(cfg)))
}

func openMigrationConnection(dialector gorm.Dialector) (*gorm.DB, error) {
	return openMigrationConnectionWith(dialector, func(dialector gorm.Dialector) (*gorm.DB, error) {
		return gorm.Open(dialector, &gorm.Config{DisableAutomaticPing: true})
	}, func(sqlDB *sql.DB) error {
		return sqlDB.Ping()
	})
}

func openMigrationConnectionWith(
	dialector gorm.Dialector,
	open func(gorm.Dialector) (*gorm.DB, error),
	ping func(*sql.DB) error,
) (*gorm.DB, error) {
	db, err := open(dialector)
	if err != nil {
		return nil, shared.NewInternalError("failed to connect to MySQL", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, shared.NewInternalError("failed to get underlying sql.DB", err)
	}

	configureMigrationPool(sqlDB)
	if err := ping(sqlDB); err != nil {
		return nil, shared.NewInternalError("failed to ping MySQL", err)
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
