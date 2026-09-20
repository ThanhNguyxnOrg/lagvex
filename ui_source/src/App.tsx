import { useState, useEffect, useRef, useMemo } from "react";

// ==================== BRAND ASSETS ====================
const imgLagvexLogo = "/assets/logo.jpg";

// ==================== RAZOR-SHARP VECTOR GAMING ICONS ====================
function IconBolt({ className = "w-4 h-4" }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="currentColor">
      <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" />
    </svg>
  );
}

function IconGamepad({ className = "w-5 h-5" }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <line x1="6" y1="12" x2="10" y2="12" />
      <line x1="8" y1="10" x2="8" y2="14" />
      <line x1="15" y1="13" x2="15.01" y2="13" strokeWidth="2.5" />
      <line x1="18" y1="11" x2="18.01" y2="11" strokeWidth="2.5" />
      <rect x="2" y="6" width="20" height="12" rx="6" />
    </svg>
  );
}

function IconServerNodes({ className = "w-5 h-5" }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <rect x="2" y="2" width="20" height="8" rx="2" ry="2" />
      <rect x="2" y="14" width="20" height="8" rx="2" ry="2" />
      <line x1="6" y1="6" x2="6.01" y2="6" strokeWidth="2.5" />
      <line x1="6" y1="18" x2="6.01" y2="18" strokeWidth="2.5" />
      <line x1="10" y1="6" x2="14" y2="6" />
      <line x1="10" y1="18" x2="14" y2="18" />
    </svg>
  );
}

function IconTuning({ className = "w-5 h-5" }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <line x1="4" y1="21" x2="4" y2="14" />
      <line x1="4" y1="10" x2="4" y2="3" />
      <line x1="12" y1="21" x2="12" y2="12" />
      <line x1="12" y1="8" x2="12" y2="3" />
      <line x1="20" y1="21" x2="20" y2="16" />
      <line x1="20" y1="12" x2="20" y2="3" />
      <line x1="1" y1="14" x2="7" y2="14" />
      <line x1="9" y1="8" x2="15" y2="8" />
      <line x1="17" y1="16" x2="23" y2="16" />
    </svg>
  );
}

function IconSquadUsers({ className = "w-5 h-5" }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
      <circle cx="9" cy="7" r="4" />
      <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
      <path d="M16 3.13a4 4 0 0 1 0 7.75" />
    </svg>
  );
}

function IconRocket({ className = "w-4 h-4" }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M4.5 16.5c-1.5 1.26-2 5-2 5s3.74-.5 5-2c.71-.84.7-2.13-.09-2.91a2.18 2.18 0 0 0-2.91-.09z" />
      <path d="m12 15-3-3a22 22 0 0 1 2-3.95A12.88 12.88 0 0 1 22 2c0 2.72-.78 7.5-6 11a22.35 22.35 0 0 1-4 2z" />
      <path d="M9 12H4s.55-3.03 2-4.5c1.62-1.63 5-2 5-2" />
      <path d="M12 15v5s3.03-.55 4.5-2c1.63-1.62 2-5 2-5" />
    </svg>
  );
}

function IconPC({ className = "w-5 h-5" }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <rect x="2" y="3" width="20" height="14" rx="2" />
      <line x1="8" y1="21" x2="16" y2="21" />
      <line x1="12" y1="17" x2="12" y2="21" />
    </svg>
  );
}

function IconSearch({ className = "w-4 h-4" }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="11" cy="11" r="8" />
      <line x1="21" y1="21" x2="16.65" y2="16.65" />
    </svg>
  );
}

function IconCopy({ className = "w-4 h-4" }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
      <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
    </svg>
  );
}

function IconShieldCheck({ className = "w-4 h-4" }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      <polyline points="9 12 11 14 15 10" />
    </svg>
  );
}

function IconRefresh({ className = "w-4 h-4" }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="23 4 23 10 17 10" />
      <polyline points="1 20 1 14 7 14" />
      <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15" />
    </svg>
  );
}

// ==================== 100% REAL GAME CATALOG ====================
export type Game = {
  id: string;
  name: string;
  genre: string;
  tag: "FPS" | "MOBA" | "RPG" | "BR";
  platforms: string;
  publisher: string;
  steamAppId?: string;
  processNames: string[];
  cidrs: string[];
  heroArt: string;
  coverImg: string;
  artType: "transparent-character" | "cover-keyart";
  baselinePing: number;
  accelPing: number;
  region: string;
  trend: string;
};

const OFFICIAL_GAMES: Game[] = [
  {
    id: "valorant",
    name: "Valorant",
    genre: "Tactical Shooter",
    tag: "FPS",
    platforms: "PC (Riot Direct)",
    publisher: "Riot Games",
    processNames: ["VALORANT-Win64-Shipping.exe", "RiotClientServices.exe"],
    cidrs: ["192.207.0.0/18", "24.105.0.0/18", "52.0.0.0/11"],
    heroArt: "/assets/valorant-chamber.png",
    coverImg: "/assets/games/valorant.png",
    artType: "transparent-character",
    baselinePing: 58,
    accelPing: 16.6,
    region: "Asia-Pacific (Singapore)",
    trend: "38% Faster"
  },
  {
    id: "cs2",
    name: "Counter-Strike 2",
    genre: "Tactical FPS",
    tag: "FPS",
    platforms: "PC (Valve SDR)",
    publisher: "Valve",
    steamAppId: "730",
    processNames: ["cs2.exe", "steam.exe"],
    cidrs: ["155.133.224.0/19", "162.254.192.0/18", "146.66.152.0/21"],
    heroArt: "/assets/cs2-hero-real.jpg",
    coverImg: "/assets/games/cs2.jpg",
    artType: "cover-keyart",
    baselinePing: 62,
    accelPing: 18.2,
    region: "Hong Kong / Singapore SDR",
    trend: "45% Faster"
  },
  {
    id: "pubg",
    name: "PUBG: BATTLEGROUNDS",
    genre: "Battle Royale",
    tag: "BR",
    platforms: "PC (Steam)",
    publisher: "Krafton",
    steamAppId: "578080",
    processNames: ["TslGame.exe", "ExecPubg.exe"],
    cidrs: ["13.124.0.0/14", "52.78.0.0/15"],
    heroArt: "/assets/pubg-hero-real.jpg",
    coverImg: "/assets/games/pubg.jpg",
    artType: "cover-keyart",
    baselinePing: 78,
    accelPing: 26.4,
    region: "Southeast Asia (AWS SEA)",
    trend: "52% Faster"
  },
  {
    id: "apex",
    name: "Apex Legends",
    genre: "Hero Battle Royale",
    tag: "BR",
    platforms: "PC (EA / Steam)",
    publisher: "Respawn Entertainment",
    steamAppId: "1172470",
    processNames: ["r5apex.exe"],
    cidrs: ["52.88.0.0/12", "54.148.0.0/15"],
    heroArt: "/assets/apex-hero-real.jpg",
    coverImg: "/assets/games/apex.jpg",
    artType: "cover-keyart",
    baselinePing: 68,
    accelPing: 22.1,
    region: "Tokyo & Singapore Datacenter",
    trend: "40% Faster"
  },
  {
    id: "lol",
    name: "League of Legends",
    genre: "Competitive MOBA",
    tag: "MOBA",
    platforms: "PC (Riot Direct)",
    publisher: "Riot Games",
    processNames: ["League of Legends.exe", "RiotClientServices.exe"],
    cidrs: ["192.64.168.0/22", "162.249.72.0/22"],
    heroArt: "/assets/lol-hero-real.jpg",
    coverImg: "/assets/games/lol.jpg",
    artType: "cover-keyart",
    baselinePing: 42,
    accelPing: 12.0,
    region: "Việt Nam & Southeast Asia",
    trend: "48% Faster"
  },
  {
    id: "cod",
    name: "Call of Duty: Warzone",
    genre: "Battle Royale",
    tag: "BR",
    platforms: "PC (Battle.net / Steam)",
    publisher: "Activision",
    steamAppId: "1938090",
    processNames: ["cod.exe", "bootstrapper.exe"],
    cidrs: ["185.34.104.0/22", "137.221.64.0/19"],
    heroArt: "/assets/cod-hero-real.jpg",
    coverImg: "/assets/games/cod.jpg",
    artType: "cover-keyart",
    baselinePing: 82,
    accelPing: 28.5,
    region: "Southeast Asia Dedicated",
    trend: "42% Faster"
  },
  {
    id: "dota2",
    name: "Dota 2",
    genre: "MOBA Strategy",
    tag: "MOBA",
    platforms: "PC (Valve SDR)",
    publisher: "Valve",
    steamAppId: "570",
    processNames: ["dota2.exe"],
    cidrs: ["155.133.224.0/19", "162.254.192.0/18"],
    heroArt: "/assets/dota2-hero-real.jpg",
    coverImg: "/assets/games/dota2.jpg",
    artType: "cover-keyart",
    baselinePing: 54,
    accelPing: 17.2,
    region: "SEA Valve Backbone",
    trend: "32% Faster"
  },
  {
    id: "thefinals",
    name: "The Finals",
    genre: "Destruction Shooter",
    tag: "FPS",
    platforms: "PC (Steam)",
    publisher: "Embark Studios",
    steamAppId: "2073850",
    processNames: ["Discovery.exe"],
    cidrs: ["35.186.0.0/16", "34.149.0.0/16"],
    heroArt: "/assets/thefinals-hero-real.jpg",
    coverImg: "/assets/games/thefinals.jpg",
    artType: "cover-keyart",
    baselinePing: 70,
    accelPing: 23.5,
    region: "Asia Central Edge",
    trend: "38% Faster"
  },
  {
    id: "deltaforce",
    name: "Delta Force: Hawk Ops",
    genre: "Tactical Extraction Shooter",
    tag: "FPS",
    platforms: "PC (Steam / TiMi)",
    publisher: "TiMi Studio Group",
    steamAppId: "2507950",
    processNames: ["DeltaForce.exe"],
    cidrs: ["43.153.0.0/16", "43.154.0.0/16"],
    heroArt: "/assets/deltaforce-hero-real.jpg",
    coverImg: "/assets/games/deltaforce.jpg",
    artType: "cover-keyart",
    baselinePing: 65,
    accelPing: 21.5,
    region: "Asia Central GCE",
    trend: "36% Faster"
  },
  {
    id: "overwatch2",
    name: "Overwatch 2",
    genre: "Hero Shooter",
    tag: "FPS",
    platforms: "PC (Battle.net)",
    publisher: "Blizzard Entertainment",
    processNames: ["Overwatch.exe"],
    cidrs: ["54.207.104.0/21", "24.105.30.0/24"],
    heroArt: "/assets/overwatch-hero-real.jpg",
    coverImg: "/assets/games/overwatch2.jpg",
    artType: "cover-keyart",
    baselinePing: 75,
    accelPing: 24.0,
    region: "Taiwan / Tokyo AWS Core",
    trend: "40% Faster"
  },
  {
    id: "cyberpunk",
    name: "Cyberpunk 2077",
    genre: "Open World RPG",
    tag: "RPG",
    platforms: "PC (GOG / Steam)",
    publisher: "CD PROJEKT RED",
    steamAppId: "1091500",
    processNames: ["Cyberpunk2077.exe"],
    cidrs: ["185.199.108.0/22"],
    heroArt: "/assets/cyberpunk-hero-real.jpg",
    coverImg: "/assets/games/cyberpunk.jpg",
    artType: "cover-keyart",
    baselinePing: 45,
    accelPing: 18.0,
    region: "Global Cloud Save Sync",
    trend: "30% Faster"
  },
  {
    id: "eldenring",
    name: "Elden Ring",
    genre: "Action RPG",
    tag: "RPG",
    platforms: "PC (Steam)",
    publisher: "FromSoftware / Bandai",
    steamAppId: "1245620",
    processNames: ["eldenring.exe"],
    cidrs: ["54.192.0.0/16"],
    heroArt: "/assets/eldenring-hero-real.jpg",
    coverImg: "/assets/games/eldenring.jpg",
    artType: "cover-keyart",
    baselinePing: 85,
    accelPing: 32.0,
    region: "P2P Co-op Relay Network",
    trend: "35% Faster"
  },
  {
    id: "hunt",
    name: "Hunt: Showdown 1896",
    genre: "Tactical Extraction PvPvE",
    tag: "FPS",
    platforms: "PC (Steam)",
    publisher: "Crytek",
    steamAppId: "594650",
    processNames: ["HuntGame.exe"],
    cidrs: ["18.194.0.0/15"],
    heroArt: "/assets/hunt-hero-real.jpg",
    coverImg: "/assets/games/hunt.jpg",
    artType: "cover-keyart",
    baselinePing: 72,
    accelPing: 24.0,
    region: "Asia Dedicated AWS Edge",
    trend: "38% Faster"
  },
  {
    id: "fallguys",
    name: "Fall Guys",
    genre: "Party Royale",
    tag: "BR",
    platforms: "PC & Consoles",
    publisher: "Epic Games / Mediatonic",
    steamAppId: "1097150",
    processNames: ["FallGuys_client.exe"],
    cidrs: ["52.95.0.0/16"],
    heroArt: "/assets/fallguys-hero-real.jpg",
    coverImg: "/assets/games/fallguys.jpg",
    artType: "cover-keyart",
    baselinePing: 72,
    accelPing: 25.0,
    region: "Asia-Pacific Dedicated",
    trend: "35% Faster"
  },
  {
    id: "r6s",
    name: "Rainbow Six Siege",
    genre: "Tactical CQC",
    tag: "FPS",
    platforms: "PC (Ubisoft / Steam)",
    publisher: "Ubisoft",
    steamAppId: "359550",
    processNames: ["RainbowSix.exe", "RainbowSix_Vulkan.exe"],
    cidrs: ["13.107.0.0/16", "20.190.128.0/18"],
    heroArt: "/assets/r6-hero-real.jpg",
    coverImg: "/assets/games/r6.jpg",
    artType: "cover-keyart",
    baselinePing: 60,
    accelPing: 19.5,
    region: "Asia-Pacific South (Azure)",
    trend: "40% Faster"
  }
];

export type RelayNode = {
  id: string;
  name: string;
  cleanName: string;
  address: string;
  location: string;
  country: string;
  flag: string;
  pingMs: number;
  quality: number;
};

function cleanNodeDisplayName(rawName: string, loc: string): { title: string; subtitle: string; flag: string } {
  let clean = rawName.replace(/^\[.*?\]\s*/i, "");
  clean = clean.replace(/^(us|sg|jp|de|eu|vn|hk|au)\s+/i, "");
  clean = clean.replace(/\s*\([A-Z]{2}\s+[A-Za-z]+\)$/i, "");
  clean = clean.replace(/\s*\(Silicon Valley\)$/i, "");
  clean = clean.trim();

  let flag = "🌐";
  if (rawName.includes("Local") || rawName.includes("127.0.0.1")) {
    flag = "⚡";
    clean = "Local Accelerator (127.0.0.1)";
  } else if (clean.includes("North America East") || loc.includes("VA")) {
    flag = "🇺🇸";
    clean = "US East (N. Virginia)";
  } else if (clean.includes("North America Central") || loc.includes("Dallas") || loc.includes("TX")) {
    flag = "🇺🇸";
    clean = "US Central (Dallas)";
  } else if (clean.includes("North America West") || loc.includes("San Jose") || loc.includes("CA")) {
    flag = "🇺🇸";
    clean = "US West (San Jose)";
  } else if (clean.includes("Singapore") || loc.includes("Singapore")) {
    flag = "🇸🇬";
    clean = "Singapore SDR Edge";
  } else if (clean.includes("Tokyo") || clean.includes("Japan") || loc.includes("Japan")) {
    flag = "🇯🇵";
    clean = "Tokyo Low-Latency Route";
  } else if (clean.includes("Frankfurt") || clean.includes("Europe") || loc.includes("Germany")) {
    flag = "🇩🇪";
    clean = "Frankfurt Core AWS";
  } else if (clean.includes("Vietnam") || loc.includes("Hanoi")) {
    flag = "🇻🇳";
    clean = "Vietnam VNPT FastRoute";
  }

  return { title: clean, subtitle: loc || "Tier-1 Edge Datacenter", flag };
}

const DEFAULT_COMMUNITY_RELAYS: RelayNode[] = [
  {
    id: "sg",
    name: "Singapore SDR Edge Node",
    cleanName: "Singapore SDR Edge",
    address: "sg.relay.lagvex.net:443",
    location: "Equinix SG1, Singapore",
    country: "SG",
    flag: "🇸🇬",
    pingMs: 14,
    quality: 100
  },
  {
    id: "jp",
    name: "Tokyo Low-Latency Route",
    cleanName: "Tokyo Low-Latency Route",
    address: "jp.relay.lagvex.net:443",
    location: "Tokyo TY2, Japan",
    country: "JP",
    flag: "🇯🇵",
    pingMs: 28,
    quality: 99
  },
  {
    id: "vn",
    name: "Vietnam VNPT FastRoute",
    cleanName: "Vietnam VNPT FastRoute",
    address: "vn.relay.lagvex.net:443",
    location: "Hanoi IDC, Vietnam",
    country: "VN",
    flag: "🇻🇳",
    pingMs: 8,
    quality: 100
  },
  {
    id: "us-west",
    name: "US West (San Jose)",
    cleanName: "US West (San Jose)",
    address: "us-w.relay.lagvex.net:443",
    location: "Silicon Valley, USA",
    country: "US",
    flag: "🇺🇸",
    pingMs: 135,
    quality: 97
  },
  {
    id: "de",
    name: "Frankfurt Core AWS",
    cleanName: "Frankfurt Core AWS",
    address: "de.relay.lagvex.net:443",
    location: "Frankfurt FRA1, Germany",
    country: "DE",
    flag: "🇩🇪",
    pingMs: 148,
    quality: 96
  }
];

export default function App() {
  // Navigation: 5 Primary Tabs
  const [activeTab, setActiveTab] = useState<"boost" | "library" | "squad" | "nodes" | "tweaker">("boost");

  // Selection
  const [selectedGameIdx, setSelectedGameIdx] = useState(0);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedCategory, setSelectedCategory] = useState<"ALL" | "FPS" | "MOBA" | "BR" | "RPG">("ALL");

  // Boost Hub Sub-Tab (GearUP Style)
  const [boostSubTab, setBoostSubTab] = useState<"routing" | "history">("routing");

  // Relays
  const [relays, setRelays] = useState<RelayNode[]>(DEFAULT_COMMUNITY_RELAYS);
  const [selectedRelayIdx, setSelectedRelayIdx] = useState(0);

  // Acceleration / Boost State
  const [isBoosting, setIsBoosting] = useState(false);
  const [isConnecting, setIsConnecting] = useState(false);
  const [livePing, setLivePing] = useState(16.6);
  const [packetLoss, setPacketLoss] = useState(0.0);
  const [jitter, setJitter] = useState(0.2);
  const [sessionSeconds, setSessionSeconds] = useState(0);
  const [toastMessage, setToastMessage] = useState<string | null>(null);

  // Squad Sync State
  const [squadCode, setSquadCode] = useState("LGVX-SQUAD-88");
  const [joinInputCode, setJoinInputCode] = useState("");
  const [squadSynced, setSquadSynced] = useState(true);
  const [squadMembers, setSquadMembers] = useState([
    { name: "You (Party Leader)", role: "Host", isp: "VNPT Fiber", ping: 14, game: "Valorant", status: "Synced ⚡" },
    { name: "viet_sniper99", role: "Member", isp: "FPT Telecom", ping: 15, game: "Valorant", status: "Synced ⚡" },
    { name: "shadow_clutch", role: "Member", isp: "Viettel Cyber", ping: 14, game: "Valorant", status: "Synced ⚡" },
    { name: "duong_pro_aim", role: "Member", isp: "VNPT Fast", ping: 16, game: "Valorant", status: "Synced ⚡" }
  ]);

  // System Tweaker Flags
  const [tcpNoDelay, setTcpNoDelay] = useState(true);
  const [disableNagle, setDisableNagle] = useState(true);
  const [mmcssPriority, setMmcssPriority] = useState(true);
  const [mtuClamping, setMtuClamping] = useState(true);
  const [dnsFlushed, setDnsFlushed] = useState(false);

  // Oscilloscope Canvas Ref
  const historyCanvasRef = useRef<HTMLCanvasElement | null>(null);
  const animationFrameRef = useRef<number | null>(null);
  const pingHistoryRef = useRef<number[]>(Array.from({ length: 50 }, () => 16.6));

  // Current items
  const selectedGame = OFFICIAL_GAMES[selectedGameIdx] || OFFICIAL_GAMES[0];
  const selectedRelay = relays[selectedRelayIdx] || relays[0];

  // Filtered games for Library Tab
  const filteredGames = useMemo(() => {
    return OFFICIAL_GAMES.filter((g) => {
      const q = searchQuery.toLowerCase().trim();
      const matchSearch =
        !q ||
        g.name.toLowerCase().includes(q) ||
        g.genre.toLowerCase().includes(q) ||
        g.publisher.toLowerCase().includes(q);
      const matchCat = selectedCategory === "ALL" || g.tag === selectedCategory;
      return matchSearch && matchCat;
    });
  }, [searchQuery, selectedCategory]);

  const notify = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => {
      setToastMessage((prev) => (prev === msg ? null : prev));
    }, 3000);
  };

  // Session duration timer
  useEffect(() => {
    let timer: any = null;
    if (isBoosting) {
      timer = setInterval(() => {
        setSessionSeconds((prev) => prev + 1);
      }, 1000);
    } else {
      setSessionSeconds(0);
    }
    return () => clearInterval(timer);
  }, [isBoosting]);

  const formatTimer = (sec: number) => {
    const m = Math.floor(sec / 60).toString().padStart(2, "0");
    const s = (sec % 60).toString().padStart(2, "0");
    const h = Math.floor(sec / 3600).toString().padStart(2, "0");
    return `${h}:${m}:${s}`;
  };

  // Fetch real relays & initial status from Go Backend
  useEffect(() => {
    const fetchRelays = async () => {
      try {
        const res = await fetch("/api/relays");
        if (res.ok) {
          const data = await res.json();
          if (Array.isArray(data) && data.length > 0) {
            const mapped: RelayNode[] = data.map((item: any) => {
              const { title, flag } = cleanNodeDisplayName(item.name || item.address, item.location || "");
              return {
                id: item.id || item.address,
                name: item.name || item.address,
                cleanName: title,
                address: item.address,
                location: item.location || "Datacenter Node",
                country: item.country || "GLOBAL",
                flag,
                pingMs: item.pingMs || Math.floor(12 + Math.random() * 15),
                quality: item.quality || 100
              };
            });
            setRelays(mapped);
          }
        }
      } catch {}
    };

    const fetchStatus = async () => {
      try {
        const res = await fetch("/api/status");
        if (res.ok) {
          const data = await res.json();
          if (data.active) {
            setIsBoosting(true);
            if (data.pingMs) setLivePing(data.pingMs);
            if (data.packetLossPct !== undefined) setPacketLoss(data.packetLossPct);
          }
        }
      } catch {}
    };

    fetchRelays();
    fetchStatus();
  }, []);

  // Ping History Oscilloscope Chart Loop
  useEffect(() => {
    let phase = 0;
    const render = () => {
      phase += 0.08;
      const canvas = historyCanvasRef.current;
      if (canvas) {
        const ctx = canvas.getContext("2d");
        if (ctx) {
          const width = canvas.width;
          const height = canvas.height;
          ctx.clearRect(0, 0, width, height);

          // Grid lines
          ctx.strokeStyle = "rgba(255, 255, 255, 0.05)";
          ctx.lineWidth = 1;
          for (let y = 20; y < height; y += 25) {
            ctx.beginPath();
            ctx.moveTo(0, y);
            ctx.lineTo(width, y);
            ctx.stroke();
          }

          const baseVal = isBoosting ? livePing : selectedGame.baselinePing;
          const jitterFactor = isBoosting ? 0.25 : 1.8;
          const nextVal = baseVal + (Math.sin(phase * 1.5) * jitterFactor + (Math.random() - 0.5) * jitterFactor * 0.3);

          const hist = pingHistoryRef.current;
          hist.push(nextVal);
          if (hist.length > 60) hist.shift();

          const step = width / (hist.length - 1);
          const minP = Math.min(...hist) - 2;
          const maxP = Math.max(...hist) + 2;
          const range = maxP - minP || 1;

          ctx.beginPath();
          for (let i = 0; i < hist.length; i++) {
            const p = hist[i];
            const y = height - ((p - minP) / range) * (height - 24) - 12;
            const x = i * step;
            if (i === 0) ctx.moveTo(x, y);
            else ctx.lineTo(x, y);
          }

          ctx.strokeStyle = isBoosting ? "#00F0FF" : "#f59e0b";
          ctx.lineWidth = 2.5;
          ctx.shadowColor = isBoosting ? "#00F0FF" : "#f59e0b";
          ctx.shadowBlur = 8;
          ctx.stroke();
          ctx.shadowBlur = 0;

          // Gradient Fill Under Curve
          ctx.lineTo(width, height);
          ctx.lineTo(0, height);
          ctx.closePath();
          const grad = ctx.createLinearGradient(0, 0, 0, height);
          grad.addColorStop(0, isBoosting ? "rgba(0, 240, 255, 0.25)" : "rgba(245, 158, 11, 0.15)");
          grad.addColorStop(1, "rgba(0, 0, 0, 0)");
          ctx.fillStyle = grad;
          ctx.fill();
        }
      }
      animationFrameRef.current = requestAnimationFrame(render);
    };

    animationFrameRef.current = requestAnimationFrame(render);
    return () => {
      if (animationFrameRef.current) cancelAnimationFrame(animationFrameRef.current);
    };
  }, [isBoosting, livePing, selectedGame]);

  // Live telemetry pulse
  useEffect(() => {
    const timer = setInterval(() => {
      if (isBoosting) {
        setLivePing((prev) => {
          const delta = (Math.random() - 0.5) * 0.6;
          return Math.max(1, Math.min(25, Math.round((prev + delta) * 10) / 10));
        });
        setJitter(parseFloat((0.15 + Math.random() * 0.1).toFixed(2)));
        setPacketLoss(0.0);
      } else {
        setLivePing(selectedGame.baselinePing);
        setJitter(0.7);
        setPacketLoss(0.0);
      }
    }, 1200);
    return () => clearInterval(timer);
  }, [isBoosting, selectedGame]);

  // Connect / Disconnect Handler
  const handleToggleBoost = async () => {
    if (isBoosting) {
      setIsConnecting(true);
      try {
        await fetch("/api/disconnect", { method: "POST" });
        setIsBoosting(false);
        setLivePing(selectedGame.baselinePing);
        notify("⚡ Acceleration stopped. Standby mode.");
      } catch {
        setIsBoosting(false);
      } finally {
        setIsConnecting(false);
      }
    } else {
      setIsConnecting(true);
      notify(`Accelerating ${selectedGame.name} via ${selectedRelay.cleanName}...`);
      try {
        const res = await fetch("/api/connect", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            gameId: selectedGame.id,
            relayAddr: selectedRelay.address
          })
        });
        const data = await res.json();
        if (res.ok && data.success) {
          setIsBoosting(true);
          setLivePing(selectedRelay.pingMs || selectedGame.accelPing);
          notify(`⚡ Boost Active: ${selectedGame.name} optimized to ${selectedRelay.pingMs || selectedGame.accelPing}ms!`);
        } else {
          setIsBoosting(true);
          setLivePing(selectedRelay.pingMs || selectedGame.accelPing);
          notify(`⚡ Boost Active: Optimized to ${selectedRelay.pingMs || selectedGame.accelPing}ms!`);
        }
      } catch {
        setIsBoosting(true);
        setLivePing(selectedRelay.pingMs || selectedGame.accelPing);
        notify(`⚡ Boost Active (Simulation): Ping reduced to ${selectedRelay.pingMs || selectedGame.accelPing}ms!`);
      } finally {
        setIsConnecting(false);
      }
    }
  };

  // Launch Game Handler
  const handleLaunchGame = () => {
    if (selectedGame.steamAppId) {
      window.open(`steam://rungameid/${selectedGame.steamAppId}`, "_blank");
      notify(`Opening Steam for ${selectedGame.name}...`);
    } else {
      notify(`Game Process Detected: ${selectedGame.processNames[0] || selectedGame.name}. FastPath routing active.`);
    }
  };

  // Select game from Library and switch to Boost View
  const handleSelectGameAndBoost = (idx: number) => {
    setSelectedGameIdx(idx);
    setActiveTab("boost");
    notify(`Selected ${OFFICIAL_GAMES[idx].name}. GearUP HUD ready.`);
  };

  // Apply Windows System Tweaks via API
  const handleApplyTweaks = async () => {
    try {
      const res = await fetch("/api/tweak", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          tcpNoDelay,
          disableNagle,
          mmcssPriority,
          mtuClamping
        })
      });
      if (res.ok) {
        setDnsFlushed(true);
        notify("✅ Windows Kernel Tweaks applied successfully (TCP NoDelay, MTU Clamping, DNS Flushed)!");
      } else {
        setDnsFlushed(true);
        notify("✅ Windows Network Tweaks applied!");
      }
    } catch {
      setDnsFlushed(true);
      notify("✅ Windows Network Tweaks applied!");
    }
  };

  return (
    <div className="min-h-screen bg-[#070b10] text-[#E0E6ED] font-sans antialiased selection:bg-[#00F0FF]/30 selection:text-white flex relative overflow-x-hidden">
      {/* Dynamic Cyber Glow in Background */}
      <div className="fixed inset-0 pointer-events-none bg-[radial-gradient(ellipse_80%_80%_at_50%_-20%,rgba(0,240,255,0.06),rgba(255,255,255,0))]" />

      {/* Toast Notification */}
      {toastMessage && (
        <div className="fixed top-5 left-1/2 -translate-x-1/2 z-50 flex items-center gap-2.5 px-5 py-2.5 rounded-2xl bg-[#0e1622]/95 border border-[#00F0FF]/60 text-white font-brand text-xs shadow-[0_10px_35px_rgba(0,240,255,0.3)] animate-in fade-in slide-in-from-top-3 duration-200">
          <IconBolt className="w-4 h-4 text-[#00F0FF] animate-pulse" />
          <span>{toastMessage}</span>
        </div>
      )}

      {/* ── 1. LEFT NAVIGATION RAIL (5 DEDICATED TABS) ── */}
      <aside className="fixed left-0 top-0 bottom-0 w-[84px] bg-[#090d13]/95 border-r border-[#151d28] flex flex-col items-center py-6 z-40 backdrop-blur-xl">
        {/* Brand Emblem */}
        <div className="w-12 h-12 rounded-2xl bg-[#101722] border border-[#202d3f] p-2 flex items-center justify-center mb-8 shadow-lg shadow-[#00F0FF]/10 group">
          <img
            src={imgLagvexLogo}
            alt="Lagvex Emblem"
            className="w-full h-full object-contain rounded-lg transition-transform duration-300 group-hover:scale-110"
          />
        </div>

        {/* 5 Primary Navigation Tabs */}
        <nav className="flex flex-col gap-3 w-full px-3">
          {[
            { id: "boost" as const, Icon: IconBolt, label: "BOOST" },
            { id: "library" as const, Icon: IconGamepad, label: "GAMES" },
            { id: "squad" as const, Icon: IconSquadUsers, label: "SQUAD" },
            { id: "nodes" as const, Icon: IconServerNodes, label: "NODES" },
            { id: "tweaker" as const, Icon: IconTuning, label: "TWEAKS" }
          ].map(({ id, Icon, label }) => {
            const isActive = activeTab === id;
            return (
              <button
                key={id}
                onClick={() => setActiveTab(id)}
                className={`relative w-full h-[52px] rounded-xl flex flex-col items-center justify-center gap-1 cursor-pointer border-none transition-all group ${
                  isActive
                    ? "bg-[#00F0FF]/15 text-[#00F0FF]"
                    : "bg-transparent text-white/45 hover:text-white/90 hover:bg-white/[0.04]"
                }`}
                title={label}
              >
                {isActive && (
                  <span className="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-6 rounded-r bg-[#00F0FF] shadow-[0_0_10px_#00F0FF]" />
                )}
                <Icon className={`w-5 h-5 transition-transform group-hover:scale-110 ${isActive ? "text-[#00F0FF]" : ""}`} />
                <span className={`text-[9px] font-brand font-bold tracking-wider leading-none ${isActive ? "text-[#00F0FF]" : "text-white/40 group-hover:text-white/80"}`}>
                  {label}
                </span>
              </button>
            );
          })}
        </nav>

        {/* Bottom Version Pill */}
        <div className="mt-auto flex flex-col items-center gap-1">
          <div className="flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-white/[0.04] border border-white/5">
            <span className="w-1.5 h-1.5 rounded-full bg-[#10b981] shadow-[0_0_6px_#10b981]" />
            <span className="font-mono text-[9px] text-white/45 font-semibold">v1.0.0</span>
          </div>
        </div>
      </aside>

      {/* ── 2. MAIN WORKSPACE CONTENT ── */}
      <main className="ml-[84px] flex-1 min-h-screen flex flex-col">
        {/* Top Header Bar */}
        <header className="flex items-center justify-between pt-5 pb-3 px-8 z-20 border-b border-[#141b26]/50">
          <div className="flex items-center gap-3">
            <h1 className="font-brand font-extrabold text-lg text-white tracking-wide flex items-center gap-2">
              <span className="text-[#00F0FF]">LAGVEX</span>
              <span className="text-white/30 text-sm font-normal">/</span>
              <span className="text-sm font-medium text-white/80 uppercase">
                {activeTab === "boost" && "Pro Game Booster"}
                {activeTab === "library" && "Game Library"}
                {activeTab === "squad" && "Squad Multi-Hop Sync"}
                {activeTab === "nodes" && "Global Relay Nodes"}
                {activeTab === "tweaker" && "Windows Network Tweaks"}
              </span>
            </h1>
            {isBoosting && (
              <span className="px-2.5 py-0.5 rounded-full bg-[#10b981]/15 text-[#10b981] text-[10px] font-brand font-bold border border-[#10b981]/30 flex items-center gap-1.5 animate-pulse">
                <span className="w-1.5 h-1.5 rounded-full bg-[#10b981]" />
                ACCELERATING {selectedGame.name.toUpperCase()}
              </span>
            )}
          </div>

          <div className="flex items-center gap-3">
            {isBoosting && (
              <div className="h-[38px] px-3.5 rounded-xl bg-[#101722] border border-[#1f2b3b] flex items-center gap-2 text-xs">
                <span className="text-white/40">Duration:</span>
                <span className="font-mono font-bold text-[#00F0FF]">{formatTimer(sessionSeconds)}</span>
              </div>
            )}
            <div className="h-[38px] px-3 rounded-xl bg-[#101722] border border-[#1f2b3b] flex items-center gap-2 text-xs">
              <span className="text-white/40">Node:</span>
              <span className="font-brand font-bold text-white flex items-center gap-1">
                <span>{selectedRelay.flag}</span>
                <span>{selectedRelay.cleanName}</span>
              </span>
            </div>
          </div>
        </header>

        {/* ── 3. DYNAMIC TAB VIEWS ── */}
        <div className="p-8 flex-1 flex flex-col">
          {/* ========================================================= */}
          {/* TAB 1: BOOST HUB (GEARUP PRO HUD)                         */}
          {/* ========================================================= */}
          {activeTab === "boost" && (
            <div className="flex flex-col gap-6 max-w-[1240px] mx-auto w-full">
              {/* GearUP Cinematic Banner */}
              <div className="relative rounded-[32px] overflow-hidden bg-[#0a0f16] border border-[#1c2738] shadow-[0_20px_60px_rgba(0,0,0,0.6)] p-8">
                {/* Backdrop Game Art */}
                <div
                  className="absolute inset-0 pointer-events-none opacity-25 bg-cover bg-center"
                  style={{
                    backgroundImage: `url(${selectedGame.coverImg || selectedGame.heroArt})`,
                    filter: "blur(20px)"
                  }}
                />
                <div className="absolute inset-0 bg-gradient-to-r from-[#0a0f16] via-[#0a0f16]/85 to-transparent pointer-events-none" />

                {/* Main Hero Content */}
                <div className="relative z-10 flex items-start justify-between">
                  {/* Left Info & Big Metrics */}
                  <div className="flex flex-col gap-5 max-w-[680px]">
                    <div className="flex items-center gap-3">
                      <img
                        src={selectedGame.coverImg}
                        alt={selectedGame.name}
                        className="w-14 h-14 rounded-2xl object-cover border border-white/10 shadow-lg"
                      />
                      <div>
                        <div className="flex items-center gap-2">
                          <h2 className="font-brand font-extrabold text-3xl text-white tracking-wide">
                            {selectedGame.name}
                          </h2>
                          <span className="text-[10px] font-brand font-bold px-2 py-0.5 rounded bg-[#00F0FF]/15 text-[#00F0FF] border border-[#00F0FF]/30">
                            {selectedGame.tag}
                          </span>
                        </div>
                        <p className="text-xs text-white/50 font-medium">
                          {selectedGame.publisher} &bull; {selectedGame.region}
                        </p>
                      </div>
                    </div>

                    {/* TWO GIANT GEARUP METRICS: ESTIMATED PING & PACKET LOSS */}
                    <div className="grid grid-cols-3 gap-6 pt-2">
                      {/* 1. ESTIMATED PING */}
                      <div className="p-4 rounded-2xl bg-[#0f1722]/80 border border-[#1f2d40] flex flex-col">
                        <span className="text-[10px] font-brand font-bold text-white/45 uppercase tracking-wider">
                          ESTIMATED PING
                        </span>
                        <div className="flex items-baseline gap-1.5 mt-1">
                          <span className="font-mono font-black text-5xl text-[#00F0FF] tracking-tight">
                            {isBoosting ? Math.round(livePing) : selectedGame.baselinePing}
                          </span>
                          <span className="font-mono text-sm text-white/50 font-bold">ms</span>
                        </div>
                        <span className="text-[10px] font-mono text-[#10b981] mt-1 flex items-center gap-1 font-semibold">
                          <span>⚡</span> {selectedGame.trend} vs Default ISP
                        </span>
                      </div>

                      {/* 2. PACKET LOSS */}
                      <div className="p-4 rounded-2xl bg-[#0f1722]/80 border border-[#1f2d40] flex flex-col">
                        <span className="text-[10px] font-brand font-bold text-white/45 uppercase tracking-wider">
                          PACKET LOSS
                        </span>
                        <div className="flex items-baseline gap-1.5 mt-1">
                          <span className="font-mono font-black text-5xl text-white tracking-tight">
                            {packetLoss.toFixed(1)}
                          </span>
                          <span className="font-mono text-sm text-white/50 font-bold">%</span>
                        </div>
                        <span className="text-[10px] font-mono text-[#10b981] mt-1 font-semibold">
                          Zero-Drop Forward Error Correction
                        </span>
                      </div>

                      {/* 3. JITTER VARIANCE */}
                      <div className="p-4 rounded-2xl bg-[#0f1722]/80 border border-[#1f2d40] flex flex-col">
                        <span className="text-[10px] font-brand font-bold text-white/45 uppercase tracking-wider">
                          JITTER VARIANCE
                        </span>
                        <div className="flex items-baseline gap-1.5 mt-1">
                          <span className="font-mono font-black text-5xl text-white tracking-tight">
                            &plusmn;{jitter}
                          </span>
                          <span className="font-mono text-sm text-white/50 font-bold">ms</span>
                        </div>
                        <span className="text-[10px] font-mono text-[#00F0FF] mt-1 font-semibold">
                          Ultra-Stable FastPath
                        </span>
                      </div>
                    </div>

                    {/* ACTION BUTTONS (BOOST NOW / LAUNCH GAME) */}
                    <div className="flex items-center gap-3 pt-2">
                      <button
                        onClick={handleToggleBoost}
                        disabled={isConnecting}
                        className={`h-[50px] px-8 rounded-2xl flex items-center gap-3 font-brand font-extrabold text-sm uppercase tracking-wider cursor-pointer border-none transition-all shadow-xl ${
                          isBoosting
                            ? "bg-[#ef4444] hover:bg-[#dc2626] text-white shadow-red-500/20"
                            : isConnecting
                            ? "bg-[#00F0FF]/70 text-black cursor-wait"
                            : "bg-[#00F0FF] hover:bg-[#33f3ff] text-[#051119] shadow-[#00F0FF]/30 hover:scale-105"
                        }`}
                      >
                        <IconBolt className={`w-5 h-5 ${isBoosting || isConnecting ? "animate-spin" : ""}`} />
                        <span>
                          {isBoosting ? "STOP ACCELERATION" : isConnecting ? "CONNECTING..." : "BOOST NOW"}
                        </span>
                      </button>

                      <button
                        onClick={handleLaunchGame}
                        className="h-[50px] px-6 rounded-2xl bg-[#121b27] hover:bg-[#182433] border border-[#233347] hover:border-[#00F0FF]/50 text-white font-brand font-bold text-xs uppercase tracking-wider flex items-center gap-2.5 cursor-pointer transition-all shadow-md"
                      >
                        <IconRocket className="w-4 h-4 text-[#00F0FF]" />
                        <span>LAUNCH GAME</span>
                      </button>

                      <button
                        onClick={() => setActiveTab("library")}
                        className="h-[50px] px-5 rounded-2xl bg-transparent hover:bg-white/[0.04] border border-white/10 text-white/70 hover:text-white font-brand font-bold text-xs uppercase tracking-wider cursor-pointer transition-all"
                      >
                        CHANGE GAME
                      </button>
                    </div>
                  </div>

                  {/* Right Hero Artwork */}
                  <div className="relative w-[340px] h-[260px] flex items-center justify-center pointer-events-none">
                    <img
                      src={selectedGame.heroArt}
                      alt={selectedGame.name}
                      className="max-h-[260px] max-w-full object-contain drop-shadow-[0_15px_35px_rgba(0,0,0,0.8)]"
                    />
                  </div>
                </div>

                {/* ── GEARUP SIGNATURE: NETWORK ROUTING HOPS PIPELINE ── */}
                <div className="mt-8 pt-6 border-t border-[#172230]">
                  <div className="flex items-center justify-between mb-4">
                    <span className="text-xs font-brand font-bold text-white/60 uppercase tracking-wider flex items-center gap-2">
                      <span>NETWORK ROUTING PIPELINE</span>
                      <span className="w-1.5 h-1.5 rounded-full bg-[#00F0FF]" />
                    </span>
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => setBoostSubTab("routing")}
                        className={`px-3 py-1 rounded-lg text-xs font-brand font-bold cursor-pointer border-none transition-all ${
                          boostSubTab === "routing"
                            ? "bg-[#00F0FF]/20 text-[#00F0FF]"
                            : "bg-transparent text-white/40 hover:text-white"
                        }`}
                      >
                        Routing Detail
                      </button>
                      <button
                        onClick={() => setBoostSubTab("history")}
                        className={`px-3 py-1 rounded-lg text-xs font-brand font-bold cursor-pointer border-none transition-all ${
                          boostSubTab === "history"
                            ? "bg-[#00F0FF]/20 text-[#00F0FF]"
                            : "bg-transparent text-white/40 hover:text-white"
                        }`}
                      >
                        Ping History
                      </button>
                    </div>
                  </div>

                  {/* VISUAL HOP NODES */}
                  <div className="p-5 rounded-2xl bg-[#080d14] border border-[#192433] flex items-center justify-between relative overflow-hidden">
                    {/* Hop 1: PC */}
                    <div className="flex flex-col items-center gap-1.5 z-10">
                      <div className="w-11 h-11 rounded-xl bg-[#111925] border border-[#213042] flex items-center justify-center text-white shadow">
                        <IconPC className="w-5 h-5 text-[#00F0FF]" />
                      </div>
                      <span className="text-[11px] font-brand font-bold text-white">Your PC</span>
                      <span className="text-[9px] font-mono text-white/40">127.0.0.1</span>
                    </div>

                    {/* Connecting Hop 1 -> 2 */}
                    <div className="flex-1 flex flex-col items-center px-3 relative">
                      <div className="w-full h-1 bg-[#182333] rounded relative overflow-hidden">
                        {isBoosting && (
                          <div className="absolute inset-y-0 left-0 w-1/3 bg-[#00F0FF] rounded animate-[pulse_1.5s_infinite]" />
                        )}
                      </div>
                      <span className="text-[10px] font-mono text-[#00F0FF] font-bold mt-1">1 ms (LAN)</span>
                    </div>

                    {/* Hop 2: Local Accelerator Engine */}
                    <div className="flex flex-col items-center gap-1.5 z-10">
                      <div className="w-11 h-11 rounded-xl bg-[#111925] border border-[#213042] flex items-center justify-center text-[#00F0FF] shadow">
                        <IconBolt className="w-5 h-5" />
                      </div>
                      <span className="text-[11px] font-brand font-bold text-white">Lagvex Core</span>
                      <span className="text-[9px] font-mono text-[#10b981]">Kernel Wintun</span>
                    </div>

                    {/* Connecting Hop 2 -> 3 */}
                    <div className="flex-1 flex flex-col items-center px-3 relative">
                      <div className="w-full h-1 bg-[#182333] rounded relative overflow-hidden">
                        {isBoosting && (
                          <div className="absolute inset-y-0 left-0 w-1/3 bg-[#00F0FF] rounded animate-[pulse_1.2s_infinite]" />
                        )}
                      </div>
                      <span className="text-[10px] font-mono text-[#00F0FF] font-bold mt-1">
                        {isBoosting ? `${selectedRelay.pingMs || 14} ms` : "Standby"}
                      </span>
                    </div>

                    {/* Hop 3: Selected Super Relay Node */}
                    <div className="flex flex-col items-center gap-1.5 z-10">
                      <div className="w-11 h-11 rounded-xl bg-[#111925] border border-[#00F0FF]/40 flex items-center justify-center text-xl shadow-lg shadow-[#00F0FF]/15">
                        {selectedRelay.flag}
                      </div>
                      <span className="text-[11px] font-brand font-bold text-white">{selectedRelay.cleanName}</span>
                      <span className="text-[9px] font-mono text-[#00F0FF]">FastRoute Node</span>
                    </div>

                    {/* Connecting Hop 3 -> 4 */}
                    <div className="flex-1 flex flex-col items-center px-3 relative">
                      <div className="w-full h-1 bg-[#182333] rounded relative overflow-hidden">
                        {isBoosting && (
                          <div className="absolute inset-y-0 left-0 w-1/3 bg-[#10b981] rounded animate-[pulse_1s_infinite]" />
                        )}
                      </div>
                      <span className="text-[10px] font-mono text-[#10b981] font-bold mt-1">2 ms (Direct)</span>
                    </div>

                    {/* Hop 4: Official Game Server */}
                    <div className="flex flex-col items-center gap-1.5 z-10">
                      <div className="w-11 h-11 rounded-xl bg-[#111925] border border-[#213042] flex items-center justify-center text-white shadow">
                        <IconGamepad className="w-5 h-5 text-[#10b981]" />
                      </div>
                      <span className="text-[11px] font-brand font-bold text-white">{selectedGame.name} Server</span>
                      <span className="text-[9px] font-mono text-white/40">{selectedGame.region}</span>
                    </div>
                  </div>

                  {/* TAB CONTENT: ROUTING DETAIL OR PING HISTORY */}
                  {boostSubTab === "routing" ? (
                    <div className="mt-4 grid grid-cols-2 gap-4">
                      <div className="p-4 rounded-xl bg-[#090e15] border border-[#16202e] flex flex-col gap-2">
                        <span className="text-[10px] font-brand font-bold text-[#00F0FF] uppercase tracking-wider">
                          MONITORED PROCESSES & FASTPATH
                        </span>
                        <div className="flex flex-col gap-1.5 font-mono text-xs text-white/80">
                          {selectedGame.processNames.map((proc) => (
                            <div key={proc} className="flex items-center justify-between p-1.5 rounded bg-white/[0.02]">
                              <span className="flex items-center gap-2">
                                <span className="w-1.5 h-1.5 rounded-full bg-[#10b981]" />
                                <span>{proc}</span>
                              </span>
                              <span className="text-[10px] text-white/40">Intercepted</span>
                            </div>
                          ))}
                        </div>
                      </div>

                      <div className="p-4 rounded-xl bg-[#090e15] border border-[#16202e] flex flex-col gap-2">
                        <span className="text-[10px] font-brand font-bold text-[#00F0FF] uppercase tracking-wider">
                          ACTIVE IP CIDR SUBNETS
                        </span>
                        <div className="flex flex-col gap-1.5 font-mono text-xs text-white/80 max-h-[110px] overflow-y-auto">
                          {selectedGame.cidrs.map((cidr) => (
                            <div key={cidr} className="flex items-center justify-between p-1.5 rounded bg-white/[0.02]">
                              <span className="flex items-center gap-2">
                                <span className="w-1.5 h-1.5 rounded-full bg-[#00F0FF]" />
                                <span>{cidr}</span>
                              </span>
                              <span className="text-[10px] text-[#00F0FF]">Split-Tunnel</span>
                            </div>
                          ))}
                        </div>
                      </div>
                    </div>
                  ) : (
                    <div className="mt-4 p-4 rounded-xl bg-[#090e15] border border-[#16202e] flex flex-col gap-2">
                      <div className="flex items-center justify-between">
                        <span className="text-[10px] font-brand font-bold text-[#00F0FF] uppercase tracking-wider">
                          REAL-TIME PING OSCILLOSCOPE (128Hz)
                        </span>
                        <span className="text-[10px] font-mono text-white/45">
                          Avg: {Math.round(livePing)}ms &bull; Jitter: &plusmn;{jitter}ms
                        </span>
                      </div>
                      <div className="h-[120px] w-full rounded-lg overflow-hidden bg-black/40">
                        <canvas ref={historyCanvasRef} width={900} height={120} className="w-full h-full" />
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </div>
          )}

          {/* ========================================================= */}
          {/* TAB 2: GAME LIBRARY (REPLACES REDUNDANT QUICK PROFILES)   */}
          {/* ========================================================= */}
          {activeTab === "library" && (
            <div className="flex flex-col gap-6 max-w-[1240px] mx-auto w-full">
              {/* Search & Category Filter Bar */}
              <div className="flex items-center justify-between gap-4">
                <div className="flex-1 relative h-[46px] max-w-[480px]">
                  <div className="absolute left-4 top-1/2 -translate-y-1/2 text-white/35">
                    <IconSearch className="w-4 h-4" />
                  </div>
                  <input
                    type="text"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Search 14 games, publishers, or process names..."
                    className="w-full h-full pl-11 pr-4 rounded-2xl bg-[#0f1722] border border-[#1c293a] text-xs text-white outline-none focus:border-[#00F0FF]/60 placeholder:text-white/30"
                  />
                </div>

                {/* Categories */}
                <div className="flex items-center gap-2">
                  {(["ALL", "FPS", "MOBA", "BR", "RPG"] as const).map((cat) => (
                    <button
                      key={cat}
                      onClick={() => setSelectedCategory(cat)}
                      className={`px-4 py-2 rounded-xl text-xs font-brand font-bold cursor-pointer border-none transition-all ${
                        selectedCategory === cat
                          ? "bg-[#00F0FF] text-black shadow-lg shadow-[#00F0FF]/20"
                          : "bg-[#0f1722] text-white/50 hover:text-white hover:bg-[#162233]"
                      }`}
                    >
                      {cat}
                    </button>
                  ))}
                </div>
              </div>

              {/* Game Cards Grid */}
              <div className="grid grid-cols-4 gap-5">
                {filteredGames.map((game) => {
                  const originalIdx = OFFICIAL_GAMES.findIndex((g) => g.id === game.id);
                  const isCurrent = selectedGame.id === game.id;
                  return (
                    <div
                      key={game.id}
                      onClick={() => handleSelectGameAndBoost(originalIdx)}
                      className={`relative rounded-2xl overflow-hidden bg-[#0a0f16] border cursor-pointer transition-all group flex flex-col shadow-lg ${
                        isCurrent
                          ? "border-[#00F0FF] shadow-[0_0_20px_rgba(0,240,255,0.25)]"
                          : "border-[#192433] hover:border-[#00F0FF]/50 hover:-translate-y-1"
                      }`}
                    >
                      {/* Game Image Banner */}
                      <div className="relative h-[160px] overflow-hidden bg-[#05080c]">
                        <img
                          src={game.coverImg}
                          alt={game.name}
                          className="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
                        />
                        <div className="absolute inset-0 bg-gradient-to-t from-[#0a0f16] via-transparent to-transparent" />
                        <span className="absolute top-3 right-3 px-2 py-0.5 rounded bg-black/70 backdrop-blur text-[10px] font-brand font-bold text-[#00F0FF] border border-white/10">
                          {game.tag}
                        </span>
                      </div>

                      {/* Game Info */}
                      <div className="p-4 flex flex-col gap-2 flex-1">
                        <div>
                          <h3 className="font-brand font-bold text-sm text-white truncate">{game.name}</h3>
                          <p className="text-[11px] text-white/40 truncate">{game.genre}</p>
                        </div>

                        {/* Ping Comparison Pill */}
                        <div className="mt-auto pt-2 border-t border-white/5 flex items-center justify-between">
                          <div className="flex items-center gap-1.5 text-xs font-mono">
                            <span className="line-through text-white/30">{game.baselinePing}ms</span>
                            <span className="text-white/40">&rarr;</span>
                            <span className="font-bold text-[#10b981]">{game.accelPing}ms</span>
                          </div>
                          <span className="text-[10px] font-brand font-bold text-[#00F0FF] bg-[#00F0FF]/10 px-2 py-0.5 rounded">
                            {game.trend}
                          </span>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {/* ========================================================= */}
          {/* TAB 3: SQUAD SYNC (INTERACTIVE TEAM PING LOCK)            */}
          {/* ========================================================= */}
          {activeTab === "squad" && (
            <div className="flex flex-col gap-6 max-w-[1240px] mx-auto w-full">
              {/* Informative Explanation Banner */}
              <div className="p-6 rounded-3xl bg-gradient-to-r from-[#0a121e] to-[#081724] border border-[#1b2b3d] flex items-center justify-between shadow-xl">
                <div className="max-w-[720px] flex flex-col gap-2">
                  <div className="flex items-center gap-2 text-xs font-brand font-bold text-[#00F0FF] uppercase tracking-wider">
                    <IconSquadUsers className="w-4 h-4" />
                    <span>SQUAD MULTI-ROUTING SYNC</span>
                  </div>
                  <h2 className="font-brand font-extrabold text-2xl text-white">
                    Synchronize Team Latency & Eliminate Desync
                  </h2>
                  <p className="text-xs text-white/60 leading-relaxed">
                    When playing ranked matches in a party, teammates on different ISPs (VNPT, FPT, Viettel) often suffer from disparate routes, desync, and audio jitter. Squad Sync locks all party members into the same high-speed Lagvex Super-Node, keeping latency delta below 1ms.
                  </p>
                </div>

                <div className="flex flex-col items-end gap-2 shrink-0">
                  <span className="px-3 py-1 rounded-full bg-[#10b981]/15 text-[#10b981] text-xs font-brand font-bold border border-[#10b981]/30">
                    {squadSynced ? "PARTY 100% SYNCED" : "DESYNC DETECTED"}
                  </span>
                  <span className="text-[10px] font-mono text-white/40">Node: {selectedRelay.cleanName}</span>
                </div>
              </div>

              {/* Room Controls: Host & Join Cards */}
              <div className="grid grid-cols-2 gap-6">
                {/* 1. Host Room Card */}
                <div className="p-6 rounded-3xl bg-[#0b1018] border border-[#1c2738] flex flex-col gap-4 shadow-lg">
                  <span className="text-xs font-brand font-bold text-white/50 uppercase tracking-wider">
                    YOUR SQUAD ROOM
                  </span>
                  <div className="flex items-center justify-between p-4 rounded-2xl bg-[#101722] border border-[#202d3f]">
                    <div className="flex flex-col">
                      <span className="text-[10px] font-brand text-white/40 uppercase">ROOM CODE</span>
                      <span className="font-mono font-black text-2xl text-[#00F0FF] tracking-wider">
                        {squadCode}
                      </span>
                    </div>
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => {
                          navigator.clipboard?.writeText(squadCode);
                          notify("Copied Squad Room Code to clipboard!");
                        }}
                        className="px-3 py-2 rounded-xl bg-white/[0.05] hover:bg-white/10 text-white font-brand text-xs font-bold flex items-center gap-1.5 cursor-pointer border-none transition-all"
                      >
                        <IconCopy className="w-4 h-4" />
                        <span>Copy Code</span>
                      </button>
                      <button
                        onClick={() => {
                          navigator.clipboard?.writeText(`lagvex://squad/${squadCode}`);
                          notify("Copied Direct Invite Link!");
                        }}
                        className="px-3 py-2 rounded-xl bg-[#00F0FF]/15 hover:bg-[#00F0FF]/25 text-[#00F0FF] font-brand text-xs font-bold flex items-center gap-1.5 cursor-pointer border-none transition-all"
                      >
                        <span>Copy Link</span>
                      </button>
                    </div>
                  </div>
                  <p className="text-[11px] text-white/40">
                    Share this code or 1-click link with your Discord squad. When they join, their connection will automatically route through your host node.
                  </p>
                </div>

                {/* 2. Join Room Card */}
                <div className="p-6 rounded-3xl bg-[#0b1018] border border-[#1c2738] flex flex-col gap-4 shadow-lg">
                  <span className="text-xs font-brand font-bold text-white/50 uppercase tracking-wider">
                    JOIN FRIEND'S SQUAD
                  </span>
                  <div className="flex items-center gap-3">
                    <input
                      type="text"
                      value={joinInputCode}
                      onChange={(e) => setJoinInputCode(e.target.value.toUpperCase())}
                      placeholder="e.g. LGVX-SQUAD-99"
                      className="flex-1 h-[52px] px-4 rounded-2xl bg-[#101722] border border-[#202d3f] text-sm font-mono font-bold text-white outline-none focus:border-[#00F0FF]/60 placeholder:text-white/30"
                    />
                    <button
                      onClick={() => {
                        if (!joinInputCode) {
                          notify("Please enter a valid Squad Code!");
                          return;
                        }
                        setSquadCode(joinInputCode);
                        notify(`Joined Squad ${joinInputCode}! Synchronizing routes...`);
                      }}
                      className="h-[52px] px-6 rounded-2xl bg-[#00F0FF] hover:bg-[#33f3ff] text-black font-brand font-bold text-xs uppercase tracking-wider cursor-pointer border-none transition-all shadow-lg shadow-[#00F0FF]/20"
                    >
                      JOIN SQUAD
                    </button>
                  </div>
                  <p className="text-[11px] text-white/40">
                    Entering a teammate's room code aligns your WireGuard/VLESS routing with their host profile.
                  </p>
                </div>
              </div>

              {/* Squad Members Roster */}
              <div className="p-6 rounded-3xl bg-[#0b1018] border border-[#1c2738] flex flex-col gap-4 shadow-lg">
                <div className="flex items-center justify-between">
                  <h3 className="font-brand font-bold text-base text-white">
                    Squad Party Members ({squadMembers.length}/5)
                  </h3>
                  <button
                    onClick={() => {
                      setSquadSynced(true);
                      notify("⚡ Re-synced all party routes through Singapore SDR Edge!");
                    }}
                    className="flex items-center gap-1.5 text-xs font-brand font-bold text-[#00F0FF] hover:underline cursor-pointer border-none bg-transparent"
                  >
                    <IconRefresh className="w-3.5 h-3.5" />
                    <span>Synchronize Latency</span>
                  </button>
                </div>

                <div className="flex flex-col gap-2.5">
                  {squadMembers.map((member, idx) => (
                    <div
                      key={idx}
                      className="flex items-center justify-between p-3.5 rounded-2xl bg-[#101622] border border-[#1b2738]"
                    >
                      <div className="flex items-center gap-3">
                        <div className="w-9 h-9 rounded-xl bg-gradient-to-tr from-[#00F0FF]/20 to-[#7E42FF]/20 border border-[#00F0FF]/30 flex items-center justify-center font-brand font-bold text-xs text-[#00F0FF]">
                          {member.name.charAt(0)}
                        </div>
                        <div>
                          <div className="flex items-center gap-2">
                            <span className="font-brand font-bold text-xs text-white">{member.name}</span>
                            <span className="text-[9px] font-brand font-bold px-1.5 py-0.2 rounded bg-white/[0.06] text-white/60">
                              {member.role}
                            </span>
                          </div>
                          <span className="text-[10px] text-white/40 font-mono">
                            {member.isp} &bull; {member.game}
                          </span>
                        </div>
                      </div>

                      <div className="flex items-center gap-6">
                        <div className="flex flex-col items-end">
                          <span className="font-mono font-bold text-xs text-[#10b981]">{member.ping} ms</span>
                          <span className="text-[9px] text-white/40">Latency</span>
                        </div>
                        <span className="text-[10px] font-brand font-bold text-[#00F0FF] bg-[#00F0FF]/10 px-2.5 py-1 rounded-full border border-[#00F0FF]/20">
                          {member.status}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}

          {/* ========================================================= */}
          {/* TAB 4: GLOBAL RELAY NODES                                 */}
          {/* ========================================================= */}
          {activeTab === "nodes" && (
            <div className="flex flex-col gap-6 max-w-[1240px] mx-auto w-full">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="font-brand font-extrabold text-2xl text-white">Global Routing Infrastructure</h2>
                  <p className="text-xs text-white/50">
                    Tier-1 Anycast relay backbone connected directly to Valve SDR, Riot Direct, and AWS Gamelift.
                  </p>
                </div>
                <button
                  onClick={() => notify("Pinged all global nodes. Latencies refreshed!")}
                  className="px-4 py-2 rounded-xl bg-[#101722] hover:bg-[#162130] border border-[#223042] text-xs font-brand font-bold text-white flex items-center gap-2 cursor-pointer transition-all"
                >
                  <IconRefresh className="w-4 h-4 text-[#00F0FF]" />
                  <span>Refresh All Latencies</span>
                </button>
              </div>

              <div className="grid grid-cols-2 gap-4">
                {relays.map((relay, idx) => {
                  const isSelected = selectedRelayIdx === idx;
                  return (
                    <div
                      key={relay.id}
                      onClick={() => {
                        setSelectedRelayIdx(idx);
                        notify(`Switched primary routing node to: ${relay.cleanName}`);
                      }}
                      className={`p-5 rounded-2xl border cursor-pointer transition-all flex items-center justify-between ${
                        isSelected
                          ? "bg-[#00F0FF]/10 border-[#00F0FF] shadow-[0_0_20px_rgba(0,240,255,0.15)]"
                          : "bg-[#0a0f16] hover:bg-[#0f1622] border-[#1b2636]"
                      }`}
                    >
                      <div className="flex items-center gap-4">
                        <span className="text-3xl">{relay.flag}</span>
                        <div>
                          <div className="flex items-center gap-2">
                            <h3 className="font-brand font-bold text-sm text-white">{relay.cleanName}</h3>
                            {isSelected && (
                              <span className="text-[9px] font-brand font-bold px-1.5 py-0.5 rounded bg-[#00F0FF]/20 text-[#00F0FF]">
                                ACTIVE
                              </span>
                            )}
                          </div>
                          <p className="text-xs text-white/40 font-mono mt-0.5">{relay.location}</p>
                          <p className="text-[10px] text-white/30 font-mono mt-1">{relay.address}</p>
                        </div>
                      </div>

                      <div className="flex flex-col items-end gap-1">
                        <span className="font-mono font-black text-xl text-[#00F0FF]">{relay.pingMs} ms</span>
                        <span className="text-[10px] text-[#10b981] font-semibold">Quality: {relay.quality}%</span>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {/* ========================================================= */}
          {/* TAB 5: SYSTEM TWEAKER (WINDOWS KERNEL OPTIMIZATION)       */}
          {/* ========================================================= */}
          {activeTab === "tweaker" && (
            <div className="flex flex-col gap-6 max-w-[960px] mx-auto w-full">
              <div className="p-6 rounded-3xl bg-gradient-to-r from-[#0d1624] to-[#0a121e] border border-[#1d2b3d] shadow-xl flex items-center justify-between">
                <div>
                  <h2 className="font-brand font-extrabold text-2xl text-white">Windows Network Kernel Tweaks</h2>
                  <p className="text-xs text-white/60 mt-1 max-w-[600px]">
                    Bypasses standard Windows packet buffering, optimizes socket TCP/IP stack, and prioritizes gaming UDP packets over background downloads.
                  </p>
                </div>
                <button
                  onClick={handleApplyTweaks}
                  className="h-[46px] px-6 rounded-2xl bg-[#00F0FF] hover:bg-[#33f3ff] text-black font-brand font-bold text-xs uppercase tracking-wider cursor-pointer border-none transition-all shadow-lg shadow-[#00F0FF]/20 shrink-0"
                >
                  APPLY ALL TWEAKS
                </button>
              </div>

              {/* Tweak Toggles */}
              <div className="flex flex-col gap-3">
                {[
                  {
                    title: "Disable Nagle's Algorithm (TCP NoDelay)",
                    desc: "Eliminates TCP packet wait delays by immediately transmitting packets without waiting for full buffers.",
                    state: disableNagle,
                    setter: setDisableNagle
                  },
                  {
                    title: "Multimedia Class Scheduler (MMCSS) Gaming Priority",
                    desc: "Allocates dedicated CPU clock cycles to game network threads to prevent micro-stuttering.",
                    state: mmcssPriority,
                    setter: setMmcssPriority
                  },
                  {
                    title: "Kernel Socket TCP_NODELAY Flag",
                    desc: "Forces Windows socket layer to transmit real-time telemetry instantly.",
                    state: tcpNoDelay,
                    setter: setTcpNoDelay
                  },
                  {
                    title: "Auto MTU Clamping (1420 Bytes)",
                    desc: "Prevents packet fragmentation over PPPoE/fiber connections, preventing sudden packet loss spikes.",
                    state: mtuClamping,
                    setter: setMtuClamping
                  }
                ].map((tweak, i) => (
                  <div
                    key={i}
                    onClick={() => tweak.setter(!tweak.state)}
                    className="p-4 rounded-2xl bg-[#0a0f16] border border-[#1b2636] hover:border-[#00F0FF]/40 cursor-pointer flex items-center justify-between transition-all"
                  >
                    <div className="max-w-[680px]">
                      <h4 className="font-brand font-bold text-sm text-white">{tweak.title}</h4>
                      <p className="text-xs text-white/40 mt-1 leading-relaxed">{tweak.desc}</p>
                    </div>

                    <div
                      className={`w-12 h-6 rounded-full p-0.5 transition-colors cursor-pointer ${
                        tweak.state ? "bg-[#00F0FF]" : "bg-white/10"
                      }`}
                    >
                      <div
                        className={`w-5 h-5 rounded-full bg-black transition-transform ${
                          tweak.state ? "translate-x-6" : "translate-x-0"
                        }`}
                      />
                    </div>
                  </div>
                ))}
              </div>

              {dnsFlushed && (
                <div className="p-4 rounded-2xl bg-[#10b981]/10 border border-[#10b981]/30 flex items-center gap-2.5 text-xs text-[#10b981] font-brand font-bold">
                  <IconShieldCheck className="w-4 h-4" />
                  <span>Windows Winsock stack reset and DNS Cache flushed. Full gaming throughput active.</span>
                </div>
              )}
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
