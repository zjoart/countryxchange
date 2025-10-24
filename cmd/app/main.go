package main

import (
	"fmt"
	"net/http"

	"github.com/zjoart/countryxchange/cmd/routes"
	"github.com/zjoart/countryxchange/internal/config"
	"github.com/zjoart/countryxchange/internal/database"
	"github.com/zjoart/countryxchange/pkg/logger"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		logger.Warn("No .env file found", logger.WithError(err))
	}

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db, errDb := database.InitDB(&cfg.DB)
	if errDb != nil {
		logger.Fatal("Failed to initialize database", logger.WithError(errDb))
	}

	defer db.Close()

	// Initialize the application

	router := routes.SetUpRoutes(db, cfg)

	// Initialize the application
	addr := fmt.Sprintf(":%s", cfg.Port)
	logger.Info("Service starting", logger.Fields{
		"port": cfg.Port,
	})

	if err := http.ListenAndServe(addr, router); err != nil {
		logger.Fatal("Server failed", logger.WithError(err))
	}
}
