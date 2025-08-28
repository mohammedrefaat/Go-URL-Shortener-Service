package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/domain"
)

type URLRepository struct {
	db *sqlx.DB
}

func NewURLRepository(databaseURL string) (*URLRepository, error) {
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	repo := &URLRepository{db: db}

	// Run migrations
	if err := repo.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return repo, nil
}

func (r *URLRepository) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		short_code VARCHAR(12) UNIQUE NOT NULL,
		original_url TEXT NOT NULL,
		click_count BIGINT DEFAULT 0,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		expires_at TIMESTAMP WITH TIME ZONE,
		last_access TIMESTAMP WITH TIME ZONE
	);

	CREATE INDEX IF NOT EXISTS idx_urls_short_code ON urls(short_code);
	CREATE INDEX IF NOT EXISTS idx_urls_original_url ON urls(original_url);
	CREATE INDEX IF NOT EXISTS idx_urls_expires_at ON urls(expires_at) WHERE expires_at IS NOT NULL;

	CREATE TABLE IF NOT EXISTS url_analytics (
		id SERIAL PRIMARY KEY,
		short_code VARCHAR(12) NOT NULL,
		clicked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		ip_address INET,
		user_agent TEXT,
		referrer TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_analytics_short_code ON url_analytics(short_code);
	CREATE INDEX IF NOT EXISTS idx_analytics_clicked_at ON url_analytics(clicked_at);
	`

	_, err := r.db.Exec(query)
	return err
}

func (r *URLRepository) CreateURL(ctx context.Context, url *domain.URL) error {
	query := `
	INSERT INTO urls (short_code, original_url, created_at, expires_at)
	VALUES (:short_code, :original_url, :created_at, :expires_at)
	RETURNING id
	`

	rows, err := r.db.NamedQueryContext(ctx, query, url)
	if err != nil {
		return fmt.Errorf("failed to insert URL: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Scan(&url.ID)
	}

	return fmt.Errorf("failed to get inserted ID")
}

func (r *URLRepository) GetURLByShortCode(ctx context.Context, shortCode string) (*domain.URL, error) {
	var url domain.URL
	query := `
	SELECT id, short_code, original_url, click_count, created_at, expires_at, last_access
	FROM urls
	WHERE short_code = $1
	`

	err := r.db.GetContext(ctx, &url, query, shortCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("URL not found")
		}
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}

	return &url, nil
}

func (r *URLRepository) GetURLByOriginalURL(ctx context.Context, originalURL string) (*domain.URL, error) {
	var url domain.URL
	query := `
	SELECT id, short_code, original_url, click_count, created_at, expires_at, last_access
	FROM urls
	WHERE original_url = $1
	ORDER BY created_at DESC
	LIMIT 1
	`

	err := r.db.GetContext(ctx, &url, query, originalURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("URL not found")
		}
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}

	return &url, nil
}

func (r *URLRepository) UpdateClickCount(ctx context.Context, shortCode string) error {
	query := `
	UPDATE urls
	SET click_count = click_count + 1, last_access = NOW()
	WHERE short_code = $1
	`

	result, err := r.db.ExecContext(ctx, query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to update click count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("URL not found")
	}

	return nil
}

func (r *URLRepository) GetAnalytics(ctx context.Context, shortCode string, days int) (*domain.AnalyticsResponse, error) {
	// Get basic URL info
	url, err := r.GetURLByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	// Get daily stats
	dailyQuery := `
SELECT DATE(clicked_at) as date, COUNT(*) as clicks
FROM url_analytics
WHERE short_code = $1 AND clicked_at >= NOW() - ($2 * INTERVAL '1 day')
GROUP BY DATE(clicked_at)
ORDER BY date DESC
`
	var dailyStats []domain.DailyStat
	err = r.db.SelectContext(ctx, &dailyStats, dailyQuery, shortCode, days) // Pass days as parameter to prevent SQL injection

	if err != nil {
		return nil, fmt.Errorf("failed to get daily stats: %w", err)
	}

	return &domain.AnalyticsResponse{
		ShortCode:    url.ShortCode,
		OriginalURL:  url.OriginalURL,
		ClickCount:   url.ClickCount,
		CreatedAt:    url.CreatedAt,
		LastAccessed: url.LastAccess,
		DailyStats:   dailyStats,
	}, nil
}

func (r *URLRepository) DeleteExpiredURLs(ctx context.Context) error {
	query := `DELETE FROM urls WHERE expires_at IS NOT NULL AND expires_at < NOW()`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete expired URLs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected > 0 {
		// Log the cleanup (would use proper logger in real implementation)
		fmt.Printf("Deleted %d expired URLs\n", rowsAffected)
	}

	return nil
}

func (r *URLRepository) HealthCheck(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *URLRepository) Close() error {
	return r.db.Close()
}

func (r *URLRepository) Cleanup(ctx context.Context) error {
	query := `DELETE FROM urls`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete expired URLs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected > 0 {
		// Log the cleanup (would use proper logger in real implementation)
		fmt.Printf("Deleted %d expired URLs\n", rowsAffected)
	}

	return nil
}

func (r *URLRepository) RecordClick(ctx context.Context, analytics *domain.URLAnalytics) error {
	query := `
		INSERT INTO url_analytics (short_code, clicked_at, user_agent, ip_address, referer, country)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		analytics.ShortCode, analytics.ClickedAt, analytics.UserAgent,
		analytics.IPAddress, analytics.Referer, analytics.Country)
	return err
}

func (r *URLRepository) GetClickCount(ctx context.Context, shortCode string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM url_analytics WHERE short_code = $1`
	err := r.db.GetContext(ctx, &count, query, shortCode)
	return count, err
}

func (r *URLRepository) GetDailyStats(ctx context.Context, shortCode string, days int) ([]domain.DailyStat, error) {
	var stats []domain.DailyStat
	query := `
		SELECT 
			DATE(clicked_at) as date,
			COUNT(*) as clicks
		FROM url_analytics 
		WHERE short_code = $1 AND clicked_at >= NOW() - INTERVAL '%d days'
		GROUP BY DATE(clicked_at)
		ORDER BY date DESC
	`
	err := r.db.SelectContext(ctx, &stats, fmt.Sprintf(query, days), shortCode)
	return stats, err
}

func (r *URLRepository) GetLastAccessed(ctx context.Context, shortCode string) (*time.Time, error) {
	var lastAccessed sql.NullTime
	query := `
		SELECT MAX(clicked_at) 
		FROM url_analytics 
		WHERE short_code = $1
	`
	err := r.db.GetContext(ctx, &lastAccessed, query, shortCode)
	if err != nil {
		return nil, err
	}
	if !lastAccessed.Valid {
		return nil, nil
	}
	return &lastAccessed.Time, nil
}
func (r *URLRepository) IsShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = $1)`
	err := r.db.GetContext(ctx, &exists, query, shortCode)
	return exists, err
}
