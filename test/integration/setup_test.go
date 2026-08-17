//go:build integration

package integration_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/muriloperosa/soat-architecture/internal/infrastructure/config"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/database"
	httpinfra "github.com/muriloperosa/soat-architecture/internal/infrastructure/http"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/migration"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
)

const testDatabaseName = "oficina_integration"

var (
	testRouter    *gin.Engine
	testDB        *gorm.DB
	testContainer *wiring.Container
)

// TestMain sobe um container MySQL real (testcontainers), aplica as
// migrations de produção e monta wiring.Container/router exatamente como em
// produção. Sobe uma vez só pra todo o pacote (subir um container por teste seria caro demais); os testes
// compartilham o container e usam resetDB pra isolamento.
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()

	mysqlC, err := tcmysql.Run(ctx, "mysql:8",
		tcmysql.WithUsername("integration"),
		tcmysql.WithPassword("integration"),
		tcmysql.WithDatabase(testDatabaseName),
	)
	if err != nil {
		log.Fatalf("erro ao subir container mysql: %v", err)
	}
	defer func() {
		if err := mysqlC.Terminate(ctx); err != nil {
			log.Printf("erro ao encerrar container mysql: %v", err)
		}
	}()

	dsn := mysqlC.MustConnectionString(ctx, "parseTime=True", "loc=Local", "charset=utf8mb4")

	db, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("erro ao conectar no mysql do container: %v", err)
	}
	testDB = db

	if err := aplicarMigrations(db); err != nil {
		log.Fatalf("erro ao aplicar migrations: %v", err)
	}

	cfg := &config.Config{
		JWTSecret:                "segredo-de-teste-integracao",
		JWTAccessTokenTTLMinutes: 15,
		JWTRefreshTokenTTLHours:  1,
	}

	testContainer = wiring.NewContainer(cfg, db)
	testRouter = httpinfra.NewRouter(testContainer)

	os.Exit(m.Run())
}

// aplicarMigrations roda migrations/mysql contra db, mesmo runner usado por
// migrations/main.go (internal/infrastructure/migration).
func aplicarMigrations(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	driver, err := migration.NewDriver(database.DriverType("mysql"), sqlDB, testDatabaseName)
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://../../migrations/mysql", "mysql", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
