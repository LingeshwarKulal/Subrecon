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

// AlienVault queries otx.alienvault.com API
type AlienVault struct {
	config *SourceConfig
	client *http.Client
}

// NewAlienVault creates a new AlienVault source
func NewAlienVault(config *SourceConfig) *AlienVault {
	if config == nil {
		config = DefaultConfig()
		config.RateLimit = 10
	}
	
	return &AlienVault{
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

type alienVaultResponse struct {
	PassiveDNS []struct {
		Hostname string `json:"hostname"`
	} `json:"passive_dns"`
}

// Run executes the AlienVault source
func (av *AlienVault) Run(ctx context.Context, domain string) ([]string, error) {
	// Check if API key is provided
	if av.config.APIKey == "" {
		return nil, fmt.Errorf("AlienVault requires an API key")
	}
	
	// Build URL
	apiURL := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/passive_dns", 
		url.QueryEscape(domain))
	
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("User-Agent", av.config.UserAgent)
	req.Header.Set("X-OTX-API-KEY", av.config.APIKey)
	
	// Execute request with retry
	var resp *http.Response
	var lastErr error
	
	for i := 0; i < av.config.Retry; i++ {
		resp, lastErr = av.client.Do(req)
		if lastErr == nil && resp.StatusCode == http.StatusOK {
			break
		}
		
		if resp != nil {
			resp.Body.Close()
		}
		
		// Handle rate limiting
		if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
			if i < av.config.Retry-1 {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(time.Duration(i+1) * 2 * time.Second):
				}
				continue
			}
		}
		
		// Exponential backoff
		if i < av.config.Retry-1 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(i+1) * time.Second):
			}
		}
	}
	
	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", av.config.Retry, lastErr)
	}
	
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		return nil, fmt.Errorf("invalid API key")
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
	
	var result alienVaultResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	
	// Extract unique subdomains
	subdomainMap := make(map[string]bool)
	for _, record := range result.PassiveDNS {
		hostname := strings.TrimSpace(record.Hostname)
		hostname = strings.ToLower(hostname)
		
		if hostname != "" && (strings.HasSuffix(hostname, "."+domain) || hostname == domain) {
			subdomainMap[hostname] = true
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
func (av *AlienVault) Name() string {
	return "alienvault"
}

// NeedsKey indicates if API key is required
func (av *AlienVault) NeedsKey() bool {
	return true
}
