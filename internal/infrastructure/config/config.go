package config

import "github.com/caarlos0/env/v9"

// Config carrega as variáveis de ambiente da aplicação.
type Config struct {
	AppPort string `env:"APP_PORT" envDefault:"8080"`

	DBHost     string `env:"DB_HOST,required"`
	DBPort     string `env:"DB_PORT,required"`
	DBUser     string `env:"DB_USER,required"`
	DBPassword string `env:"DB_PASSWORD,required"`
	DBName     string `env:"DB_NAME,required"`
	DBDriver   string `env:"DB_DRIVER" envDefault:"mysql"`

	DBMaxOpenConns           int `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	DBMaxIdleConns           int `env:"DB_MAX_IDLE_CONNS" envDefault:"5"`
	DBConnMaxLifetimeMinutes int `env:"DB_CONN_MAX_LIFETIME_MINUTES" envDefault:"5"`

	JWTSecret                string `env:"JWT_SECRET,required"`
	JWTAccessTokenTTLMinutes int    `env:"JWT_ACCESS_TOKEN_TTL_MINUTES" envDefault:"60"`
	JWTRefreshTokenTTLHours  int    `env:"JWT_REFRESH_TOKEN_TTL_HOURS" envDefault:"168"`
}

// Load lê as variáveis de ambiente do processo e retorna um Config validado.
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
