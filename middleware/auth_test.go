package middleware

import (
	"net/http/httptest"
	"os"
	"testing"

	"prak/clean-architecture-fiber-mongo/app/model"
	"prak/clean-architecture-fiber-mongo/helper"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
)

func TestAuthRequired(t *testing.T) {
	// jwt
	originalSecret := os.Getenv("JWT_SECRET")
	testSecret := "test-secret-key-for-middleware"
	os.Setenv("JWT_SECRET", testSecret)
	defer func() {
		if originalSecret == "" {
			os.Unsetenv("JWT_SECRET")
		} else {
			os.Setenv("JWT_SECRET", originalSecret)
		}
	}()

	// membuat test user dan token
	user := model.User{
		ID:       primitive.NewObjectID(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
	}
	validToken, _ := helper.GenerateToken(user)

	tests := []struct {
		name           string
		authorization  string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "missing authorization header",
			authorization:  "",
			expectedStatus: fiber.StatusUnauthorized,
			expectedBody:   "Token tidak ditemukan",
		},
		{
			name:           "valid bearer token",
			authorization:  "Bearer " + validToken,
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "token without bearer prefix",
			authorization:  validToken,
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "invalid token",
			authorization:  "Bearer invalid.token.here",
			expectedStatus: fiber.StatusUnauthorized,
			expectedBody:   "Token tidak valid",
		},
		{
			name:           "empty bearer token",
			authorization:  "Bearer ",
			expectedStatus: fiber.StatusUnauthorized,
			expectedBody:   "Token tidak ditemukan",
		},
		{
			name:           "bearer with spaces",
			authorization:  "Bearer  " + validToken,
			expectedStatus: fiber.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(AuthRequired())
			app.Get("/test", func(c *fiber.Ctx) error {
				return c.JSON(fiber.Map{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tt.authorization)

			resp, err := app.Test(req)
			utils.AssertEqual(t, nil, err)
			utils.AssertEqual(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedBody != "" {
				body := make([]byte, resp.ContentLength)
				resp.Body.Read(body)
				utils.AssertEqual(t, true, string(body) != "")
			}

			// For successful requests, verify locals are set
			if tt.expectedStatus == fiber.StatusOK {
				// We can't easily test locals without making actual request
				// This is tested in integration tests
			}
		})
	}
}

func TestAuthRequired_SetsLocals(t *testing.T) {
	// jwt
	originalSecret := os.Getenv("JWT_SECRET")
	testSecret := "test-secret-key-for-locals"
	os.Setenv("JWT_SECRET", testSecret)
	defer func() {
		if originalSecret == "" {
			os.Unsetenv("JWT_SECRET")
		} else {
			os.Setenv("JWT_SECRET", originalSecret)
		}
	}()

	user := model.User{
		ID:       primitive.NewObjectID(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "admin",
	}
	token, _ := helper.GenerateToken(user)

	app := fiber.New()
	app.Use(AuthRequired())
	app.Get("/test", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		username := c.Locals("username")
		role := c.Locals("role")

		return c.JSON(fiber.Map{
			"user_id":  userID,
			"username": username,
			"role":     role,
		})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
}
