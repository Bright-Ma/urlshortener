package model

import "time"

type CreateURLRequest struct {
	OriginalURL string `json:"original_url" validate:"required,url"`
	CustomCode  string `json:"custom_code,omitempty" validate:"omitempty,min=4,max=10,alphanum"`
	Duration    *int   `json:"duration,omitempty" validate:"omitempty,min=1,max=100"`
	UserID      int    `json:"-"`
}

type CreateURLResponse struct {
	ShortURL string `json:"short_url"`
}

type GetURLsRequest struct {
	Page   uint `query:"page"`
	Size   uint `query:"size"`
	UserID int  `query:"-"`
}

type GetURLsResponse struct {
	OriginalURL string    `json:"original_url"`
	ShortURL    string    `json:"short_url"`
	ExpiredAt   time.Time `json:"expired_at"`
	IsCustom    bool      `json:"is_custom"`
	Views       uint      `json:"views"`
}

type URL struct {
	OriginalURL string
	ShortCode   string
}
