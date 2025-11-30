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

// ThreatCrowd queries threatcrowd.org API
type ThreatCrowd struct {
	config *SourceConfig
	client *http.Client
}

// NewThreatCrowd creates a new ThreatCrowd source
func NewThreatCrowd(config *SourceConfig) *ThreatCrowd {
	if config == nil {
		config = DefaultConfig()
		config.RateLimit = 1
	}
	
	return &ThreatCrowd{
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

type threatCrowdResponse struct {
	ResponseCode string   `json:"response_code"`
	Subdomains   []string `json:"subdomains"`
}

// Run executes the ThreatCrowd source
func (tc *ThreatCrowd) Run(ctx context.Context, domain string) ([]string, error) {
	// Build URL
	apiURL := fmt.Sprintf("https://www.threatcrowd.org/searchApi/v2/domain/report/?domain=%s", 
		url.QueryEscape(domain))
	
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("User-Agent", tc.config.UserAgent)
	
	// Execute request with retry
	var resp *http.Response
	var lastErr error
	
	for i := 0; i < tc.config.Retry; i++ {
		resp, lastErr = tc.client.Do(req)
		if lastErr == nil && resp.StatusCode == http.StatusOK {
			break
		}
		
		if resp != nil {
			resp.Body.Close()
		}
		
		// Exponential backoff
		if i < tc.config.Retry-1 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(i+1) * time.Second):
			}
		}
	}
	
	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", tc.config.Retry, lastErr)
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
	
	var result threatCrowdResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	
	// Check response code
	if result.ResponseCode != "1" {
		return []string{}, nil // No results found
	}
	
	// Extract unique subdomains
	subdomainMap := make(map[string]bool)
	for _, subdomain := range result.Subdomains {
		subdomain = strings.TrimSpace(subdomain)
		subdomain = strings.ToLower(subdomain)
		
		if subdomain != "" && (strings.HasSuffix(subdomain, "."+domain) || subdomain == domain) {
			subdomainMap[subdomain] = true
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
func (tc *ThreatCrowd) Name() string {
	return "threatcrowd"
}

// NeedsKey indicates if API key is required
func (tc *ThreatCrowd) NeedsKey() bool {
	return false
}
