package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	DBUri     string `mapstructure:"MONGODB_LOCAL_URI"`
	Port      string `mapstructure:"PORT"`
	ExportDir string `mapstructure:"EXPORT_DIR"`

	AccessTokenPrivateKey string `mapstructure:"ACCESS_TOKEN_PRIVATE_KEY"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigType("env")
	viper.SetConfigName("app")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
