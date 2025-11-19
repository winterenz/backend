package model

// MetaInfo 
type MetaInfo struct {
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
	Total  int64  `json:"total"`
	Pages  int    `json:"pages"`
	SortBy string `json:"sortBy"`
	Order  string `json:"order"`
	Search string `json:"search"`
}

// DTO
type AlumniListResponse struct {
	Data []Alumni `json:"data"`
	Meta MetaInfo `json:"meta"`
}

type PekerjaanListResponse struct {
	Data []Pekerjaan `json:"data"`
	Meta MetaInfo    `json:"meta"`
}

type AlumniResponse struct {
	Success bool    `json:"success"`
	Message string  `json:"message,omitempty"`
	Data    *Alumni `json:"data"`
}

type PekerjaanResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	Data    *Pekerjaan `json:"data"`
}

type AlumniListByJurusanResponse struct {
	Success bool     `json:"success"`
	Count   int      `json:"count"`
	Data    []Alumni `json:"data"`
}

type PekerjaanListByAlumniResponse struct {
	Success bool        `json:"success"`
	Count   int         `json:"count,omitempty"`
	Data    []Pekerjaan `json:"data"`
}

type SuccessMessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type SuccessResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
}

type FileUploadResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Data    FileUploadData      `json:"data"`
}

type FileUploadData struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

type FileListResponse struct {
	Success bool    `json:"success"`
	Data    []FileDoc `json:"data"`
}

type FileResponse struct {
	Success bool    `json:"success"`
	Data    *FileDoc `json:"data"`
}

type ProfileResponse struct {
	Success bool              `json:"success"`
	Data    ProfileData       `json:"data"`
}

type ProfileData struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}