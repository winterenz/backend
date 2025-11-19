package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"prak/clean-architecture-fiber-mongo/app/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
)

// mock AlumniRepository
type mockAlumniRepository struct {
	alumni map[string]*model.Alumni
}

func newMockAlumniRepository() *mockAlumniRepository {
	return &mockAlumniRepository{
		alumni: make(map[string]*model.Alumni),
	}
}

func (m *mockAlumniRepository) List(ctx context.Context) ([]model.Alumni, error) {
	result := make([]model.Alumni, 0, len(m.alumni))
	for _, a := range m.alumni {
		result = append(result, *a)
	}
	return result, nil
}

func (m *mockAlumniRepository) GetByID(ctx context.Context, id string) (*model.Alumni, error) {
	_, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	if a, ok := m.alumni[id]; ok {
		return a, nil
	}
	return nil, nil
}

func (m *mockAlumniRepository) GetByUserID(ctx context.Context, userID string) (*model.Alumni, error) {
	for _, a := range m.alumni {
		if a.UserID.Hex() == userID {
			return a, nil
		}
	}
	return nil, nil
}

func (m *mockAlumniRepository) Create(ctx context.Context, in model.CreateAlumniReq, userID string) (string, error) {
	uid, _ := primitive.ObjectIDFromHex(userID)
	newID := primitive.NewObjectID()
	alumni := &model.Alumni{
		ID:         newID,
		UserID:     uid,
		NIM:        in.NIM,
		Nama:       in.Nama,
		Jurusan:    in.Jurusan,
		Angkatan:   in.Angkatan,
		TahunLulus: in.TahunLulus,
		Email:      in.Email,
		NoTelepon:  in.NoTelepon,
		Alamat:     in.Alamat,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	m.alumni[newID.Hex()] = alumni
	return newID.Hex(), nil
}

func (m *mockAlumniRepository) Update(ctx context.Context, id string, in model.UpdateAlumniReq) error {
	if _, ok := m.alumni[id]; !ok {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (m *mockAlumniRepository) Delete(ctx context.Context, id string) error {
	if _, ok := m.alumni[id]; !ok {
		return mongo.ErrNoDocuments
	}
	delete(m.alumni, id)
	return nil
}

func (m *mockAlumniRepository) ListByJurusan(ctx context.Context, jurusan string) ([]model.Alumni, error) {
	result := make([]model.Alumni, 0)
	for _, a := range m.alumni {
		if a.Jurusan == jurusan {
			result = append(result, *a)
		}
	}
	return result, nil
}

func (m *mockAlumniRepository) ListPaged(ctx context.Context, search, sortBy, order string, limit, offset int) ([]model.Alumni, error) {
	result := make([]model.Alumni, 0, len(m.alumni))
	for _, a := range m.alumni {
		if search == "" || containsString(a.Nama, search) || containsString(a.Email, search) {
			result = append(result, *a)
		}
	}

	start := offset
	end := offset + limit
	if start > len(result) {
		return []model.Alumni{}, nil
	}
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], nil
}

func (m *mockAlumniRepository) Count(ctx context.Context, search string) (int64, error) {
	count := int64(0)
	for _, a := range m.alumni {
		if search == "" || containsString(a.Nama, search) || containsString(a.Email, search) {
			count++
		}
	}
	return count, nil
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0)
}

func TestAlumniService_List(t *testing.T) {
	app := fiber.New()
	repo := newMockAlumniRepository()
	service := NewAlumniService(repo)

	// test data
	alumniID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	repo.alumni[alumniID.Hex()] = &model.Alumni{
		ID:       alumniID,
		UserID:   userID,
		NIM:      "12345",
		Nama:     "Test Alumni",
		Jurusan:  "Teknik Informatika",
		Email:    "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	app.Get("/alumni", service.List)

	req := httptest.NewRequest("GET", "/alumni", nil)

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
}

func TestAlumniService_Get(t *testing.T) {
	app := fiber.New()
	repo := newMockAlumniRepository()
	service := NewAlumniService(repo)

	alumniID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	repo.alumni[alumniID.Hex()] = &model.Alumni{
		ID:       alumniID,
		UserID:   userID,
		NIM:      "12345",
		Nama:     "Test Alumni",
		Jurusan:  "Teknik Informatika",
		Email:    "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	app.Get("/alumni/:id", service.Get)

	tests := []struct {
		name           string
		id             string
		expectedStatus int
	}{
		{
			name:           "valid id",
			id:             alumniID.Hex(),
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "invalid id",
			id:             "invalid",
			expectedStatus: 500,
		},
		{
			name:           "not found",
			id:             primitive.NewObjectID().Hex(),
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/alumni/"+tt.id, nil)

			resp, err := app.Test(req)
			utils.AssertEqual(t, nil, err)
			utils.AssertEqual(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestAlumniService_Create(t *testing.T) {
	app := fiber.New()
	repo := newMockAlumniRepository()
	service := NewAlumniService(repo)

	userID := primitive.NewObjectID()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", userID.Hex())
		return c.Next()
	})

	app.Post("/alumni", service.Create)

	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
	}{
		{
			name: "valid request",
			body: map[string]interface{}{
				"nim":      "12345",
				"nama":     "Test Alumni",
				"jurusan":  "Teknik Informatika",
				"email":    "test@example.com",
				"angkatan": 2020,
			},
			expectedStatus: fiber.StatusCreated,
		},
		{
			name: "missing required fields",
			body: map[string]interface{}{
				"nim": "12345",
			},
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			name:           "empty body",
			body:           map[string]interface{}{},
			expectedStatus: fiber.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/alumni", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			utils.AssertEqual(t, nil, err)
			utils.AssertEqual(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestAlumniService_Update(t *testing.T) {
	app := fiber.New()
	repo := newMockAlumniRepository()
	service := NewAlumniService(repo)

	alumniID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	repo.alumni[alumniID.Hex()] = &model.Alumni{
		ID:       alumniID,
		UserID:   userID,
		NIM:      "12345",
		Nama:     "Test Alumni",
		Jurusan:  "Teknik Informatika",
		Email:    "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	app.Put("/alumni/:id", service.Update)

	tests := []struct {
		name           string
		id             string
		body           map[string]interface{}
		expectedStatus int
	}{
		{
			name: "valid update",
			id:   alumniID.Hex(),
			body: map[string]interface{}{
				"nama": "Updated Name",
			},
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "invalid id",
			id:             "invalid",
			body:           map[string]interface{}{"nama": "Test"},
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("PUT", "/alumni/"+tt.id, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			utils.AssertEqual(t, nil, err)
			utils.AssertEqual(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestAlumniService_Delete(t *testing.T) {
	app := fiber.New()
	repo := newMockAlumniRepository()
	service := NewAlumniService(repo)

	alumniID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	repo.alumni[alumniID.Hex()] = &model.Alumni{
		ID:       alumniID,
		UserID:   userID,
		NIM:      "12345",
		Nama:     "Test Alumni",
		Jurusan:  "Teknik Informatika",
		Email:    "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	app.Delete("/alumni/:id", service.Delete)

	tests := []struct {
		name           string
		id             string
		expectedStatus int
	}{
		{
			name:           "valid delete",
			id:             alumniID.Hex(),
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "invalid id",
			id:             "invalid",
			expectedStatus: 500,
		},
		{
			name:           "not found",
			id:             primitive.NewObjectID().Hex(),
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/alumni/"+tt.id, nil)

			resp, err := app.Test(req)
			utils.AssertEqual(t, nil, err)
			utils.AssertEqual(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestAlumniService_ListByJurusan(t *testing.T) {
	app := fiber.New()
	repo := newMockAlumniRepository()
	service := NewAlumniService(repo)

	alumniID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	repo.alumni[alumniID.Hex()] = &model.Alumni{
		ID:       alumniID,
		UserID:   userID,
		NIM:      "12345",
		Nama:     "Test Alumni",
		Jurusan:  "Teknik Informatika",
		Email:    "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	app.Get("/alumni/jurusan/:jurusan", service.ListByJurusan)

	tests := []struct {
		name           string
		jurusan        string
		expectedStatus int
	}{
		{
			name:           "valid jurusan",
			jurusan:        "Teknik Informatika",
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "valid jurusan with special chars",
			jurusan:        "Teknik%20Informatika",
			expectedStatus: fiber.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encodedJurusan := url.PathEscape(tt.jurusan)
			req := httptest.NewRequest("GET", "/alumni/jurusan/"+encodedJurusan, nil)

			resp, err := app.Test(req)
			utils.AssertEqual(t, nil, err)
			utils.AssertEqual(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

