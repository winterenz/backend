package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"prak/clean-architecture-fiber-mongo/app/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
)

// mock PekerjaanRepository
type mockPekerjaanRepository struct {
	pekerjaan map[string]*model.Pekerjaan
	alumni    map[string]*model.Alumni
}

func newMockPekerjaanRepository() *mockPekerjaanRepository {
	return &mockPekerjaanRepository{
		pekerjaan: make(map[string]*model.Pekerjaan),
		alumni:    make(map[string]*model.Alumni),
	}
}

func (m *mockPekerjaanRepository) List(ctx context.Context) ([]model.Pekerjaan, error) {
	result := make([]model.Pekerjaan, 0, len(m.pekerjaan))
	for _, p := range m.pekerjaan {
		if !p.IsDeleted {
			result = append(result, *p)
		}
	}
	return result, nil
}

func (m *mockPekerjaanRepository) GetByID(ctx context.Context, id string) (*model.Pekerjaan, error) {
	if p, ok := m.pekerjaan[id]; ok && !p.IsDeleted {
		return p, nil
	}
	return nil, nil
}

func (m *mockPekerjaanRepository) GetByIDAny(ctx context.Context, id string) (*model.Pekerjaan, error) {
	if p, ok := m.pekerjaan[id]; ok {
		return p, nil
	}
	return nil, nil
}

func (m *mockPekerjaanRepository) ListByAlumniID(ctx context.Context, alumniID string) ([]model.Pekerjaan, error) {
	result := make([]model.Pekerjaan, 0)
	aid, _ := primitive.ObjectIDFromHex(alumniID)
	for _, p := range m.pekerjaan {
		if p.AlumniID == aid && !p.IsDeleted {
			result = append(result, *p)
		}
	}
	return result, nil
}

func (m *mockPekerjaanRepository) Create(ctx context.Context, in model.CreatePekerjaanReq) (string, error) {
	aid, _ := primitive.ObjectIDFromHex(in.AlumniID)
	newID := primitive.NewObjectID()
	pekerjaan := &model.Pekerjaan{
		ID:                newID,
		AlumniID:          aid,
		NamaPerusahaan:    in.NamaPerusahaan,
		PosisiJabatan:     in.PosisiJabatan,
		BidangIndustri:    in.BidangIndustri,
		LokasiKerja:       in.LokasiKerja,
		GajiRange:         in.GajiRange,
		TanggalMulaiKerja: in.TanggalMulaiKerja,
		TanggalSelesaiKerja: in.TanggalSelesaiKerja,
		StatusPekerjaan:   in.StatusPekerjaan,
		Deskripsi:         in.Deskripsi,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		IsDeleted:         false,
	}
	m.pekerjaan[newID.Hex()] = pekerjaan
	return newID.Hex(), nil
}

func (m *mockPekerjaanRepository) Update(ctx context.Context, id string, in model.UpdatePekerjaanReq) error {
	if _, ok := m.pekerjaan[id]; !ok {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (m *mockPekerjaanRepository) ListPaged(ctx context.Context, search, sortBy, order string, limit, offset int) ([]model.Pekerjaan, error) {
	result := make([]model.Pekerjaan, 0)
	for _, p := range m.pekerjaan {
		if !p.IsDeleted {
			if search == "" || containsPek(p.NamaPerusahaan, search) {
				result = append(result, *p)
			}
		}
	}
	start := offset
	end := offset + limit
	if start > len(result) {
		return []model.Pekerjaan{}, nil
	}
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], nil
}

func (m *mockPekerjaanRepository) Count(ctx context.Context, search string) (int64, error) {
	count := int64(0)
	for _, p := range m.pekerjaan {
		if !p.IsDeleted {
			if search == "" || containsPek(p.NamaPerusahaan, search) {
				count++
			}
		}
	}
	return count, nil
}

func (m *mockPekerjaanRepository) SoftDeleteAdmin(ctx context.Context, id string, username string) error {
	if p, ok := m.pekerjaan[id]; ok && !p.IsDeleted {
		p.IsDeleted = true
		p.DeletedBy = &username
		now := time.Now()
		p.DeletedAt = &now
		return nil
	}
	return mongo.ErrNoDocuments
}

func (m *mockPekerjaanRepository) SoftDeleteOwned(ctx context.Context, id string, alumniID string, username string) error {
	if p, ok := m.pekerjaan[id]; ok && !p.IsDeleted {
		aid, _ := primitive.ObjectIDFromHex(alumniID)
		if p.AlumniID == aid {
			p.IsDeleted = true
			p.DeletedBy = &username
			now := time.Now()
			p.DeletedAt = &now
			return nil
		}
	}
	return mongo.ErrNoDocuments
}

func (m *mockPekerjaanRepository) ListTrashed(ctx context.Context, search, sortBy, order string, limit, offset int) ([]model.Pekerjaan, error) {
	result := make([]model.Pekerjaan, 0)
	for _, p := range m.pekerjaan {
		if p.IsDeleted {
			if search == "" || containsPek(p.NamaPerusahaan, search) {
				result = append(result, *p)
			}
		}
	}
	start := offset
	end := offset + limit
	if start > len(result) {
		return []model.Pekerjaan{}, nil
	}
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], nil
}

func (m *mockPekerjaanRepository) CountTrashed(ctx context.Context, search string) (int64, error) {
	count := int64(0)
	for _, p := range m.pekerjaan {
		if p.IsDeleted {
			if search == "" || containsPek(p.NamaPerusahaan, search) {
				count++
			}
		}
	}
	return count, nil
}

func (m *mockPekerjaanRepository) RestoreAdmin(ctx context.Context, id string, username string) error {
	if p, ok := m.pekerjaan[id]; ok && p.IsDeleted {
		p.IsDeleted = false
		p.DeletedBy = nil
		p.DeletedAt = nil
		return nil
	}
	return mongo.ErrNoDocuments
}

func (m *mockPekerjaanRepository) RestoreOwned(ctx context.Context, id string, alumniID string, username string) error {
	if p, ok := m.pekerjaan[id]; ok && p.IsDeleted {
		aid, _ := primitive.ObjectIDFromHex(alumniID)
		if p.AlumniID == aid {
			p.IsDeleted = false
			p.DeletedBy = nil
			p.DeletedAt = nil
			return nil
		}
	}
	return mongo.ErrNoDocuments
}

func (m *mockPekerjaanRepository) HardDeleteAdmin(ctx context.Context, id string) error {
	if _, ok := m.pekerjaan[id]; !ok {
		return mongo.ErrNoDocuments
	}
	delete(m.pekerjaan, id)
	return nil
}

func (m *mockPekerjaanRepository) HardDeleteOwned(ctx context.Context, id string, alumniID string) error {
	if p, ok := m.pekerjaan[id]; ok {
		aid, _ := primitive.ObjectIDFromHex(alumniID)
		if p.AlumniID == aid {
			delete(m.pekerjaan, id)
			return nil
		}
	}
	return mongo.ErrNoDocuments
}

func (m *mockPekerjaanRepository) GetAlumniByUserID(ctx context.Context, userID string) (*model.Alumni, error) {
	for _, a := range m.alumni {
		if a.UserID.Hex() == userID {
			return a, nil
		}
	}
	return nil, nil
}

func containsPek(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0)
}

func TestPekerjaanService_List(t *testing.T) {
	app := fiber.New()
	repo := newMockPekerjaanRepository()
	service := NewPekerjaanService(repo)

	pekerjaanID := primitive.NewObjectID()
	alumniID := primitive.NewObjectID()
	repo.pekerjaan[pekerjaanID.Hex()] = &model.Pekerjaan{
		ID:              pekerjaanID,
		AlumniID:        alumniID,
		NamaPerusahaan:  "Test Company",
		PosisiJabatan:   "Developer",
		BidangIndustri:  "IT",
		LokasiKerja:     "Jakarta",
		StatusPekerjaan: "active",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		IsDeleted:       false,
	}

	app.Get("/pekerjaan", service.List)

	req := httptest.NewRequest("GET", "/pekerjaan", nil)

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
}

func TestPekerjaanService_Get(t *testing.T) {
	app := fiber.New()
	repo := newMockPekerjaanRepository()
	service := NewPekerjaanService(repo)

	pekerjaanID := primitive.NewObjectID()
	alumniID := primitive.NewObjectID()
	repo.pekerjaan[pekerjaanID.Hex()] = &model.Pekerjaan{
		ID:              pekerjaanID,
		AlumniID:        alumniID,
		NamaPerusahaan:  "Test Company",
		PosisiJabatan:   "Developer",
		BidangIndustri:  "IT",
		LokasiKerja:     "Jakarta",
		StatusPekerjaan: "active",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		IsDeleted:       false,
	}

	app.Get("/pekerjaan/:id", service.Get)

	tests := []struct {
		name           string
		id             string
		expectedStatus int
	}{
		{
			name:           "valid id",
			id:             pekerjaanID.Hex(),
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "not found",
			id:             primitive.NewObjectID().Hex(),
			expectedStatus: fiber.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/pekerjaan/"+tt.id, nil)

			resp, err := app.Test(req)
			utils.AssertEqual(t, nil, err)
			utils.AssertEqual(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestPekerjaanService_Create(t *testing.T) {
	app := fiber.New()
	repo := newMockPekerjaanRepository()
	service := NewPekerjaanService(repo)

	alumniID := primitive.NewObjectID()

	app.Post("/pekerjaan", service.Create)

	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
	}{
		{
			name: "valid request",
			body: map[string]interface{}{
				"alumni_id":         alumniID.Hex(),
				"nama_perusahaan":   "Test Company",
				"posisi_jabatan":    "Developer",
				"bidang_industri":   "IT",
				"lokasi_kerja":      "Jakarta",
				"status_pekerjaan":  "active",
				"tanggal_mulai_kerja": time.Now().Format(time.RFC3339),
			},
			expectedStatus: fiber.StatusCreated,
		},
		{
			name: "missing required fields",
			body: map[string]interface{}{
				"alumni_id": alumniID.Hex(),
			},
			expectedStatus: fiber.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/pekerjaan", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			utils.AssertEqual(t, nil, err)
			utils.AssertEqual(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestPekerjaanService_SoftDelete(t *testing.T) {
	app := fiber.New()
	repo := newMockPekerjaanRepository()
	service := NewPekerjaanService(repo)

	pekerjaanID := primitive.NewObjectID()
	alumniID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	repo.pekerjaan[pekerjaanID.Hex()] = &model.Pekerjaan{
		ID:              pekerjaanID,
		AlumniID:        alumniID,
		NamaPerusahaan:  "Test Company",
		PosisiJabatan:   "Developer",
		BidangIndustri:  "IT",
		LokasiKerja:     "Jakarta",
		StatusPekerjaan: "active",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		IsDeleted:       false,
	}
	repo.alumni[alumniID.Hex()] = &model.Alumni{
		ID:     alumniID,
		UserID: userID,
		Nama:   "Test Alumni",
	}

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", userID.Hex())
		c.Locals("username", "testuser")
		c.Locals("role", "user")
		return c.Next()
	})

	app.Delete("/pekerjaan/:id", service.SoftDelete)

	req := httptest.NewRequest("DELETE", "/pekerjaan/"+pekerjaanID.Hex(), nil)

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
}

