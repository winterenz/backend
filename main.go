// @title Alumni & Jobs API
// @version 1.0
// @description API untuk manajemen alumni, pekerjaan, file, dan autentikasi
// @host localhost:3000
// @BasePath /api
// @schemes http

// Gunakan skema HTTP Bearer agar Swagger otomatis tambahkan prefix "Bearer "
// @securityDefinitions.type http
// @securityDefinitions.scheme bearer
// @securityDefinitions.bearerFormat JWT

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"prak3/clean-architecture-fiber-mongo/config"
	"prak3/clean-architecture-fiber-mongo/database"
	"prak3/clean-architecture-fiber-mongo/route"
	_"prak3/clean-architecture-fiber-mongo/docs" // untuk swagger documentation
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
	_ = client.Disconnect(ctx)
}
