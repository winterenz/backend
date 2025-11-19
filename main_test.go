package main

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	_ "prak/clean-architecture-fiber-mongo/docs" // Penting untuk swagger

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	fiberSwagger "github.com/swaggo/fiber-swagger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func TestSwaggerRoute(t *testing.T) {
	app := fiber.New()
	app.Get("/swagger/*", fiberSwagger.FiberWrapHandler())

	// mencoba akses ke index.html swagger
	req := httptest.NewRequest("GET", "/swagger/index.html", nil)
	resp, err := app.Test(req, -1)

	// asertion
	assert.NoError(t, err)
	assert.NotEqual(t, 404, resp.StatusCode, "404 swagger error")
}

// mock connection mongodb
func TestDisconnectMongo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	
	if err != nil || client.Ping(ctx, readpref.Primary()) != nil {
		t.Skip("mongodb tidak terdeteksi, melewati TestDisconnectMongo")
	}

	disconnectMongo(client)

	// assertion ping mongodb
	ctxPing, cancelPing := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelPing()
	errPing := client.Ping(ctxPing, readpref.Primary())

	assert.Error(t, errPing, "client harusnya sudah disconnect, ping error")
}

func TestAppStaticRoute(t *testing.T) {
	app := fiber.New()
	app.Static("/uploads", "./uploads")
	
	req := httptest.NewRequest("GET", "/uploads/", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}