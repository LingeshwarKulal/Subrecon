package sources

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HackerTarget queries hackertarget.com API
type HackerTarget struct {
	config *SourceConfig
	client *http.Client
}

// NewHackerTarget creates a new HackerTarget source
func NewHackerTarget(config *SourceConfig) *HackerTarget {
	if config == nil {
		config = DefaultConfig()
		config.RateLimit = 2
	}
	
	return &HackerTarget{
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

// Run executes the HackerTarget source
func (ht *HackerTarget) Run(ctx context.Context, domain string) ([]string, error) {
	// Build URL
	apiURL := fmt.Sprintf("https://api.hackertarget.com/hostsearch/?q=%s", url.QueryEscape(domain))
	
	// Add API key if provided
	if ht.config.APIKey != "" {
		apiURL += "&apikey=" + url.QueryEscape(ht.config.APIKey)
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("User-Agent", ht.config.UserAgent)
	
	// Execute request with retry
	var resp *http.Response
	var lastErr error
	
	for i := 0; i < ht.config.Retry; i++ {
		resp, lastErr = ht.client.Do(req)
		if lastErr == nil && resp.StatusCode == http.StatusOK {
			break
		}
		
		if resp != nil {
			resp.Body.Close()
		}
		
		// Handle rate limiting
		if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
			if i < ht.config.Retry-1 {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(time.Duration(i+1) * 2 * time.Second):
				}
				continue
			}
		}
		
		// Exponential backoff
		if i < ht.config.Retry-1 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(i+1) * time.Second):
			}
		}
	}
	
	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", ht.config.Retry, lastErr)
	}
	
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	defer resp.Body.Close()
	
	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	content := string(body)
	
	// Check for error messages
	if strings.Contains(content, "error") || strings.Contains(content, "API count exceeded") {
		return nil, fmt.Errorf("API error: %s", content)
	}
	
	// Parse response (format: subdomain.domain.com,ip_address)
	lines := strings.Split(content, "\n")
	subdomainMap := make(map[string]bool)
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Split by comma
		parts := strings.Split(line, ",")
		if len(parts) > 0 {
			subdomain := strings.TrimSpace(parts[0])
			subdomain = strings.ToLower(subdomain)
			
			if subdomain != "" && (strings.HasSuffix(subdomain, "."+domain) || subdomain == domain) {
				subdomainMap[subdomain] = true
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
func (ht *HackerTarget) Name() string {
	return "hackertarget"
}

// NeedsKey indicates if API key is required
func (ht *HackerTarget) NeedsKey() bool {
	return false // Optional
}
