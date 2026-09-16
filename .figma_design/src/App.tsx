import { useMemo, useState } from "react"

/* ------------------------------------------------------------------ */
/* Icons                                                              */
/* ------------------------------------------------------------------ */

type IconProps = { className?: string }

function Bolt({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="none" className={className}>
      <path d="M13 2 4.5 13.5H11l-1 8.5 9.5-12H13l0-8Z" fill="currentColor" />
    </svg>
  )
}

function Search({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="none" className={className}>
      <circle cx="11" cy="11" r="6.5" stroke="currentColor" strokeWidth="1.6" />
      <path d="m20 20-3.4-3.4" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
    </svg>
  )
}

function Chevron({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="none" className={className}>
      <path d="m6 9 6 6 6-6" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

function Shield({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="none" className={className}>
      <path d="M12 3 5 5.5V11c0 4.5 3 8 7 9.5 4-1.5 7-5 7-9.5V5.5L12 3Z" stroke="currentColor" strokeWidth="1.5" strokeLinejoin="round" />
      <path d="m9 12 2 2 4-4.5" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

function Plus({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="none" className={className}>
      <path d="M12 5v14M5 12h14" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" />
    </svg>
  )
}

function Star({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" className={className}>
      <path d="m12 3 2.5 5.4 5.9.7-4.4 4 1.2 5.8L12 16.9 6.8 19.9 8 14.1 3.6 10l5.9-.7L12 3Z" />
    </svg>
  )
}

function ArrowUp({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" fill="none" className={className}>
      <path d="M12 19V6m0 0-5 5m5-5 5 5" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

/* ------------------------------------------------------------------ */
/* Data                                                               */
/* ------------------------------------------------------------------ */

const IMG =
  "?auto=format&fit=crop&q=80&w=600&h=760"

type Game = {
  id: string
  title: string
  studio: string
  category: string
  cover: string
}

const GAMES: Game[] = [
  { id: "valorant", title: "Valorant", studio: "Riot Games", category: "Tactical FPS", cover: "https://images.unsplash.com/flagged/photo-1560177776-55a762c5c000" },
  { id: "cs2", title: "Counter-Strike 2", studio: "Valve", category: "Tactical FPS", cover: "https://images.unsplash.com/photo-1573511860313-d333c8022170" },
  { id: "apex", title: "Apex Legends", studio: "Respawn", category: "Battle Royale", cover: "https://images.unsplash.com/photo-1672872476232-da16b45c9001" },
  { id: "pubg", title: "PUBG: Battlegrounds", studio: "Krafton", category: "Battle Royale", cover: "https://images.unsplash.com/photo-1514124838563-9243ab7791a6" },
  { id: "lol", title: "League of Legends", studio: "Riot Games", category: "MOBA", cover: "https://images.unsplash.com/photo-1566410824233-a8011929225c" },
  { id: "dota", title: "Dota 2", studio: "Valve", category: "MOBA", cover: "https://images.unsplash.com/photo-1560671021-cb36f70ce82d" },
  { id: "fortnite", title: "Fortnite", studio: "Epic Games", category: "Battle Royale", cover: "https://images.unsplash.com/photo-1519608487953-e999c86e7455" },
  { id: "ow2", title: "Overwatch 2", studio: "Blizzard", category: "Tactical FPS", cover: "https://images.unsplash.com/photo-1530919424169-4b95f917e937" },
]

const FILTERS = ["All Games", "Tactical FPS", "Battle Royale", "MOBA"] as const

const REGIONS = ["Asia Pacific", "Europe West", "North America", "South America"]

type Relay = { name: string; ping: number; tag: string }
const RELAYS: Relay[] = [
  { name: "Singapore #1", ping: 18, tag: "OPTIMAL" },
  { name: "Tokyo #3", ping: 24, tag: "STABLE" },
  { name: "Hong Kong #2", ping: 27, tag: "STABLE" },
  { name: "Seoul #4", ping: 33, tag: "BUSY" },
]

/* ------------------------------------------------------------------ */
/* Small building blocks                                              */
/* ------------------------------------------------------------------ */

function Card({ className = "", children }: { className?: string; children: React.ReactNode }) {
  return (
    <div
      className={`rounded-xl border border-[var(--color-line)] bg-[var(--color-surface)] ${className}`}
    >
      {children}
    </div>
  )
}

function Label({ children }: { children: React.ReactNode }) {
  return (
    <span className="text-[10px] font-semibold uppercase tracking-[0.18em] text-[var(--color-ink-faint)]">
      {children}
    </span>
  )
}

/* Menu: region + relay selectors */
function Dropdown({
  value,
  children,
  open,
  onToggle,
  wide,
}: {
  value: React.ReactNode
  children: React.ReactNode
  open: boolean
  onToggle: () => void
  wide?: boolean
}) {
  return (
    <div className="relative">
      <button
        onClick={onToggle}
        className="flex w-full items-center justify-between gap-3 rounded-lg border border-[var(--color-line)] bg-[var(--color-base-elev)]/70 px-3.5 py-2.5 text-left transition-colors hover:border-[#2f3a50]"
      >
        {value}
        <Chevron className={`h-4 w-4 shrink-0 text-[var(--color-ink-faint)] transition-transform ${open ? "rotate-180" : ""}`} />
      </button>
      {open && (
        <div
          className={`absolute z-30 mt-2 overflow-hidden rounded-lg border border-[var(--color-line)] bg-[var(--color-surface-2)] p-1 shadow-2xl shadow-black/60 ${wide ? "w-full" : "w-full"}`}
        >
          {children}
        </div>
      )}
    </div>
  )
}

/* ------------------------------------------------------------------ */
/* App                                                                */
/* ------------------------------------------------------------------ */

export default function App() {
  const [selectedId, setSelectedId] = useState("valorant")
  const [boosting, setBoosting] = useState(false)
  const [region, setRegion] = useState(REGIONS[0])
  const [relay, setRelay] = useState(RELAYS[0])
  const [openRegion, setOpenRegion] = useState(false)
  const [openRelay, setOpenRelay] = useState(false)
  const [filter, setFilter] = useState<(typeof FILTERS)[number]>("All Games")
  const [query, setQuery] = useState("")

  const game = GAMES.find((g) => g.id === selectedId)!

  const filtered = useMemo(() => {
    return GAMES.filter((g) => {
      const matchCat = filter === "All Games" || g.category === filter
      const matchQ =
        g.title.toLowerCase().includes(query.toLowerCase()) ||
        g.studio.toLowerCase().includes(query.toLowerCase())
      return matchCat && matchQ
    })
  }, [filter, query])

  const ispPing = 58
  const accel = boosting ? relay.ping + 10 : ispPing
  const delta = ispPing - accel

  return (
    <div className="flex min-h-screen items-center justify-center bg-[var(--color-base)] p-6">
      <div
        className="relative flex flex-col overflow-hidden rounded-2xl border border-[var(--color-line)] bg-[var(--color-base)] shadow-[0_40px_120px_-20px_rgba(0,0,0,0.9)]"
        style={{ width: 1280, height: 800 }}
      >
        {/* ambient glow */}
        <div
          className="pointer-events-none absolute -top-40 right-10 h-96 w-96 rounded-full opacity-[0.14] blur-3xl"
          style={{ background: boosting ? "var(--color-emerald)" : "var(--color-cyan)" }}
        />

        {/* ============ HEADER ============ */}
        <header className="relative z-20 flex items-center justify-between border-b border-[var(--color-line-soft)] px-7 py-4">
          <div className="flex items-center gap-2.5">
            <div className="grid h-8 w-8 place-items-center rounded-lg bg-[var(--color-cyan)] text-[var(--color-base)] shadow-[0_0_20px_-2px_var(--color-cyan)]">
              <Bolt className="h-4.5 w-4.5" />
            </div>
            <div className="flex items-baseline gap-1.5 leading-none">
              <span className="text-[15px] font-extrabold tracking-tight">LAGVEX</span>
              <span className="rounded bg-[var(--color-line)] px-1.5 py-0.5 text-[9px] font-bold tracking-[0.14em] text-[var(--color-ink-dim)]">
                PRO
              </span>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <StatusPill boosting={boosting} />
            <button className="flex items-center gap-1.5 rounded-lg border border-[var(--color-line)] px-3 py-2 text-[12px] font-semibold text-[var(--color-ink-dim)] transition-colors hover:border-[#2f3a50] hover:text-[var(--color-ink)]">
              <Plus className="h-3.5 w-3.5" />
              Custom VPS
            </button>
          </div>
        </header>

        {/* ============ BODY ============ */}
        <div className="grid flex-1 grid-cols-[1fr_360px] overflow-hidden">
          {/* ---- LEFT column: hero + library ---- */}
          <div className="flex min-h-0 flex-col gap-5 overflow-y-auto px-7 py-6">
            {/* Hero HUD */}
            <section className="relative shrink-0 overflow-hidden rounded-2xl border border-[var(--color-line)]">
              <div className="absolute inset-0 bg-[var(--color-surface)]">
                <img
                  src={game.cover + IMG}
                  alt={`${game.title} cover art`}
                  className="h-full w-full object-cover opacity-70"
                />
              </div>
              <div className="absolute inset-0 bg-gradient-to-r from-[#0b0e14] via-[#0b0e14]/85 to-[#0b0e14]/30" />
              <div className="absolute inset-0 bg-gradient-to-t from-[#0b0e14] via-transparent to-transparent" />

              <div className="relative flex items-end justify-between gap-6 p-7" style={{ minHeight: 250 }}>
                <div className="max-w-[58%]">
                  <div className="mb-3 flex items-center gap-2">
                    <span className="flex items-center gap-1.5 rounded-full border border-[var(--color-emerald)]/30 bg-[var(--color-emerald)]/10 px-2.5 py-1 text-[10px] font-bold uppercase tracking-wider text-[var(--color-emerald)]">
                      <Shield className="h-3 w-3" /> Anti-Cheat Safe
                    </span>
                    <Label>Selected Game</Label>
                  </div>
                  <h1 className="text-[42px] font-extrabold leading-none tracking-tight">{game.title}</h1>
                  <p className="mt-2 text-[13px] font-medium text-[var(--color-ink-dim)]">
                    {game.studio} <span className="text-[var(--color-ink-faint)]">•</span> {game.category}
                  </p>

                  <div className="mt-6 grid grid-cols-2 gap-3">
                    {/* Region */}
                    <div>
                      <div className="mb-1.5">
                        <Label>Target Region</Label>
                      </div>
                      <Dropdown
                        open={openRegion}
                        onToggle={() => {
                          setOpenRegion((v) => !v)
                          setOpenRelay(false)
                        }}
                        value={<span className="text-[13px] font-semibold text-[var(--color-ink)]">{region}</span>}
                      >
                        {REGIONS.map((r) => (
                          <button
                            key={r}
                            onClick={() => {
                              setRegion(r)
                              setOpenRegion(false)
                            }}
                            className={`flex w-full items-center justify-between rounded-md px-3 py-2 text-[13px] font-medium transition-colors hover:bg-[var(--color-line)]/60 ${r === region ? "text-[var(--color-cyan)]" : "text-[var(--color-ink-dim)]"}`}
                          >
                            {r}
                            {r === region && <span className="h-1.5 w-1.5 rounded-full bg-[var(--color-cyan)]" />}
                          </button>
                        ))}
                      </Dropdown>
                    </div>

                    {/* Relay */}
                    <div>
                      <div className="mb-1.5">
                        <Label>Relay Node</Label>
                      </div>
                      <Dropdown
                        open={openRelay}
                        onToggle={() => {
                          setOpenRelay((v) => !v)
                          setOpenRegion(false)
                        }}
                        value={
                          <span className="flex items-center gap-2 truncate text-[13px] font-semibold">
                            {relay.name}
                            <span className="tabular text-[11px] font-bold text-[var(--color-emerald)]">{relay.ping}ms</span>
                          </span>
                        }
                      >
                        {RELAYS.map((r) => (
                          <button
                            key={r.name}
                            onClick={() => {
                              setRelay(r)
                              setOpenRelay(false)
                            }}
                            className={`flex w-full items-center justify-between gap-2 rounded-md px-3 py-2 text-[13px] transition-colors hover:bg-[var(--color-line)]/60 ${r.name === relay.name ? "bg-[var(--color-line)]/40" : ""}`}
                          >
                            <span className="font-medium text-[var(--color-ink)]">{r.name}</span>
                            <span className="flex items-center gap-2">
                              <span className="tabular text-[11px] font-bold text-[var(--color-emerald)]">{r.ping}ms</span>
                              <span className="text-[9px] font-bold uppercase tracking-wider text-[var(--color-ink-faint)]">{r.tag}</span>
                            </span>
                          </button>
                        ))}
                      </Dropdown>
                      <div className="mt-1.5 flex items-center gap-1 pl-0.5">
                        <Star className="h-3 w-3 text-[var(--color-emerald)]" />
                        <span className="text-[11px] font-semibold text-[var(--color-emerald)]">{relay.name} — {relay.ping}ms · {relay.tag}</span>
                      </div>
                    </div>
                  </div>
                </div>

                {/* CTA */}
                <button
                  onClick={() => setBoosting((v) => !v)}
                  className={`group relative flex h-[104px] w-[176px] shrink-0 flex-col items-center justify-center gap-2 overflow-hidden rounded-xl font-extrabold transition-all ${
                    boosting
                      ? "border border-[var(--color-emerald)]/50 bg-[var(--color-emerald)]/10 text-[var(--color-emerald)]"
                      : "bg-[var(--color-cyan)] text-[var(--color-base)] shadow-[0_0_40px_-6px_var(--color-cyan)] hover:shadow-[0_0_50px_-4px_var(--color-cyan)]"
                  }`}
                >
                  {!boosting && (
                    <span
                      className="pointer-events-none absolute inset-0 -skew-x-12 bg-white/25"
                      style={{ animation: "lv-sweep 2.6s ease-in-out infinite", width: "40%" }}
                    />
                  )}
                  <Bolt className="relative h-7 w-7" />
                  <span className="relative text-[14px] tracking-wide">
                    {boosting ? "STOP BOOST" : "ACTIVATE BOOST"}
                  </span>
                </button>
              </div>
            </section>

            {/* Library */}
            <section className="shrink-0">
              <div className="mb-4 flex items-center justify-between gap-4">
                <div className="flex items-center gap-2">
                  <h2 className="text-[15px] font-bold tracking-tight">Game Library</h2>
                  <span className="tabular rounded-full bg-[var(--color-line)] px-2 py-0.5 text-[10px] font-bold text-[var(--color-ink-dim)]">
                    {filtered.length}
                  </span>
                </div>
                <div className="relative w-56">
                  <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--color-ink-faint)]" />
                  <input
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    placeholder="Search games…"
                    className="w-full rounded-lg border border-[var(--color-line)] bg-[var(--color-surface)] py-2 pl-9 pr-3 text-[13px] font-medium text-[var(--color-ink)] outline-none placeholder:text-[var(--color-ink-faint)] focus:border-[var(--color-cyan)]/50"
                  />
                </div>
              </div>

              <div className="mb-4 flex items-center gap-2">
                {FILTERS.map((f) => (
                  <button
                    key={f}
                    onClick={() => setFilter(f)}
                    className={`rounded-full border px-3.5 py-1.5 text-[12px] font-semibold transition-colors ${
                      filter === f
                        ? "border-[var(--color-cyan)]/40 bg-[var(--color-cyan)]/10 text-[var(--color-cyan)]"
                        : "border-[var(--color-line)] text-[var(--color-ink-dim)] hover:border-[#2f3a50] hover:text-[var(--color-ink)]"
                    }`}
                  >
                    {f}
                  </button>
                ))}
              </div>

              <div className="grid grid-cols-4 gap-3.5">
                {filtered.map((g) => {
                  const active = g.id === selectedId
                  return (
                    <button
                      key={g.id}
                      onClick={() => setSelectedId(g.id)}
                      className={`group relative aspect-[3/4] overflow-hidden rounded-xl border text-left transition-all ${
                        active
                          ? "border-[var(--color-cyan)] shadow-[0_0_0_1px_var(--color-cyan),0_0_28px_-6px_var(--color-cyan)]"
                          : "border-[var(--color-line)] hover:border-[#2f3a50]"
                      }`}
                    >
                      <div className="absolute inset-0 bg-[var(--color-surface)]">
                        <img
                          src={g.cover + "?auto=format&fit=crop&q=80&w=320&h=420"}
                          alt={`${g.title} poster`}
                          className={`h-full w-full object-cover transition-all duration-500 ${active ? "opacity-95" : "opacity-70 group-hover:opacity-90"}`}
                        />
                      </div>
                      <div className="absolute inset-0 bg-gradient-to-t from-[#0b0e14] via-[#0b0e14]/20 to-transparent" />

                      <span className="absolute left-2 top-2 flex h-6 w-6 items-center justify-center rounded-md bg-[var(--color-base)]/70 text-[var(--color-emerald)] backdrop-blur-sm" title="Anti-cheat safe">
                        <Shield className="h-3.5 w-3.5" />
                      </span>
                      {active && (
                        <span className="absolute right-2 top-2 rounded-md bg-[var(--color-cyan)] px-1.5 py-0.5 text-[9px] font-extrabold uppercase tracking-wider text-[var(--color-base)]">
                          Active
                        </span>
                      )}

                      <div className="absolute inset-x-0 bottom-0 p-2.5">
                        <p className="truncate text-[12.5px] font-bold leading-tight">{g.title}</p>
                        <p className="truncate text-[10px] font-medium text-[var(--color-ink-dim)]">{g.category}</p>
                      </div>
                    </button>
                  )
                })}
              </div>
            </section>
          </div>

          {/* ---- RIGHT column: telemetry ---- */}
          <aside className="flex min-h-0 flex-col gap-3.5 overflow-y-auto border-l border-[var(--color-line-soft)] bg-[var(--color-base-elev)] px-5 py-6">
            <div className="flex items-center justify-between">
              <h2 className="text-[13px] font-bold uppercase tracking-[0.16em] text-[var(--color-ink-dim)]">Telemetry</h2>
              <span className={`flex items-center gap-1.5 text-[11px] font-bold ${boosting ? "text-[var(--color-emerald)]" : "text-[var(--color-ink-faint)]"}`}>
                <span className={`h-1.5 w-1.5 rounded-full ${boosting ? "bg-[var(--color-emerald)] dot-live" : "bg-[var(--color-ink-faint)]"}`} />
                {boosting ? "LIVE" : "IDLE"}
              </span>
            </div>

            {/* Differential */}
            <Card className="p-4">
              <div className="grid grid-cols-2 gap-3">
                <PingBlock label="Direct ISP" value={ispPing} tone="dim" />
                <PingBlock label="Lagvex Accel." value={accel} tone="cyan" active={boosting} />
              </div>
              <div className="mt-3 flex items-center justify-center">
                <span
                  className={`tabular rounded-lg px-3 py-1.5 text-[15px] font-extrabold transition-colors ${
                    boosting
                      ? "bg-[var(--color-emerald)]/12 text-[var(--color-emerald)]"
                      : "bg-[var(--color-line)]/50 text-[var(--color-ink-faint)]"
                  }`}
                >
                  {boosting ? `−${delta}ms` : "—"} <span className="text-[10px] font-semibold tracking-wide">{boosting ? "IMPROVED" : "STANDBY"}</span>
                </span>
              </div>
            </Card>

            {/* Stability */}
            <div>
              <div className="mb-2 px-0.5">
                <Label>Stability</Label>
              </div>
              <div className="grid grid-cols-3 gap-2.5">
                <Metric label="P50" value={boosting ? "28" : "—"} unit="ms" />
                <Metric label="P95 Tail" value={boosting ? "31" : "—"} unit="ms" hint="spike guard" />
                <Metric label="Packet Loss" value={boosting ? "0.0" : "—"} unit="%" good={boosting} />
              </div>
            </div>

            {/* Throughput */}
            <Card className="p-4">
              <div className="mb-3 flex items-center justify-between">
                <Label>Live Throughput</Label>
                <span className="tabular text-[10px] font-semibold text-[var(--color-ink-faint)]">KB/s</span>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <Throughput dir="down" value={boosting ? 842 : 0} />
                <Throughput dir="up" value={boosting ? 196 : 0} />
              </div>
              <div className="mt-4 flex items-end gap-1" style={{ height: 40 }}>
                {BARS.map((h, i) => (
                  <span
                    key={i}
                    className="flex-1 rounded-sm transition-all duration-500"
                    style={{
                      height: `${boosting ? h : 6}%`,
                      background: boosting ? "var(--color-cyan)" : "var(--color-line)",
                      opacity: boosting ? 0.35 + (i / BARS.length) * 0.65 : 1,
                    }}
                  />
                ))}
              </div>
            </Card>

            {/* Routes */}
            <Card className="flex items-center justify-between p-4">
              <div>
                <Label>Active Routes</Label>
                <p className="tabular mt-1 text-[26px] font-extrabold leading-none">
                  {boosting ? "3" : "0"}
                  <span className="text-[13px] font-semibold text-[var(--color-ink-faint)]"> / 3</span>
                </p>
              </div>
              <div className="flex flex-col items-end gap-1">
                <span className="text-[11px] font-semibold text-[var(--color-ink-dim)]">Multi-Path Relay</span>
                <div className="flex gap-1">
                  {[0, 1, 2].map((i) => (
                    <span
                      key={i}
                      className={`h-1.5 w-6 rounded-full ${boosting ? "bg-[var(--color-emerald)]" : "bg-[var(--color-line)]"}`}
                    />
                  ))}
                </div>
              </div>
            </Card>

            <p className="mt-auto pt-2 text-center text-[10px] font-medium text-[var(--color-ink-faint)]">
              Route: <span className="text-[var(--color-ink-dim)]">{region} → {relay.name}</span>
            </p>
          </aside>
        </div>
      </div>
    </div>
  )
}

/* ------------------------------------------------------------------ */
/* Telemetry sub-components                                           */
/* ------------------------------------------------------------------ */

const BARS = [30, 55, 42, 68, 50, 74, 60, 82, 58, 70, 48, 64]

function StatusPill({ boosting }: { boosting: boolean }) {
  return (
    <span
      className={`flex items-center gap-2 rounded-full border px-3 py-1.5 text-[11px] font-bold uppercase tracking-wider transition-colors ${
        boosting
          ? "border-[var(--color-emerald)]/40 bg-[var(--color-emerald)]/10 text-[var(--color-emerald)]"
          : "border-[var(--color-line)] text-[var(--color-ink-dim)]"
      }`}
    >
      <span className={`h-2 w-2 rounded-full ${boosting ? "bg-[var(--color-emerald)] dot-live" : "bg-[var(--color-ink-faint)]"}`} />
      {boosting ? "Accelerating" : "Standby"}
    </span>
  )
}

function PingBlock({
  label,
  value,
  tone,
  active,
}: {
  label: string
  value: number
  tone: "dim" | "cyan"
  active?: boolean
}) {
  const color = tone === "cyan" && active ? "text-[var(--color-cyan)]" : tone === "dim" ? "text-[var(--color-ink)]" : "text-[var(--color-ink-dim)]"
  return (
    <div className="rounded-lg border border-[var(--color-line-soft)] bg-[var(--color-base)]/50 p-3">
      <p className="text-[10px] font-semibold uppercase tracking-wider text-[var(--color-ink-faint)]">{label}</p>
      <p className={`tabular mt-1 text-[30px] font-extrabold leading-none ${color}`}>
        {value}
        <span className="text-[12px] font-semibold text-[var(--color-ink-faint)]">ms</span>
      </p>
    </div>
  )
}

function Metric({
  label,
  value,
  unit,
  hint,
  good,
}: {
  label: string
  value: string
  unit: string
  hint?: string
  good?: boolean
}) {
  return (
    <div className="rounded-lg border border-[var(--color-line)] bg-[var(--color-surface)] p-3">
      <p className="truncate text-[9.5px] font-semibold uppercase tracking-wider text-[var(--color-ink-faint)]">{label}</p>
      <p className={`tabular mt-1.5 text-[19px] font-extrabold leading-none ${good ? "text-[var(--color-emerald)]" : "text-[var(--color-ink)]"}`}>
        {value}
        <span className="text-[10px] font-semibold text-[var(--color-ink-faint)]">{unit}</span>
      </p>
      {hint && <p className="mt-1 text-[9px] font-medium text-[var(--color-ink-faint)]">{hint}</p>}
    </div>
  )
}

function Throughput({ dir, value }: { dir: "up" | "down"; value: number }) {
  const isUp = dir === "up"
  return (
    <div className="rounded-lg border border-[var(--color-line-soft)] bg-[var(--color-base)]/50 p-3">
      <div className="flex items-center gap-1.5">
        <ArrowUp className={`h-3.5 w-3.5 ${isUp ? "text-[var(--color-cyan)]" : "rotate-180 text-[var(--color-emerald)]"}`} />
        <span className="text-[10px] font-semibold uppercase tracking-wider text-[var(--color-ink-faint)]">
          {isUp ? "Upload" : "Download"}
        </span>
      </div>
      <p className="tabular mt-1.5 text-[20px] font-extrabold leading-none">{value.toFixed(0)}</p>
    </div>
  )
}
