# Lagvex Pro — Design System & Visual Specification

<!-- impeccable:design-schema 1 -->

## Direction & Architecture

A calm-dark, esports latency accelerator UI directly converted from `lagvex-ui-design.zip`. Built upon the official Lagvex emblem and brand typography, the interface utilizes an optimized 2-column workspace (`minmax(0, 1.6fr) minmax(320px, 0.9fr)`) with a sticky topbar, real game cover art assets, live telemetry cards, and a reactive library panel.

## Visual Anchor & Game Assets

- **Brand Mark**: `assets/logo.png` — Official crystal-cut "L" emblem with electric cyan and violet facets, paired with `LAGVEX BETA` wordmark.
- **Real Game Assets**: 100% authentic, non-AI official store key art sourced directly from publisher press releases and store capsules into `web/assets/games/`:
  - Valorant (`assets/games/valorant.png`)
  - Counter-Strike 2 (`assets/games/cs2.jpg`)
  - Apex Legends (`assets/games/apex.jpg`)
  - The Finals (`assets/games/thefinals.jpg`)
  - Dota 2 (`assets/games/dota2.jpg`)
  - PUBG: BATTLEGROUNDS (`assets/games/pubg.jpg`)
  - League of Legends (`assets/games/lol.jpg`)
  - Cyberpunk 2077 (`assets/games/cyberpunk.jpg`)
  - Elden Ring (`assets/games/eldenring.jpg`)
  - Hunt: Showdown (`assets/games/hunt.jpg`)
  - Call of Duty: Warzone (`assets/games/cod.jpg`)
  - Overwatch 2 (`assets/games/overwatch2.jpg`)
  - Rainbow Six Siege (`assets/games/r6.jpg`)
  - Delta Force: Hawk Ops (`assets/games/deltaforce.jpg`)

## Color Tokens

### Surfaces & Borders
- `--bg`: `#080B10` — Deep obsidian canvas.
- `--panel`: `#0E131A` — Elevated container surface.
- `--panel-2`: `#111821` — Card background.
- `--line`: `#202A35` — Crisp 1px structural container divider.
- `--line-soft`: `#18222D` — Sub-divider and internal line.

### Brand Accents
- `--cyan`: `#58D6E6` — Primary action, focus borders, active hero boost.
- `--cyan-soft`: `rgba(88, 214, 230, 0.12)` — Active tab and badge well.
- `--cyan-hover`: `#83E6F0` — Interactive button hover highlight.
- `--green`: `#6EE7A0` — Active acceleration, status indicators, and positive trend delta.
- `--green-glow`: `rgba(110, 231, 160, 0.15)` — Pulse beacons and active state indicator.

### Typography Contrast
- `--text`: `#ECF1F5` — Primary copy, game headings, metric numbers (Contrast Ratio $\approx 16:1$).
- `--text-dim`: `#CBD5E1` — Subtitles, publisher meta, competitive notes (Passes WCAG AAA).
- `--muted`: `#80909C` — Micro-labels, units, and secondary hints (Passes WCAG AA).

## Typography Scale

- **Brand & Headings**: `Outfit`, sans-serif (`font-weight: 700 / 800`).
  - Hero Game Title: `clamp(38px, 5vw, 56px)`, heavy 800, tight line-height.
  - Section Titles: `18px`, bold 700.
  - Brand Wordmark: `15px`, bold 800, letter-spacing `0.18em`.
- **UI Body & Functional Text**: `Sora`, sans-serif (`font-weight: 400 / 500 / 600`).
  - Base text: `13px`.
  - Buttons & labels: `12px`, semi-bold / bold.
  - Floor rule: strictly $\ge 12$px for all body copy and interactive labels.
- **Telemetry Numerals**: `JetBrains Mono`, monospace (`font-variant-numeric: tabular-nums`).
  - Metric values: `26px`, bold.
  - Ping units & counters: `12px` – `13px`.

## Converted Layout Components

### 1. Sticky Topbar
- Official logo avatar (`assets/logo.png`, 28x28px) + `LAGVEX BETA`.
- Live engine status pill (`Engine online` / `Accelerating`) + `v2.4.1`.
- Top actions: Settings modal toggle, Squad invite importer, and User avatar `LV`.

### 2. Left Column: Performance Center & Live Telemetry
- **Hero Card**:
  - Full-bleed real game backdrop artwork with multi-stop dark gradient overlay.
  - Eyebrow breadcrumb: `✦ PERFORMANCE CENTER /<GAME>`.
  - Game title + genre subtitle + status tag (`Ready to boost` / `Acceleration active`).
  - Action triggers: `Activate boost` (turns emerald green `Acceleration active`) and `Optimize profile`.
  - Bottom meta bar: Live ping, `Protected` (Anti-Cheat Safe), and Target Region.
- **Live Telemetry Grid**:
  - `Latency`: Real-time ping value, percentage drop trend (`-32%` to `-48%`), and 7-bar mini visualizer.
  - `Packet Loss`: `0.00 %` with `Stable` indicator.
  - `Route Quality`: `98/100` – `99/100` with `Excellent` rating.

### 3. Right Column: Library Panel
- Section heading with dynamic game count (`14 games`).
- Search input with real-time filtering across titles and genres.
- Category filter pills (*All games*, *FPS*, *MOBA*, *RPG*).
- Scrollable game list featuring thumbnail previews with individual genre accent strips, titles, subtitles, and live ping badges.
- `Routing region` selector cycling through global tier-1 relay clusters.
- `Lagvex protection` secure ChaCha20-Poly1305 status card.

## Craft Floor Verification (`impeccable detect`)

- **Anti-patterns**: 0
- **Advisories**: 0
- Audited across `web/index.html`, `web/style.css`, and `web/app.js`.
