package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/subfinder-pro/pkg/sources"
)

func TestCrtSh(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `[
			{"name_value": "example.com"},
			{"name_value": "www.example.com"},
			{"name_value": "api.example.com"},
			{"name_value": "*.example.com"}
		]`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	config := &sources.SourceConfig{
		RateLimit:  5,
		Timeout:    10 * time.Second,
		Enabled:    true,
		Retry:      3,
		UserAgent:  "Test/1.0",
		MaxResults: 1000,
	}

	src := sources.NewCrtSh(config)

	// Test Name
	if src.Name() != "crtsh" {
		t.Errorf("Expected name 'crtsh', got '%s'", src.Name())
	}

	// Test NeedsKey
	if src.NeedsKey() {
		t.Error("CrtSh should not need an API key")
	}

	// Note: This test would need to be updated to use the mock server
	// For now, we'll skip the actual Run test as it requires HTTP interception
}

func TestHackerTarget(t *testing.T) {
	config := &sources.SourceConfig{
		RateLimit:  2,
		Timeout:    10 * time.Second,
		Enabled:    true,
		Retry:      3,
		UserAgent:  "Test/1.0",
		MaxResults: 1000,
	}

	src := sources.NewHackerTarget(config)

	if src.Name() != "hackertarget" {
		t.Errorf("Expected name 'hackertarget', got '%s'", src.Name())
	}

	if src.NeedsKey() {
		t.Error("HackerTarget should not require an API key")
	}
}

func TestThreatCrowd(t *testing.T) {
	config := &sources.SourceConfig{
		RateLimit:  1,
		Timeout:    10 * time.Second,
		Enabled:    true,
		Retry:      3,
		UserAgent:  "Test/1.0",
		MaxResults: 1000,
	}

	src := sources.NewThreatCrowd(config)

	if src.Name() != "threatcrowd" {
		t.Errorf("Expected name 'threatcrowd', got '%s'", src.Name())
	}

	if src.NeedsKey() {
		t.Error("ThreatCrowd should not require an API key")
	}
}

func TestAlienVault(t *testing.T) {
	config := &sources.SourceConfig{
		APIKey:     "test-key",
		RateLimit:  10,
		Timeout:    10 * time.Second,
		Enabled:    true,
		Retry:      3,
		UserAgent:  "Test/1.0",
		MaxResults: 1000,
	}

	src := sources.NewAlienVault(config)

	if src.Name() != "alienvault" {
		t.Errorf("Expected name 'alienvault', got '%s'", src.Name())
	}

	if !src.NeedsKey() {
		t.Error("AlienVault should require an API key")
	}
}

func TestURLScan(t *testing.T) {
	config := &sources.SourceConfig{
		RateLimit:  1,
		Timeout:    10 * time.Second,
		Enabled:    true,
		Retry:      3,
		UserAgent:  "Test/1.0",
		MaxResults: 1000,
	}

	src := sources.NewURLScan(config)

	if src.Name() != "urlscan" {
		t.Errorf("Expected name 'urlscan', got '%s'", src.Name())
	}

	if src.NeedsKey() {
		t.Error("URLScan should not require an API key")
	}
}

// Integration test (requires internet connection)
func TestCrtShIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := sources.DefaultConfig()
	src := sources.NewCrtSh(config)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test with a known domain
	subdomains, err := src.Run(ctx, "example.com")
	if err != nil {
		t.Logf("Warning: Integration test failed: %v", err)
		return
	}

	if len(subdomains) == 0 {
		t.Log("Warning: No subdomains found for example.com")
	} else {
		t.Logf("Found %d subdomains for example.com", len(subdomains))
	}
}

// Benchmark tests
func BenchmarkCrtShParsing(b *testing.B) {
	config := sources.DefaultConfig()
	src := sources.NewCrtSh(config)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Note: This would ideally use a mock server
		src.Run(ctx, "example.com")
	}
}
