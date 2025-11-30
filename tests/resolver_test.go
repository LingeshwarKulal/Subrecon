package tests

import (
	"context"
	"testing"
	"time"

	"github.com/yourusername/subfinder-pro/internal/resolve"
)

func TestValidateDomain(t *testing.T) {
	tests := []struct {
		domain  string
		wantErr bool
	}{
		{"example.com", false},
		{"sub.example.com", false},
		{"test.co.uk", false},
		{"", true},
		{"invalid domain", true},
		{".example.com", true},
		{"example.com.", true},
		{"example", true},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			err := resolve.ValidateDomain(tt.domain)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDomain(%q) error = %v, wantErr %v", tt.domain, err, tt.wantErr)
			}
		})
	}
}

func TestResolver(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := &resolve.Config{
		Servers: []string{"8.8.8.8:53", "1.1.1.1:53"},
		Timeout: 5 * time.Second,
	}

	resolver := resolve.NewResolver(config)

	ctx := context.Background()

	// Test resolving a known domain
	result, err := resolver.Resolve(ctx, "google.com")
	if err != nil {
		t.Errorf("Failed to resolve google.com: %v", err)
	}

	if !result.Exists {
		t.Error("Expected google.com to exist")
	}

	if len(result.IPs) == 0 {
		t.Error("Expected at least one IP for google.com")
	}

	t.Logf("google.com resolved to: %v", result.IPs)
}

func TestResolverCache(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := &resolve.Config{
		Servers: []string{"8.8.8.8:53"},
		Timeout: 5 * time.Second,
	}

	resolver := resolve.NewResolver(config)

	ctx := context.Background()

	// First resolve
	start := time.Now()
	result1, err := resolver.Resolve(ctx, "google.com")
	duration1 := time.Since(start)

	if err != nil {
		t.Errorf("First resolve failed: %v", err)
	}

	// Second resolve (should be cached)
	start = time.Now()
	result2, err := resolver.Resolve(ctx, "google.com")
	duration2 := time.Since(start)

	if err != nil {
		t.Errorf("Second resolve failed: %v", err)
	}

	// Cached result should be much faster
	if duration2 >= duration1 {
		t.Logf("Warning: Cached lookup (%v) not faster than first lookup (%v)", duration2, duration1)
	}

	// Results should be identical
	if len(result1.IPs) != len(result2.IPs) {
		t.Error("Cached result differs from original")
	}
}

func TestWildcardDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := &resolve.Config{
		Timeout: 5 * time.Second,
	}

	resolver := resolve.NewResolver(config)

	ctx := context.Background()

	// Test with a domain that likely doesn't have wildcard DNS
	isWildcard, ips, err := resolver.DetectWildcard(ctx, "google.com")
	if err != nil {
		t.Logf("Wildcard detection error (may be expected): %v", err)
	}

	t.Logf("Wildcard detection for google.com: %v, IPs: %v", isWildcard, ips)
}

func TestResolveMany(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := &resolve.Config{
		Timeout: 5 * time.Second,
	}

	resolver := resolve.NewResolver(config)

	ctx := context.Background()

	subdomains := []string{
		"google.com",
		"github.com",
		"stackoverflow.com",
		"nonexistent-domain-12345.com",
	}

	results, err := resolver.ResolveMany(ctx, subdomains, 3)
	if err != nil {
		t.Errorf("ResolveMany failed: %v", err)
	}

	if len(results) != len(subdomains) {
		t.Errorf("Expected %d results, got %d", len(subdomains), len(results))
	}

	existCount := 0
	for _, result := range results {
		if result.Exists {
			existCount++
			t.Logf("%s resolved to %v", result.Host, result.IPs)
		}
	}

	if existCount < 2 {
		t.Errorf("Expected at least 2 domains to resolve, got %d", existCount)
	}
}

func BenchmarkResolve(b *testing.B) {
	config := &resolve.Config{
		Timeout: 5 * time.Second,
	}

	resolver := resolve.NewResolver(config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resolver.Resolve(ctx, "google.com")
	}
}
