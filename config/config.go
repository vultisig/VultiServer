package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Host     string `mapstructure:"host" json:"host,omitempty"`
		Port     int64  `mapstructure:"port" json:"port,omitempty"`
		Database struct {
			DSN string `mapstructure:"dsn" json:"dsn,omitempty"`
		} `mapstructure:"database" json:"database,omitempty"`
		VaultsFilePath string `mapstructure:"vaults_file_path" json:"vaults_file_path,omitempty"`
		Mode           string `mapstructure:"mode" json:"mode,omitempty"`
		JWTSecret      string `mapstructure:"jwt_secret" json:"jwt_secret,omitempty"`
		Plugin         struct {
			Type string `mapstructure:"type" json:"type,omitempty"`
			Eth  struct {
				Rpc     string `mapstructure:"rpc" json:"rpc,omitempty"`
				Uniswap struct {
					V2Router string `mapstructure:"v2_router" json:"v2_router,omitempty"`
					Deadline int64  `mapstructure:"deadline" json:"deadline,omitempty"`
				} `mapstructure:"uniswap" json:"uniswap,omitempty"`
			} `mapstructure:"eth" json:"eth,omitempty"`
		} `mapstructure:"plugin" json:"plugin,omitempty"`
		UserAuth struct {
			JwtSecret string `mapstructure:"jwt_secret" json:"jwt_secret,omitempty"`
		} `mapstructure:"user_auth" json:"auth,omitempty"`
	} `mapstructure:"server" json:"server"`

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

	viper.SetDefault("Server.VaultsFilePath", "vaults")

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
