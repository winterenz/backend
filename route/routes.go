package route

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"

	"prak/clean-architecture-fiber-mongo/app/repository"
	"prak/clean-architecture-fiber-mongo/app/service"
	"prak/clean-architecture-fiber-mongo/middleware"
)

func Register(app *fiber.App, db *mongo.Database) {
	api := app.Group("/api")

	// Auth
	authRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(authRepo)
	api.Post("/login", authSvc.Login)

	// JWT
	protected := api.Group("", middleware.AuthRequired())
	protected.Get("/profile", authSvc.Profile)

	// Alumni
	alumniRepo := repository.NewAlumniRepository(db)
	alumniSvc := service.NewAlumniService(alumniRepo)

  	alumni := protected.Group("/alumni")

	alumni.Get("/", alumniSvc.List)
	alumni.Get("/:id", alumniSvc.Get)
	alumni.Get("/jurusan/:jurusan", alumniSvc.ListByJurusan)

	// Pekerjaan
	pekerjaanRepo := repository.NewPekerjaanRepository(db)
	pekerjaanSvc := service.NewPekerjaanService(pekerjaanRepo)

  	pekerjaan := protected.Group("/pekerjaan")

	pekerjaan.Get("/", pekerjaanSvc.List)
	pekerjaan.Get("/trash", pekerjaanSvc.ListTrash)
	pekerjaan.Get("/:id", pekerjaanSvc.Get)
	pekerjaan.Get("/alumni/:alumni_id", pekerjaanSvc.ListByAlumni)
	pekerjaan.Delete("/:id", pekerjaanSvc.SoftDelete)
	pekerjaan.Post("/:id/restore", pekerjaanSvc.Restore)
	pekerjaan.Delete("/:id/hard", pekerjaanSvc.HardDelete)

	// File
	fileRepo := repository.NewFileRepository(db)
	fileSvc := service.NewFileService(fileRepo, "./uploads")

	files := protected.Group("/files")  

	files.Get("/", fileSvc.GetAllFiles)
	files.Get("/:id", fileSvc.GetFileByID)
	files.Delete("/:id", fileSvc.DeleteFile)

	// Upload
	upload := protected.Group("/users")
	upload.Post("/:user_id/upload/foto", fileSvc.UploadFoto)
	upload.Post("/:user_id/upload/sertifikat", fileSvc.UploadSertifikat)

	// Alternatif
	upload.Post("/me/upload/foto", fileSvc.UploadFoto)
	upload.Post("/me/upload/sertifikat", fileSvc.UploadSertifikat)

	// Admin
	admin := protected.Group("", middleware.AdminOnly())

	// Alumni
	admin.Post("/alumni", alumniSvc.Create)
	admin.Put("/alumni/:id", alumniSvc.Update)
	admin.Delete("/alumni/:id", alumniSvc.Delete)

	// Pekerjaan
	admin.Post("/pekerjaan", pekerjaanSvc.Create)
	admin.Put("/pekerjaan/:id", pekerjaanSvc.Update)

	// File
	admin.Post("/users/:user_id/upload/foto", fileSvc.UploadFoto)
	admin.Post("/users/:user_id/upload/sertifikat", fileSvc.UploadSertifikat)
}