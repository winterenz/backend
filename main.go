// @title API Backend
// @version 1.0
// @description API untuk manajemen alumni, pekerjaan, file, dan autentikasi
// @host localhost:3000
// @BasePath /api
// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Masukkan token JWT

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"prak/clean-architecture-fiber-mongo/config"
	"prak/clean-architecture-fiber-mongo/database"
	_ "prak/clean-architecture-fiber-mongo/docs" // untuk swagger documentation
	"prak/clean-architecture-fiber-mongo/route"

	fiberSwagger "github.com/swaggo/fiber-swagger"
)

func main() {
	config.LoadEnv()
	config.InitLogger()

	client, db, err := database.ConnectMongo()
	if err != nil {
		log.Fatalf("mongo connect error: %v", err)
	}
	defer disconnectMongo(client)

	app := config.NewApp()
	app.Get("/swagger/*", fiberSwagger.FiberWrapHandler())
	app.Static("/uploads", "./uploads")
	route.Register(app, db)

	port := config.Env.AppPort
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Printf("fiber stopped: %v\n", err)
		}
	}()
	log.Printf("listening on :%s\n", port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("shutdown error: %v\n", err)
	}
}

func disconnectMongo(client *mongo.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Disconnect(ctx); err != nil {
		log.Printf("Warning: error disconnecting MongoDB: %v\n", err)
	}
}
