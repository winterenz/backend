package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// model
type Alumni struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"     json:"id,omitempty"`
	UserID     primitive.ObjectID `bson:"user_id"           json:"user_id"`
	NIM        string             `bson:"nim"               json:"nim"`
	Nama       string             `bson:"nama"              json:"nama"`
	Jurusan    string             `bson:"jurusan"           json:"jurusan"`
	Angkatan   int                `bson:"angkatan"          json:"angkatan"`
	TahunLulus int                `bson:"tahun_lulus"       json:"tahun_lulus"`
	Email      string             `bson:"email"             json:"email"`
	NoTelepon  *string            `bson:"no_telepon,omitempty" json:"no_telepon,omitempty"`
	Alamat     *string            `bson:"alamat,omitempty"     json:"alamat,omitempty"`
	CreatedAt  time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt  time.Time          `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}

// dto
type CreateAlumniReq struct {
	NIM        string  `json:"nim"`
	Nama       string  `json:"nama"`
	Jurusan    string  `json:"jurusan"`
	Angkatan   int     `json:"angkatan"`
	TahunLulus int     `json:"tahun_lulus"`
	Email      string  `json:"email"`
	NoTelepon  *string `json:"no_telepon"`
	Alamat     *string `json:"alamat"`
}

type UpdateAlumniReq = CreateAlumniReq
