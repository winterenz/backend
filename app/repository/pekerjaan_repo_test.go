package repository

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestNormalizePekSort(t *testing.T) {
	tests := []struct {
		name     string
		col      string
		order    string
		expected bson.D
	}{
		{
			name:     "valid column asc",
			col:      "nama_perusahaan",
			order:    "asc",
			expected: bson.D{{Key: "nama_perusahaan", Value: int32(1)}},
		},
		{
			name:     "valid column desc",
			col:      "nama_perusahaan",
			order:    "desc",
			expected: bson.D{{Key: "nama_perusahaan", Value: int32(-1)}},
		},
		{
			name:     "invalid column defaults to _id",
			col:      "invalid",
			order:    "asc",
			expected: bson.D{{Key: "_id", Value: int32(1)}},
		},
		{
			name:     "case insensitive column",
			col:      "NAMA_PERUSAHAAN",
			order:    "asc",
			expected: bson.D{{Key: "nama_perusahaan", Value: int32(1)}},
		},
		{
			name:     "case insensitive order",
			col:      "nama_perusahaan",
			order:    "DESC",
			expected: bson.D{{Key: "nama_perusahaan", Value: int32(-1)}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePekSort(tt.col, tt.order)
			if len(result) != len(tt.expected) {
				t.Errorf("normalizePekSort() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			if result[0].Key != tt.expected[0].Key {
				t.Errorf("normalizePekSort() key = %v, want %v", result[0].Key, tt.expected[0].Key)
			}
			if result[0].Value != tt.expected[0].Value {
				t.Errorf("normalizePekSort() value = %v, want %v", result[0].Value, tt.expected[0].Value)
			}
		})
	}
}

func TestSearchFilter(t *testing.T) {
	tests := []struct {
		name     string
		base     bson.M
		search   string
		hasOr    bool
	}{
		{
			name:     "empty search returns base",
			base:     bson.M{"is_deleted": false},
			search:   "",
			hasOr:    false,
		},
		{
			name:     "search adds $or",
			base:     bson.M{"is_deleted": false},
			search:   "test",
			hasOr:    true,
		},
		{
			name:     "whitespace search returns base",
			base:     bson.M{"is_deleted": false},
			search:   "   ",
			hasOr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := searchFilter(tt.base, tt.search)
			_, hasOr := result["$or"]
			if hasOr != tt.hasOr {
				t.Errorf("searchFilter() has $or = %v, want %v", hasOr, tt.hasOr)
			}
		})
	}
}

func TestOid(t *testing.T) {
	tests := []struct {
		name    string
		hex     string
		wantErr bool
	}{
		{
			name:    "valid hex",
			hex:     "507f1f77bcf86cd799439011",
			wantErr: false,
		},
		{
			name:    "invalid hex",
			hex:     "invalid",
			wantErr: true,
		},
		{
			name:    "empty string",
			hex:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := oid(tt.hex)
			if (err != nil) != tt.wantErr {
				t.Errorf("oid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

