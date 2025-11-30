# SubFinder Pro - Kali Linux Installation Guide

This guide shows you how to transfer and install the SubFinder Pro tool (already built on Windows) to Kali Linux.

## Method 1: Transfer via USB/External Drive (Offline)

### Step 1: Prepare Files on Windows

1. **Copy the entire subfinder-pro folder** to a USB drive or external storage:
   ```powershell
   # On Windows, copy the folder
   Copy-Item -Path "e:\newpro\subfinder-pro" -Destination "D:\" -Recurse
   ```
   Or simply copy the `subfinder-pro` folder manually to your USB drive.

### Step 2: Transfer to Kali Linux

2. **Plug the USB drive into Kali Linux** and copy the files:
   ```bash
   # Mount the USB (if not auto-mounted)
   # Usually auto-mounted to /media/username/USB_NAME
   
   # Copy to home directory
   cp -r /media/$USER/*/subfinder-pro ~/
   cd ~/subfinder-pro
   ```

### Step 3: Run Installation Script

3. **Make the script executable and run it**:
   ```bash
   chmod +x deploy-to-kali.sh
   sudo ./deploy-to-kali.sh
   ```

The script will:
- Check/install Go if needed
- Download dependencies
- Build the binary
- Install to `/usr/local/bin/`
- Set up configuration files
- Run tests

### Step 4: Verify Installation

```bash
subfinder-pro --help
subfinder-pro -d example.com
```

---

## Method 2: Transfer via Network (SCP/SFTP)

### Option A: Using SCP

**On Windows (PowerShell):**
```powershell
# Transfer to Kali Linux via SCP
scp -r "e:\newpro\subfinder-pro" username@kali-ip:~/
```

**On Kali Linux:**
```bash
cd ~/subfinder-pro
chmod +x deploy-to-kali.sh
sudo ./deploy-to-kali.sh
```

### Option B: Using WinSCP (GUI)

1. Download and install [WinSCP](https://winscp.net/)
2. Connect to your Kali Linux machine
3. Transfer the `subfinder-pro` folder to `/home/username/`
4. Run the installation script in Kali terminal

---

## Method 3: Transfer via Shared Folder (VM)

If Kali is running in VirtualBox/VMware:

### VirtualBox:

**On Windows:**
1. Open VirtualBox
2. Select your Kali VM → Settings → Shared Folders
3. Add `e:\newpro` as a shared folder
4. Name it `newpro`

**On Kali Linux:**
```bash
# Mount the shared folder
sudo mkdir -p /mnt/shared
sudo mount -t vboxsf newpro /mnt/shared

# Copy files
cp -r /mnt/shared/subfinder-pro ~/
cd ~/subfinder-pro

# Run installation
chmod +x deploy-to-kali.sh
sudo ./deploy-to-kali.sh
```

### VMware:

**On Windows:**
1. VM → Settings → Options → Shared Folders
2. Enable shared folders and add `e:\newpro`

**On Kali Linux:**
```bash
# Shared folders are usually auto-mounted at /mnt/hgfs/
cp -r /mnt/hgfs/newpro/subfinder-pro ~/
cd ~/subfinder-pro

# Run installation
chmod +x deploy-to-kali.sh
sudo ./deploy-to-kali.sh
```

---

## Method 4: Manual Installation (Step by Step)

If you prefer manual installation without the script:

### Step 1: Transfer Files

Use any method above to get the files to Kali.

### Step 2: Install Go (if not installed)

```bash
# Check Go version
go version

# If not installed or version < 1.21, install:
cd /tmp
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

# Verify
go version
```

### Step 3: Build the Application

```bash
cd ~/subfinder-pro

# Download dependencies
go mod download

# Build
go build -o subfinder-pro main.go

# Make executable
chmod +x subfinder-pro

# Test it
./subfinder-pro --help
```

### Step 4: Install System-Wide (Optional)

```bash
# Copy to system bin
sudo cp subfinder-pro /usr/local/bin/
sudo chmod +x /usr/local/bin/subfinder-pro

# Verify
subfinder-pro --help
```

### Step 5: Set Up Configuration

```bash
# Create config directory
mkdir -p ~/.config/subfinder-pro

# Copy configuration files
cp config.yaml ~/.config/subfinder-pro/
cp provider-config.yaml ~/.config/subfinder-pro/
cp -r examples ~/.config/subfinder-pro/
```

### Step 6: Test Installation

```bash
# Basic test
subfinder-pro -d example.com

# With verbose
subfinder-pro -d example.com -v

# Test with active DNS
subfinder-pro -d example.com -active
```

---

## Method 5: Create Portable Package

Create a ready-to-deploy package on Windows:

**On Windows (PowerShell):**
```powershell
# Navigate to the project
cd e:\newpro\subfinder-pro

# Build for Linux
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o subfinder-pro-linux main.go

# Create deployment package
$deployDir = "e:\newpro\subfinder-pro-deploy"
New-Item -ItemType Directory -Force -Path $deployDir

# Copy necessary files
Copy-Item subfinder-pro-linux $deployDir\subfinder-pro
Copy-Item config.yaml $deployDir\
Copy-Item provider-config.yaml $deployDir\
Copy-Item -Recurse examples $deployDir\
Copy-Item deploy-to-kali.sh $deployDir\
Copy-Item KALI_INSTALL.md $deployDir\

# Create installation script
@"
#!/bin/bash
sudo cp subfinder-pro /usr/local/bin/
sudo chmod +x /usr/local/bin/subfinder-pro
mkdir -p ~/.config/subfinder-pro
cp config.yaml ~/.config/subfinder-pro/
cp provider-config.yaml ~/.config/subfinder-pro/
cp -r examples ~/.config/subfinder-pro/
echo 'SubFinder Pro installed successfully!'
subfinder-pro --help
"@ | Out-File -FilePath "$deployDir\install.sh" -Encoding ASCII

Write-Host "Deployment package created at: $deployDir"
Write-Host "Transfer this folder to Kali Linux"
```

**On Kali Linux:**
```bash
# After transferring the deployment folder
cd ~/subfinder-pro-deploy
chmod +x install.sh
sudo ./install.sh
```

---

## Quick Start After Installation

```bash
# Basic usage
subfinder-pro -d example.com

# Save to file
subfinder-pro -d example.com -o results.txt

# Multiple domains
echo "example.com" > domains.txt
echo "test.com" >> domains.txt
subfinder-pro -dL domains.txt -o results.txt

# With DNS verification
subfinder-pro -d example.com -active

# Silent mode (only subdomains)
subfinder-pro -d example.com -silent

# JSON output
subfinder-pro -d example.com -json -o results.json

# Verbose mode
subfinder-pro -d example.com -v
```

---

## Configuration

After installation, edit configuration files to add API keys:

```bash
# Edit provider config to add API keys
nano ~/.config/subfinder-pro/provider-config.yaml

# Edit main config for DNS servers, timeouts, etc.
nano ~/.config/subfinder-pro/config.yaml
```

---

## Integration with Kali Tools

```bash
# With httpx
subfinder-pro -d target.com -silent | httpx -silent

# With nmap
subfinder-pro -d target.com -active -silent | sudo nmap -iL - -p 80,443

# With nuclei
subfinder-pro -d target.com -silent | httpx -silent | nuclei

# With aquatone
subfinder-pro -d target.com -silent | aquatone

# With masscan
subfinder-pro -d target.com -active -silent > subs.txt
sudo masscan -iL subs.txt -p 80,443,8080,8443 --rate 10000
```

---

## Troubleshooting

### Issue: "Permission denied" when running script
```bash
chmod +x deploy-to-kali.sh
sudo ./deploy-to-kali.sh
```

### Issue: "go: command not found"
```bash
# Install Go manually (see Step 2 in Manual Installation)
```

### Issue: "cannot find package"
```bash
cd ~/subfinder-pro
go mod tidy
go mod download
go build -o subfinder-pro main.go
```

### Issue: Binary doesn't run
```bash
# Make sure it's executable
chmod +x subfinder-pro

# Check if built for correct architecture
file subfinder-pro
# Should show: ELF 64-bit LSB executable, x86-64
```

### Issue: No subdomains found
```bash
# Try with verbose to see what's happening
subfinder-pro -d example.com -v

# Test with a known domain
subfinder-pro -d google.com -v
```

---

## Uninstallation

```bash
# Remove binary
sudo rm /usr/local/bin/subfinder-pro

# Remove config (optional)
rm -rf ~/.config/subfinder-pro

# Remove source files (optional)
rm -rf ~/subfinder-pro
```

---

## Support

For issues or questions:
- Check verbose output: `subfinder-pro -d domain.com -v`
- Review configuration files
- Check the examples in `~/.config/subfinder-pro/examples/`

---

**Happy Bug Hunting! 🎯**
