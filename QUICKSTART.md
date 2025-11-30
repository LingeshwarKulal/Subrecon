# Quick Start Guide - SubFinder Pro

## Prerequisites

1. **Install Go** (version 1.21+)
   - Windows: https://go.dev/dl/
   - Download and run the installer
   - Verify: `go version`

## Installation Steps

### 1. Navigate to Project Directory
```powershell
cd e:\newpro\subfinder-pro
```

### 2. Download Dependencies
```powershell
go mod download
```

### 3. Build the Project
```powershell
go build -o subfinder-pro.exe main.go
```

## Configuration

### Option 1: Use Configuration File (Recommended)
Edit `provider-config.yaml` and add your API keys:
```yaml
sources:
  alienvault:
    enabled: true
    api_key: "YOUR_API_KEY_HERE"  # Get from https://otx.alienvault.com/
```

### Option 2: Use Environment Variables
```powershell
$env:ALIENVAULT_API_KEY="your-key-here"
$env:URLSCAN_API_KEY="your-key-here"
```

## Basic Usage

### 1. Simple Enumeration
```powershell
.\subfinder-pro.exe -d example.com
```

### 2. Save to File
```powershell
.\subfinder-pro.exe -d example.com -o results.txt
```

### 3. Multiple Domains
```powershell
.\subfinder-pro.exe -dL examples\domains.txt -o results.txt
```

### 4. JSON Output
```powershell
.\subfinder-pro.exe -d example.com -json -o results.json
```

### 5. With DNS Verification
```powershell
.\subfinder-pro.exe -d example.com -active -v
```

### 6. Verbose Mode
```powershell
.\subfinder-pro.exe -d example.com -v
```

### 7. Silent Mode (Only Output)
```powershell
.\subfinder-pro.exe -d example.com -silent
```

## Advanced Usage

### Filter Results
```powershell
# Match only API subdomains
.\subfinder-pro.exe -d example.com -m "^api\."

# Exclude test subdomains
.\subfinder-pro.exe -d example.com -f "test|staging"
```

### Use Specific Sources
```powershell
.\subfinder-pro.exe -d example.com -s crtsh,alienvault
```

### High Performance Mode
```powershell
.\subfinder-pro.exe -d example.com -t 20 -timeout 60
```

## Testing

### Run All Tests
```powershell
go test ./... -v
```

### Run Short Tests (Skip Integration)
```powershell
go test ./... -v -short
```

### Run Benchmarks
```powershell
go test ./... -bench=. -benchmem
```

## Troubleshooting

### Issue: "go: command not found"
**Solution**: Install Go from https://go.dev/dl/

### Issue: "No subdomains found"
**Solutions**:
1. Add AlienVault API key (required for best results)
2. Check internet connection
3. Try with `-v` flag to see errors
4. Use specific sources: `-s crtsh`

### Issue: "AlienVault requires an API key"
**Solution**: 
1. Go to https://otx.alienvault.com/
2. Create free account
3. Get API key from Settings → API Integration
4. Add to `provider-config.yaml` or set environment variable:
   ```powershell
   $env:ALIENVAULT_API_KEY="your-key"
   ```

### Issue: Rate limit errors
**Solutions**:
1. Reduce workers: `-t 5`
2. Increase timeout: `--timeout 60`
3. Get API keys for higher limits

## API Keys Setup

### AlienVault OTX (Recommended)
1. Visit: https://otx.alienvault.com/
2. Sign up (free)
3. Settings → API Integration
4. Copy API key
5. Add to config or environment

### URLScan.io (Optional)
1. Visit: https://urlscan.io/
2. Sign up (free)
3. Settings & API → Generate API Key
4. Add to config or environment

### HackerTarget (Optional - Paid)
1. Visit: https://hackertarget.com/
2. Purchase API membership
3. Add API key to config

## Example Workflows

### Workflow 1: Basic Reconnaissance
```powershell
# Enumerate subdomains
.\subfinder-pro.exe -d target.com -o subdomains.txt

# Verify with DNS
.\subfinder-pro.exe -d target.com -active -o verified.txt

# Get JSON with IPs
.\subfinder-pro.exe -d target.com -active -json -o results.json
```

### Workflow 2: Multiple Targets
```powershell
# Create domain list
"target1.com" | Out-File domains.txt
"target2.com" | Out-File -Append domains.txt
"target3.com" | Out-File -Append domains.txt

# Run enumeration
.\subfinder-pro.exe -dL domains.txt -json -o all_results.json
```

### Workflow 3: Filtered Enumeration
```powershell
# Find only production APIs
.\subfinder-pro.exe -d target.com -m "^(api|prod)\." -f "test|dev" -o production_apis.txt
```

## Performance Tips

1. **Increase workers** for faster processing:
   ```powershell
   .\subfinder-pro.exe -d example.com -t 20
   ```

2. **Skip slow sources**:
   ```powershell
   .\subfinder-pro.exe -d example.com -es threatcrowd
   ```

3. **Adjust timeout**:
   ```powershell
   .\subfinder-pro.exe -d example.com -timeout 60
   ```

## Getting Help

### Show All Flags
```powershell
.\subfinder-pro.exe --help
```

### Show Version
```powershell
.\subfinder-pro.exe --version
```

### View Examples
```powershell
cd examples
cat README.md
```

## Next Steps

1. ✅ Install Go
2. ✅ Build the project
3. ✅ Get AlienVault API key
4. ✅ Run your first scan
5. ✅ Explore advanced features
6. ✅ Read full README.md for details

---

**Need more help?** Check:
- `README.md` - Full documentation
- `PROJECT_SUMMARY.md` - Project overview
- `examples/` - Example files
- `PRD.md` - Technical specifications

**Happy Hunting! 🎯**
