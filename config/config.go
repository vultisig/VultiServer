package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port int64
	}

	Redis struct {
		Host     string
		Port     string
		Password string
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

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	err := viper.Unmarshal(&AppConfig)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}
}
