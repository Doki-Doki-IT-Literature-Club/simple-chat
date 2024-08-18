package main

import (
	"log/slog"
	"os"

	"github.com/LeperGnome/simple-chat/pkg/chronicler"
	"github.com/kelseyhightower/envconfig"
)

type DBEnvConfig struct {
	DBUser     string `envconfig:"DB_USER" required:"true"`
	DBPassword string `envconfig:"DB_PASSWORD" required:"true"`
	DBHost     string `envconfig:"DB_HOST" required:"true"`
	DBName     string `envconfig:"DB_NAME" required:"true"`
}

func getDBConfig() chronicler.Config {
	dbEnvConfig := DBEnvConfig{}
	err := envconfig.Process("", &dbEnvConfig)
	if err != nil {
		slog.Error("Failed to get db config: " + err.Error())
		os.Exit(1)
	}

	return chronicler.Config{
		DBUser:     dbEnvConfig.DBUser,
		DBPassword: dbEnvConfig.DBPassword,
		DBHost:     dbEnvConfig.DBHost,
		DBName:     dbEnvConfig.DBName,
	}
}
