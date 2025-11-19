package middleware

import (
	"github.com/gofiber/fiber/v2"
	"prak/clean-architecture-fiber-mongo/app/model"
)

func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		if role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
				Success: false,
				Message: "Admin only",
			})
		}
		return c.Next()
	}
}