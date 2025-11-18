package service

import (
	"context"
	"time"
  
	"github.com/gofiber/fiber/v2"

	"prak3/clean-architecture-fiber-mongo/app/model"
	"prak3/clean-architecture-fiber-mongo/app/repository"
	"prak3/clean-architecture-fiber-mongo/helper"
)

type AuthService struct {
	repo repository.UserRepository
}

func NewAuthService(repo repository.UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

// Login godoc
// @Summary Login user
// @Description Autentikasi user dan mengembalikan JWT
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body model.LoginRequest true "Login payload"
// @Success 200 {object} model.LoginResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Router /login [post]
func (s *AuthService) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil || req.Username == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "username & password wajib"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	u, err := s.repo.GetByUsernameOrEmail(ctx, req.Username)
	if err != nil || u == nil {
		return c.Status(401).JSON(fiber.Map{"error": "username/password salah"})
	}

	if !helper.CheckPassword(req.Password, u.PasswordHash) {
		return c.Status(401).JSON(fiber.Map{"error": "username/password salah"})
	}

	token, err := helper.GenerateToken(*u) 
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "gagal generate token"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "login berhasil",
		"data": model.LoginResponse{
			User: model.LoginUser{
				ID:        u.ID.Hex(),
				Username:  u.Username,
				Email:     u.Email,
				Role:      u.Role,
				CreatedAt: u.CreatedAt,
			},
			Token: token,
		},
	})
}

// Profile godoc
// @Summary Profil user (dari JWT)
// @Description Mengambil profil singkat dari klaim JWT yang aktif
// @Tags Auth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /profile [get]
func (s *AuthService) Profile(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"user_id":  c.Locals("user_id"),  
			"username": c.Locals("username"),
			"role":     c.Locals("role"),
		},
	})
}
