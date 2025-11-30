# SubFinder Pro - Manual Installation Guide for Kali Linux

This guide provides detailed step-by-step instructions for manually installing SubFinder Pro on Kali Linux without using automated scripts.

---

## Prerequisites

- Kali Linux (2023.x or newer)
- Root or sudo privileges
- Internet connection
- The subfinder-pro folder from Windows

---

## Part 1: Transfer Files to Kali Linux

### Option 1: Using USB Drive

**On Windows:**
1. Insert USB drive
2. Open File Explorer
3. Navigate to `e:\newpro\subfinder-pro`
4. Copy the entire folder to your USB drive

**On Kali Linux:**
1. Plug in the USB drive
2. Open Terminal
3. Check where USB is mounted:
   ```bash
   lsblk
   df -h
   ```
4. Copy files to home directory:
   ```bash
   # Replace USB_NAME with your actual USB name
   cp -r /media/$USER/USB_NAME/subfinder-pro ~/
   cd ~/subfinder-pro
   ls -la
   ```

### Option 2: Using Network Transfer (SCP)

**On Kali Linux first - find your IP:**
```bash
ip addr show
# Note your IP address (e.g., 192.168.1.100)
```

**On Windows PowerShell:**
```powershell
# Transfer the folder
scp -r "e:\newpro\subfinder-pro" username@192.168.1.100:~/
# Replace username with your Kali username
# Replace 192.168.1.100 with your Kali IP
```

**Back on Kali Linux:**
```bash
cd ~/subfinder-pro
ls -la
```

### Option 3: Using VirtualBox Shared Folder

**On Windows (VirtualBox Manager):**
1. Select your Kali VM
2. Click Settings → Shared Folders
3. Click the folder icon with "+"
4. Folder Path: Browse to `e:\newpro`
5. Folder Name: `newpro`
6. Check "Auto-mount" and "Make Permanent"
7. Click OK

**On Kali Linux:**
```bash
# Install VirtualBox Guest Additions if not already installed
sudo apt update
sudo apt install virtualbox-guest-utils -y

# Create mount point
sudo mkdir -p /mnt/shared

# Mount the shared folder
sudo mount -t vboxsf newpro /mnt/shared

# Copy files to home directory
cp -r /mnt/shared/subfinder-pro ~/
cd ~/subfinder-pro
ls -la
```

---

## Part 2: Install Go Programming Language

### Step 1: Check if Go is Already Installed

```bash
go version
```

**If you see output like "go version go1.21.x"** → Skip to Part 3
**If you see "command not found"** → Continue with Step 2

### Step 2: Remove Old Go Version (if exists)

```bash
sudo apt remove golang-go -y
sudo rm -rf /usr/local/go
```

### Step 3: Download Go

```bash
# Go to tmp directory
cd /tmp

# Download Go 1.21.5 (or check https://go.dev/dl/ for latest version)
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz

# Verify download
ls -lh go1.21.5.linux-amd64.tar.gz
```

**If wget fails, try with curl:**
```bash
curl -LO https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
```

### Step 4: Extract and Install Go

```bash
# Extract to /usr/local
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# Verify extraction
ls -l /usr/local/go/bin/go
```

### Step 5: Configure Environment Variables

```bash
# Add Go to PATH in .bashrc
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc

# Reload shell configuration
source ~/.bashrc

# Verify installation
go version
# Should show: go version go1.21.5 linux/amd64
```

**If go version still doesn't work:**
```bash
# Manually set for current session
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# Try again
go version
```

### Step 6: Verify Go Installation

```bash
# Check Go environment
go env

# Should see GOROOT=/usr/local/go and GOPATH=/home/username/go
```

---

## Part 3: Install Required Dependencies

### Step 1: Update System Packages

```bash
sudo apt update
sudo apt upgrade -y
```

### Step 2: Install Build Tools (if needed)

```bash
sudo apt install -y build-essential git
```

### Step 3: Navigate to Project Directory

```bash
cd ~/subfinder-pro
pwd
# Should show: /home/username/subfinder-pro
```

### Step 4: Verify Go Module Files

```bash
# Check if go.mod exists
cat go.mod
```

You should see something like:
```
module github.com/yourusername/subfinder-pro
go 1.21
...
```

### Step 5: Download Go Dependencies

```bash
# Download all required packages
go mod download

# This will download:
# - github.com/spf13/cobra
# - golang.org/x/time
# - gopkg.in/yaml.v3
# - And their dependencies
```

**If you see errors:**
```bash
# Clean and retry
go clean -modcache
go mod tidy
go mod download
```

### Step 6: Verify Dependencies

```bash
# Check downloaded modules
go mod verify
# Should show: all modules verified
```

---

## Part 4: Build the Application

### Step 1: Clean Previous Builds (if any)

```bash
cd ~/subfinder-pro
rm -f subfinder-pro
rm -rf build/
```

### Step 2: Build the Binary

```bash
# Build with optimization flags
go build -ldflags="-s -w" -o subfinder-pro main.go
```

**Explanation of flags:**
- `-ldflags="-s -w"`: Strip debug info and reduce binary size
- `-o subfinder-pro`: Output filename
- `main.go`: Source file to build

**This will take 30-60 seconds**

### Step 3: Verify the Build

```bash
# Check if binary was created
ls -lh subfinder-pro
# Should show a file ~8-15 MB in size

# Check file type
file subfinder-pro
# Should show: ELF 64-bit LSB executable, x86-64

# Make it executable
chmod +x subfinder-pro
```

### Step 4: Test the Binary

```bash
# Test help command
./subfinder-pro --help
```

**You should see the help menu with available commands and flags**

```bash
# Test version (if implemented)
./subfinder-pro --version
```

**If you see errors:**
```bash
# Check for missing dependencies
ldd subfinder-pro

# Rebuild with verbose output
go build -v -ldflags="-s -w" -o subfinder-pro main.go
```

---

## Part 5: Install System-Wide (Optional but Recommended)

### Step 1: Copy Binary to System Path

```bash
# Copy to /usr/local/bin (requires sudo)
sudo cp subfinder-pro /usr/local/bin/

# Verify copy
ls -lh /usr/local/bin/subfinder-pro
```

### Step 2: Set Proper Permissions

```bash
# Make it executable by all users
sudo chmod +x /usr/local/bin/subfinder-pro

# Verify permissions
ls -l /usr/local/bin/subfinder-pro
# Should show: -rwxr-xr-x (executable)
```

### Step 3: Verify System-Wide Access

```bash
# Try running from anywhere
cd ~
subfinder-pro --help

# Should work without ./ prefix
```

**If "command not found":**
```bash
# Check if /usr/local/bin is in PATH
echo $PATH | grep "/usr/local/bin"

# If not found, add it
echo 'export PATH=$PATH:/usr/local/bin' >> ~/.bashrc
source ~/.bashrc

# Try again
subfinder-pro --help
```

---

## Part 6: Configure the Application

### Step 1: Create Configuration Directory

```bash
# Create config directory
mkdir -p ~/.config/subfinder-pro

# Verify creation
ls -la ~/.config/ | grep subfinder-pro
```

### Step 2: Copy Configuration Files

```bash
# Go back to source directory
cd ~/subfinder-pro

# Copy main config
cp config.yaml ~/.config/subfinder-pro/

# Copy provider config
cp provider-config.yaml ~/.config/subfinder-pro/

# Copy examples (optional)
cp -r examples ~/.config/subfinder-pro/

# Verify files were copied
ls -la ~/.config/subfinder-pro/
```

### Step 3: Edit Configuration Files

#### Configure Main Settings:

```bash
# Open main config
nano ~/.config/subfinder-pro/config.yaml
```

**Key settings to review:**
```yaml
timeout: 30s
workers: 10
dns:
  resolvers:
    - 8.8.8.8
    - 1.1.1.1
    - 8.8.4.4
  retries: 2
  timeout: 5s
```

**Press CTRL+X, then Y, then ENTER to save**

#### Configure Provider API Keys:

```bash
# Open provider config
nano ~/.config/subfinder-pro/provider-config.yaml
```

**Add your API keys (optional but recommended):**
```yaml
sources:
  alienvault:
    enabled: true
    api_key: "YOUR_ALIENVAULT_API_KEY"
    rate_limit: 10
  
  urlscan:
    enabled: true
    api_key: "YOUR_URLSCAN_API_KEY"
    rate_limit: 10
  
  crtsh:
    enabled: true
    rate_limit: 50
  
  hackertarget:
    enabled: true
    rate_limit: 50
  
  threatcrowd:
    enabled: true
    rate_limit: 50
```

**Press CTRL+X, then Y, then ENTER to save**

### Step 4: Set Proper Permissions

```bash
# Secure config files (contains API keys)
chmod 600 ~/.config/subfinder-pro/provider-config.yaml
chmod 644 ~/.config/subfinder-pro/config.yaml

# Verify permissions
ls -l ~/.config/subfinder-pro/
```

---

## Part 7: Test the Installation

### Test 1: Basic Help Command

```bash
subfinder-pro --help
```

**Expected output:** Help menu with all available flags

### Test 2: Version Check

```bash
subfinder-pro --version
```

### Test 3: Simple Domain Enumeration

```bash
# Test with a well-known domain
subfinder-pro -d google.com

# Should output subdomains like:
# www.google.com
# mail.google.com
# etc.
```

### Test 4: Verbose Mode

```bash
subfinder-pro -d google.com -v
```

**Expected output:** Detailed logs showing:
- Sources being queried
- Subdomains found per source
- Total unique subdomains

### Test 5: Save to File

```bash
# Test output to file
subfinder-pro -d google.com -o ~/test-results.txt

# Check if file was created
cat ~/test-results.txt
wc -l ~/test-results.txt
```

### Test 6: JSON Output

```bash
subfinder-pro -d google.com -json -o ~/test-results.json
cat ~/test-results.json
```

### Test 7: Active DNS Resolution

```bash
subfinder-pro -d google.com -active -v
```

**This will verify which subdomains are actually live**

### Test 8: Silent Mode

```bash
subfinder-pro -d google.com -silent
```

**Should only output subdomains, no logs**

---

## Part 8: Run Unit Tests (Optional)

### Step 1: Navigate to Project Directory

```bash
cd ~/subfinder-pro
```

### Step 2: Run All Tests

```bash
# Run tests with verbose output
go test ./... -v
```

### Step 3: Check Test Results

**You should see output like:**
```
=== RUN   TestResolver
--- PASS: TestResolver (0.50s)
=== RUN   TestRunner
--- PASS: TestRunner (1.23s)
=== RUN   TestSources
--- PASS: TestSources (2.45s)
PASS
ok      github.com/yourusername/subfinder-pro/tests    4.180s
```

### Step 4: Run Specific Test File

```bash
# Test resolver
go test ./tests/resolver_test.go -v

# Test runner
go test ./tests/runner_test.go -v

# Test sources
go test ./tests/sources_test.go -v
```

**If tests fail, it's usually okay - the tool may still work fine**

---

## Part 9: Create Useful Aliases (Optional)

### Step 1: Add Aliases to .bashrc

```bash
nano ~/.bashrc
```

**Add these at the end of the file:**
```bash
# SubFinder Pro aliases
alias sf='subfinder-pro'
alias sfd='subfinder-pro -d'
alias sfv='subfinder-pro -v'
alias sfa='subfinder-pro -active'
alias sfs='subfinder-pro -silent'
```

**Save and exit (CTRL+X, Y, ENTER)**

### Step 2: Reload Configuration

```bash
source ~/.bashrc
```

### Step 3: Test Aliases

```bash
# Now you can use short commands
sf -d google.com
sfd example.com -o results.txt
sfv example.com
```

---

## Part 10: Integration with Other Kali Tools

### Install Complementary Tools

```bash
# Install httpx (for HTTP probing)
go install -v github.com/projectdiscovery/httpx/cmd/httpx@latest

# Install nuclei (for vulnerability scanning)
go install -v github.com/projectdiscovery/nuclei/v2/cmd/nuclei@latest

# Install nmap (if not already installed)
sudo apt install nmap -y
```

### Usage Examples

#### 1. Find Live Web Servers

```bash
subfinder-pro -d target.com -silent | httpx -silent
```

#### 2. Port Scanning Found Subdomains

```bash
subfinder-pro -d target.com -active -silent > subs.txt
sudo nmap -iL subs.txt -p 80,443,8080,8443
```

#### 3. Vulnerability Scanning

```bash
subfinder-pro -d target.com -silent | httpx -silent | nuclei -t ~/nuclei-templates/
```

#### 4. Save Results with Timestamps

```bash
DATE=$(date +%Y%m%d_%H%M%S)
subfinder-pro -d target.com -o results_${DATE}.txt
```

#### 5. Multiple Domains Scanning

```bash
# Create domains file
cat > domains.txt << EOF
example.com
test.com
target.com
EOF

# Scan all domains
subfinder-pro -dL domains.txt -o all-results.txt
```

---

## Part 11: Troubleshooting

### Problem 1: "go: command not found"

**Solution:**
```bash
# Check if Go is installed
ls -l /usr/local/go/bin/go

# If exists, add to PATH
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### Problem 2: "Permission denied" when running subfinder-pro

**Solution:**
```bash
chmod +x subfinder-pro
# Or if installed system-wide:
sudo chmod +x /usr/local/bin/subfinder-pro
```

### Problem 3: "cannot find package" during build

**Solution:**
```bash
cd ~/subfinder-pro
go mod tidy
go mod download
go clean -modcache
go build -o subfinder-pro main.go
```

### Problem 4: No subdomains found

**Solution:**
```bash
# Test with verbose mode
subfinder-pro -d google.com -v

# Check internet connection
ping -c 3 8.8.8.8

# Check DNS resolution
nslookup google.com

# Try with specific source
subfinder-pro -d google.com -s crtsh -v
```

### Problem 5: "too many open files" error

**Solution:**
```bash
# Check current limit
ulimit -n

# Increase limit temporarily
ulimit -n 4096

# Or permanently
echo "* soft nofile 4096" | sudo tee -a /etc/security/limits.conf
echo "* hard nofile 8192" | sudo tee -a /etc/security/limits.conf
```

### Problem 6: Build takes too long or fails

**Solution:**
```bash
# Build with verbose output to see what's happening
go build -v -o subfinder-pro main.go

# Clean and rebuild
go clean
rm -rf ~/.cache/go-build
go build -o subfinder-pro main.go
```

### Problem 7: "rate limit exceeded" errors

**Solution:**
```bash
# Edit config to reduce workers
nano ~/.config/subfinder-pro/config.yaml
# Change: workers: 5 (instead of 10)

# Or use command line flag
subfinder-pro -d target.com -t 5
```

---

## Part 12: Maintenance and Updates

### Update the Tool

```bash
# If you get new version from Windows
cd ~/subfinder-pro
go build -ldflags="-s -w" -o subfinder-pro main.go
sudo cp subfinder-pro /usr/local/bin/
```

### Backup Configuration

```bash
# Backup your configs
cp -r ~/.config/subfinder-pro ~/subfinder-pro-backup-$(date +%Y%m%d)
```

### Clean Old Builds

```bash
cd ~/subfinder-pro
go clean -cache
go clean -modcache
```

---

## Part 13: Uninstallation (if needed)

### Remove Binary

```bash
sudo rm /usr/local/bin/subfinder-pro
```

### Remove Configuration

```bash
rm -rf ~/.config/subfinder-pro
```

### Remove Source Files

```bash
rm -rf ~/subfinder-pro
```

### Remove Go (if you want)

```bash
sudo rm -rf /usr/local/go
# Remove PATH entries from .bashrc manually
```

---

## Quick Reference Card

```bash
# Basic commands
subfinder-pro -d domain.com                    # Basic scan
subfinder-pro -d domain.com -o results.txt     # Save results
subfinder-pro -d domain.com -v                 # Verbose mode
subfinder-pro -d domain.com -active            # DNS verification
subfinder-pro -d domain.com -silent            # Silent mode
subfinder-pro -d domain.com -json              # JSON output
subfinder-pro -dL domains.txt -o out.txt       # Multiple domains
subfinder-pro -d domain.com -t 20              # 20 workers
subfinder-pro -d domain.com -s crtsh,alienvault # Specific sources

# Configuration files
~/.config/subfinder-pro/config.yaml            # Main config
~/.config/subfinder-pro/provider-config.yaml   # API keys
~/.config/subfinder-pro/examples/              # Example files

# Log location (if enabled in config)
~/.config/subfinder-pro/logs/
```

---

## Summary Checklist

- [ ] Files transferred to Kali Linux
- [ ] Go 1.21+ installed and in PATH
- [ ] Dependencies downloaded with `go mod download`
- [ ] Binary built with `go build`
- [ ] Binary copied to `/usr/local/bin/`
- [ ] Configuration directory created
- [ ] Config files copied and edited
- [ ] API keys added (optional)
- [ ] Installation tested with `subfinder-pro -d google.com`
- [ ] Integration with other tools tested

---

## Getting Help

- Run `subfinder-pro --help` for command reference
- Check verbose output with `-v` flag
- Review configuration files in `~/.config/subfinder-pro/`
- Check examples in `~/.config/subfinder-pro/examples/`

---

**Installation Complete! Happy Subdomain Hunting! 🎯**
