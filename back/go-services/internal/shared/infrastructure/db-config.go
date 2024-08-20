package infrastructure

import (
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"
)

type DBEnvConfig struct {
	DBUser     string `envconfig:"DB_USER" required:"true"`
	DBPassword string `envconfig:"DB_PASSWORD" required:"true"`
	DBHost     string `envconfig:"DB_HOST" required:"true"`
	DBName     string `envconfig:"DB_NAME" required:"true"`
}

func GetDBConfig() Config {
	dbEnvConfig := DBEnvConfig{}
	err := envconfig.Process("", &dbEnvConfig)
	if err != nil {
		slog.Error("Failed to get db config: " + err.Error())
		os.Exit(1)
	}

	return Config(dbEnvConfig)
}
