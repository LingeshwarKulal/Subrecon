package resolve

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// Resolver handles DNS resolution with caching and wildcard detection
type Resolver struct {
	servers   []string
	timeout   time.Duration
	cache     *sync.Map
	wildcards map[string][]string // domain -> wildcard IPs
	mu        sync.RWMutex
}

// Config holds resolver configuration
type Config struct {
	Servers []string
	Timeout time.Duration
}

// NewResolver creates a new DNS resolver
func NewResolver(config *Config) *Resolver {
	if config == nil {
		config = &Config{
			Servers: []string{"8.8.8.8:53", "1.1.1.1:53"},
			Timeout: 5 * time.Second,
		}
	}
	
	return &Resolver{
		servers:   config.Servers,
		timeout:   config.Timeout,
		cache:     &sync.Map{},
		wildcards: make(map[string][]string),
	}
}

// Result holds DNS resolution result
type Result struct {
	Host   string
	IPs    []string
	Exists bool
	Error  error
}

// Resolve resolves a subdomain to IP addresses
func (r *Resolver) Resolve(ctx context.Context, subdomain string) (*Result, error) {
	// Check cache
	if cached, ok := r.cache.Load(subdomain); ok {
		return cached.(*Result), nil
	}
	
	// Create resolver with timeout
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: r.timeout,
			}
			// Use custom DNS servers if provided
			if len(r.servers) > 0 {
				return d.DialContext(ctx, network, r.servers[0])
			}
			return d.DialContext(ctx, network, address)
		},
	}
	
	// Resolve with retry
	var ips []string
	var lastErr error
	
	for attempt := 0; attempt < 3; attempt++ {
		addrs, err := resolver.LookupHost(ctx, subdomain)
		if err == nil {
			ips = addrs
			break
		}
		
		lastErr = err
		
		// Check if it's a timeout or temporary error
		if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
			// Exponential backoff
			if attempt < 2 {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(time.Duration(attempt+1) * time.Second):
				}
				continue
			}
		}
		
		// Non-temporary error, don't retry
		break
	}
	
	result := &Result{
		Host:   subdomain,
		IPs:    ips,
		Exists: len(ips) > 0,
		Error:  lastErr,
	}
	
	// Cache result
	r.cache.Store(subdomain, result)
	
	return result, nil
}

// DetectWildcard detects if a domain has wildcard DNS
func (r *Resolver) DetectWildcard(ctx context.Context, domain string) (bool, []string, error) {
	// Check if already detected
	r.mu.RLock()
	if ips, exists := r.wildcards[domain]; exists {
		r.mu.RUnlock()
		return true, ips, nil
	}
	r.mu.RUnlock()
	
	// Generate 3 random subdomains
	randomSubdomains := make([]string, 3)
	for i := 0; i < 3; i++ {
		random := generateRandomString(16)
		randomSubdomains[i] = fmt.Sprintf("%s.%s", random, domain)
	}
	
	// Resolve random subdomains
	var resolvedIPs [][]string
	for _, subdomain := range randomSubdomains {
		result, err := r.Resolve(ctx, subdomain)
		if err == nil && result.Exists {
			resolvedIPs = append(resolvedIPs, result.IPs)
		}
	}
	
	// If 2 or more random subdomains resolve, it's likely a wildcard
	if len(resolvedIPs) >= 2 {
		// Find common IPs
		commonIPs := findCommonIPs(resolvedIPs)
		
		if len(commonIPs) > 0 {
			r.mu.Lock()
			r.wildcards[domain] = commonIPs
			r.mu.Unlock()
			
			return true, commonIPs, nil
		}
	}
	
	return false, nil, nil
}

// IsWildcard checks if a subdomain resolves to wildcard IPs
func (r *Resolver) IsWildcard(subdomain string, domain string) bool {
	r.mu.RLock()
	wildcardIPs, exists := r.wildcards[domain]
	r.mu.RUnlock()
	
	if !exists {
		return false
	}
	
	// Check if subdomain IPs match wildcard IPs
	result, err := r.Resolve(context.Background(), subdomain)
	if err != nil || !result.Exists {
		return false
	}
	
	// Check if any IP matches wildcard
	for _, ip := range result.IPs {
		for _, wildcardIP := range wildcardIPs {
			if ip == wildcardIP {
				return true
			}
		}
	}
	
	return false
}

// ResolveMany resolves multiple subdomains concurrently
func (r *Resolver) ResolveMany(ctx context.Context, subdomains []string, workers int) ([]*Result, error) {
	if workers <= 0 {
		workers = 10
	}
	
	results := make([]*Result, 0, len(subdomains))
	resultsChan := make(chan *Result, len(subdomains))
	
	// Worker pool
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	
	for _, subdomain := range subdomains {
		wg.Add(1)
		go func(sub string) {
			defer wg.Done()
			
			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()
			
			result, _ := r.Resolve(ctx, sub)
			if result != nil {
				resultsChan <- result
			}
		}(subdomain)
	}
	
	// Wait for all goroutines
	go func() {
		wg.Wait()
		close(resultsChan)
	}()
	
	// Collect results
	for result := range resultsChan {
		results = append(results, result)
	}
	
	return results, nil
}

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	rand.Read(b)
	
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = charset[int(b[i])%len(charset)]
	}
	
	return string(result)
}

// findCommonIPs finds IPs that appear in at least 2 lists
func findCommonIPs(ipLists [][]string) []string {
	ipCount := make(map[string]int)
	
	for _, ips := range ipLists {
		seen := make(map[string]bool)
		for _, ip := range ips {
			if !seen[ip] {
				ipCount[ip]++
				seen[ip] = true
			}
		}
	}
	
	// Find IPs that appear in at least 2 lists
	var commonIPs []string
	for ip, count := range ipCount {
		if count >= 2 {
			commonIPs = append(commonIPs, ip)
		}
	}
	
	return commonIPs
}

// FilterWildcard filters out subdomains that resolve to wildcard IPs
func (r *Resolver) FilterWildcard(ctx context.Context, subdomains []string, domain string) ([]string, error) {
	// First detect wildcard
	isWildcard, _, err := r.DetectWildcard(ctx, domain)
	if err != nil {
		return subdomains, err // Return original list on error
	}
	
	if !isWildcard {
		return subdomains, nil // No wildcard, return all
	}
	
	// Filter out wildcard matches
	filtered := make([]string, 0, len(subdomains))
	for _, subdomain := range subdomains {
		if !r.IsWildcard(subdomain, domain) {
			filtered = append(filtered, subdomain)
		}
	}
	
	return filtered, nil
}

// ValidateDomain validates domain format
func ValidateDomain(domain string) error {
	domain = strings.TrimSpace(domain)
	domain = strings.ToLower(domain)
	
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}
	
	// Basic validation
	if strings.Contains(domain, " ") {
		return fmt.Errorf("domain cannot contain spaces")
	}
	
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return fmt.Errorf("domain cannot start or end with a dot")
	}
	
	// Check for valid characters
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return fmt.Errorf("domain must have at least two parts")
	}
	
	return nil
}
