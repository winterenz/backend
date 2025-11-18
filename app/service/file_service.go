// app/service/file_service.go
package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"prak3/clean-architecture-fiber-mongo/app/model"
	"prak3/clean-architecture-fiber-mongo/app/repository"
)

type FileService struct {
	repo       repository.FileRepository
	uploadRoot string
}

func NewFileService(r repository.FileRepository, root string) *FileService {
	return &FileService{repo: r, uploadRoot: root}
}

// Upload Foto godoc
// @Summary Upload foto
// @Description Mengupload foto ke server
// @Tags File
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Foto"
// @Success 201 {object} map[string]interface{} // { success, message, data }
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /users/:user_id/upload/foto [post]
func (s *FileService) UploadFoto(c *fiber.Ctx) error {
	return s.uploadWithRule(c, model.CategoryFoto, 1*1024*1024, []string{"image/jpeg", "image/png", "image/jpg"})
}

// Upload Sertifikat godoc
// @Summary Upload sertifikat
// @Description Mengupload sertifikat ke server
// @Tags File
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Sertifikat"
// @Success 201 {object} map[string]interface{} // { success, message, data }
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /users/:user_id/upload/sertifikat [post]
func (s *FileService) UploadSertifikat(c *fiber.Ctx) error {
	return s.uploadWithRule(c, model.CategorySertifikat, 2*1024*1024, []string{"application/pdf"})
}

// Upload With Rule godoc
// @Summary Upload file dengan aturan
// @Description Mengupload file dengan aturan tertentu
// @Tags File
// @Accept multipart/form-data
// @Produce json
// @Param category query string true "Kategori file"
// @Param maxSize query int false "Ukuran maksimal file (bytes)" default(2097152)
// @Param allowed query []string false "Jenis file yang diizinkan" default([]string{"image/jpeg", "image/png", "image/jpg", "application/pdf"})
// @Success 201 {object} map[string]interface{} // { success, message, data }
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /users/:user_id/upload/with-rule [post]
func (s *FileService) uploadWithRule(c *fiber.Ctx, category string, maxSize int64, allowed []string) error {
	targetUserID := strings.TrimSpace(c.Params("user_id"))
	authUserID, _ := c.Locals("user_id").(string)
	role, _ := c.Locals("role").(string)

	// user biasa hanya boleh ke dirinya sendiri
	if targetUserID == "" {
		targetUserID = authUserID
	} else if role != "admin" && targetUserID != authUserID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "Tidak boleh upload untuk user lain"})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "No file uploaded", "error": err.Error()})
	}
	if fileHeader.Size > maxSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": fmt.Sprintf("File too large (max %d bytes)", maxSize)})
	}

	ct := fileHeader.Header.Get("Content-Type")
	ok := false
	for _, t := range allowed { if t == ct { ok = true; break } }
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "File type not allowed"})
	}

	ext := filepath.Ext(fileHeader.Filename)
	newName := uuid.New().String() + ext
	dir := filepath.Join(s.uploadRoot, category, targetUserID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Cannot create directory"})
	}
	dst := filepath.Join(dir, newName)
	if err := c.SaveFile(fileHeader, dst); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to save file", "error": err.Error()})
	}

	var ownerOID primitive.ObjectID
	if id, err := primitive.ObjectIDFromHex(targetUserID); err == nil { ownerOID = id }

	doc := &model.FileDoc{
		OwnerUserID:  ownerOID,
		Category:     category,
		FileName:     newName,
		OriginalName: fileHeader.Filename,
		FilePath:     dst,
		FileSize:     fileHeader.Size,
		FileType:     ct,
		UploadedAt:   time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.repo.Create(ctx, doc); err != nil {
		_ = os.Remove(dst)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Failed to save metadata"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "The file has been uploaded",
		"data": fiber.Map{
			"id":   doc.ID.Hex(),
			"path": fmt.Sprintf("/uploads/%s/%s/%s", category, targetUserID, newName),
		},
	})
}

// Get All Files godoc
// @Summary Get all files
// @Description Mengambil semua file dari database
// @Tags File
// @Produce json
// @Success 200 {object} map[string]interface{} // { success, data }
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /files [get]
func (s *FileService) GetAllFiles(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second); defer cancel()
	items, err := s.repo.FindAll(ctx)
	if err != nil { return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to get files"}) }
	return c.JSON(fiber.Map{"success": true, "data": items})
}

// Get File By ID godoc
// @Summary Get file by ID
// @Description Mengambil file berdasarkan ID
// @Tags File
// @Produce json
// @Param id path string true "File ID (hex)"
// @Success 200 {object} map[string]interface{} // { success, data }
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /files/{id} [get]
func (s *FileService) GetFileByID(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second); defer cancel()
	item, err := s.repo.FindByID(ctx, id)
	if err != nil { return c.Status(404).JSON(fiber.Map{"success": false, "message": "File not found"}) }
	return c.JSON(fiber.Map{"success": true, "data": item})
}

// Delete File godoc
// @Summary Delete file
// @Description Menghapus file berdasarkan ID
// @Tags File
// @Produce json
// @Param id path string true "File ID (hex)"
// @Success 200 {object} map[string]interface{} // { success, message }
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /files/{id} [delete]
func (s *FileService) DeleteFile(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second); defer cancel()
	item, err := s.repo.FindByID(ctx, id)
	if err != nil { return c.Status(404).JSON(fiber.Map{"success": false, "message": "File not found"}) }
	_ = os.Remove(item.FilePath)
	if err := s.repo.Delete(ctx, id); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to delete file"})
	}
	return c.JSON(fiber.Map{"success": true, "message": "File deleted"})
}