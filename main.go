package main

import (
	"golang_template/router"
	"golang_template/util"

	"github.com/rs/zerolog/log"
)

func main() {
	// Load config file using viper
	cfg, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("Could not load config")
		return
	}
	// Set logger configuration
	util.ConfigLogger(cfg)

	// Create new API server
	server, err := router.NewAPIServer()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not create router")
		return
	}

	// Setup router and swagger
	server.SetupRouter()
	server.SetupSwagger(cfg.SwaggerURL)

	// Start server on specified address
	err = server.Start(cfg.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not start server")
		return
	}

	// Log server address
	log.Info().Msgf("Listening and serving HTTP on %s", cfg.HTTPServerAddress)
}
