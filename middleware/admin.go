package middleware

import (

"github.com/gofiber/fiber/v2"
"fmt"
)
func AdminOnly() fiber.Handler {
    return func(c *fiber.Ctx) error {
        role, _ := c.Locals("role").(string)
        fmt.Println("[AdminOnly HIT]", c.Method(), c.Path(), "role=", role)
        if role != "admin" {
            return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
                "success": false,
                "error":   "Admin only",
            })
        }
        return c.Next()
    }
}

