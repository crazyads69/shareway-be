package util

import (
	"os"

	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/rs/zerolog"
)

func ConfigLogger(cfg Config) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Create a Lumberjack logger for file rotation
	fileLogger := &lumberjack.Logger{
		Filename:   cfg.LogFilename,   // Specify the log file name
		MaxSize:    cfg.LogMaxSize,    // Max size in megabytes before rotation
		MaxBackups: cfg.LogMaxBackups, // Max number of backup files to keep
		MaxAge:     cfg.LogMaxAge,     // Max number of days to retain old log files
		Compress:   cfg.LogCompress,   // Whether to compress old log files
	}

	// Create a MultiWriter that writes to both stdout and the log file
	multiWriter := zerolog.MultiLevelWriter(
		zerolog.ConsoleWriter{Out: os.Stdout},
		fileLogger,
	)

	// Initialize the logger with the MultiWriter
	log.Logger = zerolog.New(multiWriter).With().Timestamp().Caller().Logger()
}
