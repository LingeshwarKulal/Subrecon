package tests

import (
	"context"
	"testing"
	"time"

	"github.com/yourusername/subfinder-pro/pkg/runner"
	"github.com/yourusername/subfinder-pro/pkg/sources"
)

// MockSource implements the Source interface for testing
type MockSource struct {
	name       string
	subdomains []string
	err        error
	delay      time.Duration
}

func (m *MockSource) Run(ctx context.Context, domain string) ([]string, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	return m.subdomains, m.err
}

func (m *MockSource) Name() string {
	return m.name
}

func (m *MockSource) NeedsKey() bool {
	return false
}

func TestRunnerWithMockSources(t *testing.T) {
	// Create mock sources
	source1 := &MockSource{
		name:       "mock1",
		subdomains: []string{"api.example.com", "www.example.com"},
	}
	source2 := &MockSource{
		name:       "mock2",
		subdomains: []string{"www.example.com", "blog.example.com"},
	}

	srcs := []sources.Source{source1, source2}

	config := &runner.Config{
		Workers: 2,
		Timeout: 10 * time.Second,
		Verbose: false,
		Silent:  true,
	}

	r := runner.NewRunner(srcs, config)

	ctx := context.Background()
	subdomains, err := r.Run(ctx, "example.com")

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check for deduplication
	expected := 3 // api, www, blog
	if len(subdomains) != expected {
		t.Errorf("Expected %d unique subdomains, got %d", expected, len(subdomains))
	}

	// Verify subdomains are sorted
	for i := 1; i < len(subdomains); i++ {
		if subdomains[i-1] >= subdomains[i] {
			t.Error("Subdomains are not sorted")
			break
		}
	}
}

func TestRunnerWithError(t *testing.T) {
	// Create a source that always fails
	failSource := &MockSource{
		name: "fail",
		err:  context.DeadlineExceeded,
	}

	successSource := &MockSource{
		name:       "success",
		subdomains: []string{"api.example.com"},
	}

	srcs := []sources.Source{failSource, successSource}

	config := &runner.Config{
		Workers: 2,
		Timeout: 10 * time.Second,
		Verbose: false,
		Silent:  true,
	}

	r := runner.NewRunner(srcs, config)

	ctx := context.Background()
	subdomains, err := r.Run(ctx, "example.com")

	// Should succeed because one source succeeded
	if err != nil {
		t.Errorf("Expected no error when at least one source succeeds, got: %v", err)
	}

	if len(subdomains) != 1 {
		t.Errorf("Expected 1 subdomain, got %d", len(subdomains))
	}
}

func TestRunnerTimeout(t *testing.T) {
	// Create a slow source
	slowSource := &MockSource{
		name:       "slow",
		subdomains: []string{"api.example.com"},
		delay:      5 * time.Second,
	}

	srcs := []sources.Source{slowSource}

	config := &runner.Config{
		Workers: 1,
		Timeout: 1 * time.Second, // Very short timeout
		Verbose: false,
		Silent:  true,
	}

	r := runner.NewRunner(srcs, config)

	ctx := context.Background()
	_, err := r.Run(ctx, "example.com")

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestRunnerConcurrency(t *testing.T) {
	// Create multiple sources
	sources := make([]sources.Source, 10)
	for i := 0; i < 10; i++ {
		sources[i] = &MockSource{
			name:       "mock" + string(rune(i)),
			subdomains: []string{"subdomain" + string(rune(i)) + ".example.com"},
			delay:      100 * time.Millisecond,
		}
	}

	config := &runner.Config{
		Workers: 5,
		Timeout: 10 * time.Second,
		Verbose: false,
		Silent:  true,
	}

	r := runner.NewRunner(sources, config)

	start := time.Now()
	ctx := context.Background()
	subdomains, err := r.Run(ctx, "example.com")
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(subdomains) != 10 {
		t.Errorf("Expected 10 subdomains, got %d", len(subdomains))
	}

	// With 5 workers and 10 sources taking 100ms each,
	// it should take around 200ms (not 1000ms)
	if duration > 500*time.Millisecond {
		t.Errorf("Concurrency not working properly, took %v", duration)
	}
}

func BenchmarkRunner(b *testing.B) {
	source := &MockSource{
		name:       "bench",
		subdomains: []string{"api.example.com", "www.example.com", "blog.example.com"},
	}

	srcs := []sources.Source{source}

	config := &runner.Config{
		Workers: 10,
		Timeout: 10 * time.Second,
		Verbose: false,
		Silent:  true,
	}

	r := runner.NewRunner(srcs, config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Run(ctx, "example.com")
	}
}
