package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/domain"
	"github.com/mohammedrefaat/Go-URL-Shortener-Service/internal/store/postgres"
)

func TestURLRepository_Integration(t *testing.T) {
	// This would use a test database
	databaseURL := "postgres://postgres:244159@localhost:5432/urlshortener_test?sslmode=disable"

	repo, err := postgres.NewURLRepository(databaseURL)
	require.NoError(t, err)
	defer repo.Close()
	repo.Cleanup(context.Background())
	ctx := context.Background()

	t.Run("CreateAndGetURL", func(t *testing.T) {
		url := &domain.URL{
			ShortCode:   "test123",
			OriginalURL: "https://example.com",
			CreatedAt:   time.Now(),
		}

		err := repo.CreateURL(ctx, url)
		require.NoError(t, err)
		assert.NotZero(t, url.ID)

		// Get by short code
		retrieved, err := repo.GetURLByShortCode(ctx, "test123")
		require.NoError(t, err)
		assert.Equal(t, url.ShortCode, retrieved.ShortCode)
		assert.Equal(t, url.OriginalURL, retrieved.OriginalURL)

		// Get by original URL
		retrieved2, err := repo.GetURLByOriginalURL(ctx, "https://example.com")
		require.NoError(t, err)
		assert.Equal(t, url.ShortCode, retrieved2.ShortCode)
	})

	t.Run("UpdateClickCount", func(t *testing.T) {
		url := &domain.URL{
			ShortCode:   "click123",
			OriginalURL: "https://clicktest.com",
			CreatedAt:   time.Now(),
		}

		err := repo.CreateURL(ctx, url)
		require.NoError(t, err)

		// Update click count
		err = repo.UpdateClickCount(ctx, "click123")
		require.NoError(t, err)

		// Verify click count
		retrieved, err := repo.GetURLByShortCode(ctx, "click123")
		require.NoError(t, err)
		assert.Equal(t, int64(1), retrieved.ClickCount)
	})
}
