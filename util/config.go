package util

import (
	"github.com/spf13/viper"
)

// Config stores all application configs loaded from file or env variables
type Config struct {
	HTTPServerAddress string `mapstructure:"HTTP_SERVER_ADDRESS"`

	SwaggerURL string `mapstructure:"SWAGGER_URL"`

	LogFilename                    string `mapstructure:"LOG_FILENAME"`
	LogMaxSize                     int    `mapstructure:"LOG_MAX_SIZE"` // in megabytes
	LogMaxBackups                  int    `mapstructure:"LOG_MAX_BACKUPS"`
	LogMaxAge                      int    `mapstructure:"LOG_MAX_AGE"`
	LogCompress                    bool   `mapstructure:"LOG_COMPRESS"`
	GinMode                        string `mapstructure:"GIN_MODE"`
	DatabaseUsername               string `mapstructure:"DB_USER"`
	DatabasePassword               string `mapstructure:"DB_PASSWORD"`
	DatabaseName                   string `mapstructure:"DB_NAME"`
	DatabaseHost                   string `mapstructure:"DB_HOST"`
	DatabasePort                   int    `mapstructure:"DB_PORT"`
	PasetoSercetKey                string `mapstructure:"PASETO_SECRET_KEY"`
	PasetoExpiredDuration          int    `mapstructure:"PASETO_EXPIRED_DURATION"`
	TwilioAccountSID               string `mapstructure:"TWILIO_ACCOUNT_SID"`
	TwilioAuthToken                string `mapstructure:"TWILIO_AUTH_TOKEN"`
	TwilioServiceSID               string `mapstructure:"TWILIO_SERVICE_SID"`
	RedisHost                      string `mapstructure:"REDIS_HOST"`
	RedisPort                      int    `mapstructure:"REDIS_PORT"`
	RedisPassword                  string `mapstructure:"REDIS_PASSWORD"`
	RedisDB                        int    `mapstructure:"REDIS_DB"`
	RedisExpiredDuration           int    `mapstructure:"REDIS_EXPIRED_DURATION"`
	RedisProtocol                  int    `mapstructure:"REDIS_PROTOCOL"`
	FptAiApiKey                    string `mapstructure:"FPT_AI_API_KEY"`
	FptAiApiUrl                    string `mapstructure:"FPT_AI_API_URL"`
	EncryptionKey                  string `mapstructure:"ENCRYPTION_KEY"`
	AccessTokenExpiredDuration     int    `mapstructure:"ACCESS_TOKEN_EXPIRED_DURATION"`
	RefreshTokenExpiredDuration    int    `mapstructure:"REFRESH_TOKEN_EXPIRED_DURATION"`
	MaxOTPAttempts                 int    `mapstructure:"MAX_OTP_ATTEMPTS"`
	MaxOTPSendCount                int    `mapstructure:"MAX_OTP_SEND_COUNT"`
	OtpExpiredDuration             int    `mapstructure:"OTP_EXPIRED_DURATION"`
	OtpCooldownDuration            int    `mapstructure:"OTP_COOLDOWN_DURATION"`
	OTPSendCountDuration           int    `mapstructure:"OTP_SEND_COUNT_DURATION"`
	AmqpServerURL                  string `mapstructure:"AMQP_SERVER_URL"`
	AmqpNotificationQueue          string `mapstructure:"AMQP_NOTIFICATION_QUEUE"`
	GoongApiURL                    string `mapstructure:"GOONG_API_URL"`
	GoongAPIKey                    string `mapstructure:"GOONG_API_KEY"`
	GoongCacheExpiredDuration      int    `mapstructure:"GOONG_CACHE_EXPIRED_DURATION"`
	GoongCacheAutocompleteDuration int    `mapstructure:"GOONG_CACHE_AUTOCOMPLETE_DURATION"` // 24 hours
	GoongCachePlaceDetailDuration  int    `mapstructure:"GOONG_CACHE_PLACE_DETAIL_DURATION"` // 7 days
	GoongCacheRouteDuration        int    `mapstructure:"GOONG_CACHE_ROUTE_DURATION"`        // 12 hours
	GoongCacheDefaultDuration      int    `mapstructure:"GOONG_CACHE_DEFAULT_DURATION"`      // 1 hour (for backward compatibility)
	FCMConfigPath                  string `mapstructure:"FCM_CONFIG_PATH"`
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
