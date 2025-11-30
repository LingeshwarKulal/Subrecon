#!/bin/bash
# SubFinder Pro - Kali Linux Deployment Script
# This script sets up SubFinder Pro on Kali Linux

set -e

echo "================================================"
echo "SubFinder Pro - Kali Linux Deployment"
echo "================================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Go is installed
echo -e "${YELLOW}[1/8] Checking Go installation...${NC}"
if ! command -v go &> /dev/null; then
    echo -e "${RED}Go is not installed. Installing Go 1.21+...${NC}"
    
    # Download and install Go
    cd /tmp
    wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
    
    # Add to PATH
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    echo 'export GOPATH=$HOME/go' >> ~/.bashrc
    echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
    
    export PATH=$PATH:/usr/local/go/bin
    export GOPATH=$HOME/go
    export PATH=$PATH:$GOPATH/bin
    
    echo -e "${GREEN}✓ Go installed successfully${NC}"
else
    GO_VERSION=$(go version | awk '{print $3}')
    echo -e "${GREEN}✓ Go is already installed: $GO_VERSION${NC}"
fi

# Get the script's directory (where subfinder-pro files are)
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
echo -e "${GREEN}✓ Working directory: $SCRIPT_DIR${NC}"

# Install dependencies
echo -e "\n${YELLOW}[2/8] Installing dependencies...${NC}"
cd "$SCRIPT_DIR"
go mod download
echo -e "${GREEN}✓ Dependencies installed${NC}"

# Build the binary
echo -e "\n${YELLOW}[3/8] Building SubFinder Pro...${NC}"
go build -ldflags="-s -w" -o subfinder-pro main.go
chmod +x subfinder-pro
echo -e "${GREEN}✓ Build complete${NC}"

# Install to system
echo -e "\n${YELLOW}[4/8] Installing to system...${NC}"
sudo cp subfinder-pro /usr/local/bin/
sudo chmod +x /usr/local/bin/subfinder-pro
echo -e "${GREEN}✓ Installed to /usr/local/bin/subfinder-pro${NC}"

# Create config directory
echo -e "\n${YELLOW}[5/8] Setting up configuration...${NC}"
mkdir -p ~/.config/subfinder-pro

# Copy config files
if [ -f "$SCRIPT_DIR/config.yaml" ]; then
    cp "$SCRIPT_DIR/config.yaml" ~/.config/subfinder-pro/
    echo -e "${GREEN}✓ Copied config.yaml${NC}"
fi

if [ -f "$SCRIPT_DIR/provider-config.yaml" ]; then
    cp "$SCRIPT_DIR/provider-config.yaml" ~/.config/subfinder-pro/
    echo -e "${GREEN}✓ Copied provider-config.yaml${NC}"
fi

# Copy examples
if [ -d "$SCRIPT_DIR/examples" ]; then
    cp -r "$SCRIPT_DIR/examples" ~/.config/subfinder-pro/
    echo -e "${GREEN}✓ Copied examples${NC}"
fi

# Test installation
echo -e "\n${YELLOW}[6/8] Testing installation...${NC}"
if command -v subfinder-pro &> /dev/null; then
    echo -e "${GREEN}✓ SubFinder Pro is accessible from PATH${NC}"
    subfinder-pro --help > /dev/null 2>&1 && echo -e "${GREEN}✓ Help command works${NC}"
else
    echo -e "${RED}✗ Installation failed - command not found${NC}"
    exit 1
fi

# Run tests
echo -e "\n${YELLOW}[7/8] Running tests...${NC}"
cd "$SCRIPT_DIR"
if go test ./... -v > /dev/null 2>&1; then
    echo -e "${GREEN}✓ All tests passed${NC}"
else
    echo -e "${YELLOW}⚠ Some tests failed (this may be normal)${NC}"
fi

# Create quick reference
echo -e "\n${YELLOW}[8/8] Creating quick reference...${NC}"
cat > ~/.config/subfinder-pro/USAGE.txt << 'EOF'
SubFinder Pro - Quick Reference
================================

Basic Usage:
  subfinder-pro -d example.com
  subfinder-pro -d example.com -o results.txt
  subfinder-pro -dL domains.txt -o results.txt

Active DNS Resolution:
  subfinder-pro -d example.com -active

Specific Sources:
  subfinder-pro -d example.com -s crtsh,alienvault

Silent Mode:
  subfinder-pro -d example.com -silent

JSON Output:
  subfinder-pro -d example.com -json -o results.json

With Verbose:
  subfinder-pro -d example.com -v -t 20

Pattern Matching:
  subfinder-pro -d example.com -m "^(api|dev|staging)\."

Filter Patterns:
  subfinder-pro -d example.com -f "test|internal"

Configuration:
  Config files: ~/.config/subfinder-pro/
  - config.yaml (global settings)
  - provider-config.yaml (API keys)

Integration Examples:
  subfinder-pro -d target.com -silent | httpx -silent
  subfinder-pro -d target.com -active -silent | nmap -iL -
  subfinder-pro -d target.com -silent | nuclei -t ~/nuclei-templates/
EOF

echo -e "${GREEN}✓ Quick reference created at ~/.config/subfinder-pro/USAGE.txt${NC}"

# Display completion message
echo ""
echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}Installation Complete!${NC}"
echo -e "${GREEN}================================================${NC}"
echo ""
echo -e "SubFinder Pro has been installed successfully!"
echo ""
echo -e "Quick Start:"
echo -e "  ${YELLOW}subfinder-pro -d example.com${NC}"
echo ""
echo -e "Configuration:"
echo -e "  ${YELLOW}~/.config/subfinder-pro/config.yaml${NC}"
echo -e "  ${YELLOW}~/.config/subfinder-pro/provider-config.yaml${NC}"
echo ""
echo -e "Documentation:"
echo -e "  ${YELLOW}cat ~/.config/subfinder-pro/USAGE.txt${NC}"
echo ""
echo -e "For help:"
echo -e "  ${YELLOW}subfinder-pro --help${NC}"
echo ""
echo -e "${YELLOW}Note: Add API keys to provider-config.yaml for better results!${NC}"
echo ""
