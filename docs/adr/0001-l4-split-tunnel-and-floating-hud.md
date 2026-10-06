# ADR 0001: L4 Game Tick Filtering & Optional In-Game Mini HUD

## Status
Accepted

## Context
1. **L4 Traffic Pollution**: Game launchers and game clients frequently download background asset updates, cinematics, and in-game shop assets via TCP HTTP/HTTPS (ports 80 and 443) on the same IP subnets as game servers. Routing multi-gigabyte TCP streams through the VPS relay introduces bufferbloat for game ticks and consumes server bandwidth quota.
2. **In-Game Monitoring Ergonomics**: Players in competitive matches cannot easily switch windows (Alt-Tab) to view the Web UI Dashboard without risking game disruptions. An optional, lightweight floating HUD is needed, but it must be completely optional and dismissable.
3. **Accessibility**: Community relays and free zero-config modes must remain permanently intact.

## Decision
1. **L4 Smart Traffic Splitter**:
   - Inspect IPv4 transport protocol (byte 9) inside the client packet pump.
   - Keep UDP gaming loop packets accelerated through the ChaCha20-Poly1305 tunnel.
   - When enabled via settings (`smartL4Filter`), allow non-gaming bulk TCP streams (ports 80, 443) to bypass the tunnel.
2. **Optional In-Game Floating Mini HUD**:
   - Provide a draggable, glassmorphism floating pill widget directly in the web app UI.
   - Support one-click minimize, dismiss (`×`), and a master toggle switch in the dashboard header and settings.
   - Persist user display preferences in `localStorage`.
3. **Self-Hosted VPS Automation**:
   - Provide a POSIX-compliant 1-click deploy bash script for Ubuntu/Debian configuring Linux BBR, iptables NAT, and systemd.
   - Retain all free public community relays as default zero-config options.

## Consequences
- Preserves VPS bandwidth for actual game ticks.
- Players can monitor live latency in real-time without leaving their game.
- No forced UX: the floating HUD can be hidden or disabled at will.
