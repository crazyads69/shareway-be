package main

import (
	"os"
	"os/signal"
	"shareway/infra/crawler"
	"shareway/infra/db"
	"shareway/router"
	"shareway/service"
	"shareway/util"
	"shareway/util/token"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

// Define the init function that will be called before the main function
func init() {
	// Set the local timezone to UTC for the entire application
	time.Local = time.UTC
}

func main() {
	// Validator
	validate := validator.New()

	// Load config file using viper
	cfg, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("Could not load config")
		return
	}

	// Set logger configuration
	util.ConfigLogger(cfg)

	// // Init RabbitMQ
	// rabbitMQ, err := rabbitmq.NewRabbitMQ(cfg)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("Could not create RabbitMQ instance")
	// 	return
	// }
	// defer rabbitMQ.Close()

	// // Declare queue for notifications
	// err = rabbitMQ.DeclareQueue()
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("Could not declare queue")
	// 	return
	// }

	// // Initialize Firebase Cloud Messaging client
	// fcmClient, err := fcm.NewFCMClient(context.Background(), "")
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("Could not create FCM client")
	// 	return
	// }

	// // Create new notification worker
	// notificationWorker := worker.NewNotificationWorker(rabbitMQ, fcmClient, cfg)
	// go notificationWorker.Start()

	// // Create a new notification
	// err = rabbitMQ.PublishNotification(context.Background(), notification)
	// if err != nil {
	// 	log.Printf("Failed to publish notification: %v", err)
	// }

	// Create a scheduler
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not create task scheduler")
	}

	// Create new Paseto token maker
	maker, err := token.SetupPasetoMaker(cfg.PasetoSercetKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not create token maker")
		return
	}

	// Initialize DB
	database := db.NewDatabaseInstance(cfg)

	// Initialize Redis client
	redisClient := db.NewRedisClient(cfg)

	// Create a cron job to update the vehicle data from the VR website
	vrCrawler := crawler.NewVrCrawler(database)
	fuelCrawler := crawler.NewFuelCrawler(database)

	// Run the crawler once to populate the database
	err = vrCrawler.CrawlData()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not crawl data")
	}

	err = fuelCrawler.UpdateFuelPrices()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not update fuel prices")
	}

	// Add job to scheduler to run every week
	_, err = scheduler.NewJob(
		gocron.CronJob(`0 0 * * 0`, false), // Run every Sunday at midnight
		gocron.NewTask(
			vrCrawler.CrawlData,
		),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not create cron job")
	}

	// Add job to scheduler to run every 1 hour
	_, err = scheduler.NewJob(
		gocron.CronJob(`0 * * * *`, false), // Run every hour
		gocron.NewTask(
			fuelCrawler.UpdateFuelPrices,
		),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not create cron job")
	}

	// Start the scheduler
	scheduler.Start()

	// Initialize services using the service factory pattern (dependency injection also included repository pattern)
	serviceFactory := service.NewServiceFactory(database, cfg, maker, redisClient)
	services := serviceFactory.CreateServices()

	// Create new API server
	server, err := router.NewAPIServer(
		maker,
		cfg,
		services,
		validate,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not create router")
		return
	}

	// Setup router and swagger
	server.SetupRouter()
	server.SetupSwagger(cfg.SwaggerURL)

	// Create a channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start HTTP server in a goroutine
	go func() {
		if err := server.Start(cfg.HTTPServerAddress); err != nil {
			log.Fatal().Err(err).Msg("Could not start server")
		}
	}()

	log.Info().Msgf("Listening and serving HTTP on %s", cfg.HTTPServerAddress)

	// Wait for interrupt signal then after stop signal all the dependencies must be cleanup
	<-stop

	log.Info().Msg("Shutting down server...")

	// Perform any cleanup here

	// Cancel the scheduler
	scheduler.Shutdown()

}
