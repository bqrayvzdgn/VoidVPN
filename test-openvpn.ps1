# VoidVPN OpenVPN E2E Test Script
# Run this from an Administrator PowerShell terminal

$ErrorActionPreference = "Continue"
$ovpnBin = "C:\Program Files\OpenVPN\bin\openvpn.exe"
$projectDir = "C:\Projects\VoidVPN"
$serverConf = "$projectDir\test-pki\server.conf"
$voidvpn = "$projectDir\voidvpn.exe"

Write-Host "`n=== VoidVPN OpenVPN E2E Test ===" -ForegroundColor Cyan

# Step 1: Kill any existing openvpn processes and clear state
Write-Host "`n[1/6] Cleaning up..." -ForegroundColor Yellow
Get-Process openvpn -ErrorAction SilentlyContinue | Where-Object { $_.Path -eq $ovpnBin } | Stop-Process -Force -ErrorAction SilentlyContinue
& $voidvpn disconnect 2>&1 | Out-Null
Start-Sleep -Seconds 2

# Step 2: Start OpenVPN server in background
Write-Host "[2/6] Starting OpenVPN server..." -ForegroundColor Yellow
$server = Start-Process -FilePath $ovpnBin -ArgumentList "--config", $serverConf -PassThru -WindowStyle Hidden
Write-Host "  Server PID: $($server.Id)"
Write-Host "  Waiting for server initialization (15s)..." -ForegroundColor Gray
Start-Sleep -Seconds 15

if ($server.HasExited) {
    Write-Host "  ERROR: Server exited prematurely!" -ForegroundColor Red
    exit 1
}
Write-Host "  Server is running." -ForegroundColor Green

# Step 3: Connect VoidVPN client in background (it blocks after connecting)
Write-Host "[3/6] Connecting VoidVPN client..." -ForegroundColor Yellow
$client = Start-Process -FilePath $voidvpn -ArgumentList "connect", "test-local" -PassThru -WindowStyle Hidden
Start-Sleep -Seconds 10

# Step 4: Check connection status
Write-Host "[4/6] Checking connection status..." -ForegroundColor Yellow
& $voidvpn status 2>&1

# Step 5: Check debug log
Write-Host "`n[5/6] Client debug log (last 5 lines):" -ForegroundColor Yellow
if (Test-Path "$env:TEMP\voidvpn-ovpn-debug.log") {
    Get-Content "$env:TEMP\voidvpn-ovpn-debug.log" | Select-Object -Last 5
}

# Step 6: Cleanup
Write-Host "`n[6/6] Disconnecting and cleaning up..." -ForegroundColor Yellow
& $voidvpn disconnect 2>&1
Stop-Process -Id $client.Id -Force -ErrorAction SilentlyContinue
Stop-Process -Id $server.Id -Force -ErrorAction SilentlyContinue
Start-Sleep -Seconds 2
Write-Host "`n=== Test Complete ===" -ForegroundColor Cyan
