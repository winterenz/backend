package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
)

func TestAdminOnly(t *testing.T) {
	tests := []struct {
		name           string
		role           string
		expectedStatus int
	}{
		{
			name:           "admin role",
			role:           "admin",
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "user role",
			role:           "user",
			expectedStatus: fiber.StatusForbidden,
		},
		{
			name:           "empty role",
			role:           "",
			expectedStatus: fiber.StatusForbidden,
		},
		{
			name:           "case sensitive admin",
			role:           "Admin",
			expectedStatus: fiber.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("role", tt.role)
				return c.Next()
			})
			app.Use(AdminOnly())
			app.Get("/test", func(c *fiber.Ctx) error {
				return c.JSON(fiber.Map{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/test", nil)

			resp, err := app.Test(req)
			utils.AssertEqual(t, nil, err)
			utils.AssertEqual(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestAdminOnly_AllowsAdmin(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("role", "admin")
		return c.Next()
	})
	app.Use(AdminOnly())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "admin access granted"})
	})

	req := httptest.NewRequest("GET", "/test", nil)

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
}
