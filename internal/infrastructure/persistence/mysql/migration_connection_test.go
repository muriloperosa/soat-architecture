package mysql

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/config"
	"github.com/stretchr/testify/require"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestBuildMigrationDSN(t *testing.T) {
	cfg := &config.Config{DBHost: "127.0.0.1", DBPort: "3306", DBUser: "root", DBPassword: "root", DBName: "mecanica"}
	dsn := buildMigrationDSN(cfg)
	require.Equal(t, "root:root@tcp(127.0.0.1:3306)/mecanica?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=30s&writeTimeout=30s", dsn)
}

func TestConfigureMigrationPool(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	configureMigrationPool(sqlDB)
	require.Equal(t, 1, sqlDB.Stats().MaxOpenConnections)
}

func TestNewMigrationConnectionComConfigNil(t *testing.T) {
	db, err := NewMigrationConnection(nil)
	require.Nil(t, db)
	require.Error(t, err)
	require.Equal(t, ErrConfigNotFound, err)
}

func TestNewMigrationConnectionComBancoIndisponivel(t *testing.T) {
	cfg := &config.Config{DBHost: "127.0.0.1", DBPort: "1", DBUser: "root", DBPassword: "root", DBName: "mecanica"}
	db, err := NewMigrationConnection(cfg)
	require.Nil(t, db)
	require.Error(t, err)
}

func TestOpenMigrationConnectionFalhaAoAbrir(t *testing.T) {
	db, err := openMigrationConnectionWithOpen(nil, func(gorm.Dialector) (*gorm.DB, error) {
		return nil, errors.New("falha simulada ao abrir")
	}, func(*sql.DB) error { return nil })

	require.Nil(t, db)
	require.Error(t, err)
}

func TestOpenMigrationConnectionComPingBemSucedido(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	ormDB, err := gorm.Open(gormmysql.New(gormmysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)

	got, err := openMigrationConnectionWithOpen(nil, func(gorm.Dialector) (*gorm.DB, error) { return ormDB, nil }, func(*sql.DB) error { return nil })
	require.NoError(t, err)
	require.Same(t, ormDB, got)
	require.Equal(t, 1, sqlDB.Stats().MaxOpenConnections)
}

func TestOpenMigrationConnectionPropagaFalhaDoPing(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	db, err := openMigrationConnection(gormmysql.New(gormmysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}))
	require.Error(t, err)
	require.Nil(t, db)
	require.Equal(t, 1, sqlDB.Stats().MaxOpenConnections)
}

func TestPingMigrationConnectionFalha(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	db, err := pingMigrationConnection(&gorm.DB{}, sqlDB, func(*sql.DB) error { return errors.New("falha simulada") })
	require.Nil(t, db)
	require.Error(t, err)
}

func TestPingMigrationConnectionComSucesso(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	ormDB := &gorm.DB{}
	got, err := pingMigrationConnection(ormDB, sqlDB, func(*sql.DB) error { return nil })
	require.NoError(t, err)
	require.Same(t, ormDB, got)
}
