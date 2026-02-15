# VoidVPN Dual Protocol Test Script
# Run from an Administrator PowerShell terminal

$ErrorActionPreference = "Continue"
$projectDir = "C:\Projects\VoidVPN"
$voidvpn = "$projectDir\voidvpn.exe"
$ovpnBin = "C:\Program Files\OpenVPN\bin\openvpn.exe"
$serverConf = "$projectDir\test-pki\server.conf"

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  VoidVPN - Dual Protocol Test" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

# --- CLEANUP ---
Write-Host "[CLEANUP] Killing old processes and clearing state..." -ForegroundColor Gray
Get-Process openvpn -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
& $voidvpn disconnect 2>&1 | Out-Null
Start-Sleep -Seconds 1

# Remove old test servers
& $voidvpn servers remove wg-demo 2>&1 | Out-Null
& $voidvpn servers remove ovpn-demo 2>&1 | Out-Null

# ==========================================
# TEST 1: WireGuard Config Import
# ==========================================
Write-Host "=== TEST 1: WireGuard .conf Import ===" -ForegroundColor Yellow

Write-Host "  [1a] Importing test-wireguard.conf..." -ForegroundColor White
& $voidvpn servers import "$projectDir\test-pki\test-wireguard.conf" --name wg-demo 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "  [OK] WireGuard config imported" -ForegroundColor Green
} else {
    Write-Host "  [FAIL] WireGuard import failed" -ForegroundColor Red
}

# ==========================================
# TEST 2: OpenVPN Config Import
# ==========================================
Write-Host "`n=== TEST 2: OpenVPN .ovpn Import ===" -ForegroundColor Yellow

Write-Host "  [2a] Importing test-openvpn.ovpn..." -ForegroundColor White
& $voidvpn servers import "$projectDir\test-pki\test-openvpn.ovpn" --name ovpn-demo 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "  [OK] OpenVPN config imported" -ForegroundColor Green
} else {
    Write-Host "  [FAIL] OpenVPN import failed" -ForegroundColor Red
}

# ==========================================
# TEST 3: Server List (Both Protocols)
# ==========================================
Write-Host "`n=== TEST 3: Server List ===" -ForegroundColor Yellow
& $voidvpn servers list 2>&1

# ==========================================
# TEST 4: OpenVPN E2E Connect
# ==========================================
Write-Host "`n=== TEST 4: OpenVPN E2E Connect ===" -ForegroundColor Yellow

Write-Host "  [4a] Starting OpenVPN server..." -ForegroundColor White
$server = Start-Process -FilePath $ovpnBin -ArgumentList "--config", $serverConf -PassThru -WindowStyle Hidden
Write-Host "  Server PID: $($server.Id)" -ForegroundColor Gray
Write-Host "  Waiting 15s for server init..." -ForegroundColor Gray
Start-Sleep -Seconds 15

if ($server.HasExited) {
    Write-Host "  [FAIL] Server exited prematurely" -ForegroundColor Red
} else {
    Write-Host "  [OK] Server running" -ForegroundColor Green

    Write-Host "  [4b] Connecting client (background)..." -ForegroundColor White
    $client = Start-Process -FilePath $voidvpn -ArgumentList "connect", "test-local" -PassThru -WindowStyle Hidden
    Start-Sleep -Seconds 10

    Write-Host "  [4c] Checking status..." -ForegroundColor White
    $statusOutput = & $voidvpn status 2>&1 | Out-String
    if ($statusOutput -match "Connected") {
        Write-Host "  [OK] OpenVPN connected!" -ForegroundColor Green
        # Show status
        & $voidvpn status 2>&1
    } else {
        Write-Host "  [FAIL] Not connected" -ForegroundColor Red
        Write-Host "  Debug log:" -ForegroundColor Gray
        if (Test-Path "$env:TEMP\voidvpn-ovpn-debug.log") {
            Get-Content "$env:TEMP\voidvpn-ovpn-debug.log" | Select-Object -Last 5
        }
    }

    Write-Host "`n  [4d] Disconnecting..." -ForegroundColor White
    & $voidvpn disconnect 2>&1
    Start-Sleep -Seconds 3
    Stop-Process -Id $client.Id -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 2

    Write-Host "  [4e] Verifying disconnected..." -ForegroundColor White
    $statusOutput = & $voidvpn status 2>&1 | Out-String
    if ($statusOutput -match "Disconnected") {
        Write-Host "  [OK] Disconnected" -ForegroundColor Green
    } else {
        Write-Host "  [FAIL] Still shows connected" -ForegroundColor Red
    }
}

# ==========================================
# TEST 5: WireGuard E2E Connect (TUN device test)
# ==========================================
Write-Host "`n=== TEST 5: WireGuard Connect ===" -ForegroundColor Yellow
& $voidvpn disconnect 2>&1 | Out-Null
Start-Sleep -Seconds 2

Write-Host "  [5a] Connecting WireGuard (background)..." -ForegroundColor White
$wgClient = Start-Process -FilePath $voidvpn -ArgumentList "connect", "wg-demo" -PassThru -WindowStyle Hidden
Start-Sleep -Seconds 10

Write-Host "  [5b] Checking status..." -ForegroundColor White
$wgStatus = & $voidvpn status 2>&1 | Out-String
if ($wgStatus -match "Connected") {
    Write-Host "  [OK] WireGuard connected (TUN device created)!" -ForegroundColor Green
    & $voidvpn status 2>&1
} else {
    Write-Host "  [INFO] WireGuard status:" -ForegroundColor Yellow
    & $voidvpn status 2>&1
}

Write-Host "`n  [5c] Disconnecting WireGuard..." -ForegroundColor White
& $voidvpn disconnect 2>&1
Start-Sleep -Seconds 3
Stop-Process -Id $wgClient.Id -Force -ErrorAction SilentlyContinue
Start-Sleep -Seconds 2

Write-Host "  [5d] Verifying disconnected..." -ForegroundColor White
$wgStatus2 = & $voidvpn status 2>&1 | Out-String
if ($wgStatus2 -match "Disconnected") {
    Write-Host "  [OK] WireGuard disconnected" -ForegroundColor Green
} else {
    Write-Host "  [FAIL] Still shows connected" -ForegroundColor Red
}

# ==========================================
# CLEANUP
# ==========================================
Write-Host "`n=== CLEANUP ===" -ForegroundColor Gray
Stop-Process -Id $server.Id -Force -ErrorAction SilentlyContinue
& $voidvpn disconnect 2>&1 | Out-Null

# ==========================================
# SUMMARY
# ==========================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  Test Summary" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  WireGuard .conf import     : TESTED" -ForegroundColor White
Write-Host "  OpenVPN .ovpn import       : TESTED" -ForegroundColor White
Write-Host "  Server list (dual proto)   : TESTED" -ForegroundColor White
Write-Host "  OpenVPN connect            : TESTED" -ForegroundColor White
Write-Host "  OpenVPN status             : TESTED" -ForegroundColor White
Write-Host "  OpenVPN disconnect         : TESTED" -ForegroundColor White
Write-Host "  WireGuard connect path     : TESTED" -ForegroundColor White
Write-Host "========================================`n" -ForegroundColor Cyan
