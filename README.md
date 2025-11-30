# SubFinder Pro

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

**SubFinder Pro** is a high-performance, production-ready passive subdomain enumeration tool written in Go. It aggregates subdomain data from multiple public sources using concurrent processing, DNS verification, and advanced filtering capabilities.

## 🚀 Features

- **Multiple Data Sources**: Aggregates data from 5+ passive reconnaissance sources
  - Certificate Transparency (crt.sh)
  - HackerTarget API
  - ThreatCrowd
  - AlienVault OTX
  - URLScan.io
- **Concurrent Processing**: Worker pool pattern with configurable concurrency
- **DNS Verification**: Active DNS resolution with wildcard detection
- **Smart Filtering**: Regex-based pattern matching and exclusion
- **Rate Limiting**: Per-source configurable rate limits
- **Multiple Output Formats**: Plain text and JSON (JSONL)
- **Flexible Configuration**: YAML config files with environment variable support
- **Production Ready**: Comprehensive error handling and retry logic

## 📦 Installation

### From Source

Requires Go 1.21 or higher:

```bash
git clone https://github.com/yourusername/subfinder-pro.git
cd subfinder-pro
go mod download
go build -o subfinder-pro main.go
```

### Using Go Install

```bash
go install github.com/yourusername/subfinder-pro@latest
```

### Pre-built Binaries

Download pre-built binaries from the [Releases](https://github.com/yourusername/subfinder-pro/releases) page.

## 🎯 Quick Start

### Basic Usage

```bash
# Enumerate subdomains for a single domain
./subfinder-pro -d example.com

# Save results to file
./subfinder-pro -d example.com -o results.txt

# Multiple domains from file
./subfinder-pro -dL domains.txt -o results.txt

# JSON output
./subfinder-pro -d example.com -json -o results.json
```

### Advanced Usage

```bash
# With DNS verification
./subfinder-pro -d example.com -active

# Use specific sources only
./subfinder-pro -d example.com -s crtsh,alienvault

# Exclude specific sources
./subfinder-pro -d example.com -es threatcrowd

# Pattern matching (find only api/dev/staging subdomains)
./subfinder-pro -d example.com -m "^(api|dev|staging)\."

# Filter out test/internal subdomains
./subfinder-pro -d example.com -f "test|internal"

# Verbose output with 20 workers
./subfinder-pro -d example.com -v -t 20

# Silent mode (only output subdomains)
./subfinder-pro -d example.com -silent

# With proxy
./subfinder-pro -d example.com -proxy http://proxy.example.com:8080
```

## ⚙️ Configuration

### Config Files

SubFinder Pro uses two configuration files:

1. **config.yaml**: Global settings (timeout, workers, DNS, output)
2. **provider-config.yaml**: API keys and per-source configuration

### Example: config.yaml

```yaml
timeout: 30
workers: 10
rate_limit: 5

dns:
  enabled: false
  servers:
    - "8.8.8.8:53"
    - "1.1.1.1:53"
  timeout: 5
  retry: 3

output:
  format: text
  sort: true
  unique: true

http:
  user_agent: "SubFinder-Pro/1.0"
  timeout: 10
  proxy: ""
```

### Example: provider-config.yaml

```yaml
sources:
  crtsh:
    enabled: true
    rate_limit: 5
    timeout: 30
  
  alienvault:
    enabled: true
    api_key: "your-api-key-here"
    rate_limit: 10
    timeout: 20
```

### Environment Variables

API keys can be set via environment variables:

```bash
export ALIENVAULT_API_KEY="your-key"
export URLSCAN_API_KEY="your-key"
export HACKERTARGET_API_KEY="your-key"

# Alternative format
export SUBFINDER_ALIENVAULT_API_KEY="your-key"
```

## 🔑 API Keys Setup

### AlienVault OTX (Required)

1. Visit [https://otx.alienvault.com/](https://otx.alienvault.com/)
2. Create a free account
3. Go to Settings → API Integration
4. Copy your API key
5. Set in `provider-config.yaml` or environment variable:
   ```bash
   export ALIENVAULT_API_KEY="your-key"
   ```

### URLScan.io (Optional)

1. Visit [https://urlscan.io/](https://urlscan.io/)
2. Create a free account
3. Go to Settings & API
4. Generate API key
5. Set in `provider-config.yaml` or:
   ```bash
   export URLSCAN_API_KEY="your-key"
   ```

### HackerTarget (Optional)

1. Visit [https://hackertarget.com/](https://hackertarget.com/)
2. Purchase API membership
3. Set API key:
   ```bash
   export HACKERTARGET_API_KEY="your-key"
   ```

## 📋 CLI Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--domain` | `-d` | Target domain | - |
| `--domain-list` | `-dL` | File with domain list | - |
| `--output` | `-o` | Output file path | stdout |
| `--sources` | `-s` | Specific sources to use | all |
| `--all` | - | Use all available sources | true |
| `--exclude-sources` | `-es` | Sources to exclude | - |
| `--json` | - | JSON output format | false |
| `--silent` | - | Silent mode | false |
| `--timeout` | - | Timeout per source (seconds) | 30 |
| `--threads` | `-t` | Concurrent workers | 10 |
| `--config` | `-c` | Config file path | config.yaml |
| `--active` | - | Enable DNS verification | false |
| `--match` | `-m` | Match patterns (regex) | - |
| `--filter` | `-f` | Filter patterns (exclude) | - |
| `--rate-limit` | - | Rate limit (req/sec) | 5 |
| `--proxy` | - | HTTP proxy URL | - |
| `--verbose` | `-v` | Verbose output | false |
| `--version` | - | Show version | false |

## 📤 Output Formats

### Plain Text (Default)

```
api.example.com
blog.example.com
dev.example.com
mail.example.com
www.example.com
```

### JSON (JSONL)

```json
{"host":"api.example.com","source":"crtsh","timestamp":"2025-11-30T23:09:00Z","ips":null}
{"host":"blog.example.com","source":"alienvault","timestamp":"2025-11-30T23:09:01Z","ips":null}
{"host":"dev.example.com","source":"crtsh","timestamp":"2025-11-30T23:09:00Z","ips":null}
```

### JSON with DNS Verification

```json
{"host":"api.example.com","source":"crtsh","timestamp":"2025-11-30T23:09:00Z","ips":["192.0.2.1"]}
{"host":"blog.example.com","source":"alienvault","timestamp":"2025-11-30T23:09:01Z","ips":["192.0.2.2","192.0.2.3"]}
```

## 🔧 Advanced Features

### Wildcard Detection

SubFinder Pro automatically detects wildcard DNS:

```bash
./subfinder-pro -d example.com -active -v
```

The tool will:
1. Test 3 random subdomains
2. Detect if they resolve to the same IP
3. Filter out wildcard matches from results

### Pattern Matching

Use regex patterns to filter results:

```bash
# Match subdomains starting with "api", "dev", or "staging"
./subfinder-pro -d example.com -m "^(api|dev|staging)\."

# Patterns from file
./subfinder-pro -d example.com -m @patterns.txt
```

**patterns.txt:**
```
^api\.
^dev\.
^staging\.
```

### Exclusion Filtering

Exclude unwanted patterns:

```bash
# Exclude test and internal subdomains
./subfinder-pro -d example.com -f "test|internal"

# Filters from file
./subfinder-pro -d example.com -f @exclude.txt
```

### Rate Limiting

Configure per-source rate limits in `provider-config.yaml`:

```yaml
sources:
  crtsh:
    rate_limit: 5  # 5 requests per second
  alienvault:
    rate_limit: 10
```

## 🧪 Testing

### Run Unit Tests

```bash
go test ./... -v
```

### Run Integration Tests

```bash
go test ./tests -v -tags=integration
```

### Run Benchmarks

```bash
go test ./... -bench=. -benchmem
```

## 🐛 Troubleshooting

### No Results Found

1. **Check API keys**: Ensure AlienVault API key is set
2. **Check internet connectivity**: Test with `curl https://crt.sh`
3. **Increase timeout**: Use `--timeout 60`
4. **Try specific sources**: Use `-s crtsh` to test individual sources

### Rate Limit Errors

1. **Reduce workers**: Use `-t 5`
2. **Increase timeout**: Use `--timeout 60`
3. **Use API keys**: Some sources have higher limits with API keys

### DNS Verification Issues

1. **Check DNS servers**: Verify servers in `config.yaml`
2. **Increase DNS timeout**: Set `dns.timeout: 10` in config
3. **Disable wildcard detection**: Use without `-active` flag

### Permission Denied on Output File

```bash
# Check directory permissions
ls -la output_directory/

# Use absolute path
./subfinder-pro -d example.com -o /tmp/results.txt
```

## 📊 Performance

### Benchmarks

- **Throughput**: 100+ subdomains/second
- **Memory**: < 100MB typical usage
- **Concurrency**: Supports 50+ concurrent workers
- **Startup**: < 500ms

### Optimization Tips

1. **Adjust workers**: More workers = faster processing (up to a point)
   ```bash
   ./subfinder-pro -d example.com -t 20
   ```

2. **Use specific sources**: Skip slow sources
   ```bash
   ./subfinder-pro -d example.com -s crtsh,alienvault
   ```

3. **Disable DNS verification**: For faster passive enumeration
   ```bash
   ./subfinder-pro -d example.com
   ```

## 🤝 Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Adding New Sources

To add a new subdomain source:

1. Create a new file in `pkg/sources/` (e.g., `newsource.go`)
2. Implement the `Source` interface:
   ```go
   type NewSource struct {
       config *SourceConfig
       client *http.Client
   }
   
   func (ns *NewSource) Run(ctx context.Context, domain string) ([]string, error) {
       // Implementation
   }
   
   func (ns *NewSource) Name() string { return "newsource" }
   func (ns *NewSource) NeedsKey() bool { return false }
   ```
3. Add to `initializeSources()` in `main.go`
4. Update `provider-config.yaml` with source configuration
5. Add tests in `tests/sources_test.go`

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Certificate Transparency](https://certificate.transparency.dev/) - crt.sh data
- [AlienVault OTX](https://otx.alienvault.com/) - Threat intelligence
- [URLScan.io](https://urlscan.io/) - URL scanning service
- [HackerTarget](https://hackertarget.com/) - Security tools
- [ThreatCrowd](https://www.threatcrowd.org/) - Threat data

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/yourusername/subfinder-pro/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/subfinder-pro/discussions)

## 🗺️ Roadmap

- [ ] Add more data sources (VirusTotal, Censys, Shodan)
- [ ] Database output support (PostgreSQL, MongoDB)
- [ ] Web dashboard for visualization
- [ ] Continuous monitoring mode
- [ ] Diff detection between scans
- [ ] Integration with Nuclei/Nmap
- [ ] Docker container
- [ ] Distributed processing support

---

**Made with ❤️ for the security community**
