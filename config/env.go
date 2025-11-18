package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type EnvConfig struct {
	AppPort    string
	MongoURI   string
	MongoDB    string
	JWTSecret  string
}

var Env EnvConfig

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env vars")
	}

	Env.AppPort  = os.Getenv("APP_PORT")
	Env.MongoURI = os.Getenv("MONGO_URI")
	Env.MongoDB  = os.Getenv("MONGO_DB_NAME")
	Env.JWTSecret = os.Getenv("JWT_SECRET")

	// fallback kalau env lama masih dipakai
	if Env.MongoDB == "" {
		if alt := os.Getenv("DATABASE_NAME"); alt != "" {
			Env.MongoDB = alt
		}
	}
	if Env.MongoURI == "" {
		Env.MongoURI = "mongodb://localhost:27017"
	}
	if Env.AppPort == "" {
		Env.AppPort = "3000"
	}

	if Env.JWTSecret == "" {
		log.Println("Warning: JWT_SECRET is empty")
	}
}
