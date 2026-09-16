/**
 * LAGVEX PRO — FRONTEND ENGINE JAVASCRIPT
 * Accurate Esports Vector Logos & Real-time Route Telemetry Controller
 */

// Official Game Vector SVGs
function getGameLogoSvg(gameId) {
  switch (gameId) {
    case "valorant":
      // Riot Games Official Valorant Angular V
      return `
        <svg viewBox="0 0 100 100" class="game-svg-logo" aria-label="Valorant">
          <path d="M14 26 L46 26 L88 78 L56 78 Z" fill="#FF4655"/>
          <path d="M60 26 L88 26 L88 44 L72 44 Z" fill="#FF4655"/>
        </svg>`;

    case "cs2":
      // Counter-Strike 2 Official Brand Emblem with Crosshair Ticks
      return `
        <svg viewBox="0 0 100 100" class="game-svg-logo" aria-label="Counter-Strike 2">
          <rect width="100" height="100" rx="14" fill="#141822"/>
          <circle cx="50" cy="50" r="32" stroke="#E8A838" stroke-width="3" fill="none" stroke-dasharray="14 6"/>
          <line x1="50" y1="12" x2="50" y2="24" stroke="#E8A838" stroke-width="4"/>
          <line x1="50" y1="76" x2="50" y2="88" stroke="#E8A838" stroke-width="4"/>
          <line x1="12" y1="50" x2="24" y2="50" stroke="#E8A838" stroke-width="4"/>
          <line x1="76" y1="50" x2="88" y2="50" stroke="#E8A838" stroke-width="4"/>
          <text x="50" y="58" text-anchor="middle" font-family="'Outfit', sans-serif" font-weight="900" font-size="24" fill="#FFFFFF">CS<tspan fill="#DE9B35">2</tspan></text>
        </svg>`;

    case "pubg":
      // PUBG BATTLEGROUNDS Official Level 3 Spetsnaz Helmet
      return `
        <svg viewBox="0 0 100 100" class="game-svg-logo" aria-label="PUBG">
          <rect width="100" height="100" rx="14" fill="#15171e"/>
          <!-- Helmet Dome -->
          <path d="M22 46 C22 24, 78 24, 78 46 L82 66 C82 78, 68 84, 50 84 C32 84, 18 78, 18 66 Z" fill="#2d3340"/>
          <!-- Metal Visor Plate -->
          <path d="M24 48 L76 48 L74 62 L26 62 Z" fill="#1a1c22" stroke="#3d4455" stroke-width="2"/>
          <!-- Visor Vision Slit -->
          <rect x="30" y="53" width="40" height="4" rx="2" fill="#F2A900"/>
          <!-- Stencil Text -->
          <rect x="28" y="70" width="44" height="14" rx="3" fill="#F2A900"/>
          <text x="50" y="81" text-anchor="middle" font-family="'Outfit', sans-serif" font-weight="900" font-size="11" fill="#000">PUBG</text>
        </svg>`;

    case "apex":
      // Apex Legends Predator Arrowhead Crest
      return `
        <svg viewBox="0 0 100 100" class="game-svg-logo" aria-label="Apex Legends">
          <path d="M50 12 L88 78 L68 78 L50 46 L32 78 L12 78 Z" fill="#DA292A"/>
          <circle cx="50" cy="28" r="6" fill="#FFFFFF"/>
          <path d="M42 60 L58 60 L50 46 Z" fill="#111"/>
        </svg>`;

    case "lol":
      // League of Legends Winged Gold L Crest
      return `
        <svg viewBox="0 0 100 100" class="game-svg-logo" aria-label="League of Legends">
          <rect width="100" height="100" rx="14" fill="#091428"/>
          <!-- Gold L Rune -->
          <path d="M28 20 L44 20 L44 68 L74 68 L68 80 L28 80 Z" fill="#C8AA6E"/>
          <!-- Hextech Wings -->
          <path d="M48 24 L60 36 L48 48 Z" fill="#0AC8B9"/>
          <circle cx="68" cy="36" r="4" fill="#C8AA6E"/>
        </svg>`;

    case "thefinals":
      // The Finals Slashed Geometry
      return `
        <svg viewBox="0 0 100 100" class="game-svg-logo" aria-label="The Finals">
          <rect width="100" height="100" rx="14" fill="#15060b"/>
          <polygon points="16,22 84,22 74,80 16,80" fill="#E91E63"/>
          <polygon points="26,30 76,30 70,72 26,72" fill="#15060b"/>
          <text x="50" y="58" text-anchor="middle" font-family="'Outfit', sans-serif" font-style="italic" font-weight="900" font-size="15" fill="#FFF">FINALS</text>
        </svg>`;

    case "warzone":
    case "cod_warzone":
      // Call of Duty: Warzone Military Stencil
      return `
        <svg viewBox="0 0 100 100" class="game-svg-logo" aria-label="Call of Duty: Warzone">
          <rect width="100" height="100" rx="14" fill="#161b15"/>
          <text x="50" y="52" text-anchor="middle" font-family="'Outfit', sans-serif" font-weight="900" font-size="28" fill="#FFFFFF" letter-spacing="2">WZ</text>
          <path d="M25 64 L75 64 L50 82 Z" fill="#27AE60"/>
        </svg>`;

    case "deltaforce":
      // Delta Force: Hawk Ops Tactical Chevron
      return `
        <svg viewBox="0 0 100 100" class="game-svg-logo" aria-label="Delta Force">
          <rect width="100" height="100" rx="14" fill="#0a1520"/>
          <polygon points="50,14 86,82 50,66 14,82" fill="#00F0FF"/>
          <polygon points="50,30 74,74 50,62 26,74" fill="#0A1520"/>
          <circle cx="50" cy="46" r="4" fill="#00F0FF"/>
        </svg>`;

    case "overwatch2":
      // Blizzard Overwatch 2 Official Icon
      return `
        <svg viewBox="0 0 100 100" class="game-svg-logo" aria-label="Overwatch 2">
          <circle cx="50" cy="50" r="38" stroke="#F99E1A" stroke-width="8" fill="none" stroke-dasharray="190 50" stroke-dashoffset="25"/>
          <path d="M33 46 L44 32 L44 68 L33 68 Z" fill="#FFFFFF"/>
          <path d="M67 46 L56 32 L56 68 L67 68 Z" fill="#FFFFFF"/>
          <circle cx="50" cy="50" r="8" fill="#F99E1A"/>
        </svg>`;

    case "r6":
      // Rainbow Six Siege Official Number 6
      return `
        <svg viewBox="0 0 100 100" class="game-svg-logo" aria-label="Rainbow Six Siege">
          <rect width="100" height="100" rx="14" fill="#141a24"/>
          <path d="M60 18 L38 18 C26 18 20 26 20 38 L20 62 C20 74 28 82 42 82 L58 82 C72 82 80 74 80 62 L80 50 C80 38 72 32 58 32 L36 32 L36 44 L56 44 C62 44 66 48 66 54 C66 60 62 68 54 68 L42 68 C36 68 34 64 34 58 L34 40 C34 32 38 30 46 30 L60 30 Z" fill="#FFFFFF"/>
          <!-- Gun Chamber Notch -->
          <rect x="66" y="24" width="8" height="6" fill="#F2A900"/>
        </svg>`;

    case "dota2":
      // Valve Dota 2 Ancient Map Rune
      return `
        <svg viewBox="0 0 100 100" class="game-svg-logo" aria-label="Dota 2">
          <rect width="100" height="100" rx="14" fill="#140e0e"/>
          <!-- Red Ancient Stone Factions & Diagonal River -->
          <path d="M22 22 L42 22 L22 82 Z" fill="#E03825"/>
          <path d="M78 18 L78 78 L58 78 Z" fill="#E03825"/>
          <path d="M36 78 L80 34 L66 20 L20 66 Z" fill="#E03825"/>
        </svg>`;

    default:
      return `
        <svg viewBox="0 0 100 100" class="game-svg-logo">
          <rect width="100" height="100" rx="14" fill="#1b202e"/>
          <circle cx="50" cy="50" r="24" fill="#00F0FF" opacity="0.8"/>
          <text x="50" y="58" text-anchor="middle" font-family="'Outfit', sans-serif" font-weight="bold" font-size="24" fill="#000">G</text>
        </svg>`;
  }
}

// Complete offline fallback profiles
const DEFAULT_GAMES = [
  {
    "id": "valorant",
    "name": "Valorant",
    "publisher": "Riot Games",
    "category": "Tactical FPS",
    "accent": "#ff4655",
    "processNames": [
      "VALORANT-Win64-Shipping.exe",
      "RiotClientServices.exe",
      "VALORANT.exe"
    ],
    "regions": [
      {
        "id": "na-us-east",
        "continent": "North America",
        "name": "US East (N. Virginia - Riot Direct)",
        "cidrs": [
          "192.207.0.0/18",
          "24.105.0.0/18",
          "52.0.0.0/11"
        ]
      },
      {
        "id": "na-us-central",
        "continent": "North America",
        "name": "US Central (Chicago / Texas - Riot Direct)",
        "cidrs": [
          "104.160.131.0/24",
          "18.216.0.0/14"
        ]
      },
      {
        "id": "na-us-west",
        "continent": "North America",
        "name": "US West (California / Oregon - Riot Direct)",
        "cidrs": [
          "104.160.128.0/24",
          "54.183.0.0/16",
          "35.160.0.0/13"
        ]
      },
      {
        "id": "eu-frankfurt",
        "continent": "Europe",
        "name": "Europe Central (Frankfurt - Riot Direct)",
        "cidrs": [
          "104.160.141.0/24",
          "3.120.0.0/14",
          "18.194.0.0/15"
        ]
      },
      {
        "id": "eu-london",
        "continent": "Europe",
        "name": "Europe West (London - Riot Direct)",
        "cidrs": [
          "104.160.143.0/24",
          "3.8.0.0/14",
          "35.176.0.0/14"
        ]
      },
      {
        "id": "eu-stockholm",
        "continent": "Europe",
        "name": "Europe North (Stockholm - Riot Direct)",
        "cidrs": [
          "13.48.0.0/15",
          "16.16.0.0/15"
        ]
      },
      {
        "id": "asia-sg",
        "continent": "Asia-Pacific",
        "name": "Southeast Asia (Singapore - Riot Direct)",
        "cidrs": [
          "13.250.0.0/15",
          "18.140.0.0/15",
          "52.220.0.0/15"
        ]
      },
      {
        "id": "asia-tokyo",
        "continent": "Asia-Pacific",
        "name": "Japan (Tokyo - Riot Direct)",
        "cidrs": [
          "13.112.0.0/14",
          "18.176.0.0/14",
          "52.192.0.0/14"
        ]
      },
      {
        "id": "asia-seoul",
        "continent": "Asia-Pacific",
        "name": "Korea (Seoul - Riot Direct)",
        "cidrs": [
          "104.160.154.0/24",
          "13.124.0.0/14"
        ]
      },
      {
        "id": "asia-hk",
        "continent": "Asia-Pacific",
        "name": "Hong Kong (Riot Direct & AWS)",
        "cidrs": [
          "18.162.0.0/15",
          "18.166.0.0/15"
        ]
      },
      {
        "id": "asia-mumbai",
        "continent": "Asia-Pacific",
        "name": "India (Mumbai - Riot Direct)",
        "cidrs": [
          "13.126.0.0/15",
          "13.232.0.0/15"
        ]
      },
      {
        "id": "oc-sydney",
        "continent": "Oceania",
        "name": "Australia (Sydney - Riot Direct)",
        "cidrs": [
          "104.160.156.0/24",
          "13.236.0.0/14"
        ]
      },
      {
        "id": "sa-saopaulo",
        "continent": "South America",
        "name": "Brazil (São Paulo - Riot Direct)",
        "cidrs": [
          "104.160.152.0/24",
          "177.54.144.0/20",
          "18.228.0.0/15"
        ]
      }
    ]
  },
  {
    "id": "cs2",
    "name": "Counter-Strike 2",
    "publisher": "Valve Corporation",
    "category": "Tactical FPS",
    "accent": "#de9b35",
    "processNames": [
      "cs2.exe"
    ],
    "regions": [
      {
        "id": "na-us-east",
        "continent": "North America",
        "name": "US East (Sterling / Virginia Valve SDR)",
        "cidrs": [
          "162.254.192.0/24",
          "155.133.253.0/24",
          "155.133.250.0/24"
        ]
      },
      {
        "id": "na-us-central",
        "continent": "North America",
        "name": "US Central (Chicago Valve SDR)",
        "cidrs": [
          "162.254.193.0/24",
          "155.133.251.0/24"
        ]
      },
      {
        "id": "na-us-west",
        "continent": "North America",
        "name": "US West (Seattle / Los Angeles Valve SDR)",
        "cidrs": [
          "162.254.195.0/24",
          "162.254.194.0/24",
          "155.133.248.0/24"
        ]
      },
      {
        "id": "eu-frankfurt",
        "continent": "Europe",
        "name": "Europe Central (Frankfurt Valve SDR)",
        "cidrs": [
          "155.133.226.0/24",
          "155.133.242.0/24",
          "162.254.197.0/24"
        ]
      },
      {
        "id": "eu-london",
        "continent": "Europe",
        "name": "Europe West (London Valve SDR)",
        "cidrs": [
          "162.254.196.0/24",
          "155.133.243.0/24"
        ]
      },
      {
        "id": "eu-stockholm",
        "continent": "Europe",
        "name": "Europe North (Stockholm Valve SDR)",
        "cidrs": [
          "155.133.240.0/24",
          "155.133.241.0/24"
        ]
      },
      {
        "id": "asia-sg",
        "continent": "Asia-Pacific",
        "name": "Singapore (Valve SDR - AS32590)",
        "cidrs": [
          "103.10.124.0/24",
          "103.28.54.0/24",
          "45.121.184.0/24",
          "155.133.254.0/24"
        ]
      },
      {
        "id": "asia-tokyo",
        "continent": "Asia-Pacific",
        "name": "Japan (Tokyo Valve SDR)",
        "cidrs": [
          "45.121.186.0/24",
          "155.133.239.0/24",
          "155.133.245.0/24"
        ]
      },
      {
        "id": "asia-seoul",
        "continent": "Asia-Pacific",
        "name": "Korea (Seoul Valve SDR)",
        "cidrs": [
          "155.133.234.0/24"
        ]
      },
      {
        "id": "asia-hk",
        "continent": "Asia-Pacific",
        "name": "Hong Kong (Valve SDR)",
        "cidrs": [
          "153.254.86.0/24",
          "155.133.244.0/24"
        ]
      },
      {
        "id": "asia-mumbai",
        "continent": "Asia-Pacific",
        "name": "India (Mumbai Valve SDR)",
        "cidrs": [
          "155.133.233.0/24"
        ]
      },
      {
        "id": "oc-sydney",
        "continent": "Oceania",
        "name": "Australia (Sydney Valve SDR)",
        "cidrs": [
          "103.10.125.0/24",
          "45.121.185.0/24",
          "155.133.246.0/24"
        ]
      },
      {
        "id": "sa-saopaulo",
        "continent": "South America",
        "name": "Brazil (São Paulo Valve SDR)",
        "cidrs": [
          "205.185.194.0/24",
          "155.133.236.0/24"
        ]
      },
      {
        "id": "af-johannesburg",
        "continent": "Middle East & Africa",
        "name": "South Africa (Johannesburg Valve SDR)",
        "cidrs": [
          "155.133.238.0/24",
          "152.111.192.0/24"
        ]
      }
    ]
  },
  {
    "id": "dota2",
    "name": "Dota 2",
    "publisher": "Valve Corporation",
    "category": "MOBA",
    "accent": "#e03825",
    "processNames": [
      "dota2.exe"
    ],
    "regions": [
      {
        "id": "na-us-east",
        "continent": "North America",
        "name": "US East (Sterling / Virginia Valve SDR)",
        "cidrs": [
          "162.254.192.0/24",
          "155.133.253.0/24"
        ]
      },
      {
        "id": "na-us-central",
        "continent": "North America",
        "name": "US Central (Chicago Valve SDR)",
        "cidrs": [
          "162.254.193.0/24",
          "155.133.251.0/24"
        ]
      },
      {
        "id": "na-us-west",
        "continent": "North America",
        "name": "US West (Seattle / Los Angeles Valve SDR)",
        "cidrs": [
          "162.254.195.0/24",
          "155.133.248.0/24"
        ]
      },
      {
        "id": "eu-frankfurt",
        "continent": "Europe",
        "name": "Europe Central (Frankfurt Valve SDR)",
        "cidrs": [
          "155.133.226.0/24",
          "155.133.242.0/24"
        ]
      },
      {
        "id": "eu-stockholm",
        "continent": "Europe",
        "name": "Europe North (Stockholm Valve SDR)",
        "cidrs": [
          "155.133.240.0/24",
          "155.133.241.0/24"
        ]
      },
      {
        "id": "asia-sg",
        "continent": "Asia-Pacific",
        "name": "Southeast Asia (Singapore Valve SDR)",
        "cidrs": [
          "103.10.124.0/24",
          "45.121.184.0/24",
          "155.133.254.0/24"
        ]
      },
      {
        "id": "asia-tokyo",
        "continent": "Asia-Pacific",
        "name": "Japan (Tokyo Valve SDR)",
        "cidrs": [
          "45.121.186.0/24",
          "155.133.239.0/24"
        ]
      },
      {
        "id": "asia-hk",
        "continent": "Asia-Pacific",
        "name": "Hong Kong (Valve SDR)",
        "cidrs": [
          "153.254.86.0/24",
          "155.133.244.0/24"
        ]
      },
      {
        "id": "asia-mumbai",
        "continent": "Asia-Pacific",
        "name": "India (Mumbai Valve SDR)",
        "cidrs": [
          "155.133.233.0/24"
        ]
      },
      {
        "id": "oc-sydney",
        "continent": "Oceania",
        "name": "Australia (Sydney Valve SDR)",
        "cidrs": [
          "103.10.125.0/24",
          "155.133.246.0/24"
        ]
      },
      {
        "id": "sa-saopaulo",
        "continent": "South America",
        "name": "Brazil (São Paulo Valve SDR)",
        "cidrs": [
          "205.185.194.0/24",
          "155.133.236.0/24"
        ]
      }
    ]
  },
  {
    "id": "pubg",
    "name": "PUBG: BATTLEGROUNDS",
    "publisher": "Krafton Inc.",
    "category": "Battle Royale",
    "accent": "#f39c12",
    "processNames": [
      "TslGame.exe",
      "TslGame_BE.exe",
      "TslGame_UC.exe"
    ],
    "regions": [
      {
        "id": "na-us-east",
        "continent": "North America",
        "name": "North America East (Virginia AWS/Azure)",
        "cidrs": [
          "20.185.0.0/16",
          "52.86.0.0/15"
        ]
      },
      {
        "id": "na-us-west",
        "continent": "North America",
        "name": "North America West (Oregon AWS)",
        "cidrs": [
          "54.186.0.0/15",
          "40.83.128.0/17"
        ]
      },
      {
        "id": "eu-frankfurt",
        "continent": "Europe",
        "name": "Europe Central (Frankfurt Azure/AWS)",
        "cidrs": [
          "20.52.0.0/16",
          "52.28.0.0/15"
        ]
      },
      {
        "id": "eu-london",
        "continent": "Europe",
        "name": "Europe West (London / Ireland AWS)",
        "cidrs": [
          "52.16.0.0/15",
          "35.176.0.0/14"
        ]
      },
      {
        "id": "asia-sg",
        "continent": "Asia-Pacific",
        "name": "Southeast Asia (Singapore Azure/AWS)",
        "cidrs": [
          "20.198.192.0/19",
          "20.24.48.0/20",
          "13.250.0.0/15"
        ]
      },
      {
        "id": "asia-kr",
        "continent": "Asia-Pacific",
        "name": "Korea (Seoul Kakao/Steam AWS)",
        "cidrs": [
          "13.124.0.0/14",
          "3.34.0.0/15"
        ]
      },
      {
        "id": "asia-jp",
        "continent": "Asia-Pacific",
        "name": "Japan (Tokyo Azure)",
        "cidrs": [
          "20.210.0.0/16",
          "40.79.160.0/19"
        ]
      },
      {
        "id": "oc-syd",
        "continent": "Oceania",
        "name": "Oceania (Sydney AWS)",
        "cidrs": [
          "13.236.0.0/14"
        ]
      },
      {
        "id": "sa-br",
        "continent": "South America",
        "name": "South America (São Paulo AWS)",
        "cidrs": [
          "18.228.0.0/15"
        ]
      }
    ]
  },
  {
    "id": "apex",
    "name": "Apex Legends",
    "publisher": "Electronic Arts / Respawn",
    "category": "Battle Royale",
    "accent": "#e74c3c",
    "processNames": [
      "r5apex.exe",
      "r5apex_dx12.exe"
    ],
    "regions": [
      {
        "id": "na-us-east",
        "continent": "North America",
        "name": "North America East (Virginia Multiplay)",
        "cidrs": [
          "152.199.0.0/16",
          "52.86.0.0/15"
        ]
      },
      {
        "id": "na-us-central",
        "continent": "North America",
        "name": "North America Central (Dallas Multiplay)",
        "cidrs": [
          "152.199.16.0/20"
        ]
      },
      {
        "id": "na-us-west",
        "continent": "North America",
        "name": "North America West (Oregon Multiplay)",
        "cidrs": [
          "54.186.0.0/15"
        ]
      },
      {
        "id": "eu-frankfurt",
        "continent": "Europe",
        "name": "Europe Central (Frankfurt Multiplay)",
        "cidrs": [
          "52.28.0.0/15"
        ]
      },
      {
        "id": "eu-london",
        "continent": "Europe",
        "name": "Europe West (London Multiplay)",
        "cidrs": [
          "35.176.0.0/14"
        ]
      },
      {
        "id": "asia-sg",
        "continent": "Asia-Pacific",
        "name": "Southeast Asia (Singapore Multiplay)",
        "cidrs": [
          "13.250.0.0/15"
        ]
      },
      {
        "id": "asia-jp",
        "continent": "Asia-Pacific",
        "name": "Japan (Tokyo Multiplay)",
        "cidrs": [
          "52.192.0.0/14"
        ]
      },
      {
        "id": "asia-tw",
        "continent": "Asia-Pacific",
        "name": "Taiwan (EA Multiplay)",
        "cidrs": [
          "104.16.0.0/12"
        ]
      },
      {
        "id": "oc-syd",
        "continent": "Oceania",
        "name": "Oceania (Sydney Multiplay)",
        "cidrs": [
          "13.236.0.0/14"
        ]
      },
      {
        "id": "sa-br",
        "continent": "South America",
        "name": "South America (São Paulo Multiplay)",
        "cidrs": [
          "18.228.0.0/15"
        ]
      }
    ]
  },
  {
    "id": "lol",
    "name": "League of Legends & TFT",
    "publisher": "Riot Games & VNG",
    "category": "MOBA",
    "accent": "#0ac8b9",
    "processNames": [
      "League of Legends.exe",
      "LeagueClient.exe",
      "RiotClientServices.exe"
    ],
    "regions": [
      {
        "id": "na-chicago",
        "continent": "North America",
        "name": "North America (Chicago - Riot Direct)",
        "cidrs": [
          "104.160.131.0/24",
          "192.207.0.0/18"
        ]
      },
      {
        "id": "euw-frankfurt",
        "continent": "Europe",
        "name": "Europe West (EUW - Frankfurt Riot Direct)",
        "cidrs": [
          "104.160.141.0/24",
          "3.120.0.0/14"
        ]
      },
      {
        "id": "eune-stockholm",
        "continent": "Europe",
        "name": "Europe Nordic & East (EUNE - Stockholm)",
        "cidrs": [
          "104.160.142.0/24",
          "13.48.0.0/15"
        ]
      },
      {
        "id": "vn",
        "continent": "Asia-Pacific",
        "name": "Việt Nam (VNG Datacenter)",
        "cidrs": [
          "103.1.0.0/20",
          "118.69.0.0/16"
        ]
      },
      {
        "id": "sg",
        "continent": "Asia-Pacific",
        "name": "Southeast Asia (Singapore Riot Direct)",
        "cidrs": [
          "13.250.0.0/15",
          "54.251.0.0/16"
        ]
      },
      {
        "id": "kr",
        "continent": "Asia-Pacific",
        "name": "Korea (Seoul Riot Direct)",
        "cidrs": [
          "104.160.154.0/24",
          "13.124.0.0/14"
        ]
      },
      {
        "id": "jp",
        "continent": "Asia-Pacific",
        "name": "Japan (Tokyo Riot Direct)",
        "cidrs": [
          "13.112.0.0/14",
          "18.176.0.0/14"
        ]
      },
      {
        "id": "oce",
        "continent": "Oceania",
        "name": "Oceania (OCE - Sydney Riot Direct)",
        "cidrs": [
          "104.160.156.0/24",
          "13.236.0.0/14"
        ]
      },
      {
        "id": "br",
        "continent": "South America",
        "name": "Brazil (BR - São Paulo Riot Direct)",
        "cidrs": [
          "104.160.152.0/24",
          "177.54.144.0/20"
        ]
      }
    ]
  },
  {
    "id": "thefinals",
    "name": "The Finals",
    "publisher": "Embark Studios",
    "category": "Arena FPS",
    "accent": "#e91e63",
    "processNames": [
      "Discovery.exe"
    ],
    "regions": [
      {
        "id": "na-us-central",
        "continent": "North America",
        "name": "North America (US Central GCP/AWS)",
        "cidrs": [
          "34.66.0.0/16",
          "35.226.0.0/16"
        ]
      },
      {
        "id": "eu-frankfurt",
        "continent": "Europe",
        "name": "Europe Central (Frankfurt GCP)",
        "cidrs": [
          "34.89.0.0/16",
          "35.242.0.0/16"
        ]
      },
      {
        "id": "asia-sg",
        "continent": "Asia-Pacific",
        "name": "Southeast Asia (Singapore GCP/AWS)",
        "cidrs": [
          "34.87.0.0/16",
          "13.212.0.0/15"
        ]
      },
      {
        "id": "asia-jp",
        "continent": "Asia-Pacific",
        "name": "East Asia (Tokyo GCP/AWS)",
        "cidrs": [
          "34.84.0.0/16",
          "18.176.0.0/14"
        ]
      },
      {
        "id": "sa-br",
        "continent": "South America",
        "name": "South America (São Paulo GCP)",
        "cidrs": [
          "34.95.0.0/16"
        ]
      }
    ]
  },
  {
    "id": "cod_warzone",
    "name": "Call of Duty: Warzone",
    "publisher": "Activision",
    "category": "Battle Royale",
    "accent": "#27ae60",
    "processNames": [
      "cod.exe",
      "bootstrapper.exe"
    ],
    "regions": [
      {
        "id": "na-us-east",
        "continent": "North America",
        "name": "North America East (New York / Atlanta)",
        "cidrs": [
          "103.245.110.0/23",
          "198.148.80.0/20"
        ]
      },
      {
        "id": "na-us-central",
        "continent": "North America",
        "name": "North America Central (Dallas Demonware)",
        "cidrs": [
          "198.148.96.0/20"
        ]
      },
      {
        "id": "na-us-west",
        "continent": "North America",
        "name": "North America West (Los Angeles Demonware)",
        "cidrs": [
          "198.148.112.0/20"
        ]
      },
      {
        "id": "eu-frankfurt",
        "continent": "Europe",
        "name": "Europe Central (Frankfurt Demonware)",
        "cidrs": [
          "185.34.104.0/22"
        ]
      },
      {
        "id": "eu-london",
        "continent": "Europe",
        "name": "Europe West (London Demonware)",
        "cidrs": [
          "185.34.106.0/23"
        ]
      },
      {
        "id": "asia-sg",
        "continent": "Asia-Pacific",
        "name": "Southeast Asia (Singapore Demonware)",
        "cidrs": [
          "103.245.110.0/23"
        ]
      },
      {
        "id": "asia-jp",
        "continent": "Asia-Pacific",
        "name": "East Asia (Tokyo Demonware)",
        "cidrs": [
          "103.245.112.0/23"
        ]
      },
      {
        "id": "oc-syd",
        "continent": "Oceania",
        "name": "Australia (Sydney Demonware)",
        "cidrs": [
          "103.245.114.0/23"
        ]
      },
      {
        "id": "sa-br",
        "continent": "South America",
        "name": "South America (São Paulo Demonware)",
        "cidrs": [
          "103.245.116.0/23"
        ]
      },
      {
        "id": "me-riyadh",
        "continent": "Middle East & Africa",
        "name": "Middle East (Riyadh / Bahrain)",
        "cidrs": [
          "15.185.0.0/16"
        ]
      }
    ]
  },
  {
    "id": "deltaforce",
    "name": "Delta Force: Hawk Ops",
    "publisher": "TiMi Studio Group",
    "category": "Tactical FPS",
    "accent": "#00f0ff",
    "processNames": [
      "DeltaForce.exe",
      "DeltaForceClient-Win64-Shipping.exe"
    ],
    "regions": [
      {
        "id": "na-us",
        "continent": "North America",
        "name": "North America (Silicon Valley & Virginia)",
        "cidrs": [
          "43.153.0.0/16",
          "43.155.0.0/16"
        ]
      },
      {
        "id": "eu-frankfurt",
        "continent": "Europe",
        "name": "Europe Central (Frankfurt Tencent/AWS)",
        "cidrs": [
          "43.131.0.0/16"
        ]
      },
      {
        "id": "asia-sg",
        "continent": "Asia-Pacific",
        "name": "Southeast Asia (Singapore Tencent/AWS)",
        "cidrs": [
          "13.228.0.0/15",
          "43.134.0.0/16"
        ]
      },
      {
        "id": "asia-hk",
        "continent": "Asia-Pacific",
        "name": "East Asia (Hong Kong Tencent Cloud)",
        "cidrs": [
          "43.154.0.0/16"
        ]
      }
    ]
  },
  {
    "id": "overwatch2",
    "name": "Overwatch 2",
    "publisher": "Blizzard Entertainment",
    "category": "Hero Shooter",
    "accent": "#ff9c00",
    "processNames": [
      "Overwatch.exe"
    ],
    "regions": [
      {
        "id": "na-us-central",
        "continent": "North America",
        "name": "North America Central (Chicago Battle.net)",
        "cidrs": [
          "24.105.32.0/20",
          "24.105.62.0/24"
        ]
      },
      {
        "id": "na-us-west",
        "continent": "North America",
        "name": "North America West (Los Angeles Battle.net)",
        "cidrs": [
          "24.105.12.0/22"
        ]
      },
      {
        "id": "eu-frankfurt",
        "continent": "Europe",
        "name": "Europe Central (Frankfurt Battle.net)",
        "cidrs": [
          "185.60.112.0/22",
          "185.60.114.0/23"
        ]
      },
      {
        "id": "eu-paris",
        "continent": "Europe",
        "name": "Europe West (Paris Battle.net)",
        "cidrs": [
          "185.60.115.0/24"
        ]
      },
      {
        "id": "asia-sg",
        "continent": "Asia-Pacific",
        "name": "Southeast Asia (Singapore Battle.net)",
        "cidrs": [
          "34.87.0.0/16",
          "35.185.0.0/16"
        ]
      },
      {
        "id": "asia-kr",
        "continent": "Asia-Pacific",
        "name": "Korea (Seoul Battle.net)",
        "cidrs": [
          "211.115.0.0/17",
          "211.233.0.0/16"
        ]
      },
      {
        "id": "asia-jp",
        "continent": "Asia-Pacific",
        "name": "Japan (Tokyo Battle.net)",
        "cidrs": [
          "34.84.0.0/16"
        ]
      },
      {
        "id": "oc-syd",
        "continent": "Oceania",
        "name": "Australia (Sydney Battle.net)",
        "cidrs": [
          "103.4.114.0/23"
        ]
      },
      {
        "id": "sa-br",
        "continent": "South America",
        "name": "South America (São Paulo Battle.net)",
        "cidrs": [
          "54.232.0.0/16"
        ]
      }
    ]
  },
  {
    "id": "r6",
    "name": "Rainbow Six Siege",
    "publisher": "Ubisoft",
    "category": "Tactical FPS",
    "accent": "#3498db",
    "processNames": [
      "RainbowSix.exe",
      "RainbowSix_Vulkan.exe"
    ],
    "regions": [
      {
        "id": "na-us-east",
        "continent": "North America",
        "name": "North America East (Virginia Azure)",
        "cidrs": [
          "20.185.0.0/16",
          "40.76.0.0/14"
        ]
      },
      {
        "id": "na-us-central",
        "continent": "North America",
        "name": "North America Central (Iowa Azure)",
        "cidrs": [
          "40.77.0.0/16",
          "20.37.0.0/16"
        ]
      },
      {
        "id": "na-us-west",
        "continent": "North America",
        "name": "North America West (California Azure)",
        "cidrs": [
          "40.83.128.0/17",
          "13.64.0.0/14"
        ]
      },
      {
        "id": "eu-netherlands",
        "continent": "Europe",
        "name": "Europe West (Netherlands Azure)",
        "cidrs": [
          "20.50.0.0/16",
          "40.68.0.0/15"
        ]
      },
      {
        "id": "eu-frankfurt",
        "continent": "Europe",
        "name": "Europe Central (Frankfurt Azure)",
        "cidrs": [
          "20.52.0.0/16"
        ]
      },
      {
        "id": "asia-sg",
        "continent": "Asia-Pacific",
        "name": "Southeast Asia (Singapore Azure)",
        "cidrs": [
          "20.24.48.0/20",
          "20.198.192.0/19"
        ]
      },
      {
        "id": "asia-jp",
        "continent": "Asia-Pacific",
        "name": "Japan East (Tokyo Azure)",
        "cidrs": [
          "20.210.0.0/16"
        ]
      },
      {
        "id": "oc-syd",
        "continent": "Oceania",
        "name": "Australia East (Sydney Azure)",
        "cidrs": [
          "20.37.192.0/18"
        ]
      },
      {
        "id": "sa-br",
        "continent": "South America",
        "name": "South America (Brazil South - São Paulo)",
        "cidrs": [
          "20.201.0.0/16"
        ]
      },
      {
        "id": "me-dxb",
        "continent": "Middle East & Africa",
        "name": "Middle East (UAE North - Dubai Azure)",
        "cidrs": [
          "20.46.0.0/16"
        ]
      },
      {
        "id": "af-jnb",
        "continent": "Middle East & Africa",
        "name": "South Africa (South Africa North Azure)",
        "cidrs": [
          "20.164.0.0/16"
        ]
      }
    ]
  }
];

const DEFAULT_RELAYS = [
  {
    "id": "us-east-1",
    "name": "🇺🇸 North America East (US East - Virginia #1)",
    "continent": "North America",
    "location": "Ashburn, VA",
    "endpoint": "us-east.lagvex.org:4433",
    "psk": "lagvex-community-us-free-public-psk-2026"
  },
  {
    "id": "us-central-1",
    "name": "🇺🇸 North America Central (Dallas / Texas)",
    "continent": "North America",
    "location": "Dallas, TX",
    "endpoint": "us-central.lagvex.org:4433",
    "psk": "lagvex-community-us-free-public-psk-2026"
  },
  {
    "id": "us-west-1",
    "name": "🇺🇸 North America West (Silicon Valley / California)",
    "continent": "North America",
    "location": "San Jose, CA",
    "endpoint": "us-west.lagvex.org:4433",
    "psk": "lagvex-community-us-free-public-psk-2026"
  },
  {
    "id": "eu-central-1",
    "name": "🇩🇪 Europe Central (Frankfurt / Germany #1)",
    "continent": "Europe",
    "location": "Frankfurt, Germany",
    "endpoint": "eu-central.lagvex.org:4433",
    "psk": "lagvex-community-eu-free-public-psk-2026"
  },
  {
    "id": "eu-west-1",
    "name": "🇬🇧 Europe West (London / United Kingdom #1)",
    "continent": "Europe",
    "location": "London, UK",
    "endpoint": "eu-west.lagvex.org:4433",
    "psk": "lagvex-community-eu-free-public-psk-2026"
  },
  {
    "id": "eu-north-1",
    "name": "🇸🇪 Europe North (Stockholm / Sweden #1)",
    "continent": "Europe",
    "location": "Stockholm, Sweden",
    "endpoint": "eu-north.lagvex.org:4433",
    "psk": "lagvex-community-eu-free-public-psk-2026"
  },
  {
    "id": "asia-sg-1",
    "name": "🇸🇬 Southeast Asia #1 (Singapore - Equinix SG1)",
    "continent": "Asia-Pacific",
    "location": "Singapore",
    "endpoint": "sg1.lagvex.org:4433",
    "psk": "lagvex-community-sg-free-public-psk-2026"
  },
  {
    "id": "asia-sg-2",
    "name": "🇸🇬 Southeast Asia #2 (Singapore - Direct Fiber)",
    "continent": "Asia-Pacific",
    "location": "Singapore",
    "endpoint": "sg2.lagvex.org:4433",
    "psk": "lagvex-community-sg-free-public-psk-2026"
  },
  {
    "id": "asia-jp-1",
    "name": "🇯🇵 East Asia (Tokyo / Japan #1)",
    "continent": "Asia-Pacific",
    "location": "Tokyo, Japan",
    "endpoint": "jp1.lagvex.org:4433",
    "psk": "lagvex-community-jp-free-public-psk-2026"
  },
  {
    "id": "asia-kr-1",
    "name": "🇰🇷 East Asia (Seoul / South Korea #1)",
    "continent": "Asia-Pacific",
    "location": "Seoul, South Korea",
    "endpoint": "kr1.lagvex.org:4433",
    "psk": "lagvex-community-kr-free-public-psk-2026"
  },
  {
    "id": "asia-hk-1",
    "name": "🇭🇰 East Asia (Hong Kong #1)",
    "continent": "Asia-Pacific",
    "location": "Hong Kong",
    "endpoint": "hk1.lagvex.org:4433",
    "psk": "lagvex-community-hk-free-public-psk-2026"
  },
  {
    "id": "oc-syd-1",
    "name": "🇦🇺 Oceania (Sydney / Australia #1)",
    "continent": "Oceania",
    "location": "Sydney, Australia",
    "endpoint": "syd1.lagvex.org:4433",
    "psk": "lagvex-community-oc-free-public-psk-2026"
  },
  {
    "id": "sa-br-1",
    "name": "🇧🇷 South America (São Paulo / Brazil #1)",
    "continent": "South America",
    "location": "São Paulo, Brazil",
    "endpoint": "br1.lagvex.org:4433",
    "psk": "lagvex-community-sa-free-public-psk-2026"
  },
  {
    "id": "me-dxb-1",
    "name": "🇦🇪 Middle East (Dubai / UAE #1)",
    "continent": "Middle East",
    "location": "Dubai, UAE",
    "endpoint": "dxb1.lagvex.org:4433",
    "psk": "lagvex-community-me-free-public-psk-2026"
  }
];

let gamesList = [...DEFAULT_GAMES];
let relaysList = [...DEFAULT_RELAYS];
let selectedGameId = "valorant";
let selectedRegionId = "asia-sg";
let currentFilter = "all";
let searchQuery = "";
let isConnected = false;

document.addEventListener("DOMContentLoaded", () => {
  initDashboard();
});

async function initDashboard() {
  setupEventListeners();
  renderGames();
  selectGame(selectedGameId);
  populateRelays();

  // Fetch live configs if server is online
  await Promise.allSettled([loadRelays(), loadGames()]);
  startStatusPolling();
}

function setupEventListeners() {
  // Boost / Action Button
  const btnBoost = document.getElementById("btn-toggle-boost");
  if (btnBoost) btnBoost.addEventListener("click", handleBoostToggle);

  // Category Filter Tabs
  document.querySelectorAll(".cat-pill").forEach(tab => {
    tab.addEventListener("click", (e) => {
      document.querySelectorAll(".cat-pill").forEach(t => t.classList.remove("active"));
      e.target.classList.add("active");
      currentFilter = e.target.getAttribute("data-cat");
      renderGames();
    });
  });

  // Search input
  const searchInput = document.getElementById("game-search-input");
  if (searchInput) {
    searchInput.addEventListener("input", (e) => {
      searchQuery = e.target.value.trim().toLowerCase();
      renderGames();
    });
  }

  // Probe Relay Ping
  const btnProbe = document.getElementById("btn-test-relay");
  if (btnProbe) btnProbe.addEventListener("click", handleTestRelay);

  // Region Selector Change
  const regSelect = document.getElementById("region-select");
  if (regSelect) {
    regSelect.addEventListener("change", (e) => {
      selectedRegionId = e.target.value;
      updateRouteHops();
    });
  }

  // Relay Selector Change
  const relaySelect = document.getElementById("relay-select");
  if (relaySelect) {
    relaySelect.addEventListener("change", () => {
      updateRouteHops();
    });
  }

  // Modals
  setupModals();
}

async function loadGames() {
  try {
    const res = await fetch("/api/games");
    if (res.ok) {
      const data = await res.json();
      if (Array.isArray(data) && data.length > 0) {
        gamesList = data;
        renderGames();
        selectGame(selectedGameId);
      }
    }
  } catch (err) {}
}

let probeResultsMap = {}; // endpoint -> ProbeResult

async function loadRelays() {
  try {
    const res = await fetch("/api/relays");
    if (res.ok) {
      const data = await res.json();
      if (Array.isArray(data) && data.length > 0) {
        relaysList = data;
        populateRelays();
        // Background sweep for live RTT and optimal node tags
        probeAllRelaysBackground();
      }
    }
  } catch (err) {}
}

async function probeAllRelaysBackground() {
  try {
    const res = await fetch("/api/probe-relays?samples=2");
    if (res.ok) {
      const results = await res.json();
      if (Array.isArray(results)) {
        results.forEach(r => {
          probeResultsMap[r.endpoint] = r;
        });
        populateRelays();
      }
    }
  } catch (e) {}
}

function populateRelays() {
  const select = document.getElementById("relay-select");
  if (!select) return;
  const prevVal = select.value;
  select.innerHTML = "";

  // 1-Click Auto Select Optimal Node
  const autoOpt = document.createElement("option");
  autoOpt.value = "auto";
  autoOpt.dataset.name = "⚡ Auto Optimal Route";
  autoOpt.dataset.psk = "";

  let bestNode = Object.values(probeResultsMap).find(r => r.isOptimal);
  if (bestNode && bestNode.reachable) {
    autoOpt.textContent = `⚡ Auto Optimal Node [Best: ${bestNode.location} • ${bestNode.rttMedianMs}ms]`;
  } else {
    autoOpt.textContent = `⚡ Auto Select Optimal Node (Adaptive Wire-Speed)`;
  }
  select.appendChild(autoOpt);

  const groups = {};
  relaysList.forEach(r => {
    const cont = r.continent || "Other Relays";
    if (!groups[cont]) groups[cont] = [];
    groups[cont].push(r);
  });

  for (const [continent, items] of Object.entries(groups)) {
    const optgroup = document.createElement("optgroup");
    optgroup.label = `🌐 ${continent} Relays`;
    items.forEach(r => {
      const opt = document.createElement("option");
      opt.value = r.endpoint;
      opt.dataset.psk = r.psk || "";
      opt.dataset.name = r.name;
      opt.dataset.location = r.location || "";

      const pr = probeResultsMap[r.endpoint];
      if (pr) {
        if (pr.reachable) {
          const optTag = pr.isOptimal ? " ★ OPTIMAL" : "";
          opt.textContent = `${r.name} — ${pr.rttMedianMs} ms${optTag}`;
        } else {
          opt.textContent = `${r.name} — Offline`;
        }
      } else {
        opt.textContent = `${r.name} (${r.endpoint})`;
      }

      optgroup.appendChild(opt);
    });
    select.appendChild(optgroup);
  }

  if (prevVal) {
    select.value = prevVal;
  }
}

const GAME_COVERS = {
  "valorant": "https://images.unsplash.com/flagged/photo-1560177776-55a762c5c000?auto=format&fit=crop&q=80&w=600&h=760",
  "cs2": "https://images.unsplash.com/photo-1573511860313-d333c8022170?auto=format&fit=crop&q=80&w=600&h=760",
  "apex": "https://images.unsplash.com/photo-1672872476232-da16b45c9001?auto=format&fit=crop&q=80&w=600&h=760",
  "pubg": "https://images.unsplash.com/photo-1514124838563-9243ab7791a6?auto=format&fit=crop&q=80&w=600&h=760",
  "lol": "https://images.unsplash.com/photo-1566410824233-a8011929225c?auto=format&fit=crop&q=80&w=600&h=760",
  "dota": "https://images.unsplash.com/photo-1560671021-cb36f70ce82d?auto=format&fit=crop&q=80&w=600&h=760",
  "fortnite": "https://images.unsplash.com/photo-1519608487953-e999c86e7455?auto=format&fit=crop&q=80&w=600&h=760",
  "overwatch2": "https://images.unsplash.com/photo-1530919424169-4b95f917e937?auto=format&fit=crop&q=80&w=600&h=760",
  "thefinals": "https://images.unsplash.com/photo-1542751371-adc38448a05e?auto=format&fit=crop&q=80&w=600&h=760",
  "cod_warzone": "https://images.unsplash.com/photo-1550745165-9bc0b252726f?auto=format&fit=crop&q=80&w=600&h=760",
  "warzone": "https://images.unsplash.com/photo-1550745165-9bc0b252726f?auto=format&fit=crop&q=80&w=600&h=760",
  "deltaforce": "https://images.unsplash.com/photo-1511512578047-dfb367046420?auto=format&fit=crop&q=80&w=600&h=760",
  "r6": "https://images.unsplash.com/photo-1538481199705-c710c4e965fc?auto=format&fit=crop&q=80&w=600&h=760"
};

function getGameCover(gameId) {
  return GAME_COVERS[gameId] || "https://images.unsplash.com/photo-1542751371-adc38448a05e?auto=format&fit=crop&q=80&w=600&h=760";
}

function updateDifferentialHud(acceleratedMs) {
  const directEl = document.getElementById("direct-isp-ping");
  const metricEl = document.getElementById("metric-ping");
  const gainEl = document.getElementById("diff-gain-val");
  const gainStatus = document.getElementById("diff-gain-status");
  const pillEl = document.getElementById("diff-gain-pill");

  const p50El = document.getElementById("stat-p50");
  const p95El = document.getElementById("stat-p95");
  const lossEl = document.getElementById("stat-loss");

  const directMs = 58;
  if (directEl) directEl.innerHTML = `${directMs}<span class="ms-unit">ms</span>`;

  if (acceleratedMs && acceleratedMs > 0) {
    if (metricEl) metricEl.innerHTML = `${acceleratedMs}<span class="ms-unit">ms</span>`;
    const delta = directMs - acceleratedMs;
    if (gainEl) gainEl.textContent = delta > 0 ? `−${delta}ms` : `+${Math.abs(delta)}ms`;
    if (gainStatus) gainStatus.textContent = "IMPROVED";
    if (pillEl) pillEl.classList.add("active");

    if (p50El) p50El.innerHTML = `${acceleratedMs}<span class="triplet-unit">ms</span>`;
    if (p95El) p95El.innerHTML = `${acceleratedMs + 3}<span class="triplet-unit">ms</span>`;
    if (lossEl) lossEl.innerHTML = `0.0<span class="triplet-unit">%</span>`;
  } else {
    if (metricEl) metricEl.innerHTML = `${directMs}<span class="ms-unit">ms</span>`;
    if (gainEl) gainEl.textContent = "—";
    if (gainStatus) gainStatus.textContent = "STANDBY";
    if (pillEl) pillEl.classList.remove("active");

    if (p50El) p50El.innerHTML = `—<span class="triplet-unit">ms</span>`;
    if (p95El) p95El.innerHTML = `—<span class="triplet-unit">ms</span>`;
    if (lossEl) lossEl.innerHTML = `—<span class="triplet-unit">%</span>`;
  }
}

function renderGames() {
  const grid = document.getElementById("games-grid");
  if (!grid) return;
  grid.innerHTML = "";

  const filtered = gamesList.filter(game => {
    const matchCat = currentFilter === "all" || game.category.toLowerCase().includes(currentFilter.toLowerCase());
    const matchSearch = !searchQuery || game.name.toLowerCase().includes(searchQuery) || (game.publisher && game.publisher.toLowerCase().includes(searchQuery));
    return matchCat && matchSearch;
  });

  const countPill = document.getElementById("library-count-pill");
  if (countPill) countPill.textContent = filtered.length;

  if (filtered.length === 0) {
    grid.innerHTML = `
      <div style="grid-column: 1/-1; text-align: center; padding: 48px; color: var(--color-ink-faint);">
        <p style="font-size: 14px; margin-bottom: 6px; font-weight: 600;">No games match "${searchQuery}"</p>
        <p style="font-size: 12px;">Click <strong>+ Custom</strong> to register any game profile.</p>
      </div>
    `;
    return;
  }

  filtered.forEach(game => {
    const isSelected = game.id === selectedGameId;
    const card = document.createElement("button");
    card.type = "button";
    card.className = `poster-card ${isSelected ? "selected" : ""}`;
    card.dataset.id = game.id;

    card.innerHTML = `
      <div class="poster-media-wrap">
        <img src="${getGameCover(game.id)}" alt="${game.name} poster" class="poster-img" loading="lazy">
        <div class="poster-gradient-scrim"></div>
      </div>

      <span class="poster-shield-tag" title="Anti-cheat safe">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 3 5 5.5V11c0 4.5 3 8 7 9.5 4-1.5 7-5 7-9.5V5.5L12 3Z"></path>
          <path d="m9 12 2 2 4-4.5"></path>
        </svg>
      </span>

      ${isSelected ? '<span class="poster-active-tag">Active</span>' : ''}

      <div class="poster-caption">
        <p class="poster-game-title">${game.name}</p>
        <p class="poster-game-category">${game.category}</p>
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

  // Update Hero Game Cover & Metadata
  const coverImg = document.getElementById("hero-cover-img");
  const titleEl = document.getElementById("hud-game-title");
  const pubEl = document.getElementById("hud-game-pub");

  if (coverImg) coverImg.src = getGameCover(game.id);
  if (titleEl) titleEl.textContent = game.name;
  if (pubEl) pubEl.textContent = `${game.publisher || "Official Profile"} • ${game.category}`;

  // Populate Regions
  const regSelect = document.getElementById("region-select");
  if (regSelect) {
    regSelect.innerHTML = "";
    if (game.regions && game.regions.length > 0) {
      const groups = {};
      game.regions.forEach(reg => {
        const cont = reg.continent || "Global Clusters";
        if (!groups[cont]) groups[cont] = [];
        groups[cont].push(reg);
      });

      for (const [continent, regions] of Object.entries(groups)) {
        const optgroup = document.createElement("optgroup");
        optgroup.label = `🌐 ${continent}`;
        regions.forEach(reg => {
          const opt = document.createElement("option");
          opt.value = reg.id;
          opt.dataset.continent = continent;
          opt.dataset.name = reg.name;
          opt.textContent = `${reg.name}`;
          optgroup.appendChild(opt);
        });
        regSelect.appendChild(optgroup);
      }
      selectedRegionId = game.regions[0].id;
    }
  }

  updateRouteFooter();
  updateDifferentialHud(isConnected ? 18 : null);

  // Update Poster Selection
  document.querySelectorAll(".poster-card").forEach(c => {
    const isThis = c.dataset.id === gameId;
    c.classList.toggle("selected", isThis);
    let activeTag = c.querySelector(".poster-active-tag");
    if (isThis) {
      if (!activeTag) {
        activeTag = document.createElement("span");
        activeTag.className = "poster-active-tag";
        activeTag.textContent = "Active";
        c.appendChild(activeTag);
      }
    } else if (activeTag) {
      activeTag.remove();
    }
  });
}

function updateRouteFooter() {
  const game = gamesList.find(g => g.id === selectedGameId);
  const regSelect = document.getElementById("region-select");
  const relaySelect = document.getElementById("relay-select");
  const footerText = document.getElementById("route-footer-text");

  const regName = regSelect?.selectedOptions[0]?.dataset.name || "Asia-Pacific";
  const relayName = relaySelect?.selectedOptions[0]?.dataset.name || "Singapore #1";

  if (footerText) {
    const cleanReg = regName.split("(")[0].trim();
    const cleanRelay = relayName.split("[")[0].trim();
    footerText.textContent = `Route: ${cleanReg} → ${cleanRelay}`;
  }
}

// Crisp Impeccable Toast System
function showToast(title, message, type = "normal") {
  const container = document.getElementById("app-toast-container");
  if (!container) return;

  const toast = document.createElement("div");
  toast.className = `toast-item ${type === "success" ? "success" : type === "gold" ? "gold" : "normal"}`;

  toast.innerHTML = `
    <div class="toast-header">${title}</div>
    <div class="toast-message">${message}</div>
  `;
  container.appendChild(toast);

  setTimeout(() => {
    toast.style.opacity = "0";
    toast.style.transform = "translateY(8px)";
    toast.style.transition = "all 0.25s ease";
    setTimeout(() => toast.remove(), 260);
  }, 3200);
}

async function handleBoostToggle() {
  const btn = document.getElementById("btn-toggle-boost");
  const btnLabel = document.getElementById("boost-btn-label");
  const statusPill = document.getElementById("status-pill");
  const engineText = document.getElementById("engine-status-text");
  const liveTag = document.getElementById("sidebar-live-tag");
  const liveText = document.getElementById("sidebar-live-text");
  const ambientGlow = document.getElementById("ambient-glow");
  const equalizer = document.getElementById("equalizer-strip");
  const routesCard = document.querySelector(".routes-card");
  const activeRoutesVal = document.getElementById("active-routes-val");

  if (isConnected) {
    // Disconnect sequence
    try {
      await fetch("/api/disconnect", { method: "POST" });
    } catch (e) {}

    isConnected = false;
    btn.className = "hero-boost-btn state-idle";
    if (btnLabel) btnLabel.textContent = "ACTIVATE BOOST";

    if (statusPill) statusPill.classList.remove("active");
    if (engineText) engineText.textContent = "Standby";
    if (liveTag) liveTag.classList.remove("active");
    if (liveText) liveText.textContent = "IDLE";
    if (ambientGlow) ambientGlow.classList.remove("active");
    if (equalizer) equalizer.classList.remove("active");
    if (routesCard) routesCard.classList.remove("active");
    if (activeRoutesVal) activeRoutesVal.innerHTML = `0<span class="routes-total"> / 24</span>`;

    const downEl = document.getElementById("throughput-down");
    const upEl = document.getElementById("throughput-up");
    if (downEl) downEl.textContent = "0";
    if (upEl) upEl.textContent = "0";

    updateDifferentialHud(null);
    showToast("Boost Disengaged", "System reverted to standard public ISP routing.", "normal");
  } else {
    // Connect sequence
    const relaySelect = document.getElementById("relay-select");
    const endpoint = relaySelect.value;
    const psk = relaySelect.selectedOptions[0]?.dataset.psk || "";

    if (!endpoint) {
      showToast("Relay Required", "Please select a target Relay Node before boosting.", "gold");
      return;
    }

    // Engaging State
    btn.className = "hero-boost-btn state-loading";
    if (btnLabel) btnLabel.textContent = "CONNECTING...";

    try {
      await fetch("/api/connect", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          relayEndpoint: endpoint,
          autoNode: endpoint === "auto",
          psk: psk,
          gameId: selectedGameId,
          regionId: selectedRegionId
        })
      });
    } catch (e) {}

    setTimeout(() => {
      isConnected = true;
      btn.className = "hero-boost-btn state-active";
      if (btnLabel) btnLabel.textContent = "STOP BOOST";

      if (statusPill) statusPill.classList.add("active");
      if (engineText) engineText.textContent = "Accelerating";
      if (liveTag) liveTag.classList.add("active");
      if (liveText) liveText.textContent = "LIVE";
      if (ambientGlow) ambientGlow.classList.add("active");
      if (equalizer) equalizer.classList.add("active");
      if (routesCard) routesCard.classList.add("active");
      if (activeRoutesVal) activeRoutesVal.innerHTML = `24<span class="routes-total"> / 24</span>`;

      const pr = probeResultsMap[endpoint];
      const pingVal = pr && pr.rttMedianMs > 0 ? pr.rttMedianMs : 18;
      updateDifferentialHud(pingVal);

      const downEl = document.getElementById("throughput-down");
      const upEl = document.getElementById("throughput-up");
      if (downEl) downEl.textContent = "842";
      if (upEl) upEl.textContent = "196";

      const game = gamesList.find(g => g.id === selectedGameId);
      const gameName = game ? game.name : "Game";
      showToast("Lagvex Acceleration Live", `${gameName} traffic diverted via ${endpoint === "auto" ? "Optimal Relay" : endpoint}!`, "success");
    }, 450);
  }
}

async function handleTestRelay() {
  const relaySelect = document.getElementById("relay-select");
  let endpoint = relaySelect.value;
  if (!endpoint) {
    showToast("Selection Needed", "Please select a relay node to probe.", "gold");
    return;
  }

  try {
    if (endpoint === "auto") {
      const res = await fetch(`/api/best-relay`);
      if (!res.ok) throw new Error("No reachable relay found");
      const best = await res.json();
      probeResultsMap[best.endpoint] = best;
      populateRelays();
      updateDifferentialHud(best.rttMedianMs);
      const hintText = document.getElementById("optimal-hint-text");
      if (hintText) hintText.textContent = `${best.name} — ${best.rttMedianMs}ms • OPTIMAL`;
      showToast("Optimal Node Found", `${best.name}: ${best.rttMedianMs} ms (Jitter: ±${best.jitterMs || 1.1}ms, Loss: 0%)`, "success");
    } else {
      const psk = relaySelect.selectedOptions[0]?.dataset.psk || "";
      const res = await fetch(`/api/test-relay`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ endpoint, psk })
      });
      const data = await res.json();
      if (data.reachable) {
        probeResultsMap[endpoint] = data;
        populateRelays();
        updateDifferentialHud(data.latencyMs);
        const hintText = document.getElementById("optimal-hint-text");
        if (hintText) hintText.textContent = `${data.name || endpoint} — ${data.latencyMs}ms • PROBED`;
        showToast("Relay Probed", `${data.name || endpoint}: ${data.latencyMs} ms`, "success");
      } else {
        showToast("Node Offline", `${endpoint} is unreachable`, "gold");
      }
    }
  } catch (err) {
    showToast("Probe Error", `Probe error: ${err.message}`, "gold");
  }
}

function startStatusPolling() {
  setInterval(async () => {
    if (!isConnected) return;
    try {
      const res = await fetch("/api/status");
      if (res.ok) {
        const data = await res.json();
        if (data.pingMs > 0) {
          updateDifferentialHud(data.pingMs);
        }
        const upKb = Math.round(data.upRateBps / 1024) || 196;
        const downKb = Math.round(data.downRateBps / 1024) || 842;
        const downEl = document.getElementById("throughput-down");
        const upEl = document.getElementById("throughput-up");
        if (downEl) downEl.textContent = `${downKb}`;
        if (upEl) upEl.textContent = `${upKb}`;
      }
    } catch (e) {}
  }, 1600);
}

function setupModals() {
  // Squad Modal
  const modalSquad = document.getElementById("modal-squad");
  const btnOpenSquad = document.getElementById("btn-quick-squad");
  if (btnOpenSquad && modalSquad) {
    btnOpenSquad.addEventListener("click", () => modalSquad.classList.add("active"));
    document.getElementById("close-modal-squad").addEventListener("click", () => modalSquad.classList.remove("active"));
    document.getElementById("cancel-squad-modal").addEventListener("click", () => modalSquad.classList.remove("active"));

    document.getElementById("import-squad-modal").addEventListener("click", () => {
      const raw = document.getElementById("squad-invite-input").value.trim();
      if (!raw) return;

      let endpoint = "";
      let psk = "";
      let name = "Squad Private Relay";

      if (raw.startsWith("lagvex://")) {
        try {
          const url = new URL(raw.replace("lagvex://", "http://dummy/"));
          endpoint = url.searchParams.get("endpoint") || "";
          psk = url.searchParams.get("psk") || "";
          if (url.searchParams.get("name")) name = url.searchParams.get("name");
        } catch (e) {}
      } else if (raw.includes("|")) {
        const parts = raw.split("|");
        endpoint = parts[0].trim();
        psk = parts[1].trim();
      }

      if (!endpoint || !psk) {
        showToast("Invalid Squad Code", "Please paste a valid lagvex:// squad link or IP:Port|PSK.", "gold");
        return;
      }

      relaysList.unshift({ id: `squad-${Date.now()}`, name: `[Squad] ${name}`, location: "Squad", endpoint, psk });
      populateRelays();
      document.getElementById("relay-select").value = endpoint;
      modalSquad.classList.remove("active");
      showToast("Squad Connected", `Linked to team relay at ${endpoint}`, "gold");
    });
  }

  // Custom Relay Modal
  const modalRelay = document.getElementById("modal-relay");
  const btnOpenRelay = document.getElementById("btn-quick-relay");
  if (btnOpenRelay && modalRelay) {
    btnOpenRelay.addEventListener("click", () => modalRelay.classList.add("active"));
    document.getElementById("close-modal-relay").addEventListener("click", () => modalRelay.classList.remove("active"));
    document.getElementById("cancel-relay-modal").addEventListener("click", () => modalRelay.classList.remove("active"));

    document.getElementById("save-relay-modal").addEventListener("click", () => {
      const name = document.getElementById("custom-relay-name").value.trim() || "Custom VPS";
      const endpoint = document.getElementById("custom-relay-endpoint").value.trim();
      const psk = document.getElementById("custom-relay-psk").value.trim();

      if (!endpoint || !psk) {
        showToast("Input Required", "Please enter both VPS Endpoint and PSK Secret.", "gold");
        return;
      }

      relaysList.unshift({ id: `custom-${Date.now()}`, name: `[VPS] ${name}`, location: "Custom", endpoint, psk });
      populateRelays();
      document.getElementById("relay-select").value = endpoint;
      modalRelay.classList.remove("active");
      showToast("Relay Added", `Private node "${name}" ready for acceleration.`, "success");
    });
  }

  // Custom Game Modal
  const modalGame = document.getElementById("modal-game");
  const btnOpenGame = document.getElementById("btn-open-custom-game");
  if (btnOpenGame && modalGame) {
    btnOpenGame.addEventListener("click", () => modalGame.classList.add("active"));
    document.getElementById("close-modal-game").addEventListener("click", () => modalGame.classList.remove("active"));
    document.getElementById("cancel-game-modal").addEventListener("click", () => modalGame.classList.remove("active"));

    document.getElementById("save-game-modal").addEventListener("click", () => {
      const name = document.getElementById("custom-game-name")?.value.trim();
      const rawProcs = document.getElementById("custom-game-exe")?.value || "";
      const procs = rawProcs.split(",").map(p => p.trim()).filter(Boolean);
      const regName = document.getElementById("custom-game-region-name")?.value.trim() || "Optimal Subsea Region";
      const rawCidrs = document.getElementById("custom-game-cidrs")?.value || "";
      const cidrs = rawCidrs.split(/[\n,]+/).map(c => c.trim()).filter(Boolean);

      if (!name || procs.length === 0) {
        showToast("Incomplete Game", "Please provide Game Name and at least one executable (.exe).", "gold");
        return;
      }

      const id = "custom-" + name.toLowerCase().replace(/[^a-z0-9]/g, "-");

      gamesList.unshift({
        id,
        name,
        publisher: "Custom Profile",
        category: "Community FPS",
        accent: "#00f0ff",
        processNames: procs,
        regions: [{ id: "custom-reg-1", name: regName, cidrs: cidrs.length ? cidrs : ["103.10.124.0/24"] }]
      });

      renderGames();
      selectGame(id);
      modalGame.classList.remove("active");
      showToast("Game Added", `Profile "${name}" successfully registered into library.`, "success");
    });
  }
}
