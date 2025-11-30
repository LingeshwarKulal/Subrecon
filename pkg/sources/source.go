package sources

import (
	"context"
	"time"
)

// Source represents a subdomain enumeration source
type Source interface {
	// Run executes the source and returns discovered subdomains
	Run(ctx context.Context, domain string) ([]string, error)
	
	// Name returns the name of the source
	Name() string
	
	// NeedsKey indicates if the source requires an API key
	NeedsKey() bool
}

// SourceConfig holds configuration for a source
type SourceConfig struct {
	APIKey      string `yaml:"api_key"`
	RateLimit   int    `yaml:"rate_limit"`   // requests per second
	Timeout     int    `yaml:"timeout"`       // in seconds
	Enabled     bool   `yaml:"enabled"`
	Retry       int    `yaml:"retry"`
	UserAgent   string `yaml:"user_agent"`
	MaxResults  int    `yaml:"max_results"`
}

// GetTimeout returns timeout as time.Duration
func (sc *SourceConfig) GetTimeout() time.Duration {
	if sc.Timeout <= 0 {
		return 30 * time.Second
	}
	return time.Duration(sc.Timeout) * time.Second
}

// DefaultConfig returns default source configuration
func DefaultConfig() *SourceConfig {
	return &SourceConfig{
		RateLimit:  5,
		Timeout:    30,
		Enabled:    true,
		Retry:      3,
		UserAgent:  "SubFinder-Pro/1.0",
		MaxResults: 10000,
	}
}
