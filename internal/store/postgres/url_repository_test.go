package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"

	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/domain"
)

func setupTestPostgres(t *testing.T) (databaseURL string, pool *dockertest.Pool, resource *dockertest.Resource, cleanup func()) {
	t.Helper()

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("could not connect to docker: %v", err)
	}

	// Pull & run postgres
	opts := &dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15-alpine",
		Env: []string{
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_DB=testdb",
		},
	}
	// expose default port 5432
	resource, err = pool.RunWithOptions(opts, func(hostConfig *docker.HostConfig) {
		// Auto-remove container after test end
		hostConfig.AutoRemove = true
	})
	if err != nil {
		t.Fatalf("could not start resource: %v", err)
	}

	// Expire container after timeout to avoid leaking
	resource.Expire(120) // seconds

	// Build DB URL using mapped port
	hostPort := resource.GetPort("5432/tcp")
	databaseURL = fmt.Sprintf("postgres://postgres:secret@localhost:%s/testdb?sslmode=disable", hostPort)

	// wait for Postgres to be ready
	if err := pool.Retry(func() error {
		db, err := sqlx.Connect("postgres", databaseURL)
		if err != nil {
			return err
		}
		defer db.Close()
		return db.Ping()
	}); err != nil {
		// cleanup before fail
		_ = pool.Purge(resource)
		t.Fatalf("could not connect to database: %v", err)
	}

	cleanup = func() {
		// purge container
		_ = pool.Purge(resource)
	}

	return databaseURL, pool, resource, cleanup
}

func TestURLRepository_Integration(t *testing.T) {
	// skip test if running without docker in some CI that cannot run docker
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dbURL, _, _, cleanup := setupTestPostgres(t)
	defer cleanup()

	// create repo which also runs migrations
	repo, err := NewURLRepository(dbURL)
	if err != nil {
		t.Fatalf("NewURLRepository error: %v", err)
	}
	defer repo.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// HealthCheck
	if err := repo.HealthCheck(ctx); err != nil {
		t.Fatalf("healthcheck failed: %v", err)
	}

	// Create a URL
	// IMPORTANT: adjust the domain.URL field names/types to match your actual struct if different.
	now := time.Now().UTC()
	u := &domain.URL{
		// Example fields — adapt names/types if your domain.URL differs
		ShortCode:   "abc123",
		OriginalURL: "https://example.com/path",
		CreatedAt:   now,
		ExpiresAt:   nil, // nil means no expiry; change if your type is sql.NullTime or *time.Time
	}

	if err := repo.CreateURL(ctx, u); err != nil {
		t.Fatalf("CreateURL error: %v", err)
	}
	if u.ID == 0 {
		t.Fatalf("expected inserted ID to be set (>0), got 0")
	}

	// Get by short code
	got, err := repo.GetURLByShortCode(ctx, "abc123")
	if err != nil {
		t.Fatalf("GetURLByShortCode error: %v", err)
	}
	if got.ShortCode != u.ShortCode || got.OriginalURL != u.OriginalURL {
		t.Fatalf("mismatch: got %+v want %+v", got, u)
	}

	// Update click count
	if err := repo.UpdateClickCount(ctx, "abc123"); err != nil {
		t.Fatalf("UpdateClickCount error: %v", err)
	}
	again, err := repo.GetURLByShortCode(ctx, "abc123")
	if err != nil {
		t.Fatalf("GetURLByShortCode after update error: %v", err)
	}
	if again.ClickCount != got.ClickCount+1 {
		t.Fatalf("expected click count %d, got %d", got.ClickCount+1, again.ClickCount)
	}

	// Insert analytics rows directly (simulate clicks)
	// url_analytics columns: short_code, clicked_at, ip_address, user_agent, referrer
	// use repo.db to insert rows directly:
	insertQuery := `INSERT INTO url_analytics (short_code, clicked_at, ip_address, user_agent, referrer) VALUES ($1, $2, $3, $4, $5)`
	if _, err := repo.db.ExecContext(ctx, insertQuery, "abc123", time.Now().UTC(), "127.0.0.1", "go-test-agent", "https://ref.example"); err != nil {
		t.Fatalf("insert analytics error: %v", err)
	}
	// insert second row older than 2 days to ensure date filtering works
	if _, err := repo.db.ExecContext(ctx, insertQuery, "abc123", time.Now().Add(-48*time.Hour).UTC(), "127.0.0.1", "go-test-agent", ""); err != nil {
		t.Fatalf("insert analytics error: %v", err)
	}

	// Get analytics for last 7 days
	analytics, err := repo.GetAnalytics(ctx, "abc123", 7)
	if err != nil {
		t.Fatalf("GetAnalytics error: %v", err)
	}
	if analytics.ShortCode != "abc123" {
		t.Fatalf("analytics short code mismatch: %v", analytics.ShortCode)
	}
	// ensure DailyStats contains at least one entry for today (or the recent inserted)
	if len(analytics.DailyStats) == 0 {
		t.Fatalf("expected daily stats, got 0")
	}

	// Create an expired URL and test DeleteExpiredURLs
	expiredAt := time.Now().Add(-24 * time.Hour).UTC()
	expURL := &domain.URL{
		ShortCode:   "expired1",
		OriginalURL: "https://expired.example",
		CreatedAt:   now,
		ExpiresAt:   &expiredAt, // adjust if your type differs
	}
	if err := repo.CreateURL(ctx, expURL); err != nil {
		t.Fatalf("CreateURL for expired failed: %v", err)
	}

	// Call DeleteExpiredURLs
	if err := repo.DeleteExpiredURLs(ctx); err != nil {
		t.Fatalf("DeleteExpiredURLs error: %v", err)
	}
	// Try to fetch expired URL — should return not found
	if _, err := repo.GetURLByShortCode(ctx, "expired1"); err == nil {
		t.Fatalf("expected error for expired url after delete, got nil")
	}

	// Cleanup database (delete all urls)
	if err := repo.Cleanup(ctx); err != nil {
		t.Fatalf("Cleanup error: %v", err)
	}
	// After cleanup, attempt to get main url -> expect not found
	if _, err := repo.GetURLByShortCode(ctx, "abc123"); err == nil {
		t.Fatalf("expected not found after cleanup for abc123, got nil")
	}
}

func TestGetURLByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &URLRepository{db: sqlx.NewDb(db, "postgres")}

	// simulate no rows
	mock.ExpectQuery(`SELECT (.+) FROM urls WHERE short_code = \$1`).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetURLByShortCode(context.Background(), "nonexistent")
	require.Error(t, err)
	require.Equal(t, "URL not found", err.Error())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteExpiredURLs_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &URLRepository{db: sqlx.NewDb(db, "postgres")}

	mock.ExpectExec(`DELETE FROM urls`).
		WillReturnError(sql.ErrConnDone)

	err = repo.DeleteExpiredURLs(context.Background())
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHealthCheck_Error(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer db.Close()

	repo := &URLRepository{db: sqlx.NewDb(db, "postgres")}

	// mock Ping fails
	mock.ExpectPing().WillReturnError(sql.ErrConnDone)

	err = repo.HealthCheck(context.Background())
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
