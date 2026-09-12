let gamesList = [];
let relaysList = [];
let selectedGameId = "valorant";
let selectedRegionId = "asia-sg";
let currentFilter = "all";
let isConnected = false;

document.addEventListener("DOMContentLoaded", () => {
  initDashboard();
});

async function initDashboard() {
  setupEventListeners();
  await Promise.all([loadRelays(), loadGames()]);
  startStatusPolling();
}

function setupEventListeners() {
  // Boost Toggle Button
  document.getElementById("btn-toggle-boost").addEventListener("click", handleBoostToggle);

  // Filter Tabs
  document.querySelectorAll(".filter-tab").forEach(tab => {
    tab.addEventListener("click", (e) => {
      document.querySelectorAll(".filter-tab").forEach(t => t.classList.remove("active"));
      e.target.classList.add("active");
      currentFilter = e.target.getAttribute("data-cat");
      renderGames();
    });
  });

  // Relay Latency Tester
  document.getElementById("btn-test-relay").addEventListener("click", handleTestRelay);

  // Region Selector Change
  document.getElementById("region-select").addEventListener("change", (e) => {
    selectedRegionId = e.target.value;
  });

  // Modals
  setupModals();
}

async function loadGames() {
  try {
    const res = await fetch("/api/games");
    gamesList = await res.json();
    renderGames();
    selectGame(selectedGameId || (gamesList[0] ? gamesList[0].id : null));
  } catch (err) {
    console.error("Failed loading games:", err);
  }
}

async function loadRelays() {
  try {
    const res = await fetch("/api/relays");
    relaysList = await res.json();
    const select = document.getElementById("relay-select");
    select.innerHTML = "";

    relaysList.forEach(r => {
      const opt = document.createElement("option");
      opt.value = r.endpoint;
      opt.dataset.psk = r.psk || "";
      opt.textContent = `${r.name} (${r.location}) - ${r.endpoint}`;
      select.appendChild(opt);
    });
  } catch (err) {
    console.error("Failed loading relays:", err);
  }
}

function renderGames() {
  const grid = document.getElementById("games-grid");
  grid.innerHTML = "";

  const filtered = currentFilter === "all" 
    ? gamesList 
    : gamesList.filter(g => g.category.toLowerCase().includes(currentFilter.toLowerCase()));

  filtered.forEach(game => {
    const card = document.createElement("div");
    card.className = `game-card ${game.id === selectedGameId ? "selected" : ""}`;
    card.dataset.id = game.id;

    card.innerHTML = `
      <div class="game-top">
        <div class="game-icon-box">${game.icon || "🎮"}</div>
        <div class="game-info">
          <div class="game-name">${game.name}</div>
          <span class="game-cat-tag">${game.category}</span>
        </div>
      </div>
      <div class="game-regions-info">
        <span>🌐</span>
        <span>${game.regions ? game.regions.length : 0} Server Regions</span>
      </div>
      <div class="game-card-action">
        <button class="game-select-btn">${game.id === selectedGameId ? "Selected" : "Select Game"}</button>
      </div>
    `;

    card.addEventListener("click", () => selectGame(game.id));
    grid.appendChild(card);
  });
}

function selectGame(gameId) {
  selectedGameId = gameId;
  const game = gamesList.find(g => g.id === gameId);
  if (!game) return;

  // Update HUD Title
  document.getElementById("hud-game-title").textContent = game.name;
  document.getElementById("hud-game-sub").textContent = `Monitored: ${game.processNames.join(", ")}`;

  // Populate Regions
  const regSelect = document.getElementById("region-select");
  regSelect.innerHTML = "";
  if (game.regions && game.regions.length > 0) {
    game.regions.forEach(reg => {
      const opt = document.createElement("option");
      opt.value = reg.id;
      opt.textContent = `${reg.name} (${reg.cidrs ? reg.cidrs.length : 0} IP blocks)`;
      regSelect.appendChild(opt);
    });
    selectedRegionId = game.regions[0].id;
  }

  // Update selected card styling
  document.querySelectorAll(".game-card").forEach(c => {
    const isThis = c.dataset.id === gameId;
    c.classList.toggle("selected", isThis);
    const btn = c.querySelector(".game-select-btn");
    if (btn) btn.textContent = isThis ? "Selected" : "Select Game";
  });
}

async function handleBoostToggle() {
  const boostBtn = document.getElementById("btn-toggle-boost");

  if (isConnected) {
    // Disconnect
    try {
      await fetch("/api/disconnect", { method: "POST" });
    } catch (err) {
      console.error("Disconnect error:", err);
    }
  } else {
    // Connect
    const relaySelect = document.getElementById("relay-select");
    const endpoint = relaySelect.value;
    const psk = relaySelect.selectedOptions[0]?.dataset.psk || "";
    const forceNow = document.getElementById("force-routes-toggle").checked;

    if (!endpoint) {
      alert("Please select or add a Relay VPS server first.");
      return;
    }

    try {
      await fetch("/api/connect", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          relayEndpoint: endpoint,
          psk: psk,
          gameId: selectedGameId,
          regionId: selectedRegionId,
          forceNow: forceNow
        })
      });
    } catch (err) {
      console.error("Connect error:", err);
    }
  }
}

async function handleTestRelay() {
  const relaySelect = document.getElementById("relay-select");
  const endpoint = relaySelect.value;
  const psk = relaySelect.selectedOptions[0]?.dataset.psk || "";
  const testBtn = document.getElementById("btn-test-relay");

  if (!endpoint) return;

  testBtn.textContent = "Testing...";
  testBtn.disabled = true;

  try {
    const res = await fetch("/api/test-relay", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ endpoint, psk })
    });
    const data = await res.json();
    if (data.success) {
      testBtn.textContent = `${data.latencyMs} ms`;
    } else {
      testBtn.textContent = "Offline / Timeout";
    }
  } catch (err) {
    testBtn.textContent = "Error";
  } finally {
    testBtn.disabled = false;
    setTimeout(() => {
      testBtn.textContent = "Test Ping";
    }, 4000);
  }
}

function startStatusPolling() {
  setInterval(async () => {
    try {
      const res = await fetch("/api/status");
      const stats = await res.json();
      updateHUD(stats);
    } catch (err) {
      // Server down or reconnecting
    }
  }, 1000);
}

function updateHUD(stats) {
  const statusBadge = document.getElementById("header-status-badge");
  const headerPing = document.getElementById("header-ping-val");
  const pingDot = document.getElementById("ping-indicator");
  const metricPing = document.getElementById("metric-ping");
  const metricRates = document.getElementById("metric-rates");
  const metricRoutes = document.getElementById("metric-routes");
  const metricGameState = document.getElementById("metric-gamestate");
  const boostBtn = document.getElementById("btn-toggle-boost");
  const boostText = document.getElementById("boost-btn-text");
  const boostSub = document.getElementById("boost-btn-sub");

  isConnected = (stats.state === "connected");

  if (isConnected) {
    statusBadge.textContent = "ACCELERATING";
    statusBadge.className = "stat-val badge-connected";
    boostBtn.className = "boost-btn boost-active";
    boostText.textContent = "DISCONNECT";
    boostSub.textContent = "Tunnel Active";

    // Ping
    const ping = stats.pingMs || 0;
    headerPing.textContent = `${ping} ms`;
    metricPing.innerHTML = `${ping} <small>ms</small>`;

    pingDot.className = "ping-dot " + (ping <= 60 ? "ping-good" : (ping <= 120 ? "ping-medium" : "ping-bad"));

    // Rates
    const upKB = Math.round(stats.upRateBps / 1024);
    const downKB = Math.round(stats.downRateBps / 1024);
    metricRates.innerHTML = `${upKB} / ${downKB} <small>KB/s</small>`;

    // Routes
    metricRoutes.innerHTML = `${stats.routeCount} <small>active</small>`;

    // Game state
    if (stats.gameRunning) {
      metricGameState.innerHTML = `<span style="color: #00ff88;">🎮 In Game</span>`;
    } else {
      metricGameState.innerHTML = `<span style="color: #ffaa00;">⏳ Watching Process</span>`;
    }
  } else if (stats.state === "connecting") {
    statusBadge.textContent = "CONNECTING...";
    statusBadge.className = "stat-val badge-connecting";
    boostText.textContent = "CONNECTING";
    boostSub.textContent = "Handshaking VPS...";
  } else {
    statusBadge.textContent = "DISCONNECTED";
    statusBadge.className = "stat-val badge-idle";
    boostBtn.className = "boost-btn boost-idle";
    boostText.textContent = "BOOST NOW";
    boostSub.textContent = "Click to Activate";

    headerPing.textContent = "-- ms";
    pingDot.className = "ping-dot";
    metricPing.innerHTML = `-- <small>ms</small>`;
    metricRates.innerHTML = `0 / 0 <small>KB/s</small>`;
    metricRoutes.innerHTML = `0 <small>active</small>`;
    metricGameState.textContent = "Idle";
  }
}

function setupModals() {
  // Relay Modal
  const modalRelay = document.getElementById("modal-relay");
  document.getElementById("btn-open-custom-relay").addEventListener("click", () => {
    modalRelay.classList.add("active");
  });
  document.getElementById("close-modal-relay").addEventListener("click", () => {
    modalRelay.classList.remove("active");
  });
  document.getElementById("cancel-relay-modal").addEventListener("click", () => {
    modalRelay.classList.remove("active");
  });

  document.getElementById("save-relay-modal").addEventListener("click", () => {
    const name = document.getElementById("custom-relay-name").value.trim();
    const endpoint = document.getElementById("custom-relay-endpoint").value.trim();
    const psk = document.getElementById("custom-relay-psk").value.trim();

    if (!name || !endpoint || !psk) {
      alert("Please fill in all fields.");
      return;
    }

    const select = document.getElementById("relay-select");
    const opt = document.createElement("option");
    opt.value = endpoint;
    opt.dataset.psk = psk;
    opt.textContent = `${name} (Custom) - ${endpoint}`;
    select.appendChild(opt);
    select.value = endpoint;

    modalRelay.classList.remove("active");
    alert("Relay added successfully!");
  });

  // Squad Modal
  const modalSquad = document.getElementById("modal-squad");
  const btnOpenSquad = document.getElementById("btn-open-squad-modal");
  if (btnOpenSquad && modalSquad) {
    btnOpenSquad.addEventListener("click", () => {
      modalSquad.classList.add("active");
    });
    document.getElementById("close-modal-squad").addEventListener("click", () => {
      modalSquad.classList.remove("active");
    });
    document.getElementById("cancel-squad-modal").addEventListener("click", () => {
      modalSquad.classList.remove("active");
    });

    document.getElementById("import-squad-modal").addEventListener("click", () => {
      const raw = document.getElementById("squad-invite-input").value.trim();
      if (!raw) return;

      let endpoint = "";
      let psk = "";
      let name = "Squad Relay";

      if (raw.startsWith("lagvex://")) {
        try {
          const url = new URL(raw.replace("lagvex://", "http://dummy/"));
          endpoint = url.searchParams.get("endpoint") || "";
          psk = url.searchParams.get("psk") || "";
          if (url.searchParams.get("name")) {
            name = url.searchParams.get("name");
          }
        } catch (e) {
          console.error("Failed parsing lagvex URL:", e);
        }
      } else if (raw.includes("|")) {
        const parts = raw.split("|");
        endpoint = parts[0].trim();
        psk = parts[1].trim();
      } else if (raw.includes(" ")) {
        const parts = raw.split(/\s+/);
        endpoint = parts[0].trim();
        psk = parts[1].trim();
      }

      if (!endpoint || !psk) {
        alert("Invalid invite format. Please paste lagvex:// link or IP:Port|PSK");
        return;
      }

      const select = document.getElementById("relay-select");
      const opt = document.createElement("option");
      opt.value = endpoint;
      opt.dataset.psk = psk;
      opt.textContent = `🤝 ${name} (${endpoint})`;
      select.appendChild(opt);
      select.value = endpoint;

      modalSquad.classList.remove("active");
      alert(`🎉 Successfully joined squad relay: ${name} (${endpoint})!\nSelect your game and click BOOST NOW!`);
    });
  }

  // Game Modal
  const modalGame = document.getElementById("modal-game");
  document.getElementById("btn-open-custom-game").addEventListener("click", () => {
    modalGame.classList.add("active");
  });
  document.getElementById("close-modal-game").addEventListener("click", () => {
    modalGame.classList.remove("active");
  });
  document.getElementById("cancel-game-modal").addEventListener("click", () => {
    modalGame.classList.remove("active");
  });

  document.getElementById("save-game-modal").addEventListener("click", async () => {
    const id = document.getElementById("custom-game-id").value.trim().toLowerCase();
    const name = document.getElementById("custom-game-name").value.trim();
    const procs = document.getElementById("custom-game-procs").value.split(",").map(p => p.trim()).filter(Boolean);
    const regName = document.getElementById("custom-game-region-name").value.trim() || "Default Region";
    const cidrs = document.getElementById("custom-game-cidrs").value.split(/[\n,]+/).map(c => c.trim()).filter(Boolean);

    if (!id || !name || procs.length === 0) {
      alert("Please fill in Game ID, Name, and at least one executable.");
      return;
    }

    const newGame = {
      id,
      name,
      category: "Custom Games",
      icon: "🕹️",
      processNames: procs,
      regions: [
        {
          id: "default",
          name: regName,
          cidrs: cidrs
        }
      ]
    };

    try {
      const res = await fetch("/api/add-game", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(newGame)
      });
      if (res.ok) {
        modalGame.classList.remove("active");
        await loadGames();
        selectGame(id);
      } else {
        alert("Failed adding custom game.");
      }
    } catch (err) {
      console.error(err);
    }
  });
}
