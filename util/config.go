package util

import (
	"github.com/spf13/viper"
)

// Config stores all application configs loaded from file or env variables
type Config struct {
	HTTPServerAddress string `mapstructure:"HTTP_SERVER_ADDRESS"`

	SwaggerURL string `mapstructure:"SWAGGER_URL"`

	LogFilename           string `mapstructure:"LOG_FILENAME"`
	LogMaxSize            int    `mapstructure:"LOG_MAX_SIZE"` // in megabytes
	LogMaxBackups         int    `mapstructure:"LOG_MAX_BACKUPS"`
	LogMaxAge             int    `mapstructure:"LOG_MAX_AGE"`
	LogCompress           bool   `mapstructure:"LOG_COMPRESS"`
	DatabaseUsername      string `mapstructure:"DB_USER"`
	DatabasePassword      string `mapstructure:"DB_PASSWORD"`
	DatabaseName          string `mapstructure:"DB_NAME"`
	DatabaseHost          string `mapstructure:"DB_HOST"`
	DatabasePort          int    `mapstructure:"DB_PORT"`
	PasetoSercetKey       string `mapstructure:"PASETO_SECRET_KEY"`
	PasetoExpiredDuration int    `mapstructure:"PASETO_EXPIRED_DURATION"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("HTTP_SERVER_ADDRESS", ":8080")

	viper.SetDefault("SWAGGER_URL", "/docs")

	viper.SetDefault("LOG_FILENAME", "log/app.log")
	viper.SetDefault("LOG_MAX_SIZE", 10)
	viper.SetDefault("LOG_MAX_BACKUPS", 5)
	viper.SetDefault("LOG_MAX_AGE", 28)
	viper.SetDefault("LOG_COMPRESS", true)

	// Read config
	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	// Unmarshal config
	err = viper.Unmarshal(&config)
	return
}
