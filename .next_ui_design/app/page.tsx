'use client'

import { useMemo, useState } from 'react'
import { Activity, ChevronDown, Crosshair, Gauge, Gamepad2, Globe2, Search, Settings2, ShieldCheck, Sparkles, Zap } from 'lucide-react'

type Game = { name: string; genre: string; tag: string; accent: string; image: string; ping: string }

const games: Game[] = [
  { name: 'Valorant', genre: 'Tactical shooter', tag: 'FPS', accent: '#ef4444', image: '/valorant-hero.png', ping: '18 ms' },
  { name: 'Apex Legends', genre: 'Battle royale', tag: 'FPS', accent: '#f59e0b', image: '/game-library.png', ping: '27 ms' },
  { name: 'Cyberpunk 2077', genre: 'Action RPG', tag: 'RPG', accent: '#a855f7', image: '/game-library.png', ping: '34 ms' },
  { name: 'The Finals', genre: 'Arena shooter', tag: 'FPS', accent: '#22d3ee', image: '/game-library.png', ping: '22 ms' },
  { name: 'Elden Ring', genre: 'Action RPG', tag: 'RPG', accent: '#eab308', image: '/game-library.png', ping: '41 ms' },
  { name: 'Hunt: Showdown', genre: 'Extraction shooter', tag: 'FPS', accent: '#84cc16', image: '/game-library.png', ping: '31 ms' },
]

export default function Page() {
  const [selected, setSelected] = useState(games[0])
  const [query, setQuery] = useState('')
  const [category, setCategory] = useState('All games')
  const [region, setRegion] = useState('Auto region')
  const [active, setActive] = useState(false)
  const filteredGames = useMemo(() => games.filter((game) => (category === 'All games' || game.tag === category) && game.name.toLowerCase().includes(query.toLowerCase())), [category, query])

  return (
    <main className="lagvex-shell">
      <header className="topbar">
        <div className="brand"><span className="brand-mark"><Zap size={15} fill="currentColor" /></span><span>LAGVEX</span><span className="brand-beta">BETA</span></div>
        <div className="engine-status"><span className="status-dot" /> Engine online <span className="divider" /> v2.4.1</div>
        <div className="top-actions"><button className="icon-button" aria-label="Settings"><Settings2 size={17} /></button><div className="avatar">JD</div></div>
      </header>

      <div className="workspace">
        <section className="main-column">
          <div className="eyebrow"><Sparkles size={14} /> PERFORMANCE CENTER <span>/{selected.name.toUpperCase()}</span></div>
          <div className="hero-card">
            <img src={selected.image} alt="" className="hero-image" />
            <div className="hero-overlay" />
            <div className="hero-content">
              <div className="hero-kicker"><span className="live-pip" /> READY TO BOOST</div>
              <h1>{selected.name}</h1><p>{selected.genre} <span>•</span> Competitive profile</p>
              <div className="hero-actions"><button className={`boost-button ${active ? 'is-active' : ''}`} onClick={() => setActive(!active)}><Zap size={17} fill="currentColor" /> {active ? 'Acceleration active' : 'Activate boost'}</button><button className="ghost-button"><Crosshair size={16} /> Optimize profile</button></div>
            </div>
            <div className="hero-meta"><span><Gauge size={14} /> {active ? '12 ms' : selected.ping}</span><span><ShieldCheck size={14} /> Protected</span></div>
          </div>

          <div className="telemetry-head"><div><div className="section-label">LIVE TELEMETRY</div><div className="section-title">Network performance</div></div><div className="telemetry-live"><span className="status-dot" /> {active ? 'Boosting now' : 'Monitoring'}</div></div>
          <div className="telemetry-grid">
            <Metric icon={<Gauge />} label="Latency" value={active ? '12' : '18'} unit="ms" trend="− 32%" good />
            <Metric icon={<Activity />} label="Packet loss" value="0.00" unit="%" trend="Stable" good />
            <Metric icon={<Globe2 />} label="Route quality" value="98" unit="/100" trend="Excellent" good />
          </div>
        </section>

        <aside className="library-panel">
          <div className="library-heading"><div><div className="section-label">YOUR LIBRARY</div><div className="section-title">Choose a game</div></div><span className="game-count">{games.length} games</span></div>
          <div className="search-box"><Search size={16} /><input aria-label="Search games" placeholder="Search library" value={query} onChange={(event) => setQuery(event.target.value)} /></div>
          <div className="filters">{['All games', 'FPS', 'RPG'].map((item) => <button key={item} className={category === item ? 'filter active-filter' : 'filter'} onClick={() => setCategory(item)}>{item}</button>)}</div>
          <div className="game-list">{filteredGames.map((game, index) => <button key={game.name} className={`game-row ${selected.name === game.name ? 'selected-game' : ''}`} onClick={() => setSelected(game)}><span className="game-thumb" style={{ backgroundImage: `url(${game.image})`, backgroundPosition: `${index * 18}% ${index * 12}%` }}><span style={{ backgroundColor: game.accent }} /></span><span className="game-info"><strong>{game.name}</strong><small>{game.genre}</small></span><span className="game-ping">{game.ping}</span></button>)}</div>
          <div className="region-select"><span><Globe2 size={15} /> Routing region</span><button onClick={() => setRegion(region === 'Auto region' ? 'US West' : 'Auto region')}>{region}<ChevronDown size={14} /></button></div>
          <div className="pro-note"><ShieldCheck size={17} /><div><strong>Lagvex protection</strong><span>Secure routing is active for all games.</span></div></div>
        </aside>
      </div>
      <footer><span>Lagvex Engine <b>●</b> All systems operational</span><span>Privacy <i /> Terms <i /> Support</span></footer>
    </main>
  )
}

function Metric({ icon, label, value, unit, trend, good }: { icon: React.ReactNode; label: string; value: string; unit: string; trend: string; good?: boolean }) {
  return <div className="metric-card"><div className="metric-top"><span className="metric-icon">{icon}</span><span className={good ? 'metric-trend good' : 'metric-trend'}>{trend}</span></div><div className="metric-label">{label}</div><div className="metric-value">{value}<small>{unit}</small></div><div className="mini-chart"><span /><span /><span /><span /><span /><span /><span /></div></div>
}
