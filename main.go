package main

import (
	"shareway/infra/db"
	"shareway/router"
	"shareway/service"
	"shareway/util"
	"shareway/util/token"

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

	// Create new Paseto token maker
	maker, _ := token.NewPasetoMaker(cfg.PasetoSercetKey)
	// Initialize DB
	db := db.NewDatabaseInstance(cfg)

	// Initialize services using the service factory pattern (dependency injection also included repository pattern)
	serviceFactory := service.NewServiceFactory(db, cfg, maker)
	services := serviceFactory.CreateServices()

	// Create new API server
	server, err := router.NewAPIServer(
		maker,
		cfg,
		services,
	)
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
