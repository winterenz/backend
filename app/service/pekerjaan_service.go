package service

import (
	"context"
	"errors"
	"time"
    "strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/bson/primitive"
	"prak3/clean-architecture-fiber-mongo/app/model"
	"prak3/clean-architecture-fiber-mongo/app/repository"
)

type PekerjaanService struct {
	Repo repository.PekerjaanRepository
}

func NewPekerjaanService(repo repository.PekerjaanRepository) *PekerjaanService {
	return &PekerjaanService{Repo: repo}
}

// List Pekerjaan godoc
// @Summary Daftar pekerjaan
// @Description Mengambil daftar pekerjaan dengan pagination, sort, dan pencarian
// @Tags Pekerjaan
// @Accept json
// @Produce json
// @Param page query int false "Halaman" default(1)
// @Param limit query int false "Jumlah per halaman" default(10)
// @Param sortBy query string false "Kolom sort (default: created_at)"
// @Param order query string false "asc/desc (default: desc)"
// @Param search query string false "Kata kunci pencarian"
// @Success 200 {object} model.PekerjaanListResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /pekerjaan [get]
func (s *PekerjaanService) List(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	sortBy := c.Query("sortBy", "created_at")
	order := c.Query("order", "desc")
	search := c.Query("search", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	items, err := s.Repo.ListPaged(ctx, search, sortBy, order, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	total, err := s.Repo.Count(ctx, search)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.JSON(model.PekerjaanListResponse{
		Data: items,
		Meta: model.MetaInfo{
			Page:   page,
			Limit:  limit,
			Total:  total,
			Pages:  int((total + int64(limit) - 1) / int64(limit)),
			SortBy: sortBy,
			Order:  order,
			Search: search,
		},
	})
}

// Get Pekerjaan godoc
// @Summary Detail pekerjaan
// @Description Mengambil detail pekerjaan berdasarkan ID (hex string)
// @Tags Pekerjaan
// @Produce json
// @Param id path string true "Pekerjaan ID (hex)"
// @Success 200 {object} map[string]interface{}  // { success, data }
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /pekerjaan/{id} [get]
func (s *PekerjaanService) Get(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := c.Params("id")
	item, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	if item == nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "message": "Pekerjaan tidak ditemukan"})
	}
	return c.JSON(fiber.Map{"success": true, "data": item})
}

// List By Alumni godoc
// @Summary Daftar pekerjaan berdasarkan alumni
// @Description Mengambil daftar pekerjaan berdasarkan alumni ID
// @Tags Pekerjaan
// @Produce json
// @Param alumni_id path string true "Alumni ID (hex)"
// @Success 200 {object} map[string]interface{} // { success, data }
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /pekerjaan/alumni/{alumni_id} [get]
func (s *PekerjaanService) ListByAlumni(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	alumniID := c.Params("alumni_id")
	data, err := s.Repo.ListByAlumniID(ctx, alumniID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": data})
}

// Create Pekerjaan godoc
// @Summary Tambah pekerjaan
// @Description Menambahkan data pekerjaan baru (alumni_id dari JWT)
// @Tags Pekerjaan
// @Accept json
// @Produce json
// @Param body body model.CreatePekerjaanReq true "Body"
// @Success 201 {object} map[string]interface{} // { success, data }
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /pekerjaan [post]
func (s *PekerjaanService) Create(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    var in model.CreatePekerjaanReq
    if err := c.BodyParser(&in); err != nil {
        return c.Status(400).JSON(fiber.Map{"success": false, "message": "Body tidak valid"})
    }

    if strings.TrimSpace(in.AlumniID) == "" {
        return c.Status(400).JSON(fiber.Map{"success": false, "message": "alumni_id wajib"})
    }
    if _, err := primitive.ObjectIDFromHex(in.AlumniID); err != nil {
        return c.Status(400).JSON(fiber.Map{"success": false, "message": "alumni_id tidak valid (harus ObjectID hex)"})
    }

    if strings.TrimSpace(in.NamaPerusahaan) == "" ||
        strings.TrimSpace(in.PosisiJabatan) == "" ||
        strings.TrimSpace(in.BidangIndustri) == "" ||
        strings.TrimSpace(in.LokasiKerja) == "" {
        return c.Status(400).JSON(fiber.Map{"success": false, "message": "Field wajib belum lengkap"})
    }

    id, err := s.Repo.Create(ctx, in) // id string (hex) dikembalikan repo
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
    }

    item, _ := s.Repo.GetByID(ctx, id)
    return c.Status(201).JSON(fiber.Map{"success": true, "data": item})
}

// Update Pekerjaan godoc
// @Summary Ubah pekerjaan
// @Description Mengubah data pekerjaan berdasarkan ID
// @Tags Pekerjaan
// @Accept json
// @Produce json
// @Param id path string true "Pekerjaan ID (hex)"
// @Param body body model.UpdatePekerjaanReq true "Body"
// @Success 200 {object} map[string]interface{} // { success, data }
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /pekerjaan/{id} [put]
func (s *PekerjaanService) Update(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := c.Params("id")
	var in model.UpdatePekerjaanReq
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Body tidak valid"})
	}
	if err := s.Repo.Update(ctx, id, in); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return c.Status(404).JSON(fiber.Map{"success": false, "message": "Data tidak ditemukan"})
		}
		return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	item, _ := s.Repo.GetByID(ctx, id)
	return c.JSON(fiber.Map{"success": true, "data": item})
}

// Soft Delete Pekerjaan godoc
// @Summary Hapus pekerjaan secara soft
// @Description Menghapus data pekerjaan secara soft (tidak benar-benar dihapus dari database)
// @Tags Pekerjaan
// @Produce json
// @Param id path string true "Pekerjaan ID (hex)"
// @Success 200 {object} map[string]interface{} // { success, message }
// @Failure 400 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /pekerjaan/{id} [delete]
func (s *PekerjaanService) SoftDelete(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := c.Params("id")
	role, _ := c.Locals("role").(string)
	username, _ := c.Locals("username").(string)
	userID, _ := c.Locals("user_id").(string)

	job, err := s.Repo.GetByID(ctx, id)
	if err != nil || job == nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "message": "Pekerjaan tidak ditemukan"})
	}

	// Admin bisa hapus langsung
	if role == "admin" {
		if err := s.Repo.SoftDeleteAdmin(ctx, id, username); err != nil {
			return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
		}
		return c.JSON(fiber.Map{"success": true, "message": "Pekerjaan berhasil dihapus (admin)"})
	}

	// User biasa
	alumni, err := s.Repo.GetAlumniByUserID(ctx, userID)
	if err != nil || alumni == nil {
		return c.Status(403).JSON(fiber.Map{"success": false, "message": "Data Alumni tidak ditemukan"})
	}
	if job.AlumniID != alumni.ID {
		return c.Status(403).JSON(fiber.Map{"success": false, "message": "Anda tidak berhak menghapus data ini"})
	}

	if err := s.Repo.SoftDeleteOwned(ctx, id, alumni.ID.Hex(), username); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Pekerjaan berhasil dihapus (owner)"})
}

// List Trash Pekerjaan godoc
// @Summary Daftar pekerjaan yang terhapus
// @Description Mengambil daftar pekerjaan yang terhapus dengan pagination, sort, dan pencarian
// @Tags Pekerjaan
// @Accept json
// @Produce json
// @Param page query int false "Halaman" default(1)
// @Param limit query int false "Jumlah per halaman" default(10)
// @Param sortBy query string false "Kolom sort (default: deleted_at)"
// @Param order query string false "asc/desc (default: desc)"
// @Param search query string false "Kata kunci pencarian"
// @Success 200 {object} model.PekerjaanListResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /pekerjaan/trash [get]
func (s *PekerjaanService) ListTrash(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	sortBy := c.Query("sortBy", "deleted_at")
	order := c.Query("order", "desc")
	search := c.Query("search", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	items, err := s.Repo.ListTrashed(ctx, search, sortBy, order, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	total, err := s.Repo.CountTrashed(ctx, search)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.JSON(model.PekerjaanListResponse{
		Data: items,
		Meta: model.MetaInfo{
			Page:   page,
			Limit:  limit,
			Total:  total,
			Pages:  int((total + int64(limit) - 1) / int64(limit)),
			SortBy: sortBy,
			Order:  order,
			Search: search,
		},
	})
}

// Restore Pekerjaan godoc
// @Summary Restore pekerjaan yang terhapus
// @Description Mengembalikan data pekerjaan yang terhapus berdasarkan ID
// @Tags Pekerjaan
// @Produce json
// @Param id path string true "Pekerjaan ID (hex)"
// @Success 200 {object} map[string]interface{} // { success, message }
// @Failure 400 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /pekerjaan/{id}/restore [post]
func (s *PekerjaanService) Restore(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := c.Params("id")
	role, _ := c.Locals("role").(string)
	username, _ := c.Locals("username").(string)
	userID, _ := c.Locals("user_id").(string)

	job, err := s.Repo.GetByIDAny(ctx, id)
	if err != nil || job == nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "message": "Data tidak ditemukan"})
	}

	if role == "admin" {
		if err := s.Repo.RestoreAdmin(ctx, id, username); err != nil {
			return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
		}
		return c.JSON(fiber.Map{"success": true, "message": "Data berhasil direstore (admin)"})
	}

	alumni, err := s.Repo.GetAlumniByUserID(ctx, userID)
	if err != nil || alumni == nil {
		return c.Status(403).JSON(fiber.Map{"success": false, "message": "Data Alumni tidak ditemukan"})
	}
	if job.AlumniID != alumni.ID {
		return c.Status(403).JSON(fiber.Map{"success": false, "message": "Tidak berhak me-restore data ini"})
	}

	if err := s.Repo.RestoreOwned(ctx, id, alumni.ID.Hex(), username); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Data berhasil direstore"})
}

// Hard Delete Pekerjaan godoc
// @Summary Hapus pekerjaan secara hard
// @Description Menghapus data pekerjaan secara hard (benar-benar dihapus dari database)
// @Tags Pekerjaan
// @Produce json
// @Param id path string true "Pekerjaan ID (hex)"
// @Success 204 {object} map[string]interface{} // { success, message }
// @Failure 400 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /pekerjaan/{id}/hard [delete]
func (s *PekerjaanService) HardDelete(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := c.Params("id")
	role, _ := c.Locals("role").(string)
	userID, _ := c.Locals("user_id").(string)

	job, err := s.Repo.GetByIDAny(ctx, id)
	if err != nil || job == nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "message": "Data tidak ditemukan"})
	}

	if role == "admin" {
		if err := s.Repo.HardDeleteAdmin(ctx, id); err != nil {
			return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
		}
		return c.Status(204).Send(nil)
	}

	alumni, err := s.Repo.GetAlumniByUserID(ctx, userID)
	if err != nil || alumni == nil {
		return c.Status(403).JSON(fiber.Map{"success": false, "message": "Data Alumni tidak ditemukan"})
	}
	if job.AlumniID != alumni.ID {
		return c.Status(403).JSON(fiber.Map{"success": false, "message": "Tidak berhak menghapus permanen data ini"})
	}

	if err := s.Repo.HardDeleteOwned(ctx, id, alumni.ID.Hex()); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	return c.Status(204).Send(nil)
}
