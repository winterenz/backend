package model

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	CategoryFoto       = "foto"
	CategorySertifikat = "sertifikat"
)

type FileDoc struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"     json:"id"`
	OwnerUserID  primitive.ObjectID `bson:"owner_user_id"     json:"owner_user_id"`
	Category     string             `bson:"category"          json:"category"`
	FileName     string             `bson:"file_name"         json:"file_name"`
	OriginalName string             `bson:"original_name"     json:"original_name"`
	FilePath     string             `bson:"file_path"         json:"file_path"`
	FileSize     int64              `bson:"file_size"         json:"file_size"`
	FileType     string             `bson:"file_type"         json:"file_type"`
	UploadedAt   time.Time          `bson:"uploaded_at"       json:"uploaded_at"`
}