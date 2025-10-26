package routes

import (
	"database/sql"

	"github.com/zjoart/countryxchange/internal/config"

	"github.com/zjoart/countryxchange/internal/countries"

	"net/http"

	"github.com/zjoart/countryxchange/internal/middleware"

	"github.com/zjoart/countryxchange/internal/docs"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/gorilla/mux"
)

//	@title			Countries Xchange API
//	@version		1.0
//	@description	This is the backend API for the Countries Xchange an HNG Stage 2 Task.
//	@termsOfService	https://example.com/terms/

//	@contact.name	API Support
//	@contact.url	https://example.com/support
//	@contact.email	support@countryxchange.com

//	@license.name	MIT License
//	@license.url	https://opensource.org/licenses/MIT

//	@host		localhost:8080
//	@BasePath	/

// @schemes	http https
func SetUpRoutes(db *sql.DB, cfg *config.Config) http.Handler {

	allowedOrigins := []string{
		"*",
	}

	// Create a new Gorilla Mux router
	router := mux.NewRouter()

	//Use cors middleware
	router.Use(middleware.CorsMiddleware(allowedOrigins))

	// Dynamically set Swagger host and schemes from config
	if cfg.Swagger.Host != "" {
		docs.SwaggerInfo.Host = cfg.Swagger.Host
	}
	if len(cfg.Swagger.Schemes) > 0 {
		docs.SwaggerInfo.Schemes = cfg.Swagger.Schemes
	}

	if cfg.AppEnv != "production" {
		// Serve Swagger UI only in non-production environments
		router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

		// Optional: Redirect /swagger to /swagger/index.html
		router.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
		})
	}

	//Handle health
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Service is up and running"))
	}).Methods("GET")

	// Register country feature routes
	// keep feature based routing in internal/countries
	countries.RegisterRoutes(router, db)

	return router
}
