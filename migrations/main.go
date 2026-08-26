package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/muriloperosa/soat-architecture/internal/infrastructure/config"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/database"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/migration"
	persistencemysql "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql"
)

const (
	upCommand      = "up"
	downCommand    = "down"
	versionCommand = "version"
	forceCommand   = "force"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("uso: go run ./migrations [up|down|version|force]")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("erro ao carregar configuração: %v", err)
	}

	gormDB, err := persistencemysql.NewMigrationConnection(cfg)
	if err != nil {
		log.Fatalf("erro ao conectar ao banco: %v", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("erro ao obter sql.DB: %v", err)
	}

	driverType := database.DriverType(cfg.DBDriver)
	driver, err := migration.NewDriver(driverType, sqlDB, cfg.DBName)
	if err != nil {
		log.Fatalf("erro ao criar driver de migration: %v", err)
	}

	migrationsPath := fmt.Sprintf("file://migrations/%s", cfg.DBDriver)

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, driverType.String(), driver)
	if err != nil {
		log.Fatalf("erro ao criar instância de migration: %v", err)
	}

	defer func() {
		sourceErr, databaseErr := m.Close()

		if sourceErr != nil {
			log.Printf("erro ao fechar source: %v", sourceErr)
		}

		if databaseErr != nil {
			log.Printf("erro ao fechar banco: %v", databaseErr)
		}
	}()

	switch os.Args[1] {
	case upCommand:
		runUp(m)

	case downCommand:
		runDown(m)

	case versionCommand:
		showVersion(m)

	case forceCommand:
		runForce(m)

	default:
		log.Fatalf("comando inválido: %s", os.Args[1])
	}
}

func runUp(migration *migrate.Migrate) {
	err := migration.Up()

	if errors.Is(err, migrate.ErrNoChange) {
		log.Println("nenhuma migration pendente")
		return
	}

	if err != nil {
		log.Fatalf("erro ao aplicar migrations: %v", err)
	}

	log.Println("migrations aplicadas com sucesso")
}

func runDown(migration *migrate.Migrate) {
	err := migration.Steps(-1)

	if errors.Is(err, migrate.ErrNoChange) {
		log.Println("nenhuma migration para desfazer")
		return
	}

	if err != nil {
		log.Fatalf("erro ao desfazer migration: %v", err)
	}

	log.Println("última migration desfeita com sucesso")
}

func showVersion(migration *migrate.Migrate) {
	version, dirty, err := migration.Version()

	if errors.Is(err, migrate.ErrNilVersion) {
		fmt.Println("nenhuma migration aplicada")
		return
	}

	if err != nil {
		log.Fatalf("erro ao consultar versão das migrations: %v", err)
	}

	fmt.Printf("versão atual: %d | dirty: %t\n", version, dirty)
}

func runForce(migration *migrate.Migrate) {
	if len(os.Args) < 3 {
		log.Fatal("uso: go run ./migrations force <versao>")
	}

	version, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatalf("versão inválida: %v", err)
	}

	if err := migration.Force(version); err != nil {
		log.Fatalf("erro ao forçar versão: %v", err)
	}

	log.Printf("versão ajustada para %d", version)
}
