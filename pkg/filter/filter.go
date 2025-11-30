package filter

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Filter represents a subdomain filter
type Filter struct {
	matchPatterns   []*regexp.Regexp
	excludePatterns []*regexp.Regexp
}

// NewFilter creates a new filter
func NewFilter() *Filter {
	return &Filter{
		matchPatterns:   make([]*regexp.Regexp, 0),
		excludePatterns: make([]*regexp.Regexp, 0),
	}
}

// AddMatchPattern adds a match pattern (regex)
func (f *Filter) AddMatchPattern(pattern string) error {
	// Check if it's a file reference
	if strings.HasPrefix(pattern, "@") {
		return f.loadPatternsFromFile(pattern[1:], true)
	}
	
	// Split by comma for multiple patterns
	patterns := strings.Split(pattern, ",")
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		
		regex, err := regexp.Compile(p)
		if err != nil {
			return fmt.Errorf("invalid match pattern '%s': %w", p, err)
		}
		f.matchPatterns = append(f.matchPatterns, regex)
	}
	
	return nil
}

// AddExcludePattern adds an exclude pattern (regex)
func (f *Filter) AddExcludePattern(pattern string) error {
	// Check if it's a file reference
	if strings.HasPrefix(pattern, "@") {
		return f.loadPatternsFromFile(pattern[1:], false)
	}
	
	// Split by comma for multiple patterns
	patterns := strings.Split(pattern, ",")
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		
		regex, err := regexp.Compile(p)
		if err != nil {
			return fmt.Errorf("invalid exclude pattern '%s': %w", p, err)
		}
		f.excludePatterns = append(f.excludePatterns, regex)
	}
	
	return nil
}

// loadPatternsFromFile loads patterns from a file
func (f *Filter) loadPatternsFromFile(path string, isMatch bool) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open pattern file: %w", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		pattern := strings.TrimSpace(scanner.Text())
		if pattern == "" || strings.HasPrefix(pattern, "#") {
			continue
		}
		
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid pattern in file '%s': %w", pattern, err)
		}
		
		if isMatch {
			f.matchPatterns = append(f.matchPatterns, regex)
		} else {
			f.excludePatterns = append(f.excludePatterns, regex)
		}
	}
	
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read pattern file: %w", err)
	}
	
	return nil
}

// Apply applies the filter to a list of subdomains
func (f *Filter) Apply(subdomains []string) []string {
	if len(f.matchPatterns) == 0 && len(f.excludePatterns) == 0 {
		return subdomains
	}
	
	filtered := make([]string, 0, len(subdomains))
	
	for _, subdomain := range subdomains {
		// Check match patterns (must match at least one if any are specified)
		if len(f.matchPatterns) > 0 {
			matched := false
			for _, pattern := range f.matchPatterns {
				if pattern.MatchString(subdomain) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		
		// Check exclude patterns (must not match any)
		excluded := false
		for _, pattern := range f.excludePatterns {
			if pattern.MatchString(subdomain) {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}
		
		filtered = append(filtered, subdomain)
	}
	
	return filtered
}

// Deduplicate removes duplicate subdomains
func Deduplicate(subdomains []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(subdomains))
	
	for _, subdomain := range subdomains {
		subdomain = strings.ToLower(strings.TrimSpace(subdomain))
		if subdomain != "" && !seen[subdomain] {
			seen[subdomain] = true
			result = append(result, subdomain)
		}
	}
	
	return result
}
