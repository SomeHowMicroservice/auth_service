package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		ServerHost string `mapstructure:"server_host"`
		GRPCPort   int  `mapstructure:"grpc_port"`
	} `mapstructure:"app"`

	Jwt struct {
		SecretKey        string        `mapstructure:"secret_key"`
		AccessExpiresIn  time.Duration `mapstructure:"access_expires_in"`
		RefreshExpiresIn time.Duration `mapstructure:"refresh_expires_in"`
	}

	Services struct {
		UserPort int `mapstructure:"user_port"`
	} `mapstructure:"services"`

	Cache struct {
		CHost     string `mapstructure:"rd_host"`
		CPort     int    `mapstructure:"rd_port"`
		CPassword string `mapstructure:"rd_password"`
	} `mapstructure:"cache"`

	MessageQueue struct {
		RHost     string `mapstructure:"rb_host"`
		RUser     string `mapstructure:"rb_user"`
		RPassword string `mapstructure:"rb_password"`
	} `mapstructure:"message_queue"`

	SMTP struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	} `mapstructure:"smtp"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile("config/config.yaml")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
