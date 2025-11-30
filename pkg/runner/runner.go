package runner

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/yourusername/subfinder-pro/pkg/sources"
	"golang.org/x/time/rate"
)

// Runner manages the execution of multiple sources
type Runner struct {
	sources      []sources.Source
	workers      int
	timeout      time.Duration
	rateLimiters map[string]*rate.Limiter
	verbose      bool
	silent       bool
}

// Config holds runner configuration
type Config struct {
	Workers int
	Timeout time.Duration
	Verbose bool
	Silent  bool
}

// NewRunner creates a new runner
func NewRunner(srcs []sources.Source, config *Config) *Runner {
	if config == nil {
		config = &Config{
			Workers: 10,
			Timeout: 30 * time.Second,
			Verbose: false,
			Silent:  false,
		}
	}
	
	return &Runner{
		sources:      srcs,
		workers:      config.Workers,
		timeout:      config.Timeout,
		rateLimiters: make(map[string]*rate.Limiter),
		verbose:      config.Verbose,
		silent:       config.Silent,
	}
}

// SetRateLimit sets the rate limit for a specific source
func (r *Runner) SetRateLimit(sourceName string, requestsPerSecond int) {
	r.rateLimiters[sourceName] = rate.NewLimiter(rate.Limit(requestsPerSecond), 1)
}

// Result holds the result from a source
type Result struct {
	Source     string
	Subdomains []string
	Error      error
}

// Run executes all sources and returns unique subdomains
func (r *Runner) Run(ctx context.Context, domain string) ([]string, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain cannot be empty")
	}
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	
	// Channels for communication
	resultsChan := make(chan Result, len(r.sources))
	
	// Worker pool using semaphore
	sem := make(chan struct{}, r.workers)
	var wg sync.WaitGroup
	
	// Launch goroutines for each source
	for _, source := range r.sources {
		wg.Add(1)
		go func(src sources.Source) {
			defer wg.Done()
			
			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()
			
			// Apply rate limiting if configured
			if limiter, ok := r.rateLimiters[src.Name()]; ok {
				if err := limiter.Wait(ctx); err != nil {
					resultsChan <- Result{
						Source: src.Name(),
						Error:  fmt.Errorf("rate limit wait failed: %w", err),
					}
					return
				}
			}
			
			if r.verbose && !r.silent {
				fmt.Printf("[*] Running source: %s\n", src.Name())
			}
			
			// Execute source
			subdomains, err := src.Run(ctx, domain)
			
			if err != nil && r.verbose && !r.silent {
				fmt.Printf("[-] Error from %s: %v\n", src.Name(), err)
			}
			
			resultsChan <- Result{
				Source:     src.Name(),
				Subdomains: subdomains,
				Error:      err,
			}
		}(source)
	}
	
	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()
	
	// Collect results and deduplicate
	subdomainMap := make(map[string]string) // subdomain -> source
	var errors []error
	
	for result := range resultsChan {
		if result.Error != nil {
			errors = append(errors, fmt.Errorf("%s: %w", result.Source, result.Error))
			continue
		}
		
		for _, subdomain := range result.Subdomains {
			if _, exists := subdomainMap[subdomain]; !exists {
				subdomainMap[subdomain] = result.Source
			}
		}
		
		if r.verbose && !r.silent {
			fmt.Printf("[+] %s found %d subdomains\n", result.Source, len(result.Subdomains))
		}
	}
	
	// Convert map to sorted slice
	subdomains := make([]string, 0, len(subdomainMap))
	for subdomain := range subdomainMap {
		subdomains = append(subdomains, subdomain)
	}
	sort.Strings(subdomains)
	
	// Return error only if all sources failed
	if len(subdomains) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("all sources failed: %v", errors)
	}
	
	return subdomains, nil
}

// RunWithMetadata executes all sources and returns subdomains with metadata
func (r *Runner) RunWithMetadata(ctx context.Context, domain string) ([]SubdomainResult, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain cannot be empty")
	}
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	
	// Channels for communication
	resultsChan := make(chan Result, len(r.sources))
	
	// Worker pool using semaphore
	sem := make(chan struct{}, r.workers)
	var wg sync.WaitGroup
	
	// Launch goroutines for each source
	for _, source := range r.sources {
		wg.Add(1)
		go func(src sources.Source) {
			defer wg.Done()
			
			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()
			
			// Apply rate limiting if configured
			if limiter, ok := r.rateLimiters[src.Name()]; ok {
				if err := limiter.Wait(ctx); err != nil {
					resultsChan <- Result{
						Source: src.Name(),
						Error:  fmt.Errorf("rate limit wait failed: %w", err),
					}
					return
				}
			}
			
			if r.verbose && !r.silent {
				fmt.Printf("[*] Running source: %s\n", src.Name())
			}
			
			// Execute source
			subdomains, err := src.Run(ctx, domain)
			
			if err != nil && r.verbose && !r.silent {
				fmt.Printf("[-] Error from %s: %v\n", src.Name(), err)
			}
			
			resultsChan <- Result{
				Source:     src.Name(),
				Subdomains: subdomains,
				Error:      err,
			}
		}(source)
	}
	
	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()
	
	// Collect results with metadata
	subdomainMap := make(map[string]*SubdomainResult)
	var errors []error
	
	for result := range resultsChan {
		if result.Error != nil {
			errors = append(errors, fmt.Errorf("%s: %w", result.Source, result.Error))
			continue
		}
		
		for _, subdomain := range result.Subdomains {
			if _, exists := subdomainMap[subdomain]; !exists {
				subdomainMap[subdomain] = &SubdomainResult{
					Host:      subdomain,
					Source:    result.Source,
					Timestamp: time.Now(),
				}
			}
		}
		
		if r.verbose && !r.silent {
			fmt.Printf("[+] %s found %d subdomains\n", result.Source, len(result.Subdomains))
		}
	}
	
	// Convert map to sorted slice
	results := make([]SubdomainResult, 0, len(subdomainMap))
	for _, result := range subdomainMap {
		results = append(results, *result)
	}
	
	// Sort by host
	sort.Slice(results, func(i, j int) bool {
		return results[i].Host < results[j].Host
	})
	
	// Return error only if all sources failed
	if len(results) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("all sources failed: %v", errors)
	}
	
	return results, nil
}

// SubdomainResult holds subdomain with metadata
type SubdomainResult struct {
	Host      string    `json:"host"`
	Source    string    `json:"source"`
	Timestamp time.Time `json:"timestamp"`
	IPs       []string  `json:"ips,omitempty"`
}
