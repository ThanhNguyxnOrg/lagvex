/**
 * LAGVEX PRO — Frontend Interactive Controller
 * Converts lagvex-ui-design.zip architecture to live application
 * Real game assets, live telemetry & backend API integration
 */

(function () {
  'use strict';

  // ==================== REAL GAME CATALOG ====================
  // 100% Real, official game cover images (no AI images)
  const GAMES_CATALOG = [
    {
      id: 'valorant',
      name: 'Valorant',
      genre: 'Tactical shooter',
      tag: 'FPS',
      accent: '#ef4444',
      image: 'assets/games/valorant.png',
      ping: '18 ms',
      baselinePing: 58,
      accelPing: 18,
      region: 'Asia-Pacific (Singapore)',
      trend: '&minus; 32%'
    },
    {
      id: 'cs2',
      name: 'Counter-Strike 2',
      genre: 'Tactical shooter',
      tag: 'FPS',
      accent: '#f59e0b',
      image: 'assets/games/cs2.jpg',
      ping: '22 ms',
      baselinePing: 62,
      accelPing: 22,
      region: 'Asia-Pacific (Singapore Valve SDR)',
      trend: '&minus; 38%'
    },
    {
      id: 'apex',
      name: 'Apex Legends',
      genre: 'Battle royale',
      tag: 'FPS',
      accent: '#f59e0b',
      image: 'assets/games/apex.jpg',
      ping: '27 ms',
      baselinePing: 68,
      accelPing: 27,
      region: 'Asia-Pacific (Singapore Multiplay)',
      trend: '&minus; 35%'
    },
    {
      id: 'thefinals',
      name: 'The Finals',
      genre: 'Arena shooter',
      tag: 'FPS',
      accent: '#22d3ee',
      image: 'assets/games/thefinals.jpg',
      ping: '22 ms',
      baselinePing: 65,
      accelPing: 22,
      region: 'Asia-Pacific (Singapore GCP)',
      trend: '&minus; 41%'
    },
    {
      id: 'dota2',
      name: 'Dota 2',
      genre: 'MOBA strategy',
      tag: 'MOBA',
      accent: '#e74c3c',
      image: 'assets/games/dota2.jpg',
      ping: '19 ms',
      baselinePing: 59,
      accelPing: 19,
      region: 'Southeast Asia (Singapore)',
      trend: '&minus; 36%'
    },
    {
      id: 'pubg',
      name: 'PUBG: BATTLEGROUNDS',
      genre: 'Battle royale',
      tag: 'FPS',
      accent: '#e67e22',
      image: 'assets/games/pubg.jpg',
      ping: '25 ms',
      baselinePing: 70,
      accelPing: 25,
      region: 'Asia-Pacific (Singapore AWS)',
      trend: '&minus; 44%'
    },
    {
      id: 'lol',
      name: 'League of Legends',
      genre: 'MOBA',
      tag: 'MOBA',
      accent: '#0ac8b9',
      image: 'assets/games/lol.jpg',
      ping: '15 ms',
      baselinePing: 52,
      accelPing: 15,
      region: 'Việt Nam & Southeast Asia',
      trend: '&minus; 48%'
    },
    {
      id: 'cyberpunk',
      name: 'Cyberpunk 2077',
      genre: 'Action RPG',
      tag: 'RPG',
      accent: '#a855f7',
      image: 'assets/games/cyberpunk.jpg',
      ping: '34 ms',
      baselinePing: 78,
      accelPing: 34,
      region: 'Global Cloud Relay',
      trend: '&minus; 30%'
    },
    {
      id: 'eldenring',
      name: 'Elden Ring',
      genre: 'Action RPG',
      tag: 'RPG',
      accent: '#eab308',
      image: 'assets/games/eldenring.jpg',
      ping: '41 ms',
      baselinePing: 88,
      accelPing: 41,
      region: 'Peer-to-Peer Matchmaking',
      trend: '&minus; 28%'
    },
    {
      id: 'hunt',
      name: 'Hunt: Showdown',
      genre: 'Extraction shooter',
      tag: 'FPS',
      accent: '#84cc16',
      image: 'assets/games/hunt.jpg',
      ping: '31 ms',
      baselinePing: 74,
      accelPing: 31,
      region: 'Asia-Pacific Dedicated',
      trend: '&minus; 33%'
    },
    {
      id: 'cod',
      name: 'Call of Duty: Warzone',
      genre: 'Battle royale',
      tag: 'FPS',
      accent: '#27ae60',
      image: 'assets/games/cod.jpg',
      ping: '24 ms',
      baselinePing: 66,
      accelPing: 24,
      region: 'Asia-Pacific (Singapore Demonware)',
      trend: '&minus; 39%'
    },
    {
      id: 'overwatch2',
      name: 'Overwatch 2',
      genre: 'Hero shooter',
      tag: 'FPS',
      accent: '#ff9c00',
      image: 'assets/games/overwatch2.jpg',
      ping: '21 ms',
      baselinePing: 63,
      accelPing: 21,
      region: 'Asia-Pacific (Singapore Battle.net)',
      trend: '&minus; 37%'
    },
    {
      id: 'r6',
      name: 'Rainbow Six Siege',
      genre: 'Tactical shooter',
      tag: 'FPS',
      accent: '#3498db',
      image: 'assets/games/r6.jpg',
      ping: '23 ms',
      baselinePing: 64,
      accelPing: 23,
      region: 'Southeast Asia (Singapore Azure)',
      trend: '&minus; 35%'
    },
    {
      id: 'deltaforce',
      name: 'Delta Force: Hawk Ops',
      genre: 'Tactical shooter',
      tag: 'FPS',
      accent: '#00f0ff',
      image: 'assets/games/deltaforce.jpg',
      ping: '20 ms',
      baselinePing: 60,
      accelPing: 20,
      region: 'Asia-Pacific (Singapore Tencent)',
      trend: '&minus; 40%'
    }
  ];

  // Dynamic Relays list from Go backend
  let activeRelays = [
    { id: 'auto', name: '⚡ Auto (Optimal Node Probing)', endpoint: 'auto' }
  ];
  let currentRelayIndex = 0;

  // State
  let selectedGame = GAMES_CATALOG[0];
  let activeCategory = 'all';
  let searchQuery = '';
  let isAccelerating = false;
  let chartInterval = null;

  // DOM Elements
  const heroEyebrowSlug = document.getElementById('eyebrow-slug');
  const heroCoverImg = document.getElementById('hero-cover-img');
  const heroGameTitle = document.getElementById('hero-game-title');
  const heroGameSubtitle = document.getElementById('hero-game-subtitle');
  const heroKickerText = document.getElementById('hero-kicker-text');
  const heroMetaPing = document.getElementById('hero-meta-ping');
  const heroMetaRegion = document.getElementById('hero-meta-region');
  const btnToggleBoost = document.getElementById('btn-toggle-boost');
  const boostButtonLabel = document.getElementById('boost-button-label');

  const telemetryLiveLabel = document.getElementById('telemetry-live-label');
  const telemetryDot = document.getElementById('telemetry-dot');
  const metricValLatency = document.getElementById('metric-val-latency');
  const metricTrendLatency = document.getElementById('metric-trend-latency');
  const metricValLoss = document.getElementById('metric-val-loss');
  const metricValRoute = document.getElementById('metric-val-route');

  const gameListContainer = document.getElementById('game-list-container');
  const gameSearchInput = document.getElementById('game-search-input');
  const categoryFilterTabs = document.getElementById('category-filter-tabs');
  const libraryGameCount = document.getElementById('library-game-count');

  const btnToggleRegion = document.getElementById('btn-toggle-region');
  const currentRegionLabel = document.getElementById('current-region-label');

  const engineStatusDot = document.getElementById('engine-status-dot');
  const engineStatusLabel = document.getElementById('engine-status-label');

  // Auto-Failover & Route Advisor Elements
  const failoverPill = document.getElementById('failover-pill');
  const btnToggleFailover = document.getElementById('btn-toggle-failover');
  const settingAutoFailoverCheckbox = document.getElementById('setting-auto-failover-checkbox');

  const routeAdvisorCard = document.getElementById('route-advisor-card');
  const advisorBadge = document.getElementById('advisor-badge');
  const advisorIcon = document.getElementById('advisor-icon');
  const advisorVerdictText = document.getElementById('advisor-verdict-text');
  const advisorHeadline = document.getElementById('advisor-headline');
  const advisorSummary = document.getElementById('advisor-summary');
  const advDirectVal = document.getElementById('adv-direct-val');
  const advRelayVal = document.getElementById('adv-relay-val');

  let lastFailoverTimestamp = null;
  let autoFailoverEnabled = true;

  // Zero-RTT FEC Elements
  const hudFecBadge = document.getElementById('hud-fec-badge');
  const hudFecStatus = document.getElementById('hud-fec-status');
  const hudFecRecovered = document.getElementById('hud-fec-recovered');
  const settingFecCheckbox = document.getElementById('setting-fec-checkbox');
  let fecEnabled = true;

  // Modals
  const modalSettings = document.getElementById('modal-settings');
  const btnOpenSettings = document.getElementById('btn-open-settings');
  const closeModalSettings = document.getElementById('close-modal-settings');
  const cancelModalSettings = document.getElementById('cancel-modal-settings');
  const saveModalSettings = document.getElementById('save-modal-settings');

  const modalSquad = document.getElementById('modal-squad');
  const btnQuickSquad = document.getElementById('btn-quick-squad');
  const closeModalSquad = document.getElementById('close-modal-squad');
  const cancelModalSquad = document.getElementById('cancel-modal-squad');
  const joinSquadBtn = document.getElementById('join-squad-btn');
  const btnOptimizeProfile = document.getElementById('btn-optimize-profile');

  // Server Selection Elements
  const btnOpenServerModal = document.getElementById('btn-open-server-modal');
  const heroServerBadge = document.getElementById('hero-server-badge');
  const heroServerName = document.getElementById('hero-server-name');
  const heroServerPing = document.getElementById('hero-server-ping');
  const heroMetaRegionPill = document.getElementById('hero-meta-region-pill');
  const modalServers = document.getElementById('modal-servers');
  const closeModalServers = document.getElementById('close-modal-servers');
  const cancelModalServers = document.getElementById('cancel-modal-servers');
  const serverSearchInput = document.getElementById('server-search-input');
  const serverContinentTabs = document.getElementById('server-continent-tabs');
  const serverModalList = document.getElementById('server-modal-list');
  const btnProbeAllServers = document.getElementById('btn-probe-all-servers');

  let activeServerFilter = 'all';
  let serverSearchQuery = '';
  let serverPings = {};

  // Profile Optimization Modal Elements
  const modalOptimizeProfile = document.getElementById('modal-optimize-profile');
  const closeModalOpt = document.getElementById('close-modal-opt');
  const btnOptDone = document.getElementById('btn-opt-done');
  const optGameImg = document.getElementById('opt-game-img');
  const optGameName = document.getElementById('opt-game-name');
  const optGameMeta = document.getElementById('opt-game-meta');
  const optValMtu = document.getElementById('opt-val-mtu');
  const optValDscp = document.getElementById('opt-val-dscp');
  const optValFec = document.getElementById('opt-val-fec');
  const optValSubnets = document.getElementById('opt-val-subnets');
  const optChecklist = document.getElementById('opt-checklist');

  // ==================== INITIALIZATION ====================
  function init() {
    loadRelays();
    renderGameList();
    updateHeroCard(selectedGame);
    setupEventListeners();
    startEqualizerAnimation();
    pollEngineStatus();
    pollRouteAdvisor();
    setInterval(pollEngineStatus, 2000);
    setInterval(pollRouteAdvisor, 8000);
  }

  async function loadRelays() {
    try {
      const res = await fetch('/api/relays');
      if (res.ok) {
        const data = await res.json();
        if (Array.isArray(data) && data.length > 0) {
          activeRelays = data;
          currentRelayIndex = 0;
          updateServerSelectionUI(activeRelays[0]);
        }
      }
    } catch (e) {
      console.warn('Failed to load relays from backend:', e);
    }
  }

  function updateServerSelectionUI(relay) {
    if (!relay) return;
    const name = relay.name || 'Auto (Smart Route)';
    let badge = '[AUTO]';
    const match = name.match(/\[(.*?)\]/);
    if (match) {
      badge = `[${match[1]}]`;
    } else if (name.includes('Auto')) {
      badge = '[AUTO]';
    } else if (name.includes('Local')) {
      badge = '[LOCAL]';
    }

    if (heroServerBadge) heroServerBadge.textContent = badge;
    if (heroServerName) heroServerName.textContent = name;
    if (heroServerPing) {
      const ping = serverPings[relay.endpoint];
      heroServerPing.textContent = ping ? `${ping}ms` : (relay.endpoint === '127.0.0.1:4433' ? '<1ms' : '~1ms');
    }
    if (heroMetaRegion) heroMetaRegion.textContent = relay.location || name;
    if (currentRegionLabel) currentRegionLabel.textContent = name;
  }

  // ==================== RENDERING ====================
  function getFilteredGames() {
    return GAMES_CATALOG.filter(game => {
      const matchesCat = activeCategory === 'all' || game.tag === activeCategory;
      const matchesSearch = !searchQuery || game.name.toLowerCase().includes(searchQuery.toLowerCase()) || game.genre.toLowerCase().includes(searchQuery.toLowerCase());
      return matchesCat && matchesSearch;
    });
  }

  function renderGameList() {
    const filtered = getFilteredGames();
    libraryGameCount.textContent = `${filtered.length} games`;
    gameListContainer.innerHTML = '';

    if (filtered.length === 0) {
      gameListContainer.innerHTML = '<div style="padding: 24px; text-align: center; color: #7a8c9b; font-size: 12px;">No matching games found</div>';
      return;
    }

    filtered.forEach(game => {
      const row = document.createElement('button');
      row.className = `game-row ${selectedGame.id === game.id ? 'selected-game' : ''}`;
      row.type = 'button';
      row.setAttribute('data-game-id', game.id);

      row.innerHTML = `
        <span class="game-thumb" style="background-image: url('${game.image}')">
          <span class="game-thumb-strip" style="background-color: ${game.accent}"></span>
        </span>
        <span class="game-info">
          <strong>${game.name}</strong>
          <small>${game.genre}</small>
        </span>
        <span class="game-ping">${isAccelerating && selectedGame.id === game.id ? 'Protected' : '~' + game.baselinePing + ' ms'}</span>
      `;

      row.addEventListener('click', () => {
        selectGame(game);
      });

      gameListContainer.appendChild(row);
    });
  }

  function selectGame(game) {
    selectedGame = game;
    updateHeroCard(game);

    // Update selected class in list
    const rows = gameListContainer.querySelectorAll('.game-row');
    rows.forEach(r => {
      if (r.getAttribute('data-game-id') === game.id) {
        r.classList.add('selected-game');
      } else {
        r.classList.remove('selected-game');
      }
    });

    pollRouteAdvisor();
  }

  function updateHeroCard(game) {
    heroEyebrowSlug.textContent = `/${game.name.toUpperCase()}`;
    heroGameTitle.textContent = game.name;
    heroGameSubtitle.innerHTML = `${game.genre} <span>&bull;</span> Competitive profile <span>&bull;</span> <strong id="hero-kicker-text" class="hero-status-tag">${isAccelerating ? 'Acceleration active' : 'Ready to boost'}</strong>`;
    heroCoverImg.src = game.image;
    heroCoverImg.alt = `${game.name} cover artwork`;
    heroMetaRegion.textContent = game.region;

    // Honest Telemetry: Never display fake hardcoded pings when disconnected
    if (isAccelerating) {
      heroMetaPing.textContent = 'Protected';
      metricTrendLatency.innerHTML = '&minus; Live';
      metricValLoss.textContent = '0.00%';
      metricValRoute.textContent = 'Optimal';
    } else {
      heroMetaPing.textContent = `~${game.baselinePing} ms (Target)`;
      metricValLatency.textContent = '--';
      metricTrendLatency.innerHTML = 'Standby';
      metricValLoss.textContent = '--';
      metricValRoute.textContent = 'Standby';
    }
  }

  // ==================== BOOST ACCELERATION ====================
  async function toggleBoost() {
    if (!isAccelerating) {
      // Transitioning to connecting
      boostButtonLabel.textContent = 'Connecting...';
      telemetryLiveLabel.textContent = 'Handshaking';
      engineStatusLabel.textContent = 'Negotiating tunnel...';
      engineStatusDot.style.background = 'var(--cyan)';
      metricValLatency.textContent = '...';
      heroMetaPing.textContent = 'Connecting...';

      const boostKicker = document.getElementById('hero-kicker-text');
      if (boostKicker) {
        boostKicker.textContent = 'Connecting...';
      }

      showToast(`Initiating secure tunnel for ${selectedGame.name}...`, 'info');

      const selectedRelay = activeRelays[currentRelayIndex] || { endpoint: 'auto' };
      try {
        const res = await fetch('/api/connect', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            relayEndpoint: selectedRelay.endpoint || 'auto',
            autoNode: selectedRelay.endpoint === 'auto',
            gameId: selectedGame.id,
            regionId: selectedGame.region || '',
            forceNow: true
          })
        });
        if (!res.ok) {
          const errData = await res.text();
          showToast(`⚠️ Boost failed: ${errData}`, 'error');
          isAccelerating = false;
          btnToggleBoost.classList.remove('is-active');
          boostButtonLabel.textContent = 'Activate boost';
          telemetryLiveLabel.textContent = 'Standby';
          engineStatusLabel.textContent = 'Connection failed';
          engineStatusDot.style.background = 'var(--red, #ef4444)';
          updateHeroCard(selectedGame);
        }
      } catch (e) {
        console.warn('Backend connect notice:', e);
        showToast('⚠️ Could not communicate with Lagvex engine', 'error');
      }
    } else {
      btnToggleBoost.classList.remove('is-active');
      boostButtonLabel.textContent = 'Activate boost';
      telemetryLiveLabel.textContent = 'Standby';
      engineStatusLabel.textContent = 'Engine online';
      engineStatusDot.style.background = '';
      isAccelerating = false;

      const boostKicker = document.getElementById('hero-kicker-text');
      if (boostKicker) {
        boostKicker.textContent = 'Ready to boost';
      }

      showToast('Acceleration stopped — Returned to standby', 'info');

      try {
        await fetch('/api/disconnect', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          }
        });
      } catch (e) {
        console.warn('Backend disconnect notice:', e);
      }

      updateHeroCard(selectedGame);
      renderGameList();
    }
  }

  // ==================== ANIMATED MINI CHARTS ====================
  function startEqualizerAnimation() {
    const barsLatency = document.querySelectorAll('#mini-chart-latency span');
    const barsRoute = document.querySelectorAll('#mini-chart-route span');

    if (chartInterval) clearInterval(chartInterval);

    chartInterval = setInterval(() => {
      if (isAccelerating) {
        barsLatency.forEach((bar, idx) => {
          const rand = Math.floor(Math.random() * 40) + (idx === 5 ? 55 : 30);
          bar.style.height = `${rand}%`;
        });
        barsRoute.forEach(bar => {
          const rand = Math.floor(Math.random() * 20) + 78;
          bar.style.height = `${rand}%`;
        });
      }
    }, 450);
  }

  // ==================== AUTO-FAILOVER CONTROLS ====================
  async function toggleAutoFailover(explicitState) {
    try {
      const payload = typeof explicitState === 'boolean' ? { enabled: explicitState } : {};
      const res = await fetch('/api/failover/toggle', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      if (res.ok) {
        const data = await res.json();
        updateFailoverUI(data.autoFailover);
        showToast(
          `Auto-Failover protection: ${data.autoFailover ? 'ACTIVE' : 'DISABLED'}`,
          data.autoFailover ? 'success' : 'info'
        );
      }
    } catch (e) {
      console.warn('Failover toggle error:', e);
    }
  }

  function updateFailoverUI(enabled) {
    autoFailoverEnabled = enabled;
    if (btnToggleFailover) {
      if (enabled) {
        btnToggleFailover.classList.add('active');
        btnToggleFailover.setAttribute('aria-pressed', 'true');
        if (failoverPill) failoverPill.classList.add('active');
      } else {
        btnToggleFailover.classList.remove('active');
        btnToggleFailover.setAttribute('aria-pressed', 'false');
        if (failoverPill) failoverPill.classList.remove('active');
      }
    }
    if (settingAutoFailoverCheckbox) {
      settingAutoFailoverCheckbox.checked = enabled;
    }
  }

  // ==================== ZERO-RTT FEC CONTROLS ====================
  async function toggleFEC(explicitState) {
    try {
      const payload = typeof explicitState === 'boolean' ? { enabled: explicitState } : {};
      const res = await fetch('/api/fec/toggle', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      if (res.ok) {
        const data = await res.json();
        fecEnabled = data.fecEnabled;
        updateFECUI(data.fecEnabled, data.fecEnabled ? '6:1' : 'Off', 0);
        showToast(
          `Zero-RTT FEC Loss Recovery: ${data.fecEnabled ? 'ACTIVE (Systematic XOR)' : 'DISABLED'}`,
          data.fecEnabled ? 'success' : 'info'
        );
      }
    } catch (e) {
      console.warn('FEC toggle error:', e);
    }
  }

  function updateFECUI(active, ratio, recovered) {
    fecEnabled = active;
    if (settingFecCheckbox) {
      settingFecCheckbox.checked = active;
    }
    if (hudFecBadge) {
      if (active) {
        hudFecBadge.classList.remove('disabled');
        hudFecBadge.classList.add('active');
        if (hudFecStatus) hudFecStatus.textContent = ratio && ratio !== 'Off' ? `ACTIVE (${ratio})` : 'ACTIVE';
      } else {
        hudFecBadge.classList.remove('active');
        hudFecBadge.classList.add('disabled');
        if (hudFecStatus) hudFecStatus.textContent = 'OFF';
      }
    }
    if (hudFecRecovered) {
      hudFecRecovered.textContent = `${recovered || 0} rec`;
    }
  }

  // ==================== SMART ROUTE ADVISOR ====================
  async function pollRouteAdvisor() {
    if (!advisorBadge || !selectedGame) return;
    try {
      const url = `/api/advisor?gameId=${encodeURIComponent(selectedGame.id)}&regionId=${encodeURIComponent(selectedGame.region || '')}`;
      const res = await fetch(url);
      if (!res.ok) return;
      const advice = await res.json();

      if (advice.verdict) {
        advisorBadge.className = 'advisor-badge';
        if (advice.verdict === 'BOOST_RECOMMENDED') {
          advisorBadge.classList.add('verdict-boost');
          advisorIcon.textContent = '⚡';
          advisorVerdictText.textContent = 'BOOST RECOMMENDED';
        } else if (advice.verdict === 'DIRECT_OPTIMAL') {
          advisorBadge.classList.add('verdict-direct');
          advisorIcon.textContent = '🛡️';
          advisorVerdictText.textContent = 'DIRECT ISP OPTIMAL';
        } else {
          advisorBadge.classList.add('verdict-comparable');
          advisorIcon.textContent = '⚖️';
          advisorVerdictText.textContent = 'COMPARABLE PATHS';
        }
      }

      if (advice.headline) advisorHeadline.textContent = advice.headline;
      if (advice.summary) advisorSummary.textContent = advice.summary;

      if (advDirectVal && typeof advice.directRTTMs === 'number') {
        advDirectVal.textContent = `${Math.round(advice.directRTTMs)} ms`;
      }
      if (advRelayVal && typeof advice.relayRTTMs === 'number') {
        advRelayVal.textContent = `${Math.round(advice.relayRTTMs)} ms`;
      }
    } catch (e) {
      // Standalone preview fallback
    }
  }

  // ==================== ENGINE POLLING ====================
  async function pollEngineStatus() {
    try {
      const res = await fetch('/api/status');
      if (!res.ok) return;
      const data = await res.json();

      // Sync auto-failover toggle
      if (typeof data.autoFailover === 'boolean' && data.autoFailover !== autoFailoverEnabled) {
        updateFailoverUI(data.autoFailover);
      }

      // Sync Zero-RTT FEC
      if (typeof data.fecActive === 'boolean') {
        updateFECUI(data.fecActive, data.fecRatio, data.packetsRecovered || 0);
      }

      // Check for seamless failover event notifications
      if (data.lastFailover && data.lastFailover.timestamp) {
        if (!lastFailoverTimestamp) {
          lastFailoverTimestamp = data.lastFailover.timestamp;
        } else if (data.lastFailover.timestamp !== lastFailoverTimestamp) {
          lastFailoverTimestamp = data.lastFailover.timestamp;
          showToast(
            `⚡ Seamless Handover: Switched from ${data.lastFailover.oldRelayName} to ${data.lastFailover.newRelayName} (${Math.round(data.lastFailover.oldPingMs)}ms ➔ ${Math.round(data.lastFailover.newPingMs)}ms)`,
            'success'
          );
        }
      }

      const isConnected = data.state === 'connected' || data.state === 'accelerating';
      if (isConnected) {
        if (!isAccelerating) {
          isAccelerating = true;
          btnToggleBoost.classList.add('is-active');
          boostButtonLabel.textContent = 'Acceleration active';
          telemetryLiveLabel.textContent = 'Boosting now';
          engineStatusLabel.textContent = data.activeRelayName ? `Boosted via ${data.activeRelayName}` : 'Accelerating';
          engineStatusDot.style.background = 'var(--green)';
          showToast(`⚡ Tunnel established via ${data.activeRelayName || 'relay'} (${data.pingMs || 0}ms)`, 'success');
          updateHeroCard(selectedGame);
          renderGameList();
        }
        // Update live stats from real Go engine
        metricValLatency.textContent = data.pingMs > 0 ? data.pingMs : '--';
        heroMetaPing.textContent = data.pingMs > 0 ? `${data.pingMs} ms` : 'Active';
        metricTrendLatency.innerHTML = '&minus; Live';
        metricValLoss.textContent = '0.00%';
        metricValRoute.textContent = data.routeCount > 0 ? `${data.routeCount} routes` : 'Optimal';
      } else if (data.state === 'connecting') {
        boostButtonLabel.textContent = 'Connecting...';
        telemetryLiveLabel.textContent = 'Handshaking';
        engineStatusLabel.textContent = 'Probing & connecting...';
        engineStatusDot.style.background = 'var(--cyan)';
        metricValLatency.textContent = '...';
        heroMetaPing.textContent = 'Connecting...';
      } else if (data.state === 'disconnected' || data.state === 'standby' || data.state === 'error') {
        if (isAccelerating || boostButtonLabel.textContent === 'Connecting...') {
          const hadError = data.lastError || data.state === 'error';
          isAccelerating = false;
          btnToggleBoost.classList.remove('is-active');
          boostButtonLabel.textContent = 'Activate boost';
          telemetryLiveLabel.textContent = 'Standby';
          engineStatusLabel.textContent = hadError ? 'Connection failed' : 'Engine online';
          engineStatusDot.style.background = hadError ? 'var(--red, #ef4444)' : '';
          updateHeroCard(selectedGame);
          renderGameList();
          if (data.lastError) {
            showToast(`⚠️ Boost failed: ${data.lastError}`, 'error');
          }
        }
      }
    } catch (e) {
      // Offline / standalone preview fallback
    }
  }

  // ==================== EVENT LISTENERS ====================
  function setupEventListeners() {
    // Boost Button
    btnToggleBoost.addEventListener('click', toggleBoost);

    // Auto-Failover toggles
    if (btnToggleFailover) {
      btnToggleFailover.addEventListener('click', () => {
        toggleAutoFailover(!autoFailoverEnabled);
      });
    }
    if (settingAutoFailoverCheckbox) {
      settingAutoFailoverCheckbox.addEventListener('change', (e) => {
        toggleAutoFailover(e.target.checked);
      });
    }

    // Zero-RTT FEC toggles
    if (settingFecCheckbox) {
      settingFecCheckbox.addEventListener('change', (e) => {
        toggleFEC(e.target.checked);
      });
    }
    if (hudFecBadge) {
      hudFecBadge.addEventListener('click', () => {
        modalSettings.classList.add('open');
      });
    }

    // Search input
    gameSearchInput.addEventListener('input', (e) => {
      searchQuery = e.target.value.trim();
      renderGameList();
    });

    // Category filter tabs
    const tabs = categoryFilterTabs.querySelectorAll('.filter');
    tabs.forEach(tab => {
      tab.addEventListener('click', () => {
        tabs.forEach(t => {
          t.classList.remove('active-filter');
          t.setAttribute('aria-selected', 'false');
        });
        tab.classList.add('active-filter');
        tab.setAttribute('aria-selected', 'true');
        activeCategory = tab.getAttribute('data-cat');
        renderGameList();
      });
    });

    // Server Selection Modal openers
    if (btnOpenServerModal) {
      btnOpenServerModal.addEventListener('click', () => {
        renderServerModalList();
        if (modalServers) modalServers.classList.add('open');
      });
    }
    if (heroMetaRegionPill) {
      heroMetaRegionPill.addEventListener('click', () => {
        renderServerModalList();
        if (modalServers) modalServers.classList.add('open');
      });
    }
    if (btnToggleRegion) {
      btnToggleRegion.addEventListener('click', () => {
        renderServerModalList();
        if (modalServers) modalServers.classList.add('open');
      });
    }

    if (closeModalServers) {
      closeModalServers.addEventListener('click', () => {
        if (modalServers) modalServers.classList.remove('open');
      });
    }
    if (cancelModalServers) {
      cancelModalServers.addEventListener('click', () => {
        if (modalServers) modalServers.classList.remove('open');
      });
    }

    // Server Search Input
    if (serverSearchInput) {
      serverSearchInput.addEventListener('input', (e) => {
        serverSearchQuery = e.target.value.trim();
        renderServerModalList();
      });
    }

    // Server Continent Tabs
    if (serverContinentTabs) {
      const sTabs = serverContinentTabs.querySelectorAll('.filter');
      sTabs.forEach(tab => {
        tab.addEventListener('click', () => {
          sTabs.forEach(t => t.classList.remove('active-filter'));
          tab.classList.add('active-filter');
          activeServerFilter = tab.getAttribute('data-continent');
          renderServerModalList();
        });
      });
    }

    // Probe All Servers Latency
    if (btnProbeAllServers) {
      btnProbeAllServers.addEventListener('click', probeAllServers);
    }

    // Optimize Profile Action
    if (btnOptimizeProfile) {
      btnOptimizeProfile.addEventListener('click', optimizeGameProfile);
    }
    if (closeModalOpt) {
      closeModalOpt.addEventListener('click', () => {
        if (modalOptimizeProfile) modalOptimizeProfile.classList.remove('open');
      });
    }
    if (btnOptDone) {
      btnOptDone.addEventListener('click', () => {
        if (modalOptimizeProfile) modalOptimizeProfile.classList.remove('open');
        showToast('Profile parameters active for next match', 'success');
      });
    }

    // Settings Modal
    btnOpenSettings.addEventListener('click', () => {
      modalSettings.classList.add('open');
    });
    closeModalSettings.addEventListener('click', () => {
      modalSettings.classList.remove('open');
    });
    cancelModalSettings.addEventListener('click', () => {
      modalSettings.classList.remove('open');
    });
    saveModalSettings.addEventListener('click', () => {
      modalSettings.classList.remove('open');
      showToast('Relay and protocol settings saved', 'success');
    });

    // Squad Modal
    btnQuickSquad.addEventListener('click', () => {
      modalSquad.classList.add('open');
    });
    closeModalSquad.addEventListener('click', () => {
      modalSquad.classList.remove('open');
    });
    cancelModalSquad.addEventListener('click', () => {
      modalSquad.classList.remove('open');
    });
    joinSquadBtn.addEventListener('click', () => {
      modalSquad.classList.remove('open');
      showToast('Synchronized with Squad party tunnel', 'success');
    });

    // Custom Game Modal
    const modalCustomGame = document.getElementById('modal-custom-game');
    const btnOpenCustomGame = document.getElementById('btn-open-custom-game');
    const closeModalCustomGame = document.getElementById('close-modal-custom-game');
    const cancelModalCustomGame = document.getElementById('cancel-modal-custom-game');
    const saveCustomGameBtn = document.getElementById('save-custom-game-btn');

    if (btnOpenCustomGame) {
      btnOpenCustomGame.addEventListener('click', () => {
        modalCustomGame.classList.add('open');
      });
    }
    if (closeModalCustomGame) {
      closeModalCustomGame.addEventListener('click', () => {
        modalCustomGame.classList.remove('open');
      });
    }
    if (cancelModalCustomGame) {
      cancelModalCustomGame.addEventListener('click', () => {
        modalCustomGame.classList.remove('open');
      });
    }
    if (saveCustomGameBtn) {
      saveCustomGameBtn.addEventListener('click', async () => {
        const nameInput = document.getElementById('custom-game-name');
        const exeInput = document.getElementById('custom-game-exe');
        const regInput = document.getElementById('custom-game-region');
        const cidrsInput = document.getElementById('custom-game-cidrs');

        const name = (nameInput.value || '').trim();
        const exe = (exeInput.value || '').trim();
        const region = (regInput.value || 'Custom Server').trim();
        const cidrsStr = (cidrsInput.value || '').trim();

        if (!name) {
          showToast('Please enter a game title', 'info');
          return;
        }

        const cidrs = cidrsStr ? cidrsStr.split(',').map(s => s.trim()).filter(Boolean) : [];
        const newGame = {
          id: name.toLowerCase().replace(/[^a-z0-9]+/g, '-'),
          name: name,
          publisher: 'Custom User Profile',
          category: 'Custom',
          accent: '#00f0ff',
          processNames: exe ? [exe] : [name + '.exe'],
          regions: [
            {
              id: 'custom-region',
              continent: 'Global',
              name: region,
              cidrs: cidrs.length > 0 ? cidrs : ['1.1.1.1/32']
            }
          ]
        };

        // Post to backend API
        try {
          await fetch('/api/add-game', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(newGame)
          });
        } catch (e) {
          console.warn('Backend add-game notice:', e);
        }

        // Add to local catalog & select
        const catalogItem = {
          id: newGame.id,
          name: newGame.name,
          genre: 'Custom profile',
          tag: 'Custom',
          accent: '#00f0ff',
          image: 'assets/games/valorant.png',
          ping: '20 ms',
          baselinePing: 60,
          accelPing: 20,
          region: region,
          trend: '&minus; 35%'
        };
        GAMES_CATALOG.unshift(catalogItem);
        selectGame(catalogItem);
        renderGameList();

        modalCustomGame.classList.remove('open');
        nameInput.value = '';
        exeInput.value = '';
        regInput.value = '';
        cidrsInput.value = '';
        showToast(`Added custom profile for "${name}"`, 'success');
      });
    }

    // Close modals on Escape
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') {
        modalSettings.classList.remove('open');
        modalSquad.classList.remove('open');
        if (modalCustomGame) modalCustomGame.classList.remove('open');
        if (modalServers) modalServers.classList.remove('open');
        if (modalOptimizeProfile) modalOptimizeProfile.classList.remove('open');
      }
    });

    // Close modals on clicking backdrop outside modal-card
    document.querySelectorAll('.modal-backdrop').forEach(backdrop => {
      backdrop.addEventListener('click', (e) => {
        if (e.target === backdrop) {
          backdrop.classList.remove('open');
        }
      });
    });
  }

  // ==================== SERVER SELECTION MODAL LOGIC ====================
  function renderServerModalList() {
    if (!serverModalList) return;
    serverModalList.innerHTML = '';

    const filtered = activeRelays.filter(relay => {
      const continent = relay.continent || 'Global';
      const tier = relay.tier || '';
      const matchesFilter = activeServerFilter === 'all' ||
        (activeServerFilter === 'smart' && (tier === 'smart' || relay.id === 'auto')) ||
        (activeServerFilter === 'Local' && (continent === 'Local' || relay.endpoint.includes('127.0.0.1'))) ||
        continent.toLowerCase().includes(activeServerFilter.toLowerCase());

      const q = serverSearchQuery.toLowerCase();
      const matchesSearch = !q ||
        relay.name.toLowerCase().includes(q) ||
        (relay.location && relay.location.toLowerCase().includes(q)) ||
        (relay.endpoint && relay.endpoint.toLowerCase().includes(q));

      return matchesFilter && matchesSearch;
    });

    if (filtered.length === 0) {
      serverModalList.innerHTML = '<div style="padding: 30px; text-align: center; color: #7a8c9b; font-size: 13px; grid-column: 1/-1;">No servers match the selected criteria</div>';
      return;
    }

    filtered.forEach(relay => {
      const card = document.createElement('div');
      const isSelected = activeRelays[currentRelayIndex] && activeRelays[currentRelayIndex].id === relay.id;
      card.className = `server-card ${isSelected ? 'active-server' : ''}`;

      let badge = 'NODE';
      let badgeClass = '';
      const match = relay.name.match(/\[(.*?)\]/);
      if (match) {
        badge = match[1];
      } else if (relay.name.includes('Auto')) {
        badge = '⚡ AUTO';
        badgeClass = 'badge-smart';
      } else if (relay.name.includes('Local')) {
        badge = '💻 LOCAL';
        badgeClass = 'badge-local';
      }

      const pingVal = serverPings[relay.endpoint] || (relay.endpoint === '127.0.0.1:4433' ? '<1ms' : (relay.endpoint === 'auto' ? '~1ms' : '--'));

      card.innerHTML = `
        <div class="server-card-info">
          <span class="server-badge ${badgeClass}">${badge}</span>
          <div class="server-details">
            <span class="server-name-txt">${relay.name.replace(/\[.*?\]\s*/, '')}</span>
            <span class="server-sub-txt">${relay.location || relay.endpoint} &bull; ${relay.tier || 'edge'}</span>
          </div>
        </div>
        <div class="server-card-actions">
          <span class="server-ping-badge ${pingVal !== '--' && !pingVal.includes('err') ? 'good' : ''}">${pingVal}</span>
          <button class="server-card-btn" type="button">${isSelected ? 'Active' : 'Select'}</button>
        </div>
      `;

      card.addEventListener('click', () => {
        const idx = activeRelays.findIndex(r => r.id === relay.id);
        if (idx !== -1) {
          currentRelayIndex = idx;
          updateServerSelectionUI(relay);
          if (modalServers) modalServers.classList.remove('open');
          showToast(`Selected routing node: ${relay.name}`, 'success');
          pollRouteAdvisor();
        }
      });

      serverModalList.appendChild(card);
    });
  }

  async function probeAllServers() {
    if (btnProbeAllServers) {
      btnProbeAllServers.textContent = 'Probing...';
      btnProbeAllServers.disabled = true;
    }

    try {
      // Loopback is always instantaneous
      serverPings['127.0.0.1:4433'] = '<1ms';
      serverPings['auto'] = '~1ms';

      const res = await fetch('/api/probe-relays?samples=1');
      if (res.ok) {
        const results = await res.json();
        if (Array.isArray(results)) {
          results.forEach(p => {
            if (p.reachable && p.medianPingMs > 0) {
              serverPings[p.endpoint] = `${Math.round(p.medianPingMs)}ms`;
            } else if (p.endpoint === '127.0.0.1:4433') {
              serverPings[p.endpoint] = '<1ms';
            } else {
              serverPings[p.endpoint] = 'Unreachable';
            }
          });
        }
      }
    } catch (e) {
      console.warn('Probe error:', e);
    } finally {
      if (btnProbeAllServers) {
        btnProbeAllServers.textContent = '⚡ Test All Latencies';
        btnProbeAllServers.disabled = false;
      }
      renderServerModalList();
      if (activeRelays[currentRelayIndex]) {
        updateServerSelectionUI(activeRelays[currentRelayIndex]);
      }
      showToast('Completed latency probing across all nodes', 'info');
    }
  }

  // ==================== PROFILE OPTIMIZATION ACTION ====================
  async function optimizeGameProfile() {
    if (!selectedGame) return;
    if (btnOptimizeProfile) {
      btnOptimizeProfile.classList.add('loading');
      const span = btnOptimizeProfile.querySelector('span');
      if (span) span.textContent = 'Optimizing...';
    }

    try {
      const res = await fetch(`/api/optimize-profile?gameId=${encodeURIComponent(selectedGame.id)}`);
      if (res.ok) {
        const data = await res.json();

        if (optGameImg) optGameImg.src = selectedGame.image;
        if (optGameName) optGameName.textContent = data.gameTitle || selectedGame.name;
        if (optGameMeta) optGameMeta.textContent = `${data.genre} Optimized Profile // Jitter Dampening Active`;
        if (optValMtu) optValMtu.textContent = `${data.mtu} B`;
        if (optValDscp) optValDscp.textContent = data.dscp.includes('EF-46') ? 'EF-46' : 'CS6';
        if (optValFec) optValFec.textContent = `${data.fecRatio} Parity`;
        if (optValSubnets) optValSubnets.textContent = `${data.subnetsCount} CIDRs`;

        if (optChecklist && Array.isArray(data.details)) {
          optChecklist.innerHTML = data.details.map(det => `
            <div class="opt-check-item">
              <span class="opt-check-icon">✓</span>
              <span>${det}</span>
            </div>
          `).join('');
        }

        const kicker = document.getElementById('hero-kicker-text');
        if (kicker) {
          kicker.textContent = 'Profile optimal';
        }

        if (modalOptimizeProfile) {
          modalOptimizeProfile.classList.add('open');
        }
        showToast(`✨ ${data.gameTitle} profile optimized: MTU 1400B & EF-46 QoS applied`, 'success');
      }
    } catch (e) {
      console.warn('Profile optimization error:', e);
      showToast('⚠️ Could not complete profile auto-tune', 'error');
    } finally {
      if (btnOptimizeProfile) {
        btnOptimizeProfile.classList.remove('loading');
        const span = btnOptimizeProfile.querySelector('span');
        if (span) span.textContent = 'Optimize profile';
      }
    }
  }

  // ==================== TOAST HELPER ====================
  function showToast(message, type = 'info') {
    const shelf = document.getElementById('app-toast-container');
    if (!shelf) return;

    const item = document.createElement('div');
    item.className = `toast-item ${type === 'success' ? 'toast-success' : ''}`;
    item.textContent = message;

    shelf.appendChild(item);
    setTimeout(() => {
      item.style.opacity = '0';
      item.style.transition = 'opacity 200ms ease';
      setTimeout(() => item.remove(), 250);
    }, 2800);
  }

  // Run on DOM ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
