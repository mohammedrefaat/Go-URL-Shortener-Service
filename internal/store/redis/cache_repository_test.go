package redis

import (
	"context"
	"strings"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
)

func TestCacheRepository_BasicFlow(t *testing.T) {
	// start in-memory redis
	srv, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer srv.Close()

	redisURL := "redis://" + srv.Addr()
	ctx := context.Background()

	// create repository
	r, err := NewCacheRepository(redisURL)
	if err != nil {
		t.Fatalf("NewCacheRepository error: %v", err)
	}
	defer r.Close()

	// health check
	if err := r.HealthCheck(ctx); err != nil {
		t.Fatalf("health check failed: %v", err)
	}

	// Set / Get simple string
	type dummy struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	key := "foo"
	value := "bar"

	if err := r.Set(ctx, key, value, 5*time.Second); err != nil {
		t.Fatalf("Set error: %v", err)
	}

	var got string
	if err := r.Get(ctx, key, &got); err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if got != value {
		t.Fatalf("expected %q, got %q", value, got)
	}

	// Delete and ensure not found
	if err := r.Delete(ctx, key); err != nil {
		t.Fatalf("Delete error: %v", err)
	}
	if err := r.Get(ctx, key, &got); err == nil {
		t.Fatalf("expected error after delete, got nil and value %q", got)
	} else if !strings.Contains(err.Error(), "key not found") {
		t.Fatalf("expected 'key not found' error, got: %v", err)
	}

	// Set/Get struct
	skey := "person"
	p := dummy{Name: "alice", Age: 30}
	if err := r.Set(ctx, skey, p, 5*time.Second); err != nil {
		t.Fatalf("Set struct error: %v", err)
	}
	var pgot dummy
	if err := r.Get(ctx, skey, &pgot); err != nil {
		t.Fatalf("Get struct error: %v", err)
	}
	if pgot.Name != p.Name || pgot.Age != p.Age {
		t.Fatalf("expected %+v, got %+v", p, pgot)
	}
}

func TestCacheRepository_IncrementAndNumericGet(t *testing.T) {
	srv, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer srv.Close()

	redisURL := "redis://" + srv.Addr()
	ctx := context.Background()

	r, err := NewCacheRepository(redisURL)
	if err != nil {
		t.Fatalf("NewCacheRepository error: %v", err)
	}
	defer r.Close()

	// increment on non-existing key should start from 0
	if err := r.Increment(ctx, "counter", 5); err != nil {
		t.Fatalf("Increment error: %v", err)
	}

	var val int64
	if err := r.Get(ctx, "counter", &val); err != nil {
		t.Fatalf("Get counter error: %v", err)
	}
	if val != 5 {
		t.Fatalf("expected counter 5, got %d", val)
	}

	if err := r.Increment(ctx, "counter", 3); err != nil {
		t.Fatalf("Increment second time error: %v", err)
	}
	if err := r.Get(ctx, "counter", &val); err != nil {
		t.Fatalf("Get counter after second increment error: %v", err)
	}
	if val != 8 {
		t.Fatalf("expected counter 8, got %d", val)
	}
}

func TestCacheRepository_URLMappingAndCleanup(t *testing.T) {
	srv, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer srv.Close()

	redisURL := "redis://" + srv.Addr()
	ctx := context.Background()

	r, err := NewCacheRepository(redisURL)
	if err != nil {
		t.Fatalf("NewCacheRepository error: %v", err)
	}
	defer r.Close()

	shortKey := "abc123"
	origURL := "https://example.com/some/path"

	// set mapping
	if err := r.SetURLMapping(ctx, shortKey, origURL, 10*time.Second); err != nil {
		t.Fatalf("SetURLMapping error: %v", err)
	}

	// get by short key
	u, err := r.GetURLByShortKey(ctx, shortKey)
	if err != nil {
		t.Fatalf("GetURLByShortKey error: %v", err)
	}
	if u != origURL {
		t.Fatalf("expected url %q, got %q", origURL, u)
	}

	// get short key by url
	sk, err := r.GetShortKeyByURL(ctx, origURL)
	if err != nil {
		t.Fatalf("GetShortKeyByURL error: %v", err)
	}
	if sk != shortKey {
		t.Fatalf("expected shortKey %q, got %q", shortKey, sk)
	}

	// cleanup DB and ensure keys removed
	if err := r.Cleanup(ctx); err != nil {
		t.Fatalf("Cleanup error: %v", err)
	}

	if _, err := r.GetURLByShortKey(ctx, shortKey); err == nil {
		t.Fatalf("expected error after cleanup for GetURLByShortKey, got nil")
	}
	if _, err := r.GetShortKeyByURL(ctx, origURL); err == nil {
		t.Fatalf("expected error after cleanup for GetShortKeyByURL, got nil")
	}
}
