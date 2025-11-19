package helper

import (
	"os"
	"testing"
	"time"

	"prak/clean-architecture-fiber-mongo/app/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGenerateToken(t *testing.T) {
	// jwt
	originalSecret := os.Getenv("JWT_SECRET")
	testSecret := "test-secret-key-for-jwt-generation"
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
		Role:     "user",
	}

	tests := []struct {
		name    string
		user    model.User
		wantErr bool
	}{
		{
			name:    "valid user",
			user:    user,
			wantErr: false,
		},
		{
			name: "user with admin role",
			user: model.User{
				ID:       primitive.NewObjectID(),
				Username: "admin",
				Email:    "admin@example.com",
				Role:     "admin",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == "" {
				t.Error("GenerateToken() returned empty token")
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	// jwt
	originalSecret := os.Getenv("JWT_SECRET")
	testSecret := "test-secret-key-for-jwt-validation"
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
		Role:     "user",
	}

	tests := []struct {
		name    string
		setup   func() string
		wantErr bool
	}{
		{
			name: "valid token",
			setup: func() string {
				token, _ := GenerateToken(user)
				return token
			},
			wantErr: false,
		},
		{
			name: "invalid token format",
			setup: func() string {
				return "invalid.token.format"
			},
			wantErr: true,
		},
		{
			name: "empty token",
			setup: func() string {
				return ""
			},
			wantErr: true,
		},
		{
			name: "token with wrong secret",
			setup: func() string {
				// Generate token with different secret
				os.Setenv("JWT_SECRET", "wrong-secret")
				token, _ := GenerateToken(user)
				os.Setenv("JWT_SECRET", testSecret)
				return token
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setup()
			claims, err := ValidateToken(token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if claims == nil {
					t.Error("ValidateToken() returned nil claims for valid token")
					return
				}
				if claims.UserID != user.ID.Hex() {
					t.Errorf("ValidateToken() UserID = %v, want %v", claims.UserID, user.ID.Hex())
				}
				if claims.Username != user.Username {
					t.Errorf("ValidateToken() Username = %v, want %v", claims.Username, user.Username)
				}
				if claims.Role != user.Role {
					t.Errorf("ValidateToken() Role = %v, want %v", claims.Role, user.Role)
				}
			}
		})
	}
}

func TestGenerateToken_ValidateToken_RoundTrip(t *testing.T) {
	// jwt
	originalSecret := os.Getenv("JWT_SECRET")
	testSecret := "test-secret-key-for-roundtrip"
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
		CreatedAt: time.Now(),
	}

	token, err := GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken() failed: %v", err)
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() failed: %v", err)
	}

	if claims.UserID != user.ID.Hex() {
		t.Errorf("RoundTrip: UserID = %v, want %v", claims.UserID, user.ID.Hex())
	}
	if claims.Username != user.Username {
		t.Errorf("RoundTrip: Username = %v, want %v", claims.Username, user.Username)
	}
	if claims.Role != user.Role {
		t.Errorf("RoundTrip: Role = %v, want %v", claims.Role, user.Role)
	}
}

