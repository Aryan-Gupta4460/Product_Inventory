package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	StorageBackend string
	FilePath       string
	LogLevel       string
}

func Load() (Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AddConfigPath(".")
	viper.AddConfigPath("./internal/config")

	viper.SetDefault("storage_backend", "memory")
	viper.SetDefault("file_path", "data.json")
	viper.SetDefault("log_level", "info")

	if err := viper.ReadInConfig(); err != nil {
		return Config{}, err
	}

	return Config{
		StorageBackend: viper.GetString("storage_backend"),
		FilePath:       viper.GetString("file_path"),
		LogLevel:       viper.GetString("log_level"),
	}, nil
}
