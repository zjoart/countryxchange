package database

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zjoart/countryxchange/internal/config"
	"github.com/zjoart/countryxchange/pkg/logger"
)

// InitDB initializes and returns a database connection
func InitDB(config *config.DBConfig) (*sql.DB, error) {
	logger.Info("initializing database connection")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Name,
	)

	driverName := "mysql"

	logger.Info("opening database connection",
		logger.Fields{
			"driver": driverName,
		})

	db, err := sql.Open(driverName, dsn)

	if err != nil {
		logger.Error("failed to open database connection",
			logger.Merge(
				logger.WithError(err),
				logger.Fields{
					"driver": driverName,
				},
			))
		return nil, err
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		logger.Error("failed to ping database",
			logger.Merge(
				logger.WithError(err),
				logger.Fields{
					"driver": driverName,
				},
			))

		// Close the connection if ping fails
		if closeErr := db.Close(); closeErr != nil {
			logger.Error("failed to close database connection after ping failure",
				logger.WithError(closeErr))
		}

		return nil, err
	}

	logger.Info("database connection established successfully",
		logger.Fields{
			"driver": driverName,
		})

	return db, nil
}
