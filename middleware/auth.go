package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"prak3/clean-architecture-fiber-mongo/helper"
)

func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		bearer := strings.TrimSpace(c.Get("Authorization"))
		if bearer == "" {
			return c.Status(fiber.StatusUnauthorized).
				JSON(fiber.Map{"success": false, "message": "Token tidak ditemukan"})
		}

		// format Bearer (case-insensitive)
		if len(bearer) < 7 || !strings.EqualFold(bearer[:7], "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).
				JSON(fiber.Map{"success": false, "message": "Format Authorization salah"})
		}
		token := strings.TrimSpace(bearer[7:])

		claims, err := helper.ValidateToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).
				JSON(fiber.Map{"success": false, "message": "Token tidak valid"})
		}

		if claims == nil || strings.TrimSpace(claims.UserID) == "" ||
			strings.TrimSpace(claims.Username) == "" || strings.TrimSpace(claims.Role) == "" {
			return c.Status(fiber.StatusUnauthorized).
				JSON(fiber.Map{"success": false, "message": "Klaim token tidak lengkap"})
		}

		role := strings.ToLower(claims.Role)

		c.Locals("user_id",  claims.UserID)  
		c.Locals("username", claims.Username)
		c.Locals("role",     role)
		c.Locals("user",     claims)

		return c.Next()
	}
}
