package service

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"prak/clean-architecture-fiber-mongo/app/model"
	"prak/clean-architecture-fiber-mongo/app/repository"
)

type AlumniService struct {
	Repo repository.AlumniRepository
}

func NewAlumniService(repo repository.AlumniRepository) *AlumniService {
	return &AlumniService{Repo: repo}
}

// List Alumni godoc
// @Summary daftar alumni
// @Description daftar alumni dengan pagination, sort, dan pencarian
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
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	total, err := s.Repo.Count(ctx, search)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
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
// @Summary detail alumni
// @Description detail alumni berdasarkan ID (hex string)
// @Tags Alumni
// @Produce json
// @Param id path string true "Alumni ID (hex)"
// @Success 200 {object} model.AlumniResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /alumni/{id} [get]
func (s *AlumniService) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "id tidak valid",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	item, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	if item == nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Success: false,
			Message: "data tidak ditemukan",
		})
	}
	return c.JSON(model.AlumniResponse{
		Success: true,
		Data:    item,
	})
}

// Create Alumni godoc
// @Summary tambah alumni
// @Description menambahkan data alumni baru (Admin Only)
// @Tags Alumni
// @Accept json
// @Produce json
// @Param body body model.CreateAlumniReq true "Body"
// @Success 201 {object} model.AlumniResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /alumni [post]
func (s *AlumniService) Create(c *fiber.Ctx) error {
	var in model.CreateAlumniReq
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "body tidak valid",
		})
	}
	if in.NIM == "" || in.Nama == "" || in.Jurusan == "" || in.Email == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "NIM, nama, jurusan, email wajib diisi",
		})
	}

	uidHex, _ := c.Locals("user_id").(string)
	if uidHex == "" {
		return c.Status(401).JSON(model.ErrorResponse{
			Success: false,
			Message: "user tidak ditemukan",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	newID, err := s.Repo.Create(ctx, in, uidHex)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	item, err := s.Repo.GetByID(ctx, newID)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: "gagal mengambil data setelah create: " + err.Error(),
		})
	}
	if item == nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: "data tidak ditemukan setelah create",
		})
	}
	return c.Status(201).JSON(model.AlumniResponse{
		Success: true,
		Data:    item,
	})
}

// Update Alumni godoc
// @Summary update alumni
// @Description mengupdate data alumni
// @Tags Alumni
// @Accept json
// @Produce json
// @Param id path string true "Alumni ID (hex)"
// @Param body body model.UpdateAlumniReq true "Body"
// @Success 200 {object} model.AlumniResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /alumni/{id} [put]
func (s *AlumniService) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "id tidak valid",
		})
	}

	var in model.UpdateAlumniReq
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "body tidak valid",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Repo.Update(ctx, id, in); err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	item, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: "gagal mengambil data setelah update: " + err.Error(),
		})
	}
	if item == nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Success: false,
			Message: "data tidak ditemukan setelah update",
		})
	}
	return c.JSON(model.AlumniResponse{
		Success: true,
		Data:    item,
	})
}

// Delete Alumni godoc
// @Summary hapus alumni
// @Description menghapus data alumni berdasarkan id (Admin Only)
// @Tags Alumni
// @Produce json
// @Param id path string true "Alumni ID (hex)"
// @Success 200 {object} model.SuccessMessageResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /alumni/{id} [delete]
func (s *AlumniService) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "id tidak valid",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Repo.Delete(ctx, id); err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	return c.JSON(model.SuccessMessageResponse{
		Success: true,
		Message: "data alumni berhasil dihapus",
	})
}

// List By Jurusan godoc
// @Summary daftar alumni berdasarkan jurusan
// @Description mengambil daftar alumni berdasarkan jurusan
// @Tags Alumni
// @Produce json
// @Param jurusan path string true "Nama jurusan"
// @Success 200 {object} model.AlumniListByJurusanResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /alumni/jurusan/{jurusan} [get]
func (s *AlumniService) ListByJurusan(c *fiber.Ctx) error {
	jurusan := c.Params("jurusan")
	if jurusan == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Message: "parameter jurusan wajib diisi",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	items, err := s.Repo.ListByJurusan(ctx, jurusan)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	return c.JSON(model.AlumniListByJurusanResponse{
		Success: true,
		Count:   len(items),
		Data:    items,
	})
}