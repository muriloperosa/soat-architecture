package database

type DriverType string

const (
	DriverTypeMySQL      DriverType = "mysql"
	DriverTypePostgreSQL DriverType = "postgres"
	DriverTypeSQLite     DriverType = "sqlite"
	DriverTypeSQLServer  DriverType = "sqlserver"
	DriverTypeOracle     DriverType = "oracle"
)

func (d DriverType) String() string {
	return string(d)
}

func (d DriverType) IsValid() bool {
	return (d == DriverTypeMySQL || d == DriverTypePostgreSQL || d == DriverTypeSQLite ||
		d == DriverTypeSQLServer || d == DriverTypeOracle)
}
