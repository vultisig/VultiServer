package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port           int64  `mapstructure:"port" json:"port,omitempty"`
		Host           string `mapstructure:"host" json:"host,omitempty"`
		VaultsFilePath string `mapstructure:"vaults_file_path" json:"vaults_file_path,omitempty"`
		Mode           string `mapstructure:"mode" json:"mode,omitempty"`
		Database       struct {
			DSN string `mapstructure:"dsn" json:"dsn,omitempty"`
		} `mapstructure:"database" json:"database,omitempty"`
	} `mapstructure:"server" json:"server"`

	Plugin struct {
		Type         string                 `mapstructure:"type" json:"type,omitempty"`
		PluginConfig map[string]interface{} `mapstructure:"plugin_config" json:"plugin_config,omitempty"`
	} `mapstructure:"plugin" json:"plugin,omitempty"`

	Redis struct {
		Host     string `mapstructure:"host" json:"host,omitempty"`
		Port     string `mapstructure:"port" json:"port,omitempty"`
		User     string `mapstructure:"user" json:"user,omitempty"`
		Password string `mapstructure:"password" json:"password,omitempty"`
		DB       int    `mapstructure:"db" json:"db,omitempty"`
	} `mapstructure:"redis" json:"redis,omitempty"`

	Relay struct {
		Server string `mapstructure:"server" json:"server"`
	} `mapstructure:"relay" json:"relay,omitempty"`

	EmailServer struct {
		ApiKey string `mapstructure:"api_key" json:"api_key"`
	} `mapstructure:"email_server" json:"email_server"`

	BlockStorage struct {
		Host      string `mapstructure:"host" json:"host"`
		Region    string `mapstructure:"region" json:"region"`
		AccessKey string `mapstructure:"access_key" json:"access_key"`
		SecretKey string `mapstructure:"secret" json:"secret"`
		Bucket    string `mapstructure:"bucket" json:"bucket"`
	} `mapstructure:"block_storage" json:"block_storage"`

	Datadog struct {
		Host string `mapstructure:"host" json:"host,omitempty"`
		Port string `mapstructure:"port" json:"port,omitempty"`
	} `mapstructure:"datadog" json:"datadog"`
}

func GetConfigure() (*Config, error) {
	configName := os.Getenv("VS_CONFIG_NAME")
	if configName == "" {
		configName = "config"
	}

	return ReadConfig(configName)
}

func ReadConfig(configName string) (*Config, error) {
	viper.SetConfigName(configName)
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	viper.SetDefault("Server.Port", 8080)
	viper.SetDefault("Server.Host", "localhost")
	viper.SetDefault("Server.VaultsFilePath", "vaults")
	viper.SetDefault("Server.Mode", "vultiserver")
	viper.SetDefault("Redis.Host", "localhost")
	viper.SetDefault("Redis.Port", "6379")
	viper.SetDefault("Redis.User", "")
	viper.SetDefault("Redis.Password", "")
	viper.SetDefault("Redis.DB", 0)
	viper.SetDefault("Relay.Server", "https://api.vultisig.com/router")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("fail to reading config file, %w", err)
	}
	var cfg Config
	err := viper.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %w", err)
	}
	return &cfg, nil
}
