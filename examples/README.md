# SubFinder Pro Examples

This directory contains example files for using SubFinder Pro.

## Files

### domains.txt
Example domain list file. Use with `-dL` flag:
```bash
subfinder-pro -dL examples/domains.txt -o results.txt
```

### match-patterns.txt
Example match patterns for filtering. Use with `-m` flag:
```bash
subfinder-pro -d example.com -m @examples/match-patterns.txt
```

### exclude-patterns.txt
Example exclude patterns for filtering. Use with `-f` flag:
```bash
subfinder-pro -d example.com -f @examples/exclude-patterns.txt
```

## Usage Examples

### Basic enumeration
```bash
subfinder-pro -d example.com
```

### Multiple domains with JSON output
```bash
subfinder-pro -dL examples/domains.txt -json -o results.json
```

### With pattern matching
```bash
subfinder-pro -d example.com -m @examples/match-patterns.txt -f @examples/exclude-patterns.txt
```

### With DNS verification
```bash
subfinder-pro -d example.com -active -v -o verified.txt
```

### Silent mode (only output)
```bash
subfinder-pro -d example.com -silent > subdomains.txt
```

### Using specific sources
```bash
subfinder-pro -d example.com -s crtsh,alienvault -v
```

### Excluding sources
```bash
subfinder-pro -d example.com -es threatcrowd -v
```

### High performance mode
```bash
subfinder-pro -d example.com -t 20 -timeout 60 -v
```
