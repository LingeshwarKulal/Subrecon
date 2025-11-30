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

// URLScan queries urlscan.io API
type URLScan struct {
	config *SourceConfig
	client *http.Client
}

// NewURLScan creates a new URLScan source
func NewURLScan(config *SourceConfig) *URLScan {
	if config == nil {
		config = DefaultConfig()
		config.RateLimit = 1
	}
	
	return &URLScan{
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

type urlscanResponse struct {
	Results []struct {
		Page struct {
			Domain string `json:"domain"`
		} `json:"page"`
	} `json:"results"`
}

// Run executes the URLScan source
func (us *URLScan) Run(ctx context.Context, domain string) ([]string, error) {
	// Build URL
	apiURL := fmt.Sprintf("https://urlscan.io/api/v1/search/?q=domain:%s", url.QueryEscape(domain))
	
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("User-Agent", us.config.UserAgent)
	
	// Add API key if provided
	if us.config.APIKey != "" {
		req.Header.Set("API-Key", us.config.APIKey)
	}
	
	// Execute request with retry
	var resp *http.Response
	var lastErr error
	
	for i := 0; i < us.config.Retry; i++ {
		resp, lastErr = us.client.Do(req)
		if lastErr == nil && resp.StatusCode == http.StatusOK {
			break
		}
		
		if resp != nil {
			resp.Body.Close()
		}
		
		// Handle rate limiting
		if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
			if i < us.config.Retry-1 {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(time.Duration(i+1) * 2 * time.Second):
				}
				continue
			}
		}
		
		// Exponential backoff
		if i < us.config.Retry-1 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(i+1) * time.Second):
			}
		}
	}
	
	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", us.config.Retry, lastErr)
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
	
	var result urlscanResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	
	// Extract unique subdomains
	subdomainMap := make(map[string]bool)
	for _, item := range result.Results {
		subdomain := strings.TrimSpace(item.Page.Domain)
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
func (us *URLScan) Name() string {
	return "urlscan"
}

// NeedsKey indicates if API key is required
func (us *URLScan) NeedsKey() bool {
	return false // Optional
}
