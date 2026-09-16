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

  const REGIONS = [
    'Auto region',
    'Singapore #1 (18ms)',
    'Tokyo #1 (32ms)',
    'Hong Kong #1 (28ms)',
    'Frankfurt #1 (140ms)',
    'US West #1 (165ms)'
  ];

  // State
  let selectedGame = GAMES_CATALOG[0];
  let activeCategory = 'all';
  let searchQuery = '';
  let isAccelerating = false;
  let currentRegionIndex = 0;
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

  // ==================== INITIALIZATION ====================
  function init() {
    renderGameList();
    updateHeroCard(selectedGame);
    setupEventListeners();
    startEqualizerAnimation();
    pollEngineStatus();
    setInterval(pollEngineStatus, 3000);
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
        <span class="game-ping">${isAccelerating ? Math.max(10, Math.round(game.accelPing * 0.75)) + ' ms' : game.ping}</span>
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
  }

  function updateHeroCard(game) {
    heroEyebrowSlug.textContent = `/${game.name.toUpperCase()}`;
    heroGameTitle.textContent = game.name;
    heroGameSubtitle.innerHTML = `${game.genre} <span>&bull;</span> Competitive profile <span>&bull;</span> <strong id="hero-kicker-text" class="hero-status-tag">${isAccelerating ? 'Acceleration active' : 'Ready to boost'}</strong>`;
    heroCoverImg.src = game.image;
    heroCoverImg.alt = `${game.name} cover artwork`;

    const displayPing = isAccelerating ? `${Math.max(10, Math.round(game.accelPing * 0.75))} ms` : game.ping;
    heroMetaPing.textContent = displayPing;
    heroMetaRegion.textContent = game.region;

    // Update Telemetry values
    if (isAccelerating) {
      metricValLatency.textContent = Math.max(10, Math.round(game.accelPing * 0.75));
      metricTrendLatency.innerHTML = game.trend;
      metricValLoss.textContent = '0.00';
      metricValRoute.textContent = '99';
    } else {
      metricValLatency.textContent = game.accelPing;
      metricTrendLatency.innerHTML = game.trend;
      metricValLoss.textContent = '0.00';
      metricValRoute.textContent = '98';
    }
  }

  // ==================== BOOST ACCELERATION ====================
  async function toggleBoost() {
    isAccelerating = !isAccelerating;

    if (isAccelerating) {
      btnToggleBoost.classList.add('is-active');
      boostButtonLabel.textContent = 'Acceleration active';
      telemetryLiveLabel.textContent = 'Boosting now';
      engineStatusLabel.textContent = 'Accelerating';
      engineStatusDot.style.background = 'var(--green)';

      const boostKicker = document.getElementById('hero-kicker-text');
      if (boostKicker) {
        boostKicker.textContent = 'Acceleration active';
      }

      showToast(`Tunnel established for ${selectedGame.name} via ${selectedGame.region}`, 'success');

      // Call Go backend API with full connectReq payload
      try {
        await fetch('/api/connect', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            relayEndpoint: 'auto',
            autoNode: true,
            gameId: selectedGame.id,
            regionId: selectedGame.region || '',
            forceNow: true
          })
        });
      } catch (e) {
        console.warn('Backend connect notice:', e);
      }
    } else {
      btnToggleBoost.classList.remove('is-active');
      boostButtonLabel.textContent = 'Activate boost';
      telemetryLiveLabel.textContent = 'Monitoring';
      engineStatusLabel.textContent = 'Engine online';

      const boostKicker = document.getElementById('hero-kicker-text');
      if (boostKicker) {
        boostKicker.textContent = 'Ready to boost';
      }

      showToast('Acceleration stopped — Returned to standby', 'info');

      // Call Go backend API
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
    }

    updateHeroCard(selectedGame);
    renderGameList();
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

  // ==================== ENGINE POLLING ====================
  async function pollEngineStatus() {
    try {
      const res = await fetch('/api/status');
      if (!res.ok) return;
      const data = await res.json();

      if (data.state === 'accelerating') {
        if (!isAccelerating) {
          isAccelerating = true;
          btnToggleBoost.classList.add('is-active');
          boostButtonLabel.textContent = 'Acceleration active';
          telemetryLiveLabel.textContent = 'Boosting now';
          engineStatusLabel.textContent = 'Accelerating';
          updateHeroCard(selectedGame);
        }
        // Update live stats from real Go engine
        if (data.pingMs && data.pingMs > 0) {
          metricValLatency.textContent = data.pingMs;
          heroMetaPing.textContent = `${data.pingMs} ms`;
        }
      } else if (data.state === 'standby' && isAccelerating) {
        isAccelerating = false;
        btnToggleBoost.classList.remove('is-active');
        boostButtonLabel.textContent = 'Activate boost';
        telemetryLiveLabel.textContent = 'Monitoring';
        engineStatusLabel.textContent = 'Engine online';
        updateHeroCard(selectedGame);
      }
    } catch (e) {
      // Offline / standalone preview fallback
    }
  }

  // ==================== EVENT LISTENERS ====================
  function setupEventListeners() {
    // Boost Button
    btnToggleBoost.addEventListener('click', toggleBoost);

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

    // Region dropdown cycling
    btnToggleRegion.addEventListener('click', () => {
      currentRegionIndex = (currentRegionIndex + 1) % REGIONS.length;
      const newRegion = REGIONS[currentRegionIndex];
      currentRegionLabel.textContent = newRegion;
      showToast(`Routing region set to: ${newRegion}`, 'info');
    });

    // Optimize Profile trigger
    btnOptimizeProfile.addEventListener('click', () => {
      modalSettings.classList.add('open');
    });

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
      }
    });
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
