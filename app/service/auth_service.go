package service

import (
	"context"
	"time"
  
	"github.com/gofiber/fiber/v2"

	"prak/clean-architecture-fiber-mongo/app/model"
	"prak/clean-architecture-fiber-mongo/app/repository"
	"prak/clean-architecture-fiber-mongo/helper"
)

type AuthService struct {
	repo repository.UserRepository
}

func NewAuthService(repo repository.UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

// Login godoc
// @Summary login
// @Description Autentikasi user dan klaim JWT
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body model.LoginRequest true "Login payload"
// @Success 200 {object} model.LoginSuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Router /login [post]
func (s *AuthService) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil || req.Username == "" || req.Password == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "Username dan Password wajib diisi",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	u, err := s.repo.GetByUsernameOrEmail(ctx, req.Username)
	if err != nil || u == nil {
		return c.Status(401).JSON(model.ErrorResponse{
			Success: false,
			Message: "Username atau Password salah",
		})
	}

	if !helper.CheckPassword(req.Password, u.PasswordHash) {
		return c.Status(401).JSON(model.ErrorResponse{
			Success: false,
			Message: "Username atau Password salah",
		})
	}

	token, err := helper.GenerateToken(*u) 
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: "Gagal generate token",
		})
	}

	return c.JSON(model.LoginSuccessResponse{
		Success: true,
		Message: "Login berhasil",
		Data: model.LoginResponse{
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
// @Summary profil user
// @Description Mengambil profil singkat dari klaim JWT yang aktif
// @Tags Auth
// @Produce json
// @Success 200 {object} model.ProfileResponse
// @Security BearerAuth
// @Router /profile [get]
func (s *AuthService) Profile(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	username, _ := c.Locals("username").(string)
	role, _ := c.Locals("role").(string)

	return c.JSON(model.ProfileResponse{
		Success: true,
		Data: model.ProfileData{
			UserID:   userID,
			Username: username,
			Role:     role,
		},
	})
}