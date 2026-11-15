<#
.SYNOPSIS
    Lagvex Relay — Windows is not a supported relay host.

.DESCRIPTION
    Relay hosting requires a Linux kernel TUN device (/dev/net/tun) plus
    iptables/NAT forwarding. Windows and macOS cannot host a Lagvex relay, and
    this script intentionally does not pretend otherwise. Use a small Linux VPS
    (KVM) and run scripts/install-relay.sh there instead.

    The Windows/macOS binaries in this project are *clients*, not relays.
#>

[CmdletBinding()]
param()

Write-Host "==========================================================" -ForegroundColor Yellow
Write-Host "  LAGVEX RELAY — NOT SUPPORTED ON WINDOWS                 " -ForegroundColor Yellow
Write-Host "==========================================================" -ForegroundColor Yellow
Write-Host ""
Write-Host "Relay hosting needs a Linux kernel:" -ForegroundColor White
Write-Host "  * /dev/net/tun for the tunnel interface" -ForegroundColor Gray
Write-Host "  * iptables NAT masquerade and forwarding for client traffic" -ForegroundColor Gray
Write-Host ""
Write-Host "Windows and macOS can run the Lagvex *client*, but cannot host a relay." -ForegroundColor White
Write-Host ""
Write-Host "Deploy the relay on a Linux VPS (KVM, not OpenVZ) with:" -ForegroundColor Cyan
Write-Host "  curl -fsSL https://raw.githubusercontent.com/ThanhNguyxnOrg/lagvex/main/scripts/install-relay.sh | sudo bash" -ForegroundColor Cyan
Write-Host ""
Write-Host "Full guide: docs/DEPLOYMENT.md" -ForegroundColor Gray

exit 1
