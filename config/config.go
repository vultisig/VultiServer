package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port           int64
		Host           string
		VaultsFilePath string
	}

	Redis struct {
		Host     string
		Port     string
		User     string
		Password string
		DB       int
	}

	Relay struct {
		Server string
	}
}

var AppConfig Config

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	viper.SetDefault("Server.Port", 8080)
	viper.SetDefault("Server.Host", "localhost")
	viper.SetDefault("Server.VaultsFilePath", "vaults")
	viper.SetDefault("Redis.Host", "localhost")
	viper.SetDefault("Redis.Port", "6379")
	viper.SetDefault("Redis.User", "")
	viper.SetDefault("Redis.Password", "")
	viper.SetDefault("Redis.DB", 0)
	viper.SetDefault("Relay.Server", "http://localhost:8080")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	err := viper.Unmarshal(&AppConfig)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}
}
