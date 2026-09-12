<#
.SYNOPSIS
    Lagvex Relay Windows Setup & Runner Script
.DESCRIPTION
    Sets up and starts the Lagvex Relay daemon on a Windows machine for local testing or LAN hosting.
#>

[CmdletBinding()]
param(
    [string]$Port = "4433",
    [string]$Psk = ""
)

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  LAGVEX RELAY SERVER (Windows Setup & Runner)            " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# 1. Check Administrator
$isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Warning "Running without Administrator privileges. Routing & TAP capabilities may be limited."
}

# 2. Setup PSK
$ConfDir = "$env:ProgramData\Lagvex"
if (-not (Test-Path $ConfDir)) {
    New-Item -ItemType Directory -Path $ConfDir -Force | Out-Null
}

$PskFile = Join-Path $ConfDir "psk.key"
if ([string]::IsNullOrWhiteSpace($Psk)) {
    if (Test-Path $PskFile) {
        $Psk = (Get-Content $PskFile -Raw).Trim()
        Write-Host "  Using existing PSK from $PskFile" -ForegroundColor DarkGray
    } else {
        $bytes = New-Object byte[] 24
        [Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($bytes)
        $Psk = [Convert]::ToBase64String($bytes) -replace '[^a-zA-Z0-9]', ''
        $Psk = $Psk.Substring(0, [Math]::Min(32, $Psk.Length))
        $Psk | Out-File -FilePath $PskFile -Encoding ascii
        Write-Host "  Generated new secure 32-character PSK." -ForegroundColor Green
    }
}

# 3. Detect Public / Local IP
$LocalIP = (Get-NetIPAddress -AddressFamily IPv4 | Where-Object { $_.InterfaceAlias -notmatch 'Loopback|vEthernet|Virtual' -and $_.IPAddress -notmatch '^169\.254' } | Select-Object -First 1).IPAddress
if (-not $LocalIP) { $LocalIP = "127.0.0.1" }

Write-Host ""
Write-Host "  Relay Endpoint : $LocalIP`:$Port" -ForegroundColor Green
Write-Host "  PSK Secret     : $Psk" -ForegroundColor Yellow
Write-Host ""
Write-Host "🤝 1-CLICK SQUAD SHARE LINK (Send this to your squad teammates):" -ForegroundColor Yellow
Write-Host "  lagvex://connect?endpoint=$LocalIP`:$Port&psk=$Psk&name=SquadRelay" -ForegroundColor Cyan
Write-Host ""

# 4. Locate or Build Binary
$Binary = Join-Path $PSScriptRoot "..\bin\lagvex-relay.exe"
if (-not (Test-Path $Binary)) {
    $Binary = Join-Path $PSScriptRoot "..\bin\lagvex-relay"
}

if (-not (Test-Path $Binary)) {
    Write-Host "Compiling lagvex-relay for Windows..." -ForegroundColor DarkCyan
    & go build -ldflags="-s -w" -o (Join-Path $PSScriptRoot "..\bin\lagvex-relay.exe") (Join-Path $PSScriptRoot "..\cmd\lagvex-relay")
    $Binary = Join-Path $PSScriptRoot "..\bin\lagvex-relay.exe"
}

if (Test-Path $Binary) {
    Write-Host "Starting Lagvex Relay daemon on port $Port..." -ForegroundColor Green
    & $Binary -listen "0.0.0.0:$Port" -psk "$Psk"
} else {
    Write-Warning "Could not find or compile lagvex-relay.exe. Please ensure Go is installed."
}
