package repository

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestNormalizeSort(t *testing.T) {
	tests := []struct {
		name     string
		col      string
		order    string
		expected bson.D
	}{
		{
			name:     "valid column asc",
			col:      "nama",
			order:    "asc",
			expected: bson.D{{Key: "nama", Value: int32(1)}},
		},
		{
			name:     "valid column desc",
			col:      "nama",
			order:    "desc",
			expected: bson.D{{Key: "nama", Value: int32(-1)}},
		},
		{
			name:     "invalid column defaults to _id",
			col:      "invalid",
			order:    "asc",
			expected: bson.D{{Key: "_id", Value: int32(1)}},
		},
		{
			name:     "case insensitive column",
			col:      "NAMA",
			order:    "asc",
			expected: bson.D{{Key: "nama", Value: int32(1)}},
		},
		{
			name:     "case insensitive order",
			col:      "nama",
			order:    "DESC",
			expected: bson.D{{Key: "nama", Value: int32(-1)}},
		},
		{
			name:     "empty order defaults to asc",
			col:      "nama",
			order:    "",
			expected: bson.D{{Key: "nama", Value: int32(1)}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeSort(tt.col, tt.order)
			if len(result) != len(tt.expected) {
				t.Errorf("normalizeSort() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			if result[0].Key != tt.expected[0].Key {
				t.Errorf("normalizeSort() key = %v, want %v", result[0].Key, tt.expected[0].Key)
			}
			if result[0].Value != tt.expected[0].Value {
				t.Errorf("normalizeSort() value = %v, want %v", result[0].Value, tt.expected[0].Value)
			}
		})
	}
}

func TestToOID(t *testing.T) {
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
		{
			name:    "too short",
			hex:     "507f1f77",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := toOID(tt.hex)
			if (err != nil) != tt.wantErr {
				t.Errorf("toOID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

