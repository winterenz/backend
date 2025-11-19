package service

import (
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"prak/clean-architecture-fiber-mongo/app/model"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
)

// mock FileRepository
type mockFileRepository struct {
	files map[string]*model.FileDoc
}

func newMockFileRepository() *mockFileRepository {
	return &mockFileRepository{
		files: make(map[string]*model.FileDoc),
	}
}

func (m *mockFileRepository) Create(ctx context.Context, f *model.FileDoc) error {
	if f.ID.IsZero() {
		f.ID = primitive.NewObjectID()
	}
	m.files[f.ID.Hex()] = f
	return nil
}

func (m *mockFileRepository) FindAll(ctx context.Context) ([]model.FileDoc, error) {
	result := make([]model.FileDoc, 0, len(m.files))
	for _, f := range m.files {
		result = append(result, *f)
	}
	return result, nil
}

func (m *mockFileRepository) FindByID(ctx context.Context, id string) (*model.FileDoc, error) {
	if f, ok := m.files[id]; ok {
		return f, nil
	}
	return nil, os.ErrNotExist
}

func (m *mockFileRepository) Delete(ctx context.Context, id string) error {
	if _, ok := m.files[id]; !ok {
		return os.ErrNotExist
	}
	delete(m.files, id)
	return nil
}

func TestFileService_GetAllFiles(t *testing.T) {
	app := fiber.New()
	repo := newMockFileRepository()
	tempDir := t.TempDir()
	service := NewFileService(repo, tempDir)

	// Add test data
	fileID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	repo.files[fileID.Hex()] = &model.FileDoc{
		ID:           fileID,
		OwnerUserID:  userID,
		Category:     model.CategoryFoto,
		FileName:     "test.jpg",
		OriginalName: "test.jpg",
		FilePath:     "/test/path",
		FileSize:     1024,
		FileType:     "image/jpeg",
		UploadedAt:   time.Now(),
	}

	app.Get("/files", service.GetAllFiles)

	req := httptest.NewRequest("GET", "/files", nil)

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode)
}

func TestFileService_GetFileByID(t *testing.T) {
	app := fiber.New()
	repo := newMockFileRepository()
	tempDir := t.TempDir()
	service := NewFileService(repo, tempDir)

	fileID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	repo.files[fileID.Hex()] = &model.FileDoc{
		ID:           fileID,
		OwnerUserID:  userID,
		Category:     model.CategoryFoto,
		FileName:     "test.jpg",
		OriginalName: "test.jpg",
		FilePath:     "/test/path",
		FileSize:     1024,
		FileType:     "image/jpeg",
		UploadedAt:   time.Now(),
	}

	app.Get("/files/:id", service.GetFileByID)

	tests := []struct {
		name           string
		id             string
		expectedStatus int
	}{
		{
			name:           "valid id",
			id:             fileID.Hex(),
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
			req := httptest.NewRequest("GET", "/files/"+tt.id, nil)

			resp, err := app.Test(req)
			utils.AssertEqual(t, nil, err)
			utils.AssertEqual(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestFileService_DeleteFile(t *testing.T) {
	app := fiber.New()
	repo := newMockFileRepository()
	tempDir := t.TempDir()
	service := NewFileService(repo, tempDir)

	fileID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	filePath := filepath.Join(tempDir, "test.jpg")
	repo.files[fileID.Hex()] = &model.FileDoc{
		ID:           fileID,
		OwnerUserID:  userID,
		Category:     model.CategoryFoto,
		FileName:     "test.jpg",
		OriginalName: "test.jpg",
		FilePath:     filePath,
		FileSize:     1024,
		FileType:     "image/jpeg",
		UploadedAt:   time.Now(),
	}

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", userID.Hex())
		c.Locals("role", "user")
		return c.Next()
	})

	app.Delete("/files/:id", service.DeleteFile)

	tests := []struct {
		name           string
		id             string
		expectedStatus int
	}{
		{
			name:           "valid delete",
			id:             fileID.Hex(),
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
			req := httptest.NewRequest("DELETE", "/files/"+tt.id, nil)

			resp, err := app.Test(req)
			utils.AssertEqual(t, nil, err)
			utils.AssertEqual(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestFileService_DeleteFile_Forbidden(t *testing.T) {
	app := fiber.New()
	repo := newMockFileRepository()
	tempDir := t.TempDir()
	service := NewFileService(repo, tempDir)

	fileID := primitive.NewObjectID()
	ownerID := primitive.NewObjectID()
	otherUserID := primitive.NewObjectID()
	repo.files[fileID.Hex()] = &model.FileDoc{
		ID:           fileID,
		OwnerUserID:  ownerID,
		Category:     model.CategoryFoto,
		FileName:     "test.jpg",
		OriginalName: "test.jpg",
		FilePath:     "/test/path",
		FileSize:     1024,
		FileType:     "image/jpeg",
		UploadedAt:   time.Now(),
	}

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", otherUserID.Hex())
		c.Locals("role", "user")
		return c.Next()
	})

	app.Delete("/files/:id", service.DeleteFile)

	req := httptest.NewRequest("DELETE", "/files/"+fileID.Hex(), nil)

	resp, err := app.Test(req)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, fiber.StatusForbidden, resp.StatusCode)
}

