package model

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Pekerjaan struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AlumniID           primitive.ObjectID `bson:"alumni_id" json:"alumni_id"`
	NamaPerusahaan     string             `bson:"nama_perusahaan" json:"nama_perusahaan"`
	PosisiJabatan      string             `bson:"posisi_jabatan" json:"posisi_jabatan"`
	BidangIndustri     string             `bson:"bidang_industri" json:"bidang_industri"`
	LokasiKerja        string             `bson:"lokasi_kerja" json:"lokasi_kerja"`
	GajiRange          *string            `bson:"gaji_range,omitempty" json:"gaji_range,omitempty"`
	TanggalMulaiKerja  time.Time          `bson:"tanggal_mulai_kerja" json:"tanggal_mulai_kerja"`
	TanggalSelesaiKerja *time.Time         `bson:"tanggal_selesai_kerja,omitempty" json:"tanggal_selesai_kerja,omitempty"`
	StatusPekerjaan    string             `bson:"status_pekerjaan" json:"status_pekerjaan"`
	Deskripsi          *string            `bson:"deskripsi_pekerjaan,omitempty" json:"deskripsi_pekerjaan,omitempty"`
	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"updated_at"`
	IsDeleted bool       `bson:"is_deleted" json:"is_deleted"`
	DeletedBy *string    `bson:"deleted_by,omitempty" json:"deleted_by,omitempty"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// DTO
type CreatePekerjaanReq struct {
	AlumniID            string     `json:"alumni_id" binding:"required"`
	NamaPerusahaan      string     `json:"nama_perusahaan" binding:"required"`
	PosisiJabatan       string     `json:"posisi_jabatan" binding:"required"`
	BidangIndustri      string     `json:"bidang_industri" binding:"required"`
	LokasiKerja         string     `json:"lokasi_kerja" binding:"required"`
	GajiRange           *string    `json:"gaji_range,omitempty"`
	TanggalMulaiKerja   time.Time  `json:"tanggal_mulai_kerja" binding:"required"`
	TanggalSelesaiKerja *time.Time `json:"tanggal_selesai_kerja,omitempty"`
	StatusPekerjaan     string     `json:"status_pekerjaan" binding:"required"`
	Deskripsi           *string    `json:"deskripsi_pekerjaan,omitempty"`
}

// Update
type UpdatePekerjaanReq struct {
	AlumniID            *string    `json:"alumni_id,omitempty"`
	NamaPerusahaan      *string    `json:"nama_perusahaan,omitempty"`
	PosisiJabatan       *string    `json:"posisi_jabatan,omitempty"`
	BidangIndustri      *string    `json:"bidang_industri,omitempty"`
	LokasiKerja         *string    `json:"lokasi_kerja,omitempty"`
	GajiRange           *string    `json:"gaji_range,omitempty"`
	TanggalMulaiKerja   *time.Time `json:"tanggal_mulai_kerja,omitempty"`
	TanggalSelesaiKerja *time.Time `json:"tanggal_selesai_kerja,omitempty"`
	StatusPekerjaan     *string    `json:"status_pekerjaan,omitempty"`
	Deskripsi           *string    `json:"deskripsi_pekerjaan,omitempty"`
}