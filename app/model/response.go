package model

type MetaInfo struct{
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
	Total  int64    `json:"total"`
	Pages  int    `json:"pages"`
	SortBy string `json:"sortBy"`
	Order  string `json:"order"`
	Search string `json:"search"`
}

type AlumniListResponse struct{
	Data []Alumni `json:"data"`
	Meta MetaInfo `json:"meta"`
}

type PekerjaanListResponse struct {
	Data []Pekerjaan `json:"data"`
	Meta MetaInfo    `json:"meta"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type SuccessResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}