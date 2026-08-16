package migration

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4/database"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"

	db "github.com/muriloperosa/soat-architecture/internal/infrastructure/database"
)

func NewDriver(driverType db.DriverType, sqlDB *sql.DB, databaseName string) (database.Driver, error) {
	if !driverType.IsValid() {
		return nil, fmt.Errorf("%w: %s", ErrDriverInvalido, driverType)
	}

	switch driverType {
	case db.DriverTypeMySQL:
		if sqlDB == nil {
			return nil, ErrConexaoObrigatoria
		}

		driver, err := migratemysql.WithInstance(sqlDB, &migratemysql.Config{DatabaseName: databaseName})
		if err != nil {
			return nil, fmt.Errorf("erro ao criar driver MySQL: %w", err)
		}

		return driver, nil

	default:
		return nil, fmt.Errorf("%w: %s", ErrDriverNaoSuportado, driverType)
	}
}
