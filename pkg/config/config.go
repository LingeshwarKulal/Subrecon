package config

import (
	"fmt"
	"os"
	"time"

	"github.com/yourusername/subfinder-pro/pkg/sources"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Timeout    int      `yaml:"timeout"`
	Workers    int      `yaml:"workers"`
	RateLimit  int      `yaml:"rate_limit"`
	DNS        DNSConfig `yaml:"dns"`
	Output     OutputConfig `yaml:"output"`
	HTTP       HTTPConfig `yaml:"http"`
}

// DNSConfig holds DNS resolver configuration
type DNSConfig struct {
	Enabled bool     `yaml:"enabled"`
	Servers []string `yaml:"servers"`
	Timeout int      `yaml:"timeout"`
	Retry   int      `yaml:"retry"`
}

// OutputConfig holds output configuration
type OutputConfig struct {
	Format string `yaml:"format"` // text or json
	Sort   bool   `yaml:"sort"`
	Unique bool   `yaml:"unique"`
}

// HTTPConfig holds HTTP client configuration
type HTTPConfig struct {
	UserAgent string `yaml:"user_agent"`
	Timeout   int    `yaml:"timeout"`
	Proxy     string `yaml:"proxy"`
}

// ProviderConfig represents provider-specific configuration
type ProviderConfig struct {
	Sources map[string]*sources.SourceConfig `yaml:"sources"`
}

// Load loads configuration from a file
func Load(path string) (*Config, error) {
	// Set defaults
	config := &Config{
		Timeout:   30,
		Workers:   10,
		RateLimit: 5,
		DNS: DNSConfig{
			Enabled: false,
			Servers: []string{"8.8.8.8:53", "1.1.1.1:53"},
			Timeout: 5,
			Retry:   3,
		},
		Output: OutputConfig{
			Format: "text",
			Sort:   true,
			Unique: true,
		},
		HTTP: HTTPConfig{
			UserAgent: "SubFinder-Pro/1.0",
			Timeout:   10,
			Proxy:     "",
		},
	}
	
	// If file doesn't exist, return defaults
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return config, nil
	}
	
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	// Validate
	if err := config.Validate(); err != nil {
		return nil, err
	}
	
	return config, nil
}

// LoadProviderConfig loads provider configuration from a file
func LoadProviderConfig(path string) (*ProviderConfig, error) {
	config := &ProviderConfig{
		Sources: make(map[string]*sources.SourceConfig),
	}
	
	// If file doesn't exist, return empty config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return config, nil
	}
	
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read provider config file: %w", err)
	}
	
	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse provider config file: %w", err)
	}
	
	// Override with environment variables
	for name, srcConfig := range config.Sources {
		if apiKey := getEnvAPIKey(name); apiKey != "" {
			srcConfig.APIKey = apiKey
		}
	}
	
	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}
	
	if c.Workers <= 0 {
		return fmt.Errorf("workers must be greater than 0")
	}
	
	if c.Workers > 100 {
		return fmt.Errorf("workers cannot exceed 100")
	}
	
	if c.RateLimit < 0 {
		return fmt.Errorf("rate_limit cannot be negative")
	}
	
	if c.Output.Format != "text" && c.Output.Format != "json" {
		return fmt.Errorf("output format must be 'text' or 'json'")
	}
	
	return nil
}

// GetTimeout returns timeout as duration
func (c *Config) GetTimeout() time.Duration {
	return time.Duration(c.Timeout) * time.Second
}

// GetDNSTimeout returns DNS timeout as duration
func (c *Config) GetDNSTimeout() time.Duration {
	return time.Duration(c.DNS.Timeout) * time.Second
}

// GetHTTPTimeout returns HTTP timeout as duration
func (c *Config) GetHTTPTimeout() time.Duration {
	return time.Duration(c.HTTP.Timeout) * time.Second
}

// getEnvAPIKey tries to get API key from environment variables
func getEnvAPIKey(sourceName string) string {
	// Try multiple formats
	envVars := []string{
		fmt.Sprintf("%s_API_KEY", toUpperSnakeCase(sourceName)),
		fmt.Sprintf("SUBFINDER_%s_API_KEY", toUpperSnakeCase(sourceName)),
	}
	
	for _, envVar := range envVars {
		if value := os.Getenv(envVar); value != "" {
			return value
		}
	}
	
	return ""
}

// toUpperSnakeCase converts string to UPPER_SNAKE_CASE
func toUpperSnakeCase(s string) string {
	result := ""
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result += "_"
		}
		result += string(r)
	}
	return result
}

// GetSourceConfig returns configuration for a specific source
func (pc *ProviderConfig) GetSourceConfig(name string) *sources.SourceConfig {
	if config, ok := pc.Sources[name]; ok {
		return config
	}
	
	// Return default config
	return sources.DefaultConfig()
}

// IsSourceEnabled checks if a source is enabled
func (pc *ProviderConfig) IsSourceEnabled(name string) bool {
	if config, ok := pc.Sources[name]; ok {
		return config.Enabled
	}
	return true // Default to enabled
}
