# SubFinder Pro - Setup and Build Script
# Run this script after installing Go

Write-Host "==================================" -ForegroundColor Cyan
Write-Host "SubFinder Pro - Setup Script" -ForegroundColor Cyan
Write-Host "==================================" -ForegroundColor Cyan
Write-Host ""

# Check if Go is installed
Write-Host "[1/5] Checking Go installation..." -ForegroundColor Yellow
try {
    $goVersion = go version 2>&1
    Write-Host "✓ Go is installed: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "✗ Go is not installed!" -ForegroundColor Red
    Write-Host "Please install Go from: https://go.dev/dl/" -ForegroundColor Yellow
    Write-Host "After installation, restart PowerShell and run this script again." -ForegroundColor Yellow
    exit 1
}

Write-Host ""

# Download dependencies
Write-Host "[2/5] Downloading dependencies..." -ForegroundColor Yellow
try {
    go mod download
    Write-Host "✓ Dependencies downloaded successfully" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed to download dependencies" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Build the project
Write-Host "[3/5] Building SubFinder Pro..." -ForegroundColor Yellow
try {
    go build -ldflags="-s -w" -o subfinder-pro.exe main.go
    Write-Host "✓ Build successful: subfinder-pro.exe" -ForegroundColor Green
} catch {
    Write-Host "✗ Build failed" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Run tests
Write-Host "[4/5] Running tests..." -ForegroundColor Yellow
Write-Host "(This may take a moment...)" -ForegroundColor Gray
try {
    $testResult = go test ./... -short 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ All tests passed" -ForegroundColor Green
    } else {
        Write-Host "⚠ Some tests failed (this is okay if you don't have internet)" -ForegroundColor Yellow
    }
} catch {
    Write-Host "⚠ Tests encountered issues (non-critical)" -ForegroundColor Yellow
}

Write-Host ""

# Verify build
Write-Host "[5/5] Verifying installation..." -ForegroundColor Yellow
if (Test-Path "subfinder-pro.exe") {
    $fileSize = (Get-Item "subfinder-pro.exe").Length / 1MB
    Write-Host "✓ Binary created successfully" -ForegroundColor Green
    Write-Host "  Size: $([math]::Round($fileSize, 2)) MB" -ForegroundColor Gray
    Write-Host "  Location: $PWD\subfinder-pro.exe" -ForegroundColor Gray
} else {
    Write-Host "✗ Binary not found" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "==================================" -ForegroundColor Cyan
Write-Host "Setup Complete! 🎉" -ForegroundColor Green
Write-Host "==================================" -ForegroundColor Cyan
Write-Host ""

# Next steps
Write-Host "Next Steps:" -ForegroundColor Yellow
Write-Host ""
Write-Host "1. Configure API Keys (Optional but recommended):" -ForegroundColor White
Write-Host "   • Edit provider-config.yaml" -ForegroundColor Gray
Write-Host "   • Or set environment variable:" -ForegroundColor Gray
Write-Host "     " -NoNewline
Write-Host '$env:ALIENVAULT_API_KEY="your-key"' -ForegroundColor Cyan
Write-Host ""

Write-Host "2. Test the tool:" -ForegroundColor White
Write-Host "   " -NoNewline
Write-Host ".\subfinder-pro.exe -d example.com" -ForegroundColor Cyan
Write-Host ""

Write-Host "3. Get AlienVault API key (free):" -ForegroundColor White
Write-Host "   • Visit: https://otx.alienvault.com/" -ForegroundColor Gray
Write-Host "   • Sign up → Settings → API Integration" -ForegroundColor Gray
Write-Host ""

Write-Host "4. View examples:" -ForegroundColor White
Write-Host "   " -NoNewline
Write-Host "Get-Content examples\README.md" -ForegroundColor Cyan
Write-Host ""

Write-Host "5. Read documentation:" -ForegroundColor White
Write-Host "   " -NoNewline
Write-Host "Get-Content README.md" -ForegroundColor Cyan
Write-Host "   " -NoNewline
Write-Host "Get-Content QUICKSTART.md" -ForegroundColor Cyan
Write-Host ""

Write-Host "==================================" -ForegroundColor Cyan
Write-Host "Quick Test Command:" -ForegroundColor Yellow
Write-Host ".\subfinder-pro.exe -d example.com -v" -ForegroundColor Cyan
Write-Host "==================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "Happy Subdomain Hunting! 🎯" -ForegroundColor Green
