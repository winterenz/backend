package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"prak/clean-architecture-fiber-mongo/app/model"
	"prak/clean-architecture-fiber-mongo/app/repository"
	"prak/clean-architecture-fiber-mongo/helper"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
)

// mock UserRepository
type mockUserRepository struct {
	users map[string]*model.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*model.User),
	}
}

func (m *mockUserRepository) GetByUsernameOrEmail(ctx context.Context, x string) (*model.User, error) {
	for _, user := range m.users {
		if user.Username == x || user.Email == x {
			return user, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepository) addUser(user *model.User) {
	m.users[user.Username] = user
	m.users[user.Email] = user
}

func TestAuthService_Login(t *testing.T) {
	// jwt
	originalSecret := os.Getenv("JWT_SECRET")
	testSecret := "test-secret-key-for-auth-service"
	os.Setenv("JWT_SECRET", testSecret)
	defer func() {
		if originalSecret == "" {
			os.Unsetenv("JWT_SECRET")
		} else {
			os.Setenv("JWT_SECRET", originalSecret)
		}
	}()

	// hash
	hashedPassword, _ := helper.HashPassword("password123")
	userID := primitive.NewObjectID()
	testUser := &model.User{
		ID:           userID,
		Username:     "testuser",
		Email:        "test@example.com",
		Role:         "user",
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
	}

	tests := []struct {
		name           string
		requestBody    map[string]string
		setupRepo      func() repository.UserRepository
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful login",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "password123",
			},
			setupRepo: func() repository.UserRepository {
				repo := newMockUserRepository()
				repo.addUser(testUser)
				return repo
			},
			expectedStatus: fiber.StatusOK,
		},
		{
			name: "login with email",
			requestBody: map[string]string{
				"username": "test@example.com",
				"password": "password123",
			},
			setupRepo: func() repository.UserRepository {
				repo := newMockUserRepository()
				repo.addUser(testUser)
				return repo
			},
			expectedStatus: fiber.StatusOK,
		},
		{
			name: "wrong password",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "wrongpassword",
			},
			setupRepo: func() repository.UserRepository {
				repo := newMockUserRepository()
				repo.addUser(testUser)
				return repo
			},
			expectedStatus: fiber.StatusUnauthorized,
			expectedBody:   "Username atau Password salah",
		},
		{
			name: "user not found",
			requestBody: map[string]string{
				"username": "nonexistent",
				"password": "password123",
			},
			setupRepo: func() repository.UserRepository {
				return newMockUserRepository()
			},
			expectedStatus: fiber.StatusUnauthorized,
			expectedBody:   "Username atau Password salah",
		},
		{
			name: "missing username",
			requestBody: map[string]string{
				"password": "password123",
			},
			setupRepo: func() repository.UserRepository {
				return newMockUserRepository()
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "Username dan Password wajib diisi",
		},
		{
			name: "missing password",
			requestBody: map[string]string{
				"username": "testuser",
			},
			setupRepo: func() repository.UserRepository {
				return newMockUserRepository()
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "Username dan Password wajib diisi",
		},
		{
			name: "empty body",
			requestBody: map[string]string{},
			setupRepo: func() repository.UserRepository {
				return newMockUserRepository()
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   "Username dan Password wajib diisi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			repo := tt.setupRepo()
			service := NewAuthService(repo)

			app.Post("/login", service.Login)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			utils.AssertEqual(t, nil, err)
			utils.AssertEqual(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedBody != "" {
				body := make([]byte, resp.ContentLength)
				resp.Body.Read(body)
				utils.AssertEqual(t, true, len(body) > 0)
			}
		})
	}
}

func TestAuthService_Profile(t *testing.T) {
	app := fiber.New()
	repo := newMockUserRepository()
	service := NewAuthService(repo)

	app.Get("/profile", func(c *fiber.Ctx) error {
		c.Locals("user_id", "507f1f77bcf86cd799439011")
		c.Locals("username", "testuser")
		c.Locals("role", "user")
		return service.Profile(c)
	})

	req := httptest.NewRequest("GET", "/profile", nil)

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
}

