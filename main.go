package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/subrecon/internal/resolve"
	"github.com/yourusername/subrecon/pkg/config"
	"github.com/yourusername/subrecon/pkg/filter"
	"github.com/yourusername/subrecon/pkg/output"
	"github.com/yourusername/subrecon/pkg/runner"
	"github.com/yourusername/subrecon/pkg/sources"
)

const version = "1.0.0"

const banner = `
   _____       __   ____                      
  / ___/__  __/ /_ / __ \___  _________  ____ 
  \__ \/ / / / __ \/ /_/ / _ \/ ___/ __ \/ __ \
 ___/ / /_/ / /_/ / _, _/  __/ /__/ /_/ / / / /
/____/\__,_/_.___/_/ |_|\___/\___/\____/_/ /_/ 
                                                
        Advanced Subdomain Enumeration Tool    
                 Version ` + version + `
                  By Lingesan
`

var (
	// Flags
	domain         string
	domainList     string
	outputFile     string
	sourceList     string
	useAllSources  bool
	excludeSources string
	jsonOutput     bool
	silentMode     bool
	timeoutSec     int
	workers        int
	configPath     string
	activeMode     bool
	matchPattern   string
	filterPattern  string
	rateLimit      int
	proxyURL       string
	verbose        bool
	showVersion    bool
)

var rootCmd = &cobra.Command{
	Use:   "subrecon",
	Short: "SubRecon - Advanced Subdomain Enumeration Tool",
	Long: `SubRecon is a high-performance subdomain enumeration tool that discovers
subdomains using multiple passive sources including Certificate Transparency logs,
search engines, and threat intelligence platforms.

Features:
  • Multiple passive reconnaissance sources
  • Concurrent processing with worker pools
  • DNS verification with wildcard detection
  • Pattern-based filtering
  • Rate limiting per source
  • JSON and text output formats`,
	RunE: run,
}

func init() {
	rootCmd.Flags().StringVarP(&domain, "domain", "d", "", "Target domain (e.g., example.com)")
	rootCmd.Flags().StringVar(&domainList, "domain-list", "", "File containing list of domains")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default: stdout)")
	rootCmd.Flags().StringVarP(&sourceList, "sources", "s", "", "Comma-separated list of sources to use")
	rootCmd.Flags().BoolVar(&useAllSources, "all", true, "Use all available sources")
	rootCmd.Flags().StringVar(&excludeSources, "exclude-sources", "", "Comma-separated list of sources to exclude")
	rootCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	rootCmd.Flags().BoolVar(&silentMode, "silent", false, "Suppress progress and error messages")
	rootCmd.Flags().IntVar(&timeoutSec, "timeout", 30, "Timeout in seconds per source")
	rootCmd.Flags().IntVarP(&workers, "threads", "t", 10, "Number of concurrent workers")
	rootCmd.Flags().StringVarP(&configPath, "config", "c", "config.yaml", "Path to config file")
	rootCmd.Flags().BoolVar(&activeMode, "active", false, "Enable DNS verification")
	rootCmd.Flags().StringVarP(&matchPattern, "match", "m", "", "Match patterns (regex or comma-separated)")
	rootCmd.Flags().StringVarP(&filterPattern, "filter", "f", "", "Filter patterns (exclude matches)")
	rootCmd.Flags().IntVar(&rateLimit, "rate-limit", 5, "Rate limit (requests/second)")
	rootCmd.Flags().StringVar(&proxyURL, "proxy", "", "HTTP proxy URL")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.Flags().BoolVar(&showVersion, "version", false, "Show version information")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// Show banner in verbose mode
	if verbose && !silentMode {
		fmt.Println(banner)
	}
	
	// Show version
	if showVersion {
		fmt.Print(banner)
		fmt.Printf("\nSubRecon v%s\n", version)
		return nil
	}
	
	// Validate input
	if domain == "" && domainList == "" {
		return fmt.Errorf("either -d or -dL flag is required")
	}
	
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	// Load provider configuration
	providerCfg, err := config.LoadProviderConfig("provider-config.yaml")
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load provider config: %w", err)
	}
	
	// Override config with CLI flags
	if timeoutSec > 0 {
		cfg.Timeout = timeoutSec
	}
	if workers > 0 {
		cfg.Workers = workers
	}
	if jsonOutput {
		cfg.Output.Format = "json"
	}
	
	// Get domains to process
	domains := make([]string, 0)
	if domain != "" {
		domains = append(domains, domain)
	}
	if domainList != "" {
		fileDomains, err := readDomainsFromFile(domainList)
		if err != nil {
			return fmt.Errorf("failed to read domain list: %w", err)
		}
		domains = append(domains, fileDomains...)
	}
	
	// Process each domain
	allResults := make([]runner.SubdomainResult, 0)
	
	for _, dom := range domains {
		dom = strings.TrimSpace(dom)
		if dom == "" {
			continue
		}
		
		// Validate domain
		if err := resolve.ValidateDomain(dom); err != nil {
			if !silentMode {
				fmt.Fprintf(os.Stderr, "[-] Invalid domain %s: %v\n", dom, err)
			}
			continue
		}
		
		if verbose && !silentMode {
			fmt.Printf("[*] Processing domain: %s\n", dom)
		}
		
		// Initialize sources
		srcs, err := initializeSources(providerCfg, sourceList, excludeSources)
		if err != nil {
			return err
		}
		
		// Create runner
		runnerCfg := &runner.Config{
			Workers: cfg.Workers,
			Timeout: cfg.GetTimeout(),
			Verbose: verbose,
			Silent:  silentMode,
		}
		r := runner.NewRunner(srcs, runnerCfg)
		
		// Set rate limits
		for _, src := range srcs {
			srcCfg := providerCfg.GetSourceConfig(src.Name())
			if srcCfg.RateLimit > 0 {
				r.SetRateLimit(src.Name(), srcCfg.RateLimit)
			}
		}
		
		// Run enumeration
		ctx := context.Background()
		results, err := r.RunWithMetadata(ctx, dom)
		if err != nil {
			if !silentMode {
				fmt.Fprintf(os.Stderr, "[-] Error processing %s: %v\n", dom, err)
			}
			continue
		}
		
		if verbose && !silentMode {
			fmt.Printf("[+] Found %d subdomains for %s\n", len(results), dom)
		}
		
		// Apply filtering
		if matchPattern != "" || filterPattern != "" {
			f := filter.NewFilter()
			
			if matchPattern != "" {
				if err := f.AddMatchPattern(matchPattern); err != nil {
					return fmt.Errorf("invalid match pattern: %w", err)
				}
			}
			
			if filterPattern != "" {
				if err := f.AddExcludePattern(filterPattern); err != nil {
					return fmt.Errorf("invalid filter pattern: %w", err)
				}
			}
			
			// Extract hosts for filtering
			hosts := make([]string, len(results))
			for i, r := range results {
				hosts[i] = r.Host
			}
			
			// Apply filter
			filteredHosts := f.Apply(hosts)
			
			// Rebuild results
			filteredResults := make([]runner.SubdomainResult, 0, len(filteredHosts))
			hostMap := make(map[string]runner.SubdomainResult)
			for _, r := range results {
				hostMap[r.Host] = r
			}
			for _, host := range filteredHosts {
				if r, ok := hostMap[host]; ok {
					filteredResults = append(filteredResults, r)
				}
			}
			results = filteredResults
			
			if verbose && !silentMode {
				fmt.Printf("[+] %d subdomains after filtering\n", len(results))
			}
		}
		
		// DNS verification
		if activeMode {
			if verbose && !silentMode {
				fmt.Printf("[*] Performing DNS verification...\n")
			}
			
			resolver := resolve.NewResolver(&resolve.Config{
				Servers: cfg.DNS.Servers,
				Timeout: cfg.GetDNSTimeout(),
			})
			
			// Detect wildcard
			isWildcard, wildcardIPs, _ := resolver.DetectWildcard(ctx, dom)
			if isWildcard && verbose && !silentMode {
				fmt.Printf("[!] Wildcard DNS detected for %s: %v\n", dom, wildcardIPs)
			}
			
			// Resolve all subdomains
			verifiedResults := make([]runner.SubdomainResult, 0)
			for _, result := range results {
				res, err := resolver.Resolve(ctx, result.Host)
				if err == nil && res.Exists {
					// Check if it's not a wildcard
					if !isWildcard || !resolver.IsWildcard(result.Host, dom) {
						result.IPs = res.IPs
						verifiedResults = append(verifiedResults, result)
					}
				}
			}
			
			results = verifiedResults
			
			if verbose && !silentMode {
				fmt.Printf("[+] %d subdomains verified via DNS\n", len(results))
			}
		}
		
		allResults = append(allResults, results...)
	}
	
	// Write output
	if len(allResults) == 0 {
		if !silentMode {
			fmt.Fprintln(os.Stderr, "[-] No subdomains found")
		}
		return nil
	}
	
	// Format and write output
	if jsonOutput {
		formatter := output.NewJSONFormatter(cfg.Output.Sort)
		if err := output.WriteResults(allResults, formatter, outputFile); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	} else {
		formatter := output.NewTextFormatter(cfg.Output.Sort)
		if err := output.WriteResults(allResults, formatter, outputFile); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}
	
	if !silentMode && outputFile != "" {
		fmt.Printf("[+] Results saved to %s\n", outputFile)
	}
	
	return nil
}

func initializeSources(cfg *config.ProviderConfig, sourceList, excludeSources string) ([]sources.Source, error) {
	allSources := map[string]func(*sources.SourceConfig) sources.Source{
		"crtsh":       func(c *sources.SourceConfig) sources.Source { return sources.NewCrtSh(c) },
		"hackertarget": func(c *sources.SourceConfig) sources.Source { return sources.NewHackerTarget(c) },
		"threatcrowd": func(c *sources.SourceConfig) sources.Source { return sources.NewThreatCrowd(c) },
		"alienvault":  func(c *sources.SourceConfig) sources.Source { return sources.NewAlienVault(c) },
		"urlscan":     func(c *sources.SourceConfig) sources.Source { return sources.NewURLScan(c) },
	}
	
	// Determine which sources to use
	var sourcesToUse []string
	if sourceList != "" {
		sourcesToUse = strings.Split(sourceList, ",")
		for i := range sourcesToUse {
			sourcesToUse[i] = strings.TrimSpace(sourcesToUse[i])
		}
	} else {
		// Use all sources
		for name := range allSources {
			sourcesToUse = append(sourcesToUse, name)
		}
	}
	
	// Exclude sources if specified
	if excludeSources != "" {
		excludeList := strings.Split(excludeSources, ",")
		excludeMap := make(map[string]bool)
		for _, name := range excludeList {
			excludeMap[strings.TrimSpace(name)] = true
		}
		
		filtered := make([]string, 0)
		for _, name := range sourcesToUse {
			if !excludeMap[name] {
				filtered = append(filtered, name)
			}
		}
		sourcesToUse = filtered
	}
	
	// Initialize sources
	srcs := make([]sources.Source, 0)
	for _, name := range sourcesToUse {
		factory, ok := allSources[name]
		if !ok {
			return nil, fmt.Errorf("unknown source: %s", name)
		}
		
		// Check if enabled
		if !cfg.IsSourceEnabled(name) {
			if verbose && !silentMode {
				fmt.Printf("[-] Source %s is disabled\n", name)
			}
			continue
		}
		
		srcCfg := cfg.GetSourceConfig(name)
		src := factory(srcCfg)
		
		// Check if source needs API key
		if src.NeedsKey() && srcCfg.APIKey == "" {
			if !silentMode {
				fmt.Fprintf(os.Stderr, "[!] Warning: %s requires an API key, skipping\n", name)
			}
			continue
		}
		
		srcs = append(srcs, src)
	}
	
	if len(srcs) == 0 {
		return nil, fmt.Errorf("no sources available")
	}
	
	return srcs, nil
}

func readDomainsFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	domains := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		domain := strings.TrimSpace(scanner.Text())
		if domain != "" && !strings.HasPrefix(domain, "#") {
			domains = append(domains, domain)
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	
	return domains, nil
}
