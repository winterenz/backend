package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"prak/clean-architecture-fiber-mongo/app/model"
	"prak/clean-architecture-fiber-mongo/app/repository"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
// @Description Upload foto. admin dapat upload untuk semua user_id, sedangkan user hanya dapat upload untuk dirinya sendiri.
// @Tags File
// @Accept multipart/form-data
// @Produce json
// @Param user_id path string true "User ID (Admin: bisa upload untuk semua user_id, User: gunakan 'me' atau user_id sendiri)"
// @Param file formData file true "Foto (maks 1MB, format: jpeg/png/jpg)"
// @Success 201 {object} model.FileUploadResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /users/{user_id}/upload/foto [post]
func (s *FileService) UploadFoto(c *fiber.Ctx) error {
	return s.uploadWithRule(c, model.CategoryFoto, 1*1024*1024, []string{"image/jpeg", "image/png", "image/jpg"})
}

// Upload Sertifikat godoc
// @Summary Upload sertifikat
// @Description Upload sertifikat. admin dapat upload untuk semua user_id, sedangkan user hanya dapat upload untuk dirinya sendiri.
// @Tags File
// @Accept multipart/form-data
// @Produce json
// @Param user_id path string true "User ID (Admin: bisa upload untuk semua user_id, User: gunakan 'me' atau user_id sendiri)"
// @Param file formData file true "Sertifikat (maks 2MB, format: pdf)"
// @Success 201 {object} model.FileUploadResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /users/{user_id}/upload/sertifikat [post]
func (s *FileService) UploadSertifikat(c *fiber.Ctx) error {
	return s.uploadWithRule(c, model.CategorySertifikat, 2*1024*1024, []string{"application/pdf"})
}

func (s *FileService) uploadWithRule(c *fiber.Ctx, category string, maxSize int64, allowed []string) error {
	targetUserID := strings.TrimSpace(c.Params("user_id"))
	authUserID, _ := c.Locals("user_id").(string)
	role, _ := c.Locals("role").(string)

	if targetUserID == "" || targetUserID == "me" {
		targetUserID = authUserID
	} else if role != "admin" && targetUserID != authUserID {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Success: false,
			Message: "anda tidak berhak mengupload file orang lain ya :p",
		})
	}

	if targetUserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "user_id tidak valid",
		})
	}

	ownerOID, err := primitive.ObjectIDFromHex(targetUserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: fmt.Sprintf("user_id tidak valid: %s", err.Error()),
		})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "tidak ada file yang diupload",
			Error:   err.Error(),
		})
	}
	if fileHeader.Size > maxSize {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: fmt.Sprintf("ukuran file terlalu besar (maks %d bytes)", maxSize),
		})
	}

	ct := fileHeader.Header.Get("Content-Type")
	ok := false
	for _, t := range allowed {
		if t == ct {
			ok = true
			break
		}
	}
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Success: false,
			Message: "tipe file yang di upload tidak sesuai",
		})
	}

	ext := filepath.Ext(fileHeader.Filename)
	newName := uuid.New().String() + ext
	dir := filepath.Join(s.uploadRoot, category, targetUserID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "gagal membuat direktori",
		})
	}
	dst := filepath.Join(dir, newName)
	if err := c.SaveFile(fileHeader, dst); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "gagal menyimpan file",
			Error:   err.Error(),
		})
	}

	doc := &model.FileDoc{
		OwnerUserID:  ownerOID, // Sudah divalidasi dan dikonversi di atas
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
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Success: false,
			Message: "gagal menyimpan metadata",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(model.FileUploadResponse{
		Success: true,
		Message: "file berhasil di upload",
		Data: model.FileUploadData{
			ID:   doc.ID.Hex(),
			Path: fmt.Sprintf("/uploads/%s/%s/%s", category, targetUserID, newName),
		},
	})
}

// Get All Files godoc
// @Summary get all file
// @Description mengambil semua file dari database
// @Tags File
// @Produce json
// @Success 200 {object} model.FileListResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /files [get]
func (s *FileService) GetAllFiles(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	items, err := s.repo.FindAll(ctx)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: "gagal memuat file",
		})
	}
	return c.JSON(model.FileListResponse{
		Success: true,
		Data:    items,
	})
}

// Get File By ID godoc
// @Summary get file by ID
// @Description mengambil file berdasarkan ID
// @Tags File
// @Produce json
// @Param id path string true "File ID (hex)"
// @Success 200 {object} model.FileResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /files/{id} [get]
func (s *FileService) GetFileByID(c *fiber.Ctx) error {
	id := c.Params("id")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Success: false,
			Message: "file tidak ditemukan",
		})
	}
	return c.JSON(model.FileResponse{
		Success: true,
		Data:    item,
	})
}

// Delete File godoc
// @Summary delete file
// @Description menghapus file berdasarkan id. admin dapat menghapus semua file, sedangkan user hanya dapat menghapus file miliknya sendiri.
// @Tags File
// @Produce json
// @Param id path string true "File ID (hex)"
// @Success 200 {object} model.SuccessMessageResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /files/{id} [delete]
func (s *FileService) DeleteFile(c *fiber.Ctx) error {
	id := c.Params("id")
	authUserID, _ := c.Locals("user_id").(string)
	role, _ := c.Locals("role").(string)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Success: false,
			Message: "file tidak ditemukan",
		})
	}

	// Validasi: user biasa hanya bisa hapus file miliknya sendiri
	if role != "admin" {
		authOID, err := primitive.ObjectIDFromHex(authUserID)
		if err != nil || item.OwnerUserID != authOID {
			return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
				Success: false,
				Message: "anda tidak berhak menghapus file ini ya :p",
			})
		}
	}

	_ = os.Remove(item.FilePath)
	if err := s.repo.Delete(ctx, id); err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: "gagal menghapus file",
		})
	}
	return c.JSON(model.SuccessMessageResponse{
		Success: true,
		Message: "file telah dihapus",
	})
}
