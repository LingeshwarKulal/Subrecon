# SubFinder Pro - Kali Linux Deployment Guide

This guide provides step-by-step instructions to deploy and use SubFinder Pro on Kali Linux.

## Prerequisites

Before starting, ensure you have:
- Kali Linux (2023.x or later)
- Root or sudo access
- Internet connection

## Step 1: Install Go

SubFinder Pro requires Go 1.21 or higher.

### Check if Go is installed:
```bash
go version
```

### If Go is not installed or version is below 1.21:

```bash
# Update package list
sudo apt update

# Install Go from Kali repositories (may be older version)
sudo apt install golang-go -y

# Verify installation
go version
```

### Install Latest Go Version (Recommended):

If the repository version is too old, install the latest Go:

```bash
# Remove old Go version (if exists)
sudo apt remove golang-go -y

# Download latest Go (check https://go.dev/dl/ for the latest version)
cd /tmp
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz

# Extract to /usr/local
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# Add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc

# Reload shell configuration
source ~/.bashrc

# Verify installation
go version
```

## Step 2: Install Git (if not already installed)

```bash
sudo apt install git -y
```

## Step 3: Clone the Repository

```bash
# Navigate to your projects directory
cd ~
mkdir -p projects
cd projects

# Clone the repository
git clone https://github.com/yourusername/subfinder-pro.git
cd subfinder-pro
```

**Note**: Replace `https://github.com/yourusername/subfinder-pro.git` with the actual repository URL, or copy the tool directly if you already have it.

## Step 4: Download Dependencies

```bash
# Download all Go dependencies
go mod download

# Verify dependencies
go mod verify
```

## Step 5: Build the Application

### Option A: Build using Make (Recommended)

```bash
# Build the binary
make build

# The binary will be in build/subfinder-pro
ls -lh build/subfinder-pro
```

### Option B: Build using Go directly

```bash
# Build the binary
go build -o subfinder-pro main.go

# Make it executable
chmod +x subfinder-pro

# Verify the build
./subfinder-pro --version
```

## Step 6: Install Globally (Optional)

To use `subfinder-pro` from anywhere in the system:

### Option A: Install to /usr/local/bin

```bash
# Copy binary to system path
sudo cp build/subfinder-pro /usr/local/bin/

# Or if built with go build directly:
sudo cp subfinder-pro /usr/local/bin/

# Make it executable
sudo chmod +x /usr/local/bin/subfinder-pro

# Verify installation
subfinder-pro --version
```

### Option B: Install using Make

```bash
# Install to $GOPATH/bin
make install

# Verify installation
subfinder-pro --version
```

## Step 7: Configure the Tool

### Copy Configuration Files

```bash
# Create config directory in your home
mkdir -p ~/.config/subfinder-pro

# Copy configuration files
cp config.yaml ~/.config/subfinder-pro/
cp provider-config.yaml ~/.config/subfinder-pro/

# Or keep configs in the project directory and use them with -c flag
```

### Edit Configuration (Optional)

```bash
# Edit main config
nano ~/.config/subfinder-pro/config.yaml

# Edit provider config (add API keys for better results)
nano ~/.config/subfinder-pro/provider-config.yaml
```

**Important**: Add API keys to `provider-config.yaml` for sources that require authentication (AlienVault OTX, URLScan.io, etc.) for better results.

## Step 8: Test the Installation

### Basic Test

```bash
# Test with a domain
subfinder-pro -d example.com

# Or if not installed globally:
cd ~/projects/subfinder-pro
./build/subfinder-pro -d example.com
```

### Test with Verbose Output

```bash
subfinder-pro -d example.com -v
```

### Test DNS Resolution

```bash
subfinder-pro -d example.com -active -v
```

## Step 9: Run Tests (Optional)

```bash
# Navigate to project directory
cd ~/projects/subfinder-pro

# Run all tests
make test

# Or using go directly
go test ./... -v
```

## Common Use Cases

### 1. Basic Subdomain Enumeration

```bash
subfinder-pro -d target.com -o results.txt
```

### 2. Multiple Domains

```bash
# Create a file with domains
echo "example.com" > domains.txt
echo "test.com" >> domains.txt

# Run enumeration
subfinder-pro -dL domains.txt -o results.txt
```

### 3. Active DNS Resolution

```bash
subfinder-pro -d target.com -active -o live-subdomains.txt
```

### 4. JSON Output for Processing

```bash
subfinder-pro -d target.com -json -o results.json
```

### 5. Silent Mode (for piping)

```bash
subfinder-pro -d target.com -silent | grep api
```

### 6. Pattern Filtering

```bash
# Find only dev, staging, and test subdomains
subfinder-pro -d target.com -m "^(dev|staging|test)\."
```

### 7. With Proxy

```bash
subfinder-pro -d target.com -proxy http://127.0.0.1:8080
```

## Troubleshooting

### Issue: "go: command not found"

**Solution**: Go is not properly installed or not in PATH. Follow Step 1 again.

### Issue: "permission denied" when building

**Solution**: 
```bash
chmod +x main.go
go build -o subfinder-pro main.go
```

### Issue: "cannot find package"

**Solution**: Download dependencies again:
```bash
go mod tidy
go mod download
```

### Issue: No results found

**Solutions**:
1. Check internet connection
2. Add API keys to `provider-config.yaml`
3. Try with verbose mode to see which sources are working:
   ```bash
   subfinder-pro -d target.com -v
   ```
4. Test specific sources:
   ```bash
   subfinder-pro -d target.com -s crtsh -v
   ```

### Issue: DNS resolution fails

**Solution**: 
```bash
# Check DNS servers in config.yaml
nano ~/.config/subfinder-pro/config.yaml

# Make sure you have valid DNS servers like:
# - 8.8.8.8
# - 1.1.1.1
```

## Integration with Other Kali Tools

### With HTTPx (probe for live web servers)

```bash
subfinder-pro -d target.com -silent | httpx -silent
```

### With Nmap

```bash
subfinder-pro -d target.com -active -silent | xargs -I {} nmap -sV {}
```

### With Nuclei

```bash
subfinder-pro -d target.com -silent | httpx -silent | nuclei -t ~/nuclei-templates/
```

### With Amass

```bash
# Combine results from both tools
subfinder-pro -d target.com -silent > subs1.txt
amass enum -d target.com -silent > subs2.txt
cat subs1.txt subs2.txt | sort -u > all-subs.txt
```

## Updating SubFinder Pro

```bash
# Navigate to project directory
cd ~/projects/subfinder-pro

# Pull latest changes
git pull origin main

# Rebuild
make build

# Reinstall (if installed globally)
sudo cp build/subfinder-pro /usr/local/bin/
```

## Uninstallation

```bash
# Remove binary from system path
sudo rm /usr/local/bin/subfinder-pro

# Remove config directory (optional)
rm -rf ~/.config/subfinder-pro

# Remove project directory (optional)
rm -rf ~/projects/subfinder-pro
```

## Performance Tips

1. **Increase Workers**: Use `-t` flag to increase concurrent workers
   ```bash
   subfinder-pro -d target.com -t 30
   ```

2. **Use Specific Sources**: If you only need certain sources
   ```bash
   subfinder-pro -d target.com -s crtsh,alienvault
   ```

3. **Skip DNS Resolution**: Disable active mode for faster results
   ```bash
   subfinder-pro -d target.com
   ```

4. **Use Configuration File**: Store common settings in config.yaml
   ```bash
   subfinder-pro -d target.com -c ~/.config/subfinder-pro/config.yaml
   ```

## Security Considerations

- Always get proper authorization before scanning domains
- Use rate limiting to avoid overwhelming target infrastructure
- Consider using proxy/VPN for anonymity
- Store API keys securely (use environment variables if possible)
- Be aware of the legal implications in your jurisdiction

## Additional Resources

- **Project Documentation**: See `README.md` and `QUICKSTART.md`
- **Examples**: Check the `examples/` directory
- **GitHub Issues**: Report bugs or request features

## Support

If you encounter issues:
1. Check the troubleshooting section above
2. Review verbose output with `-v` flag
3. Check GitHub issues
4. Create a new issue with detailed information

---

**Happy Hunting! 🎯**
