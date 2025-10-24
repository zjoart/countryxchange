package config

import (
	"fmt"
	"os"
	"strings"
)

type SwaggerConfig struct {
	Host    string
	Schemes []string
}

type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
}

type Config struct {
	AppEnv  string
	Port    string
	DB      DBConfig
	Swagger SwaggerConfig
}

func LoadConfig() *Config {
	config := &Config{
		Port: getEnv("PORT"),
		DB: DBConfig{
			User:     getEnv("DB_USER"),
			Password: getEnv("DB_PASS"),
			Host:     getEnv("DB_HOST"),
			Port:     getEnv("DB_PORT"),
			Name:     getEnv("DB_NAME"),
		},
		Swagger: loadSwaggerConfig(),
		AppEnv:  getEnv("APP_ENV"),
	}

	return config
}

func loadSwaggerConfig() SwaggerConfig {
	host := getEnv("API_BASE")
	schemes := getEnv("SWAGGER_SCHEMES")
	return SwaggerConfig{
		Host:    host,
		Schemes: strings.Split(schemes, ","),
	}
}

func getEnv(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	panic(fmt.Sprintf("%s is required", key))
}
