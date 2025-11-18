package service

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"prak3/clean-architecture-fiber-mongo/app/model"
	"prak3/clean-architecture-fiber-mongo/app/repository"
)

type AlumniService struct {
	Repo repository.AlumniRepository
}

func NewAlumniService(repo repository.AlumniRepository) *AlumniService {
	return &AlumniService{Repo: repo}
}

// List Alumni godoc
// @Summary Daftar alumni
// @Description Mengambil daftar alumni dengan pagination, sort, dan pencarian
// @Tags Alumni
// @Accept json
// @Produce json
// @Param page query int false "Halaman" default(1)
// @Param limit query int false "Jumlah per halaman" default(10)
// @Param sortBy query string false "Kolom sort (default: id)"
// @Param order query string false "asc/desc (default: asc)"
// @Param search query string false "Kata kunci pencarian"
// @Success 200 {object} model.AlumniListResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /alumni [get]
func (s *AlumniService) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	sortBy := c.Query("sortBy", "id")
	order := c.Query("order", "asc")
	search := c.Query("search", "")

	if page < 1 { page = 1 }
	if limit < 1 { limit = 10 }
	if limit > 100 { limit = 100 }
	if order != "asc" && order != "desc" { order = "asc" }
	offset := (page - 1) * limit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	items, err := s.Repo.ListPaged(ctx, search, sortBy, order, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	total, err := s.Repo.Count(ctx, search)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	if items == nil { items = make([]model.Alumni, 0) }

	pages := int((total + int64(limit) - 1) / int64(limit))

	return c.JSON(model.AlumniListResponse{
		Data: items,
		Meta: model.MetaInfo{
			Page:   page,
			Limit:  limit,
			Total:  total,
			Pages:  pages,
			SortBy: sortBy,
			Order:  order,
			Search: search,
		},
	})
}

// Get Alumni godoc
// @Summary Detail alumni
// @Description Mengambil detail alumni berdasarkan ID (hex string)
// @Tags Alumni
// @Produce json
// @Param id path string true "Alumni ID (hex)"
// @Success 200 {object} map[string]interface{}  // { success, data }
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /alumni/{id} [get]
func (s *AlumniService) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	item, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	if item == nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "message": "Data tidak ditemukan"})
	}
	return c.JSON(fiber.Map{"success": true, "data": item})
}

// Create Alumni godoc
// @Summary Tambah alumni
// @Description Menambahkan data alumni baru (user_id dari JWT)
// @Tags Alumni
// @Accept json
// @Produce json
// @Param body body model.CreateAlumniReq true "Body"
// @Success 201 {object} map[string]interface{} // { success, data }
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /alumni [post]
func (s *AlumniService) Create(c *fiber.Ctx) error {
	var in model.CreateAlumniReq
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "body tidak valid"})
	}
	if in.NIM == "" || in.Nama == "" || in.Jurusan == "" || in.Email == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "NIM, nama, jurusan, email wajib"})
	}

	uidHex, _ := c.Locals("user_id").(string)
	if uidHex == "" {
		return c.Status(401).JSON(fiber.Map{"success": false, "message": "User tidak terautentikasi"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	newID, err := s.Repo.Create(ctx, in, uidHex)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	item, _ := s.Repo.GetByID(ctx, newID)
	return c.Status(201).JSON(fiber.Map{"success": true, "data": item})
}

// Update Alumni godoc
// @Summary Ubah alumni
// @Description Mengubah data alumni
// @Tags Alumni
// @Accept json
// @Produce json
// @Param id path string true "Alumni ID (hex)"
// @Param body body model.UpdateAlumniReq true "Body"
// @Success 200 {object} map[string]interface{} // { success, data }
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /alumni/{id} [put]
func (s *AlumniService) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}

	var in model.UpdateAlumniReq
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "body tidak valid"})
	}
	if in.Nama == "" || in.Jurusan == "" || in.Email == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "nama, jurusan, email wajib"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Repo.Update(ctx, id, in); err != nil {
		// mongo.ErrNoDocuments → 404
		return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	item, _ := s.Repo.GetByID(ctx, id)
	return c.JSON(fiber.Map{"success": true, "data": item})
}

// Delete Alumni godoc
// @Summary Hapus alumni
// @Description Menghapus data alumni berdasarkan ID
// @Tags Alumni
// @Produce json
// @Param id path string true "Alumni ID (hex)"
// @Success 200 {object} map[string]interface{} // { success, message }
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /alumni/{id} [delete]
func (s *AlumniService) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Repo.Delete(ctx, id); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "message": "hapus ok"})
}


// List By Jurusan godoc
// @Summary Daftar alumni berdasarkan jurusan
// @Description Mengambil daftar alumni berdasarkan jurusan
// @Tags Alumni
// @Produce json
// @Param jurusan path string true "Nama jurusan"
// @Success 200 {object} map[string]interface{} // { success, count, data }
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /alumni/jurusan/{jurusan} [get]
func (s *AlumniService) ListByJurusan(c *fiber.Ctx) error {
	jurusan := c.Params("jurusan")
	if jurusan == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "message": "parameter jurusan wajib diisi",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	items, err := s.Repo.ListByJurusan(ctx, jurusan)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false, "message": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(items),
		"data":    items,
	})
}
