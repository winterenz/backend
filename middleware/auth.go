package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"prak/clean-architecture-fiber-mongo/helper"
)

func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		bearer := strings.TrimSpace(c.Get("Authorization"))
		if bearer == "" {
			return c.Status(fiber.StatusUnauthorized).
				JSON(fiber.Map{"success": false, "message": "Token tidak ditemukan"})
		}

		var token string
		// Menerima format "Bearer {token}" atau hanya "{token}"
		if len(bearer) >= 7 && strings.EqualFold(bearer[:7], "Bearer ") {
			// Format: Bearer {token}
			token = strings.TrimSpace(bearer[7:])
		} else {
			// Format: {token} saja (tanpa Bearer prefix)
			token = strings.TrimSpace(bearer)
		}
		
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).
				JSON(fiber.Map{"success": false, "message": "Token tidak ditemukan"})
		}

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

		c.Locals("user_id", claims.UserID)  
		c.Locals("username", claims.Username)
		c.Locals("role", role)
		c.Locals("user", claims)

		return c.Next()
	}
}
