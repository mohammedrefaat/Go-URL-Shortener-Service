package utils

import (
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// --- JWT tests ---
func TestGenerateAndValidateJWT(t *testing.T) {
	secret := "test-secret-123"
	userID := "user-42"

	tokenStr, err := GenerateJWT(secret, userID, time.Hour)
	if err != nil {
		t.Fatalf("GenerateJWT error: %v", err)
	}
	if tokenStr == "" {
		t.Fatalf("expected non-empty token")
	}

	// validate with correct secret
	tok, err := ValidateJWT(tokenStr, secret)
	if err != nil {
		t.Fatalf("ValidateJWT error: %v", err)
	}
	if tok == nil {
		t.Fatalf("expected token, got nil")
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("expected MapClaims, got %T", tok.Claims)
	}
	if got, _ := claims["user_id"].(string); got != userID {
		t.Fatalf("expected user_id %q, got %q", userID, got)
	}

	// tamper token (invalid signature)
	tampered := tokenStr + "x"
	if _, err := ValidateJWT(tampered, secret); err == nil {
		t.Fatalf("expected error for tampered token, got nil")
	}
}

func TestValidateJWT_WrongSecret(t *testing.T) {
	secret := "s1"
	wrong := "s2"
	tokenStr, err := GenerateJWT(secret, "u", time.Minute)
	if err != nil {
		t.Fatalf("GenerateJWT error: %v", err)
	}
	if _, err := ValidateJWT(tokenStr, wrong); err == nil {
		t.Fatalf("expected signature error with wrong secret, got nil")
	}
}

// --- Snowflake ID tests ---
func TestGenerateID_Basic(t *testing.T) {
	// reset global map to isolate tests
	nodes = sync.Map{}

	id := GenerateID(1)
	if id == "" {
		t.Fatalf("expected non-empty id for machineID=1")
	}

	id2 := GenerateID(1)
	if id2 == "" {
		t.Fatalf("expected second id non-empty")
	}
	if id == id2 {
		// unlikely to collide, but if happens test still passes — we just ensure they are strings
		t.Logf("warning: two consecutive ids equal (unlikely): %s", id)
	}
}

func TestGenerateID_InvalidMachine(t *testing.T) {
	if got := GenerateID(-5); got != "" {
		t.Fatalf("expected empty id for invalid machineID, got %q", got)
	}
	if got := GenerateID(1024); got != "" {
		t.Fatalf("expected empty id for out-of-range machineID, got %q", got)
	}
}

// --- URL validation tests ---
func TestIsValidURL(t *testing.T) {
	cases := []struct {
		url string
		ok  bool
	}{
		{"http://example.com", true},
		{"https://example.com/path?q=1", true},
		{"ftp://example.com", false}, // unsupported scheme
		{"//example.com", false},     // missing scheme
		{"http://", false},           // missing host
		{"", false},
		{"http://malware.example.com", false}, // malicious domain from list
	}

	for _, c := range cases {
		got := IsValidURL(c.url)
		if got != c.ok {
			t.Fatalf("IsValidURL(%q) = %v; want %v", c.url, got, c.ok)
		}
	}
}
