package domain

import (
	"time"
)

// URL represents a shortened URL
type URL struct {
	ID          int64      `json:"id" db:"id"`
	ShortCode   string     `json:"short_code" db:"short_code"`
	OriginalURL string     `json:"original_url" db:"original_url"`
	ClickCount  int64      `json:"click_count" db:"click_count"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	LastAccess  *time.Time `json:"last_access,omitempty" db:"last_access"`
}

// ShortenRequest represents a request to shorten a URL
type ShortenRequest struct {
	URL         string     `json:"url" binding:"required,url"`
	CustomAlias string     `json:"custom_alias,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// ShortenResponse represents a response containing the shortened URL
type ShortenResponse struct {
	ShortURL    string     `json:"short_url"`
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// AnalyticsResponse represents the analytics data for a shortened URL
type AnalyticsResponse struct {
	ShortCode    string      `json:"short_code"`
	OriginalURL  string      `json:"original_url"`
	ClickCount   int64       `json:"click_count"`
	CreatedAt    time.Time   `json:"created_at"`
	LastAccessed *time.Time  `json:"last_accessed,omitempty"`
	DailyStats   []DailyStat `json:"daily_stats,omitempty"`
}

// DailyStat represents the daily statistics for a shortened URL
type DailyStat struct {
	Date   string `json:"date"`
	Clicks int64  `json:"clicks"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}
