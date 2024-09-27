package main

import (
	"golang_template/router"
	"golang_template/util"

	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("Could not load config")
		return
	}

	util.ConfigLogger(cfg)

	server, err := router.NewAPIServer()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not create router")
		return
	}

	server.SetupRouter()
	server.SetupSwagger(cfg.SwaggerURL)

	err = server.Start(cfg.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not start server")
		return
	}

	log.Info().Msgf("Listening and serving HTTP on %s", cfg.HTTPServerAddress)
}
