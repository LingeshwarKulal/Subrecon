package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// CrtSh queries crt.sh certificate transparency logs
type CrtSh struct {
	config *SourceConfig
	client *http.Client
}

// NewCrtSh creates a new CrtSh source
func NewCrtSh(config *SourceConfig) *CrtSh {
	if config == nil {
		config = DefaultConfig()
		config.RateLimit = 5
	}
	
	return &CrtSh{
		config: config,
		client: &http.Client{
			Timeout: config.GetTimeout(),
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

type crtshResponse struct {
	NameValue string `json:"name_value"`
}

// Run executes the CrtSh source
func (c *CrtSh) Run(ctx context.Context, domain string) ([]string, error) {
	// Build URL
	apiURL := fmt.Sprintf("https://crt.sh/?q=%s&output=json", url.QueryEscape("%."+domain))
	
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("User-Agent", c.config.UserAgent)
	
	// Execute request with retry
	var resp *http.Response
	var lastErr error
	
	for i := 0; i < c.config.Retry; i++ {
		resp, lastErr = c.client.Do(req)
		if lastErr == nil && resp.StatusCode == http.StatusOK {
			break
		}
		
		if resp != nil {
			resp.Body.Close()
		}
		
		// Exponential backoff
		if i < c.config.Retry-1 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(i+1) * time.Second):
			}
		}
	}
	
	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", c.config.Retry, lastErr)
	}
	
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	defer resp.Body.Close()
	
	// Read and parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	var results []crtshResponse
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	
	// Extract unique subdomains
	subdomainMap := make(map[string]bool)
	for _, result := range results {
		// name_value can contain multiple subdomains separated by newlines
		names := strings.Split(result.NameValue, "\n")
		for _, name := range names {
			name = strings.TrimSpace(name)
			name = strings.ToLower(name)
			
			// Remove wildcard
			name = strings.TrimPrefix(name, "*.")
			
			// Validate subdomain
			if name != "" && strings.HasSuffix(name, "."+domain) || name == domain {
				subdomainMap[name] = true
			}
		}
	}
	
	// Convert map to slice
	subdomains := make([]string, 0, len(subdomainMap))
	for subdomain := range subdomainMap {
		subdomains = append(subdomains, subdomain)
	}
	
	return subdomains, nil
}

// Name returns the source name
func (c *CrtSh) Name() string {
	return "crtsh"
}

// NeedsKey indicates if API key is required
func (c *CrtSh) NeedsKey() bool {
	return false
}
