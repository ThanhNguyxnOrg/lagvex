import { useState, useEffect, useRef, useMemo } from "react";
import FloatingHUD from "./FloatingHUD";

// ==================== BRAND ASSETS ====================
const imgLagvexLogo = "/assets/logo.jpg";

// ==================== RAZOR-SHARP VECTOR GAMING ICONS ====================
function IconHome({ className = "w-5 h-5" }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
      <polyline points="9 22 9 12 15 12 15 22" />
    </svg>
  );
}

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

function IconActivity({ className = "w-4 h-4" }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
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

function IconEdit({ className = "w-3.5 h-3.5" }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z" />
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

function IconPower({ className = "w-4 h-4" }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M18.36 6.64a9 9 0 1 1-12.73 0" />
      <line x1="12" y1="2" x2="12" y2="12" />
    </svg>
  );
}

// ==================== GAME PRESENTATION METADATA ====================
// Regions and CIDRs are loaded from the backend profile catalog.
export type GameRegion = {
  id: string;
  name: string;
  cidrs: string[];
};

export type GameProfile = {
  id: string;
  name: string;
  category: string;
  processNames: string[];
  regions: GameRegion[];
};

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
  region: string;
  regions?: GameRegion[];
  profileAvailable?: boolean;
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
    region: "Asia-Pacific (Singapore)",
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
    region: "Hong Kong / Singapore SDR",
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
    region: "Southeast Asia (AWS SEA)",
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
    region: "Tokyo & Singapore Datacenter",
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
    region: "Asia-Pacific (Regional Low-Latency)",
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
    region: "Southeast Asia Dedicated",
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
    region: "SEA Valve Backbone",
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
    region: "Asia Central Edge",
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
    region: "Asia Central GCE",
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
    region: "Taiwan / Tokyo AWS Core",
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
    region: "Global Cloud Save Sync",
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
    region: "P2P Co-op Relay Network",
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
    region: "Asia Dedicated AWS Edge",
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
    region: "Asia-Pacific Dedicated",
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
    region: "Asia-Pacific South (Azure)",
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
  packetLoss: number;
  reachable: boolean;
  /** Measured locally by the prober; retained across probe sweeps. */
  rttAvgMs: number;
  rttStdMs: number;
  lossAvgPct: number;
  rttTrend: number[];
  historySamples: number;
  lastProbeAt: string;
  /** Relay's own RTT to the region anchor (measured by the relay). */
  relayAnchorRttMs: number;
  /** client→relay + relay→anchor, the measured path estimate used for ranking. */
  totalEstimateMs: number;
  /** Top-ranked relay of the last measured sweep. */
  isOptimal: boolean;
  /** Relay the user connected through recently. */
  recent: boolean;
  /** History shows high loss or a large RTT spread. */
  unstable: boolean;
  /** Optional ICMP echo RTT to the relay host (0 = not measured). */
  icmpRttMs: number;
};

export type DetectedGame = {
  gameId: string;
  name: string;
  matchedProcess?: string;
  pid?: number;
  running: boolean;
};

export type GameConnection = {
  protocol: string;
  family: string;
  local: string;
  remote: string;
  state?: string;
};

export type GameConnectionStatus = {
  gameId: string;
  name: string;
  running: boolean;
  process?: string;
  pid?: number;
  connections: GameConnection[];
};

export type GameConnectionReport = {
  supported: boolean;
  note: string;
  scannedAt: string;
  games: GameConnectionStatus[];
};

export type RouteHop = {
  id: string;
  label: string;
  target?: string;
  rttMs?: number;
  rttAvgMs?: number;
  rttStdMs?: number;
  lossPct?: number;
  samples?: number;
  measured: boolean;
  reason?: string;
  note?: string;
};

export type FailoverEventItem = {
  timestamp: string;
  oldRelayId: string;
  oldRelayName: string;
  newRelayId: string;
  newRelayName: string;
  oldPingMs: number;
  newPingMs: number;
  reason: string;
  handoverOk: boolean;
};

export type RouteAnchor = {
  id: string;
  label: string;
  target: string;
  rttMs?: number;
  measured: boolean;
  note?: string;
  relayRttMs?: number;
  relayMeasured: boolean;
  viaRelayEstimateMs?: number;
  improvementMs?: number;
  improvementMeasured: boolean;
};

export type RouteDetail = {
  measuredAt: string;
  regionName?: string;
  connected: boolean;
  hops: RouteHop[];
  overheadMs: number;
  anchor: RouteAnchor;
  notes: string[];
};

function cleanNodeDisplayName(rawName: string, loc: string): { title: string; subtitle: string; flag: string } {
  // Preserve the operator-provided name instead of rewriting it into a
  // marketing label. Only strip a leading bracketed region prefix.
  const clean = rawName.replace(/^\[.*?\]\s*/i, "").trim();
  const haystack = `${rawName} ${loc}`;

  let flag = "🌐";
  if (/local|127\.0\.0\.1|localhost/i.test(haystack)) flag = "💻";
  else if (/singapore|\bsg\b/i.test(haystack)) flag = "🇸🇬";
  else if (/tokyo|japan|\bjp\b/i.test(haystack)) flag = "🇯🇵";
  else if (/seoul|korea|\bkr\b/i.test(haystack)) flag = "🇰🇷";
  else if (/hong kong|\bhk\b/i.test(haystack)) flag = "🇭🇰";
  else if (/frankfurt|germany|\bde\b/i.test(haystack)) flag = "🇩🇪";
  else if (/london|\buk\b/i.test(haystack)) flag = "🇬🇧";
  else if (/paris|france|\bfr\b/i.test(haystack)) flag = "🇫🇷";
  else if (/stockholm|sweden|\bse\b/i.test(haystack)) flag = "🇸🇪";
  else if (/sydney|australia|\bau\b/i.test(haystack)) flag = "🇦🇺";
  else if (/são paulo|sao paulo|brazil|\bbr\b/i.test(haystack)) flag = "🇧🇷";

  return { title: clean || loc || "Relay Node", subtitle: loc || "", flag };
}

const EMPTY_RELAY: RelayNode = {
  id: "none",
  name: "No relay configured",
  cleanName: "No relay configured",
  address: "",
  location: "Deploy a relay or add a custom node",
  country: "GLOBAL",
  flag: "⚠️",
  pingMs: 0,
  packetLoss: 0,
  reachable: false,
  rttAvgMs: 0,
  rttStdMs: 0,
  lossAvgPct: 0,
  rttTrend: [],
  historySamples: 0,
  lastProbeAt: "",
  relayAnchorRttMs: 0,
  totalEstimateMs: 0,
  isOptimal: false,
  recent: false,
  unstable: false,
  icmpRttMs: 0
};

function mapProbeRelays(data: unknown): RelayNode[] {
  if (!Array.isArray(data)) return [];
  return data
    .filter((item: any) => item && (item.relayId || item.id || item.endpoint || item.address))
    .map((item: any) => {
      const rawName = item.name || item.endpoint || item.address || "Relay";
      const { title, flag } = cleanNodeDisplayName(rawName, item.location || "");
      const measuredRtt =
        typeof item.rttMedianMs === "number" && item.rttMedianMs > 0
          ? item.rttMedianMs
          : typeof item.latencyMs === "number" && item.latencyMs > 0
            ? item.latencyMs
            : typeof item.pingMs === "number" && item.pingMs > 0
              ? item.pingMs
              : 0;
      return {
        id: item.relayId || item.id || item.endpoint || item.address,
        name: rawName,
        cleanName: title,
        address: item.endpoint || item.address,
        location: typeof item.location === "string" ? item.location : "",
        country: item.continent || item.country || "GLOBAL",
        flag,
        pingMs: measuredRtt > 0 ? Math.round(measuredRtt * 10) / 10 : 0,
        packetLoss:
          typeof item.packetLoss === "number" ? Math.round(item.packetLoss * 10) / 10 : 0,
        reachable: item.reachable === true,
        rttAvgMs: typeof item.rttAvgMs === "number" ? item.rttAvgMs : 0,
        rttStdMs: typeof item.rttStdMs === "number" ? item.rttStdMs : 0,
        lossAvgPct: typeof item.lossAvgPct === "number" ? item.lossAvgPct : 0,
        rttTrend: Array.isArray(item.rttTrend)
          ? item.rttTrend.filter((v: unknown): v is number => typeof v === "number")
          : [],
        historySamples: typeof item.historySamples === "number" ? item.historySamples : 0,
        lastProbeAt: typeof item.lastProbeAt === "string" ? item.lastProbeAt : "",
        relayAnchorRttMs: typeof item.relayAnchorRttMs === "number" ? item.relayAnchorRttMs : 0,
        totalEstimateMs: typeof item.totalEstimateMs === "number" ? item.totalEstimateMs : 0,
        isOptimal: item.isOptimal === true,
        recent: item.recent === true,
        unstable: item.unstable === true,
        icmpRttMs: typeof item.icmpRttMs === "number" ? item.icmpRttMs : 0
      };
    });
}

/** Small measured-RTT sparkline. Renders a dash until two real samples exist. */
function RttSparkline({ samples, width = 110, height = 26 }: { samples: number[]; width?: number; height?: number }) {
  if (!samples || samples.length < 2) {
    return <span className="text-[10px] text-white/30">Not enough samples</span>;
  }
  const min = Math.min(...samples);
  const max = Math.max(...samples);
  const range = max - min || 1;
  const step = width / (samples.length - 1);
  const points = samples
    .map((v, i) => {
      const x = i * step;
      const y = height - 3 - ((v - min) / range) * (height - 6);
      return `${x.toFixed(1)},${y.toFixed(1)}`;
    })
    .join(" ");
  const last = samples[samples.length - 1];
  const lastY = height - 3 - ((last - min) / range) * (height - 6);
  return (
    <svg width={width} height={height} className="overflow-visible" role="img" aria-label="Measured RTT history">
      <polyline
        points={points}
        fill="none"
        stroke="#00F0FF"
        strokeWidth="1.5"
        strokeLinejoin="round"
        strokeLinecap="round"
      />
      <circle cx={width} cy={lastY} r="2" fill="#00F0FF" />
    </svg>
  );
}

// Legacy UI ids that differ from the backend profile catalog.
const GAME_PROFILE_ALIASES: Record<string, string> = {
  cod: "cod_warzone",
};

function mergeGameProfiles(catalog: GameProfile[]): Game[] {
  const byId = new Map<string, GameProfile>();
  for (const profile of catalog) {
    byId.set(profile.id.toLowerCase(), profile);
  }

  const merged = OFFICIAL_GAMES.map((game) => {
    const profileId = (GAME_PROFILE_ALIASES[game.id] || game.id).toLowerCase();
    const profile = byId.get(profileId);
    const regions = (profile?.regions || []).filter((region) => region.cidrs?.length > 0);
    return {
      ...game,
      name: profile?.name || game.name,
      processNames: profile?.processNames?.length ? profile.processNames : game.processNames,
      region: regions[0]?.name || game.region,
      regions,
      profileAvailable: regions.length > 0,
    };
  });

  const known = new Set(merged.map((game) => game.id.toLowerCase()));
  const aliased = new Set(Object.values(GAME_PROFILE_ALIASES).map((id) => id.toLowerCase()));
  for (const profile of catalog) {
    const id = profile.id.toLowerCase();
    if (known.has(id) || aliased.has(id)) continue;
    const regions = (profile.regions || []).filter((region) => region.cidrs?.length > 0);
    merged.push({
      id: profile.id,
      name: profile.name,
      genre: profile.category || "Custom",
      tag: "FPS",
      platforms: "Custom profile",
      publisher: "Local profile",
      processNames: profile.processNames || [],
      cidrs: regions.flatMap((region) => region.cidrs),
      heroArt: "/assets/logo_clean.png",
      coverImg: "/assets/logo_clean.png",
      artType: "cover-keyart",
      region: regions[0]?.name || "Global",
      regions,
      profileAvailable: regions.length > 0,
    });
  }

  return merged;
}

function pickRegionForRelay(regions: GameRegion[], relay: RelayNode): string {
  const haystack = `${relay.country} ${relay.location} ${relay.name}`.toLowerCase();
  let prefix = "";
  if (/(asia|singapore|japan|tokyo|hong|korea|seoul|mumbai|sydney|oceania|apac)/.test(haystack)) {
    prefix = "asia";
  } else if (/(europe|germany|frankfurt|london|stockholm|paris|amsterdam)/.test(haystack)) {
    prefix = "eu";
  } else if (/(america|united states|chicago|virginia|california|seattle)/.test(haystack)) {
    prefix = "na";
  }

  if (prefix) {
    const match = regions.find((region) => {
      const key = `${region.id} ${region.name}`.toLowerCase();
      return prefix === "asia"
        ? /(asia|sg|jp|tokyo|hk|kr|seoul|mumbai|sydney)/.test(key)
        : prefix === "eu"
        ? /(eu|frankfurt|london|stockholm|paris)/.test(key)
        : /(na|us-|america|chicago|virginia|california)/.test(key);
    });
    if (match) return match.id;
  }

  return regions[0]?.id || "";
}

export type SquadMember = {
  name: string;
  role: "Host" | "Member";
  isp: string;
  ping: number;
  game: string;
  isSelf?: boolean;
};

export default function App() {
  // Navigation: HOME is now the primary welcoming dashboard, BOOST opens when a game is chosen!
  const [activeTab, setActiveTab] = useState<"home" | "boost" | "library" | "squad" | "nodes" | "tweaker">("home");

  // Selection
  const [selectedGameIdx, setSelectedGameIdx] = useState(0);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedCategory, setSelectedCategory] = useState<"ALL" | "FPS" | "MOBA" | "BR" | "RPG">("ALL");

  // Nickname & Gamer Profile
  const [gamerNickname, setGamerNickname] = useState<string>(() => {
    return localStorage.getItem("lagvex_nickname") || "Gamer";
  });
  const [isEditingNickname, setIsEditingNickname] = useState(false);
  const [tempNickname, setTempNickname] = useState(gamerNickname);
  const [pcHostname, setPcHostname] = useState("Local-PC");

  // Boost Hub Sub-Tab (GearUP Style)
  const [boostSubTab, setBoostSubTab] = useState<"routing" | "history">("routing");

  // Relays
  const [relays, setRelays] = useState<RelayNode[]>([]);
  const [selectedRelayIdx, setSelectedRelayIdx] = useState(0);

  // Backend game catalog (the source of truth for regions and CIDRs)
  const [games, setGames] = useState<Game[]>(OFFICIAL_GAMES);
  const [selectedRegionId, setSelectedRegionId] = useState("");
  const [connectionError, setConnectionError] = useState<string | null>(null);

  // Custom Game & Custom Node Modals State
  const [isAddingGame, setIsAddingGame] = useState(false);
  const [newGameName, setNewGameName] = useState("");
  const [newGameProcess, setNewGameProcess] = useState("");
  const [newGameCategory, setNewGameCategory] = useState<"FPS" | "MOBA" | "BR" | "RPG">("FPS");
  const [newGameRegion, setNewGameRegion] = useState("Asia-Pacific");
  const [newGameCidrs, setNewGameCidrs] = useState("");

  const [isAddingNode, setIsAddingNode] = useState(false);
  const [newNodeName, setNewNodeName] = useState("");
  const [newNodeEndpoint, setNewNodeEndpoint] = useState("");
  const [newNodeLocation, setNewNodeLocation] = useState("");
  const [newNodeContinent, setNewNodeContinent] = useState("ASIA");
  const [newNodePsk, setNewNodePsk] = useState("");

  // Node Search & Continent Filter
  const [nodeSearchQuery, setNodeSearchQuery] = useState("");
  const [selectedNodeContinent, setSelectedNodeContinent] = useState<"ALL" | "ASIA" | "NORTH AMERICA" | "EUROPE" | "OCEANIA" | "GLOBAL">("ALL");

  // Admin & Bufferbloat Diagnostics State
  const [isAdmin, setIsAdmin] = useState(false);
  const [driverName, setDriverName] = useState<string>("Detecting adapter…");
  const [isTestingBufferbloat, setIsTestingBufferbloat] = useState(false);
  const [bufferbloatResult, setBufferbloatResult] = useState<{
    baselineMs: number;
    loadedMs: number;
    diffMs: number;
    grade: string;
    rating: string;
    advice: string;
  } | null>(null);

  // Real Hardware & Network Socket Telemetry (Zero Fake Numbers)
  const [networkDiag, setNetworkDiag] = useState<{ gateway: string; lanPingMs: number; ispPingMs: number; dnsServer: string } | null>(null);
  const recentPingsRef = useRef<number[]>([]);

  // Detected game processes (measured by one local process-table scan)
  const [detectedGames, setDetectedGames] = useState<Record<string, DetectedGame>>({});
  const [gameConnections, setGameConnections] = useState<GameConnectionReport | null>(null);
  const [isScanningConnections, setIsScanningConnections] = useState(false);
  const [routeDetail, setRouteDetail] = useState<RouteDetail | null>(null);
  const [isMeasuringRoute, setIsMeasuringRoute] = useState(false);
  const [peerRouting, setPeerRouting] = useState(false);
  const [peerRouteCount, setPeerRouteCount] = useState(0);
  const [peerRoutes, setPeerRoutes] = useState<string[]>([]);
  const [fullTunnel, setFullTunnel] = useState(false);
  const [standby, setStandby] = useState<{ id: string; name: string; rttMs: number; totalMs: number; ageSec: number } | null>(null);
  const [rateHistory, setRateHistory] = useState<{ t: number; up: number; down: number }[]>([]);
  const [sessionBytes, setSessionBytes] = useState<{ up: number; down: number }>({ up: 0, down: 0 });
  const [failoverEvents, setFailoverEvents] = useState<FailoverEventItem[]>([]);
  const trafficCanvasRef = useRef<HTMLCanvasElement | null>(null);
  const [recentPeers, setRecentPeers] = useState<{ endpoint: string; firstSeen: number }[]>([]);

  // Acceleration / Boost State
  const [isBoosting, setIsBoosting] = useState(false);
  const [isConnecting, setIsConnecting] = useState(false);
  const [livePing, setLivePing] = useState(0);
  const [packetLoss, setPacketLoss] = useState<number | null>(null);
  const [jitter, setJitter] = useState(0);
  const [sessionSeconds, setSessionSeconds] = useState(0);
  const [toastMessage, setToastMessage] = useState<string | null>(null);

  // Squad Sync State: Real state, initially NOT in a room!
  const [currentSquadRoom, setCurrentSquadRoom] = useState<string | null>(null);
  const [joinInputCode, setJoinInputCode] = useState("");
  const [squadMembers, setSquadMembers] = useState<SquadMember[]>([]);

  // System Tweaker Flags
  const [tcpNoDelay, setTcpNoDelay] = useState(true);
  const [disableNagle, setDisableNagle] = useState(true);
  const [mmcssPriority, setMmcssPriority] = useState(true);
  const [mtuClamping, setMtuClamping] = useState(true);
  const [dnsFlushed, setDnsFlushed] = useState(false);
  const [snapshotResult, setSnapshotResult] = useState<{ applied: string[]; warnings: string[] } | null>(null);
  const snapshotFileRef = useRef<HTMLInputElement | null>(null);

  // Oscilloscope Canvas Ref
  const historyCanvasRef = useRef<HTMLCanvasElement | null>(null);
  const animationFrameRef = useRef<number | null>(null);
  const pingHistoryRef = useRef<number[]>([]);

  // Current items
  const selectedGame = games[selectedGameIdx] || games[0] || OFFICIAL_GAMES[0];
  const selectedRelay = relays[selectedRelayIdx] || relays[0] || EMPTY_RELAY;
  const selectedRegion =
    (selectedGame.regions || []).find((region) => region.id === selectedRegionId) ||
    (selectedGame.regions || [])[0];
  const selectedGameConnections = gameConnections?.games.find((g) => g.gameId === selectedGame.id);

  // Filtered games for Library Tab
  const filteredGames = useMemo(() => {
    return games.filter((g) => {
      const q = searchQuery.toLowerCase().trim();
      const matchSearch =
        !q ||
        g.name.toLowerCase().includes(q) ||
        g.genre.toLowerCase().includes(q) ||
        g.publisher.toLowerCase().includes(q);
      const matchCat = selectedCategory === "ALL" || g.tag === selectedCategory;
      return matchSearch && matchCat;
    });
  }, [games, searchQuery, selectedCategory]);

  // Filtered Relays for Nodes Explorer
  const filteredRelays = useMemo(() => {
    return relays.filter((r) => {
      const q = nodeSearchQuery.toLowerCase().trim();
      const matchSearch =
        !q ||
        r.name.toLowerCase().includes(q) ||
        r.cleanName.toLowerCase().includes(q) ||
        r.location.toLowerCase().includes(q) ||
        r.country.toLowerCase().includes(q);
      const matchCont =
        selectedNodeContinent === "ALL" ||
        r.country.toUpperCase() === selectedNodeContinent ||
        (selectedNodeContinent === "ASIA" && (r.country === "SG" || r.country === "JP" || r.country === "HK" || r.country === "KR" || r.country === "ASIA")) ||
        (selectedNodeContinent === "NORTH AMERICA" && (r.country === "US" || r.country === "NORTH AMERICA")) ||
        (selectedNodeContinent === "EUROPE" && (r.country === "DE" || r.country === "GB" || r.country === "FR" || r.country === "EUROPE")) ||
        (selectedNodeContinent === "OCEANIA" && (r.country === "AU" || r.country === "OCEANIA")) ||
        (selectedNodeContinent === "GLOBAL" && r.country === "GLOBAL");
      return matchSearch && matchCont;
    });
  }, [relays, nodeSearchQuery, selectedNodeContinent]);

  // Keep a valid server region selected for the active game and relay.
  useEffect(() => {
    const regions = selectedGame.regions || [];
    if (regions.length === 0) {
      if (selectedRegionId) setSelectedRegionId("");
      return;
    }
    if (!regions.some((region) => region.id === selectedRegionId)) {
      setSelectedRegionId(pickRegionForRelay(regions, selectedRelay));
    }
  }, [selectedGame.id, selectedRelay.id, selectedRegionId, games]);

  const refreshGames = async (): Promise<Game[]> => {
    try {
      const res = await fetch("/api/games");
      if (res.ok) {
        const data = await res.json();
        if (Array.isArray(data)) {
          const merged = mergeGameProfiles(data as GameProfile[]);
          setGames(merged);
          return merged;
        }
      }
    } catch {}
    return games;
  };

  // Add Custom Game Handler
  const handleCreateCustomGame = async () => {
    if (!newGameName.trim() || !newGameProcess.trim()) {
      notify("Please provide Game Name and Executable Process Name!");
      return;
    }
    const cidrs = newGameCidrs
      .split(/[\s,]+/)
      .map((cidr) => cidr.trim())
      .filter(Boolean);
    if (cidrs.length === 0) {
      notify("Enter at least one verified game server CIDR (for example 203.0.113.0/24).");
      return;
    }
    const id = newGameName.toLowerCase().replace(/[^a-z0-9]/g, "");
    try {
      const res = await fetch("/api/add-game", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          id,
          name: newGameName.trim(),
          category: newGameCategory,
          processNames: [newGameProcess.trim()],
          regions: [{ id: "global", name: newGameRegion.trim() || "Global", cidrs }]
        })
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        notify(`Failed to add custom game: ${data.error || res.statusText}`);
        return;
      }
      const merged = await refreshGames();
      const idx = merged.findIndex((game) => game.id === id);
      if (idx >= 0) setSelectedGameIdx(idx);
      notify(`✅ Added custom game: ${newGameName}`);
      setIsAddingGame(false);
      setNewGameName("");
      setNewGameProcess("");
      setNewGameCidrs("");
      setActiveTab("boost");
    } catch {
      notify("Failed to connect to backend.");
    }
  };

  // Add Custom Node Handler
  const handleCreateCustomNode = async () => {
    if (!newNodeName.trim() || !newNodeEndpoint.trim()) {
      notify("Please provide Node Name and IP:Port Endpoint!");
      return;
    }
    if (!newNodePsk.trim()) {
      notify("Relay PSK is required. Enter the PSK printed by your relay installer.");
      return;
    }
    const id = "custom-" + Date.now();
    try {
      const res = await fetch("/api/add-relay", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          id,
          name: newNodeName.trim(),
          endpoint: newNodeEndpoint.trim(),
          location: newNodeLocation.trim() || "Dedicated Host",
          continent: newNodeContinent,
          psk: newNodePsk.trim(),
          tier: "custom"
        })
      });
      if (!res.ok) {
        const message = (await res.text().catch(() => "")) || res.statusText;
        notify(`❌ Failed to add relay: ${message.trim()}`);
        return;
      }
      notify(`✅ Added custom relay node: ${newNodeName}`);
      setIsAddingNode(false);
      setNewNodeName("");
      setNewNodeEndpoint("");
      setNewNodeLocation("");
      setNewNodePsk("");
      // Re-probe relays
      const pRes = await fetch("/api/probe-relays");
      if (pRes.ok) {
        const data = await pRes.json();
        setRelays(mapProbeRelays(data));
      }
    } catch {
      notify("Failed to connect to backend.");
    }
  };

  // Relaunch as Administrator Handler
  const handleRelaunchAdmin = async () => {
    notify("Requesting administrator elevation prompt...");
    try {
      const res = await fetch("/api/relaunch-admin", { method: "POST" });
      const data = await res.json();
      if (data.isAdmin) {
        setIsAdmin(true);
        notify("Already running with Administrator privileges!");
      } else if (data.status) {
        notify(data.status);
      } else {
        notify("Please approve the administrator authorization prompt!");
      }
    } catch {
      notify("Failed to invoke administrator elevation.");
    }
  };

  // Graceful App Shutdown Handler
  const handleShutdownApp = async () => {
    if (window.confirm("Exit Lagvex completely? All game acceleration tunnels will be safely terminated.")) {
      notify("Shutting down Lagvex application...");
      try {
        await fetch("/api/shutdown", { method: "POST" });
      } catch {
        // Ignored
      }
      setTimeout(() => {
        window.close();
      }, 300);
    }
  };

  // Run Bufferbloat Test Handler
  const handleRunBufferbloatTest = async () => {
    setIsTestingBufferbloat(true);
    notify("⚡ Testing connection latency under load (Bufferbloat Analysis)...");
    try {
      const res = await fetch("/api/diagnose-bufferbloat");
      if (!res.ok) {
        notify(`Bufferbloat test failed (${res.status}).`);
        return;
      }
      const data = await res.json();
      if (data.available === false) {
        setBufferbloatResult(null);
        notify(`⚠️ ${data.error || "Bufferbloat could not be measured."}`);
        return;
      }
      setBufferbloatResult(data);
      notify(`✅ Bufferbloat Test complete: Grade ${data.grade} (${data.rating})`);
    } catch {
      notify("Bufferbloat diagnostic failed to connect.");
    } finally {
      setIsTestingBufferbloat(false);
    }
  };

  const notify = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => {
      setToastMessage((prev) => (prev === msg ? null : prev));
    }, 3000);
  };

  // Diagnostics snapshot: export measured state, import non-secret settings
  const handleExportDiagnostics = async () => {
    try {
      const res = await fetch("/api/diagnostics/export");
      if (!res.ok) {
        notify("Could not export a diagnostics snapshot.");
        return;
      }
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = `lagvex-diagnostics-${new Date().toISOString().slice(0, 10)}.json`;
      document.body.appendChild(link);
      link.click();
      link.remove();
      URL.revokeObjectURL(url);
      notify("Diagnostics snapshot downloaded. Relay PSKs are not included.");
    } catch {
      notify("Failed to reach the local Lagvex service.");
    }
  };

  const handleImportDiagnostics = async (file: File) => {
    try {
      const text = await file.text();
      const res = await fetch("/api/diagnostics/import", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: text
      });
      const data = await res.json().catch(() => null);
      if (!res.ok) {
        notify(`❌ Import failed: ${data?.error || res.statusText}`);
        return;
      }
      setSnapshotResult({
        applied: Array.isArray(data?.applied) ? data.applied : [],
        warnings: Array.isArray(data?.warnings) ? data.warnings : []
      });
      await refreshGames();
      notify("Snapshot imported.");
    } catch {
      notify("Failed to read or import the snapshot file.");
    }
  };

  // Keep the last few distinct endpoints this game was seen talking to.
  const rememberRecentPeers = (report: GameConnectionReport, gameId: string) => {
    const status = report.games.find((g) => g.gameId === gameId);
    if (!status?.running) return;

    const ips: string[] = [];
    for (const conn of status.connections) {
      if (!conn.remote) continue;
      const host = conn.remote.includes("]")
        ? conn.remote.slice(1, conn.remote.indexOf("]"))
        : conn.remote.split(":")[0];
      if (host && !ips.includes(host)) ips.push(host);
    }
    if (ips.length === 0) return;

    setRecentPeers((prev) => {
      const merged = [...prev];
      for (const ip of ips) {
        if (!merged.some((p) => p.endpoint === ip)) {
          merged.push({ endpoint: ip, firstSeen: Date.now() });
        }
      }
      return merged.slice(-3);
    });
  };

  // A new game selection starts a fresh observed-endpoint history.
  useEffect(() => {
    setRecentPeers([]);
  }, [selectedGame.id]);

  // Game server connections: read-only scan of the local OS socket tables
  const handleScanGameConnections = async (silent = false) => {
    if (!silent) setIsScanningConnections(true);
    try {
      const res = await fetch("/api/game-connections");
      if (!res.ok) {
        if (!silent) notify(`Connection scan failed (${res.status}).`);
        return;
      }
      const data = await res.json();
      const report: GameConnectionReport = {
        supported: data.supported === true,
        note: typeof data.note === "string" ? data.note : "",
        scannedAt: typeof data.scannedAt === "string" ? data.scannedAt : "",
        games: Array.isArray(data.games) ? data.games : []
      };
      setGameConnections(report);
      rememberRecentPeers(report, selectedGame.id);
    } catch {
      if (!silent) notify("Failed to reach the local Lagvex service.");
    } finally {
      if (!silent) setIsScanningConnections(false);
    }
  };

  // Route detail: measure each hop the client can observe (on demand)
  const handleMeasureRouteDetail = async (silent = false) => {
    if (!silent) setIsMeasuringRoute(true);
    try {
      const params = new URLSearchParams();
      if (selectedGame.profileAvailable) params.set("gameId", selectedGame.id);
      if (selectedRegion?.id) params.set("regionId", selectedRegion.id);
      if (selectedRelay.address) params.set("relayEndpoint", selectedRelay.address);

      const res = await fetch(`/api/route-detail?${params.toString()}`);
      if (!res.ok) {
        if (!silent) notify(`Route measurement failed (${res.status}).`);
        return;
      }
      const data = await res.json();
      setRouteDetail({
        measuredAt: typeof data.measuredAt === "string" ? data.measuredAt : "",
        regionName: typeof data.regionName === "string" ? data.regionName : "",
        connected: data.connected === true,
        hops: Array.isArray(data.hops) ? data.hops : [],
        overheadMs: typeof data.overheadMs === "number" ? data.overheadMs : 0,
        anchor:
          data.anchor && typeof data.anchor === "object"
            ? data.anchor
            : {
                id: "",
                label: "",
                target: "",
                measured: false,
                relayMeasured: false,
                improvementMeasured: false
              },
        notes: Array.isArray(data.notes) ? data.notes : []
      });
    } catch {
      if (!silent) notify("Failed to reach the local Lagvex service.");
    } finally {
      if (!silent) setIsMeasuringRoute(false);
    }
  };

  // Peer routing: opt-in /32 routes for the peers the game actually talks to
  const handleTogglePeerRouting = async () => {
    try {
      const res = await fetch("/api/peer-routing/toggle", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ enabled: !peerRouting })
      });
      if (!res.ok) {
        const message = (await res.text().catch(() => "")) || res.statusText;
        notify(`❌ Could not change peer routing: ${message.trim()}`);
        return;
      }
      const data = await res.json();
      setPeerRouting(data.enabled === true);
      setPeerRouteCount(Array.isArray(data.peerRoutes) ? data.peerRoutes.length : 0);
      notify(
        data.enabled
          ? "Peer routing enabled: observed game peers get a /32 route."
          : "Peer routing disabled; peer routes removed."
      );
    } catch {
      notify("Failed to reach the local Lagvex service.");
    }
  };

  // Full tunnel: opt-in "all IPv4 traffic through the relay" mode
  const handleToggleFullTunnel = async () => {
    try {
      const res = await fetch("/api/full-tunnel/toggle", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ enabled: !fullTunnel })
      });
      if (!res.ok) {
        const message = (await res.text().catch(() => "")) || res.statusText;
        notify(`❌ Could not change full tunnel: ${message.trim()}`);
        return;
      }
      const data = await res.json();
      setFullTunnel(data.enabled === true);
      notify(
        data.enabled
          ? "Full tunnel enabled: all IPv4 traffic now uses the relay."
          : "Full tunnel disabled; normal routing restored."
      );
    } catch {
      notify("Failed to reach the local Lagvex service.");
    }
  };

  // Save nickname
  const handleSaveNickname = () => {
    const trimmed = tempNickname.trim();
    if (!trimmed) return;
    setGamerNickname(trimmed);
    localStorage.setItem("lagvex_nickname", trimmed);
    setIsEditingNickname(false);
    notify(`Nickname updated to: ${trimmed}`);
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

  const formatRate = (bps: number) => {
    if (!bps || bps <= 0) return "0 KB/s";
    if (bps >= 1_000_000) return `${(bps / 1_000_000).toFixed(1)} MB/s`;
    return `${Math.round(bps / 1000)} KB/s`;
  };

  const formatBytes = (bytes: number) => {
    if (!bytes || bytes <= 0) return "0 MB";
    const mb = bytes / (1024 * 1024);
    if (mb < 1024) return `${mb.toFixed(1)} MB`;
    return `${(mb / 1024).toFixed(2)} GB`;
  };

  // Fetch real relays, status, system-info, and game-telemetry from Go Backend
  useEffect(() => {
    const fetchSystemInfo = async () => {
      try {
        const res = await fetch("/api/system-info");
        if (res.ok) {
          const data = await res.json();
          if (data.hostname) setPcHostname(data.hostname);
          if (data.isAdmin !== undefined) setIsAdmin(data.isAdmin);
          if (data.driverName) setDriverName(data.driverName);
          const savedNick = localStorage.getItem("lagvex_nickname");
          if (!savedNick || savedNick === "Gamer" || savedNick === "Player") {
            const initialNick = data.suggestedNickname || (data.username ? `${data.username}#${data.discriminator || 1001}` : "Gamer#1337");
            setGamerNickname(initialNick);
            setTempNickname(initialNick);
            localStorage.setItem("lagvex_nickname", initialNick);
          }
        }
      } catch {}
    };

    const fetchNetworkDiagnostics = async () => {
      try {
        const res = await fetch("/api/diagnose-network");
        if (res.ok) {
          const data = await res.json();
          setNetworkDiag(data);
          if (data.ispPingMs && !isBoosting) {
            setLivePing(data.ispPingMs);
          }
        }
      } catch {}
    };

    const fetchRelays = async () => {
      try {
        const res = await fetch("/api/probe-relays");
        if (res.ok) {
          const data = await res.json();
          if (Array.isArray(data) && data.length > 0) {
            const mapped = mapProbeRelays(data);
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
          if (data.active || data.state === "connected") {
            setIsBoosting(true);
            if (data.pingMs && data.pingMs > 0) setLivePing(data.pingMs);
            if (data.packetLossPct !== undefined) setPacketLoss(data.packetLossPct);
          }
        }
      } catch {}
    };

    fetchSystemInfo();
    fetchNetworkDiagnostics();
    fetchRelays();
    refreshGames();
    fetchStatus();
  }, []);

  // Real process detection: which configured game processes are running now.
  // One local process-table scan per poll; no game process is touched.
  useEffect(() => {
    let stopped = false;
    const pollDetected = async () => {
      try {
        const res = await fetch("/api/detected-processes");
        if (!res.ok) return;
        const data = await res.json();
        if (stopped || !Array.isArray(data.games)) return;
        const next: Record<string, DetectedGame> = {};
        for (const g of data.games) {
          if (g && typeof g.gameId === "string") next[g.gameId] = g as DetectedGame;
        }
        setDetectedGames(next);
      } catch {}
    };
    pollDetected();
    const timer = setInterval(pollDetected, 15000);
    return () => {
      stopped = true;
      clearInterval(timer);
    };
  }, []);

  // While a tunnel is up, keep the socket view fresh (read-only scan).
  useEffect(() => {
    if (!isBoosting) return;
    let stopped = false;
    const scan = async () => {
      try {
        const res = await fetch("/api/game-connections");
        if (!res.ok) return;
        const data = await res.json();
        if (stopped) return;
        const report: GameConnectionReport = {
          supported: data.supported === true,
          note: typeof data.note === "string" ? data.note : "",
          scannedAt: typeof data.scannedAt === "string" ? data.scannedAt : "",
          games: Array.isArray(data.games) ? data.games : []
        };
        setGameConnections(report);
        rememberRecentPeers(report, selectedGame.id);
      } catch {}
    };
    scan();
    const timer = setInterval(scan, 10000);
    return () => {
      stopped = true;
      clearInterval(timer);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isBoosting, selectedGame.id]);

  // Relay failover history (measured handovers, per session).
  useEffect(() => {
    if (!isBoosting && activeTab !== "boost") return;
    let stopped = false;
    const load = async () => {
      try {
        const res = await fetch("/api/failover/history");
        if (!res.ok) return;
        const data = await res.json();
        if (stopped || !Array.isArray(data)) return;
        setFailoverEvents(data as FailoverEventItem[]);
      } catch {}
    };
    load();
    const timer = setInterval(load, 10000);
    return () => {
      stopped = true;
      clearInterval(timer);
    };
  }, [isBoosting, activeTab]);

  // Keep the route detail fresh while the routing view is open.
  useEffect(() => {
    if (activeTab !== "boost" || boostSubTab !== "routing") return;
    void handleMeasureRouteDetail(true);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [activeTab, boostSubTab, selectedGame.id, selectedRegion?.id, selectedRelay.address, isBoosting]);

  // Ping History Oscilloscope Chart Loop (renders only real engine samples)
  useEffect(() => {
    const render = () => {
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

          const hist = pingHistoryRef.current;
          if (hist.length < 2) {
            animationFrameRef.current = requestAnimationFrame(render);
            return;
          }

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
  }, [isBoosting]);

  // Tunnel throughput canvas: draws only measured engine rate samples.
  useEffect(() => {
    const canvas = trafficCanvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const width = canvas.width;
    const height = canvas.height;
    ctx.clearRect(0, 0, width, height);
    if (rateHistory.length < 2) return;

    const maxRate = Math.max(1, ...rateHistory.map((s) => Math.max(s.up, s.down)));
    const step = width / (rateHistory.length - 1);
    const drawLine = (pick: (s: { up: number; down: number }) => number, color: string) => {
      ctx.beginPath();
      rateHistory.forEach((sample, i) => {
        const x = i * step;
        const y = height - (pick(sample) / maxRate) * (height - 8) - 4;
        if (i === 0) ctx.moveTo(x, y);
        else ctx.lineTo(x, y);
      });
      ctx.strokeStyle = color;
      ctx.lineWidth = 2;
      ctx.stroke();
    };
    drawLine((s) => s.down, "#10b981");
    drawLine((s) => s.up, "#00F0FF");
  }, [rateHistory, boostSubTab, activeTab]);

  // Keep UI state in sync with the real engine, including failures and
  // disconnects triggered outside this browser tab.
  useEffect(() => {
    let stopped = false;
    const pollStatus = async () => {
      try {
        const res = await fetch("/api/status");
        if (!res.ok) return;
        const data = await res.json();
        if (stopped) return;

        setPeerRouting(data.peerRouting === true);
        setFullTunnel(data.fullTunnel === true);
        if (data.standbyRelayName) {
          setStandby({
            id: typeof data.standbyRelayId === "string" ? data.standbyRelayId : "",
            name: data.standbyRelayName,
            rttMs: typeof data.standbyRttMs === "number" ? data.standbyRttMs : 0,
            totalMs: typeof data.standbyTotalMs === "number" ? data.standbyTotalMs : 0,
            ageSec: typeof data.standbyAgeSec === "number" ? data.standbyAgeSec : 0
          });
        } else {
          setStandby(null);
        }
        if (Array.isArray(data.peerRoutes)) setPeerRoutes(data.peerRoutes);
        if (typeof data.peerRouteCount === "number") setPeerRouteCount(data.peerRouteCount);

        if (data.state === "connected") {
          setIsBoosting(true);
          setConnectionError(null);
          if (data.pingMs > 0) setLivePing(data.pingMs);
          if (data.packetLossPct !== undefined) setPacketLoss(data.packetLossPct);
        } else if (data.state === "disconnected" || data.state === "error") {
          setIsBoosting(false);
          setPeerRoutes([]);
          setPeerRouteCount(0);
          setRateHistory([]);
          setSessionBytes({ up: 0, down: 0 });
          if (data.lastError) setConnectionError(data.lastError);
        }
      } catch {}
    };

    pollStatus();
    const timer = setInterval(pollStatus, 2000);
    return () => {
      stopped = true;
      clearInterval(timer);
    };
  }, []);

  // Live real telemetry polling (Zero fake data, poll live socket engine)
  useEffect(() => {
    let timer: any;
    if (isBoosting) {
      timer = setInterval(async () => {
        try {
          const res = await fetch("/api/status");
          if (res.ok) {
            const data = await res.json();
            if (data.active || data.state === "connected") {
              if (data.pingMs > 0) {
                const currentPing = Number(data.pingMs);
                setLivePing(Number(currentPing.toFixed(1)));

                const chart = pingHistoryRef.current;
                chart.push(currentPing);
                if (chart.length > 60) chart.shift();

                // Mean absolute difference between recent measured RTT samples.
                recentPingsRef.current.push(currentPing);
                if (recentPingsRef.current.length > 8) {
                  recentPingsRef.current.shift();
                }
                const samples = recentPingsRef.current;
                if (samples.length >= 2) {
                  let sumDiff = 0;
                  for (let i = 1; i < samples.length; i++) {
                    sumDiff += Math.abs(samples[i] - samples[i - 1]);
                  }
                  setJitter(Number((sumDiff / (samples.length - 1)).toFixed(2)));
                }
              }

              // Real packet loss from engine
              if (data.packetLossPct !== undefined) {
                setPacketLoss(Number(Number(data.packetLossPct).toFixed(1)));
              }

              // Real measured throughput samples (bytes/sec from the engine)
              const upRate = Number(data.upRateBps) || 0;
              const downRate = Number(data.downRateBps) || 0;
              setRateHistory((prev) => [
                ...prev.slice(-59),
                { t: Date.now(), up: upRate, down: downRate }
              ]);
              setSessionBytes({
                up: Number(data.bytesUp) || 0,
                down: Number(data.bytesDown) || 0
              });
            }
          }
        } catch {}
      }, 1200);
    } else {
      setLivePing(0);
      setJitter(0);
      setPacketLoss(null);
      recentPingsRef.current = [];
      pingHistoryRef.current = [];
      setRateHistory([]);
      setSessionBytes({ up: 0, down: 0 });
    }
    return () => clearInterval(timer);
  }, [isBoosting, selectedGame, selectedRelay]);

  // Connect / Disconnect Handler
  const handleToggleBoost = async () => {
    if (isBoosting) {
      setIsConnecting(true);
      try {
        await fetch("/api/disconnect", { method: "POST" });
        setIsBoosting(false);
        setLivePing(0);
        setConnectionError(null);
        notify("⚡ Acceleration stopped. Standby mode.");
      } catch {
        notify("Could not reach the local Lagvex service to disconnect.");
      } finally {
        setIsConnecting(false);
      }
      return;
    }

    if (!selectedGame.profileAvailable || !selectedRegion) {
      notify("This game has no verified region profile yet. Add a custom profile with server CIDRs first.");
      return;
    }
    if (selectedRelay.id === "none" || !selectedRelay.address) {
      notify("No relay is configured. Deploy a relay or add a custom node with its endpoint and PSK first.");
      return;
    }

    setIsConnecting(true);
    setConnectionError(null);
    const targetRelayAddr = selectedRelay.address || "auto";
    notify(`Accelerating ${selectedGame.name} (${selectedRegion.name}) via ${selectedRelay.cleanName}...`);
    try {
      const res = await fetch("/api/connect", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          gameId: selectedGame.id,
          regionId: selectedRegion.id,
          relayEndpoint: targetRelayAddr,
          relayAddr: targetRelayAddr,
          autoNode: targetRelayAddr === "auto" || selectedRelay.id === "auto",
          forceNow: true
        })
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok || data.success === false || data.state === "error") {
        const message = data.error || data.lastError || `Connection failed (${res.status})`;
        setConnectionError(message);
        setIsBoosting(false);
        notify(`❌ ${message}`);
        return;
      }

      setIsBoosting(true);
      // Only show RTT reported by the tunnel engine. A relay probe performed
      // before connecting is not live tunnel latency.
      const targetPing = data.pingMs > 0 ? Number(data.pingMs) : 0;
      setLivePing(Number(targetPing.toFixed(1)));
      const routeNote = data.routeCount > 0
        ? `${data.routeCount} game routes installed`
        : "routes will activate when the game process starts";
      notify(`⚡ Boost Active: ${selectedGame.name} via ${selectedRelay.cleanName} — ${routeNote}.`);
    } catch {
      setConnectionError("Could not reach the local Lagvex service.");
      setIsBoosting(false);
      notify("❌ Could not reach the local Lagvex service.");
    } finally {
      setIsConnecting(false);
    }
  };

  // Launch Game Handler
  const handleLaunchGame = () => {
    if (selectedGame.steamAppId) {
      window.open(`steam://rungameid/${selectedGame.steamAppId}`, "_blank");
      notify(`Opening Steam for ${selectedGame.name}...`);
    } else {
      notify(`Launch ${selectedGame.processNames[0] || selectedGame.name} manually. Lagvex will detect the process while a tunnel is active.`);
    }
  };

  // Select game from Library/Home and switch to Boost View
  const handleSelectGameAndBoost = (idx: number) => {
    setSelectedGameIdx(idx);
    setActiveTab("boost");
    notify(`Selected ${games[idx]?.name || "game"}. Boost HUD ready.`);
  };

  // Create a new Squad Room (starts with ONLY YOU, real state!)
  const handleCreateSquadRoom = async () => {
    const code = `LGVX-${Math.floor(1000 + Math.random() * 9000)}`;
    try {
      const res = await fetch("/api/squad/create", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          code,
          host: gamerNickname,
          relayId: selectedRelay.id !== "none" ? selectedRelay.id : "",
          relayName: selectedRelay.id !== "none" ? selectedRelay.cleanName : "",
          game: selectedGame.name,
          ping: Math.round(livePing),
          isp: networkDiag?.gateway || ""
        })
      });
      if (!res.ok) {
        const message = (await res.text().catch(() => "")) || res.statusText;
        notify(`❌ Could not create squad room: ${message.trim()}`);
        return;
      }
      const selfMember: SquadMember = {
        name: gamerNickname,
        role: "Host",
        isp: networkDiag?.gateway || "",
        ping: Math.round(livePing),
        game: selectedGame.name,
        isSelf: true
      };
      setCurrentSquadRoom(code);
      setSquadMembers([selfMember]);
      notify(`Created Squad Room ${code}. Share the code with your teammates.`);
    } catch {
      notify("Failed to reach the local Lagvex service.");
    }
  };

  // Join a Squad Room
  const handleJoinSquadRoom = async () => {
    if (!joinInputCode.trim()) {
      notify("Please enter a Squad Room code first.");
      return;
    }
    const code = joinInputCode.trim().toUpperCase();
    try {
      const res = await fetch("/api/squad/join", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          code,
          name: gamerNickname,
          game: selectedGame.name,
          ping: Math.round(livePing),
          isp: networkDiag?.gateway || ""
        })
      });
      if (!res.ok) {
        const message = (await res.text().catch(() => "")) || res.statusText;
        notify(`❌ Could not join squad room: ${message.trim()}`);
        return;
      }
      const data = await res.json();
      setCurrentSquadRoom(code);
      if (Array.isArray(data.members) && data.members.length > 0) {
        const mapped: SquadMember[] = data.members.map((m: any) => ({
          name: m.name,
          role: m.role || "Member",
          isp: typeof m.isp === "string" ? m.isp : "",
          ping:
            typeof m.ping === "number" && m.ping > 0
              ? m.ping
              : m.name === gamerNickname && livePing > 0
                ? Math.round(livePing)
                : 0,
          game: typeof m.game === "string" ? m.game : "",
          isSelf: m.name === gamerNickname
        }));
        setSquadMembers(mapped);
      }
      notify(`Joined Squad Room ${code}.`);
    } catch {
      notify("Failed to reach the local Lagvex service.");
    }
  };

  // Leave Squad Room
  const handleLeaveSquadRoom = async () => {
    if (currentSquadRoom) {
      try {
        await fetch("/api/squad/leave", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ code: currentSquadRoom, name: gamerNickname })
        });
      } catch {}
    }
    setCurrentSquadRoom(null);
    setSquadMembers([]);
    notify("Left Squad Room. Returned to Lobby.");
  };


  // Squad Room Real-Time Heartbeat & Sync Loop (Auto purges stale peers & syncs ping)
  useEffect(() => {
    if (!currentSquadRoom) return;
    const interval = setInterval(async () => {
      try {
        await fetch("/api/squad/heartbeat", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            code: currentSquadRoom,
            name: gamerNickname,
            ping: Math.round(livePing)
          })
        });
        const res = await fetch(`/api/squad/room?code=${encodeURIComponent(currentSquadRoom)}`);
        if (res.ok) {
          const data = await res.json();
          if (Array.isArray(data.members)) {
            const mapped: SquadMember[] = data.members.map((m: any) => ({
              name: m.name,
              role: m.role || "Member",
              isp: typeof m.isp === "string" ? m.isp : "",
              ping:
                typeof m.ping === "number" && m.ping > 0
                  ? m.ping
                  : m.name === gamerNickname && livePing > 0
                    ? Math.round(livePing)
                    : 0,
              game: typeof m.game === "string" ? m.game : "",
              isSelf: m.name === gamerNickname
            }));
            setSquadMembers(mapped);
          }
        }
      } catch {}
    }, 2500);
    return () => clearInterval(interval);
  }, [currentSquadRoom, gamerNickname, livePing, selectedGame.name]);

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
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        notify(`Tweak request failed (${res.status}).`);
        return;
      }
      const applied: string[] = Array.isArray(data.applied) ? data.applied : [];
      const failed: string[] = Array.isArray(data.failed) ? data.failed : [];
      if (applied.some((item) => item.toLowerCase().includes("dns"))) {
        setDnsFlushed(true);
      }
      if (applied.length > 0 && failed.length === 0) {
        notify(`✅ ${data.status || "Optimizations applied"}: ${applied.join(", ")}`);
      } else if (applied.length > 0) {
        notify(`⚠️ ${data.status || "Partially applied"}. Applied: ${applied.join(", ")}. Failed: ${failed.join(", ")}`);
      } else if (failed.length > 0) {
        notify(`❌ No system tweak was applied. Failed: ${failed.join(", ")}`);
      } else {
        notify(data.status || "Preferences saved.");
      }
    } catch {
      notify("❌ Could not reach the local Lagvex service; no tweaks were applied.");
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

      {/* ── NICKNAME EDIT MODAL ── */}
      {isEditingNickname && (
        <div className="fixed inset-0 bg-black/75 backdrop-blur-md z-50 flex items-center justify-center p-4 modal-backdrop-animate">
          <div className="w-full max-w-[420px] rounded-3xl bg-[#0d141e] border border-[#223347] p-6 shadow-2xl flex flex-col gap-4 modal-pop-animate">
            <h3 className="font-brand font-extrabold text-lg text-white">Customize Gamer Profile</h3>
            <p className="text-xs text-white/50">
              Set your personal gaming callsign. This name appears on your dashboard and when hosting or joining Squad rooms.
            </p>

            <div className="flex flex-col gap-1.5">
              <label className="text-[10px] font-brand font-bold text-white/40 uppercase">Gamer Nickname</label>
              <input
                type="text"
                value={tempNickname}
                onChange={(e) => setTempNickname(e.target.value)}
                placeholder="e.g. TenZ, Faker, ShadowPro..."
                className="h-[46px] px-4 rounded-xl bg-[#121c2a] border border-[#23354b] text-sm text-white font-bold outline-none focus:border-[#00F0FF]"
              />
            </div>

            <div className="p-3 rounded-xl bg-white/[0.03] border border-white/5 text-[11px] text-white/40 flex items-center justify-between font-mono">
              <span>Detected Machine Name:</span>
              <span className="text-[#00F0FF] font-bold">{pcHostname}</span>
            </div>

            <div className="flex items-center justify-end gap-3 pt-2">
              <button
                onClick={() => setIsEditingNickname(false)}
                className="px-4 py-2 rounded-xl text-xs font-brand font-bold text-white/50 hover:text-white bg-transparent border-none cursor-pointer"
              >
                Cancel
              </button>
              <button
                onClick={handleSaveNickname}
                className="px-5 py-2.5 rounded-xl text-xs font-brand font-bold bg-[#00F0FF] hover:bg-[#33f3ff] text-black border-none cursor-pointer transition-all shadow-md shadow-[#00F0FF]/20"
              >
                Save Profile
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── CUSTOM GAME MODAL ── */}
      {isAddingGame && (
        <div className="fixed inset-0 bg-black/75 backdrop-blur-md z-50 flex items-center justify-center p-4 modal-backdrop-animate">
          <div className="w-full max-w-[460px] rounded-3xl bg-[#0d141e] border border-[#223347] p-6 shadow-2xl flex flex-col gap-4 modal-pop-animate">
            <div className="flex items-center justify-between">
              <h3 className="font-brand font-extrabold text-lg text-white">Add Custom Game Profile</h3>
              <button onClick={() => setIsAddingGame(false)} className="text-white/40 hover:text-white border-none bg-transparent cursor-pointer text-base">✕</button>
            </div>
            <p className="text-xs text-white/50">
              Register a game executable and the verified server CIDRs it connects to. Lagvex only routes those subnets through the tunnel; traffic outside them stays direct.
            </p>

            <div className="flex flex-col gap-1.5">
              <label className="text-[10px] font-brand font-bold text-white/40 uppercase">Game Title</label>
              <input
                type="text"
                value={newGameName}
                onChange={(e) => setNewGameName(e.target.value)}
                placeholder="e.g. Cyberpunk 2077, GTA V, Arena Breakout..."
                className="h-[44px] px-4 rounded-xl bg-[#121c2a] border border-[#23354b] text-sm text-white font-bold outline-none focus:border-[#00F0FF]"
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-[10px] font-brand font-bold text-white/40 uppercase">Executable Process Name (.exe)</label>
              <input
                type="text"
                value={newGameProcess}
                onChange={(e) => setNewGameProcess(e.target.value)}
                placeholder="e.g. Cyberpunk2077.exe, GTA5.exe..."
                className="h-[44px] px-4 rounded-xl bg-[#121c2a] border border-[#23354b] text-sm text-white font-mono outline-none focus:border-[#00F0FF]"
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-[10px] font-brand font-bold text-white/40 uppercase">Verified Server CIDRs (required)</label>
              <textarea
                value={newGameCidrs}
                onChange={(e) => setNewGameCidrs(e.target.value)}
                placeholder="e.g. 203.0.113.0/24, 198.51.100.0/24"
                rows={2}
                className="px-4 py-3 rounded-xl bg-[#121c2a] border border-[#23354b] text-xs text-white font-mono outline-none focus:border-[#00F0FF] resize-none"
              />
              <span className="text-[10px] text-white/35">
                One or more IPv4/IPv6 prefixes, separated by commas, spaces or new lines.
              </span>
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div className="flex flex-col gap-1.5">
                <label className="text-[10px] font-brand font-bold text-white/40 uppercase">Genre Category</label>
                <select
                  value={newGameCategory}
                  onChange={(e: any) => setNewGameCategory(e.target.value)}
                  className="h-[44px] px-3 rounded-xl bg-[#121c2a] border border-[#23354b] text-xs text-white outline-none focus:border-[#00F0FF]"
                >
                  <option value="FPS">FPS Shooter</option>
                  <option value="MOBA">MOBA Strategy</option>
                  <option value="BR">Battle Royale</option>
                  <option value="RPG">Action / RPG</option>
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-[10px] font-brand font-bold text-white/40 uppercase">Primary Region</label>
                <input
                  type="text"
                  value={newGameRegion}
                  onChange={(e) => setNewGameRegion(e.target.value)}
                  placeholder="e.g. Asia-Pacific, US East"
                  className="h-[44px] px-3 rounded-xl bg-[#121c2a] border border-[#23354b] text-xs text-white outline-none focus:border-[#00F0FF]"
                />
              </div>
            </div>

            <div className="flex items-center justify-end gap-3 pt-3 border-t border-white/5">
              <button
                onClick={() => setIsAddingGame(false)}
                className="px-4 py-2 rounded-xl text-xs font-brand font-bold text-white/50 hover:text-white bg-transparent border-none cursor-pointer"
              >
                Cancel
              </button>
              <button
                onClick={handleCreateCustomGame}
                className="px-5 py-2.5 rounded-xl text-xs font-brand font-bold bg-[#00F0FF] hover:bg-[#33f3ff] text-black border-none cursor-pointer transition-all shadow-md shadow-[#00F0FF]/20"
              >
                + Register Game
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── CUSTOM RELAY NODE MODAL ── */}
      {isAddingNode && (
        <div className="fixed inset-0 bg-black/75 backdrop-blur-md z-50 flex items-center justify-center p-4 modal-backdrop-animate">
          <div className="w-full max-w-[480px] rounded-3xl bg-[#0d141e] border border-[#223347] p-6 shadow-2xl flex flex-col gap-4 modal-pop-animate">
            <div className="flex items-center justify-between">
              <h3 className="font-brand font-extrabold text-lg text-white">Add Custom Routing Node</h3>
              <button onClick={() => setIsAddingNode(false)} className="text-white/40 hover:text-white border-none bg-transparent cursor-pointer text-base">✕</button>
            </div>
            <p className="text-xs text-white/50">
              Hook up your personal VPS, dedicated server, or community relay node to bypass routing congestion.
            </p>

            <div className="flex flex-col gap-1.5">
              <label className="text-[10px] font-brand font-bold text-white/40 uppercase">Node Display Name</label>
              <input
                type="text"
                value={newNodeName}
                onChange={(e) => setNewNodeName(e.target.value)}
                placeholder="e.g. My Tokyo FastPath VPS"
                className="h-[44px] px-4 rounded-xl bg-[#121c2a] border border-[#23354b] text-sm text-white font-bold outline-none focus:border-[#00F0FF]"
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-[10px] font-brand font-bold text-white/40 uppercase">Endpoint (IP:Port or Host:Port)</label>
              <input
                type="text"
                value={newNodeEndpoint}
                onChange={(e) => setNewNodeEndpoint(e.target.value)}
                placeholder="e.g. 104.28.19.42:4433"
                className="h-[44px] px-4 rounded-xl bg-[#121c2a] border border-[#23354b] text-sm text-white font-mono outline-none focus:border-[#00F0FF]"
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-[10px] font-brand font-bold text-white/40 uppercase">Relay PSK</label>
              <input
                type="password"
                value={newNodePsk}
                onChange={(e) => setNewNodePsk(e.target.value)}
                placeholder="PSK printed by your relay installer"
                autoComplete="off"
                className="h-[44px] px-4 rounded-xl bg-[#121c2a] border border-[#23354b] text-sm text-white font-mono outline-none focus:border-[#00F0FF]"
              />
              <span className="text-[10px] text-white/35">
                Stored locally and never returned by the API. A relay cannot be probed or used without its PSK.
              </span>
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div className="flex flex-col gap-1.5">
                <label className="text-[10px] font-brand font-bold text-white/40 uppercase">Continent / Region</label>
                <select
                  value={newNodeContinent}
                  onChange={(e) => setNewNodeContinent(e.target.value)}
                  className="h-[44px] px-3 rounded-xl bg-[#121c2a] border border-[#23354b] text-xs text-white outline-none focus:border-[#00F0FF]"
                >
                  <option value="ASIA">Asia-Pacific</option>
                  <option value="NORTH AMERICA">North America</option>
                  <option value="EUROPE">Europe</option>
                  <option value="OCEANIA">Oceania</option>
                  <option value="GLOBAL">Global / Other</option>
                </select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-[10px] font-brand font-bold text-white/40 uppercase">Location Datacenter</label>
                <input
                  type="text"
                  value={newNodeLocation}
                  onChange={(e) => setNewNodeLocation(e.target.value)}
                  placeholder="e.g. Equinix TY2, Tokyo"
                  className="h-[44px] px-3 rounded-xl bg-[#121c2a] border border-[#23354b] text-xs text-white outline-none focus:border-[#00F0FF]"
                />
              </div>
            </div>

            <div className="flex items-center justify-end gap-3 pt-3 border-t border-white/5">
              <button
                onClick={() => setIsAddingNode(false)}
                className="px-4 py-2 rounded-xl text-xs font-brand font-bold text-white/50 hover:text-white bg-transparent border-none cursor-pointer"
              >
                Cancel
              </button>
              <button
                onClick={handleCreateCustomNode}
                className="px-5 py-2.5 rounded-xl text-xs font-brand font-bold bg-[#00F0FF] hover:bg-[#33f3ff] text-black border-none cursor-pointer transition-all shadow-md shadow-[#00F0FF]/20"
              >
                + Add Relay Node
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── 1. LEFT NAVIGATION RAIL (6 DEDICATED TABS) ── */}
      <aside className="fixed left-0 top-0 bottom-0 w-[84px] bg-[#090d13]/95 border-r border-[#151d28] flex flex-col items-center py-6 z-40 backdrop-blur-xl">
        {/* Brand Emblem */}
        <div className="w-12 h-12 rounded-2xl bg-[#101722] border border-[#202d3f] p-2 flex items-center justify-center mb-6 shadow-lg shadow-[#00F0FF]/10 group">
          <img
            src={imgLagvexLogo}
            alt="Lagvex Emblem"
            className="w-full h-full object-contain rounded-lg transition-transform duration-300 group-hover:scale-110"
          />
        </div>

        {/* 6 Primary Navigation Tabs */}
        <nav className="flex flex-col gap-2.5 w-full px-3">
          {[
            { id: "home" as const, Icon: IconHome, label: "HOME" },
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
                className={`relative w-full h-[50px] rounded-xl flex flex-col items-center justify-center gap-1 cursor-pointer border-none transition-all group ${
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
                {activeTab === "home" && "Dashboard Overview"}
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
            {/* Driver Execution Mode Pill (Kernel vs Userspace fallback) */}
            {isAdmin ? (
              <div
                className="h-[38px] px-3.5 rounded-xl bg-[#10b981]/15 border border-[#10b981]/30 flex items-center gap-2 text-xs font-brand font-bold text-[#10b981]"
                title={`Running with administrator rights. Tunnel adapter: ${driverName}.`}
              >
                <span className="w-2 h-2 rounded-full bg-[#10b981] shadow-[0_0_8px_#10b981]" />
                <span>Elevated &bull; {driverName}</span>
              </div>
            ) : (
              <button
                onClick={handleRelaunchAdmin}
                className="h-[38px] px-3.5 rounded-xl bg-[#f59e0b]/15 hover:bg-[#f59e0b]/25 border border-[#f59e0b]/40 flex items-center gap-2 text-xs font-brand font-bold text-[#f59e0b] cursor-pointer transition-all shadow group"
                title={`Administrator rights may be required to create the ${driverName} adapter. Click to request elevation.`}
              >
                <span className="w-2 h-2 rounded-full bg-[#f59e0b] shadow-[0_0_8px_#f59e0b] animate-pulse" />
                <span>Standard user &bull; Elevate &#x2197;</span>
              </button>
            )}

            {isBoosting && (
              <div className="h-[38px] px-3.5 rounded-xl bg-[#101722] border border-[#1f2b3b] flex items-center gap-2 text-xs">
                <span className="text-white/40">Duration:</span>
                <span className="font-mono font-bold text-[#00F0FF]">{formatTimer(sessionSeconds)}</span>
              </div>
            )}

            {/* Gamer Nickname Pill (Clickable to Edit) */}
            <div
              onClick={() => {
                setTempNickname(gamerNickname);
                setIsEditingNickname(true);
              }}
              className="h-[38px] pl-3 pr-3.5 rounded-xl bg-[#101722] hover:bg-[#162130] border border-[#1f2b3b] hover:border-[#00F0FF]/50 flex items-center gap-2.5 cursor-pointer transition-all shadow group"
              title="Click to change your Gamer Nickname"
            >
              <div className="w-6 h-6 rounded-lg bg-gradient-to-tr from-[#00F0FF] to-[#7E42FF] flex items-center justify-center text-[10px] font-brand font-extrabold text-black">
                {gamerNickname.charAt(0).toUpperCase()}
              </div>
              <div className="flex flex-col">
                <span className="font-brand font-bold text-xs text-white leading-tight flex items-center gap-1.5">
                  <span>{gamerNickname}</span>
                  <IconEdit className="text-white/40 group-hover:text-[#00F0FF] transition-colors" />
                </span>
              </div>
            </div>

            {/* Quit App Button */}
            <button
              onClick={handleShutdownApp}
              className="h-[38px] px-3.5 rounded-xl bg-red-500/10 hover:bg-red-500/20 border border-red-500/30 hover:border-red-500/50 flex items-center gap-1.5 text-xs font-brand font-bold text-red-400 hover:text-red-300 cursor-pointer transition-all shadow group"
              title="Safely exit and close Lagvex application"
            >
              <IconPower className="w-3.5 h-3.5 text-red-400 group-hover:scale-110 transition-transform" />
              <span>Quit</span>
            </button>
          </div>
        </header>

        {/* ── 3. DYNAMIC TAB VIEWS ── */}
        <div className="p-8 flex-1 flex flex-col">
          {/* ========================================================= */}
          {/* TAB 0: HOME (MAIN DASHBOARD OVERVIEW)                    */}
          {/* ========================================================= */}
          {activeTab === "home" && (
            <div className="flex flex-col gap-6 max-w-[1240px] mx-auto w-full tab-enter">
              {/* Hero Banner: Welcome & Quick Launch */}
              <div className="relative rounded-[32px] overflow-hidden bg-gradient-to-r from-[#0d1624] via-[#0b121c] to-[#080d14] border border-[#1d2a3a] p-8 shadow-2xl flex items-center justify-between">
                <div className="max-w-[620px] flex flex-col gap-3 z-10">
                  <div className="flex items-center gap-2">
                    <span className="px-3 py-1 rounded-full bg-[#00F0FF]/15 text-[#00F0FF] text-xs font-brand font-bold border border-[#00F0FF]/30 flex items-center gap-1.5">
                      <span className="w-1.5 h-1.5 rounded-full bg-[#00F0FF] animate-pulse" />
                      SYSTEM READY &bull; KERNEL ACCELERATOR
                    </span>
                  </div>
                  <h2 className="font-brand font-extrabold text-3xl text-white tracking-wide leading-tight">
                    Welcome back, <span className="text-[#00F0FF]">{gamerNickname}</span>!
                  </h2>
                  <p className="text-xs text-white/60 leading-relaxed">
                    Lagvex routes only the selected game server subnets through a low-latency relay, leaving the rest of your traffic on the direct path. Connect to see live engine RTT and loss.
                  </p>
                  <div className="flex items-center gap-3 pt-2">
                    <button
                      onClick={() => setActiveTab("library")}
                      className="h-[46px] px-6 rounded-2xl bg-[#00F0FF] hover:bg-[#33f3ff] text-black font-brand font-extrabold text-xs uppercase tracking-wider cursor-pointer border-none transition-all shadow-lg shadow-[#00F0FF]/20"
                    >
                      EXPLORE GAME LIBRARY
                    </button>
                    {isBoosting && (
                      <button
                        onClick={() => setActiveTab("boost")}
                        className="h-[46px] px-5 rounded-2xl bg-[#10b981]/20 hover:bg-[#10b981]/30 border border-[#10b981]/40 text-[#10b981] font-brand font-bold text-xs uppercase tracking-wider flex items-center gap-2 cursor-pointer transition-all"
                      >
                        <IconBolt className="w-4 h-4 text-[#10b981] animate-pulse" />
                        <span>VIEW ACTIVE ACCELERATION ({selectedGame.name})</span>
                      </button>
                    )}
                  </div>
                </div>

                {/* PC Network Telemetry Card */}
                <div className="p-5 rounded-2xl bg-[#090e15]/90 border border-[#1c2738] flex flex-col gap-3 min-w-[280px] shadow-xl z-10">
                  <span className="text-[10px] font-brand font-bold text-white/50 uppercase tracking-wider flex items-center justify-between">
                    <span>YOUR LOCAL NETWORK</span>
                    <span className="text-[#10b981] font-mono">
                      {networkDiag && networkDiag.ispPingMs > 0 ? "ONLINE" : "MEASURING"}
                    </span>
                  </span>
                  <div className="flex flex-col gap-2 font-mono text-xs">
                    <div className="flex items-center justify-between p-2 rounded bg-white/[0.02]">
                      <span className="text-white/50">Hostname:</span>
                      <span className="font-bold text-white">{pcHostname}</span>
                    </div>
                    <div className="flex items-center justify-between p-2 rounded bg-white/[0.02]">
                      <span className="text-white/50">Gateway LAN:</span>
                      <span className="text-[#00F0FF] font-bold">
                        {networkDiag && networkDiag.lanPingMs > 0
                          ? `${networkDiag.lanPingMs}ms (${networkDiag.gateway})`
                          : "—"}
                      </span>
                    </div>
                    <div className="flex items-center justify-between p-2 rounded bg-white/[0.02]">
                      <span className="text-white/50">ISP Direct Latency:</span>
                      <span className="text-[#10b981] font-bold">
                        {networkDiag && networkDiag.ispPingMs > 0
                          ? `${networkDiag.ispPingMs}ms`
                          : "—"}
                      </span>
                    </div>
                    <div className="flex items-center justify-between p-2 rounded bg-white/[0.02]">
                      <span className="text-white/50">Optimal Node:</span>
                      <span className="text-white font-bold">{selectedRelay.cleanName}</span>
                    </div>
                  </div>
                </div>
              </div>

              {/* Popular Games (Quick Launch Cards) */}
              <div className="flex flex-col gap-4">
                <div className="flex items-center justify-between">
                  <h3 className="font-brand font-bold text-lg text-white">Popular Competitive Games</h3>
                  <button
                    onClick={() => setActiveTab("library")}
                    className="text-xs font-brand font-bold text-[#00F0FF] hover:underline cursor-pointer border-none bg-transparent"
                  >
                    View All Games &rarr;
                  </button>
                </div>

                <div className="grid grid-cols-4 gap-5">
                  {games.slice(0, 4).map((game) => {
                    const originalIdx = games.findIndex((g) => g.id === game.id);
                    const isCurrentActive = isBoosting && selectedGame.id === game.id;
                    const protocolRoute = game.platforms.replace(/^PC\s*\((.*)\)$/, "$1");
                    return (
                      <div
                        key={game.id}
                        onClick={() => handleSelectGameAndBoost(originalIdx)}
                        className={`relative rounded-2xl overflow-hidden bg-[#0a0f16] border cursor-pointer card-smooth group flex flex-col shadow-lg ${
                          isCurrentActive
                            ? "border-[#10b981] shadow-[0_0_24px_rgba(16,185,129,0.25)]"
                            : "border-[#192433] hover:border-[#00F0FF]/50"
                        }`}
                      >
                        <div className="relative h-[150px] overflow-hidden bg-[#05080c]">
                          <img
                            src={game.coverImg}
                            alt={game.name}
                            className="w-full h-full object-cover transition-transform duration-700 ease-[cubic-bezier(0.16,1,0.3,1)] group-hover:scale-105"
                          />
                          <div className="absolute inset-0 bg-gradient-to-t from-[#0a0f16] via-transparent to-transparent" />
                          <span className="absolute top-3 right-3 px-2 py-0.5 rounded bg-black/70 backdrop-blur text-[10px] font-brand font-bold text-[#00F0FF] border border-white/10">
                            {game.tag}
                          </span>
                          {detectedGames[game.id]?.running && (
                            <span className="absolute top-3 left-3 px-2 py-0.5 rounded bg-[#10b981] text-[10px] font-brand font-bold text-black">
                              RUNNING
                            </span>
                          )}
                        </div>

                        <div className="p-4 flex flex-col gap-2 flex-1">
                          <div>
                            <h4 className="font-brand font-bold text-sm text-white truncate">{game.name}</h4>
                            <p className="text-[11px] text-white/40 truncate">{game.genre}</p>
                          </div>

                          <div className="flex items-center justify-between text-xs mt-auto pt-2.5 border-t border-white/5">
                            {isCurrentActive ? (
                              <>
                                <div className="flex items-center gap-2 font-mono text-[#10b981] font-bold text-xs">
                                  <span className="w-2 h-2 rounded-full bg-[#10b981] animate-pulse" />
                                  <span>{livePing > 0 ? `${Math.round(livePing)} ms` : "—"}</span>
                                </div>
                                <span className="text-[10px] font-brand font-bold text-[#10b981] bg-[#10b981]/10 border border-[#10b981]/30 px-2 py-0.5 rounded-full">
                                  TUNNEL LIVE
                                </span>
                              </>
                            ) : (
                              <>
                                <div className="flex items-center gap-1.5 text-white/50 text-[11px] truncate max-w-[140px]">
                                  <span className="text-[#00F0FF]/85 font-medium truncate">{protocolRoute}</span>
                                </div>
                                <span className="text-[10px] font-brand font-semibold text-white/60 bg-white/[0.04] border border-white/10 px-2 py-0.5 rounded-md group-hover:border-[#00F0FF]/40 group-hover:text-[#00F0FF] transition-colors duration-200">
                                  {game.profileAvailable ? "Profile Ready" : "Profile Pending"}
                                </span>
                              </>
                            )}
                          </div>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            </div>
          )}

          {/* ========================================================= */}
          {/* TAB 1: BOOST HUB (GEARUP PRO HUD)                         */}
          {/* ========================================================= */}
          {activeTab === "boost" && (
            <div className="flex flex-col gap-6 max-w-[1240px] mx-auto w-full tab-enter">
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
                          {selectedGame.publisher} &bull; {selectedRegion?.name || selectedGame.region}
                        </p>
                        <div className="flex flex-wrap items-center gap-2 mt-2">
                          {selectedGame.profileAvailable ? (
                            <>
                              <span className="text-[9px] font-brand font-bold text-white/35 uppercase tracking-wider">
                                Server region
                              </span>
                              <select
                                value={selectedRegion?.id || ""}
                                onChange={(e) => setSelectedRegionId(e.target.value)}
                                disabled={isBoosting}
                                className="h-7 max-w-[260px] px-2 rounded-lg bg-[#111b28] border border-[#24364b] text-[11px] font-mono text-[#00F0FF] outline-none focus:border-[#00F0FF] disabled:opacity-60"
                              >
                                {(selectedGame.regions || []).map((region) => (
                                  <option key={region.id} value={region.id}>
                                    {region.name}
                                  </option>
                                ))}
                              </select>
                              <span className="text-[9px] font-mono text-white/30">
                                {selectedRegion?.cidrs.length || 0} CIDRs
                              </span>
                            </>
                          ) : (
                            <span className="text-[10px] font-brand font-bold text-[#f59e0b]">
                              No verified region profile — add server CIDRs with “Add Custom Game” to enable routing.
                            </span>
                          )}
                        </div>
                      </div>
                    </div>

                    {/* TWO GIANT GEARUP METRICS: ESTIMATED PING & PACKET LOSS */}
                    <div className="grid grid-cols-3 gap-6 pt-2">
                      {/* 1. ESTIMATED PING */}
                      <div className="p-4 rounded-2xl bg-[#0f1722]/80 border border-[#1f2d40] flex flex-col">
                        <span className="text-[10px] font-brand font-bold text-white/45 uppercase tracking-wider">
                          {isBoosting ? "LIVE ACCELERATED PING" : "TARGET ROUTE LATENCY"}
                        </span>
                        <div className="flex items-baseline gap-1.5 mt-1">
                          <span className="font-mono font-black text-5xl text-[#00F0FF] tracking-tight">
                            {isBoosting && livePing > 0
                              ? Math.round(livePing)
                              : selectedRelay.pingMs > 0
                              ? Math.round(selectedRelay.pingMs)
                              : "—"}
                          </span>
                          <span className="font-mono text-sm text-white/50 font-bold">ms</span>
                        </div>
                        <span className="text-[10px] font-brand text-[#10b981] mt-1 flex items-center gap-1 font-semibold">
                          <span>⚡</span> {isBoosting && livePing > 0
                            ? "Live RTT measured by the tunnel engine"
                            : isBoosting
                            ? "Waiting for the first tunnel RTT sample"
                            : `Relay probe via ${selectedRelay.name || "Auto"} • Not connected`}
                        </span>
                      </div>

                      {/* 2. PACKET LOSS */}
                      <div className="p-4 rounded-2xl bg-[#0f1722]/80 border border-[#1f2d40] flex flex-col">
                        <span className="text-[10px] font-brand font-bold text-white/45 uppercase tracking-wider">
                          PACKET LOSS
                        </span>
                        <div className="flex items-baseline gap-1.5 mt-1">
                          <span className="font-mono font-black text-5xl text-white tracking-tight">
                            {isBoosting && packetLoss !== null ? packetLoss.toFixed(1) : "—"}
                          </span>
                          <span className="font-mono text-sm text-white/50 font-bold">%</span>
                        </div>
                        <span className="text-[10px] font-mono text-[#10b981] mt-1 font-semibold">
                          {isBoosting && packetLoss !== null
                            ? "Estimated from missed relay keepalives"
                            : "Connect to measure"}
                        </span>
                      </div>

                      {/* 3. JITTER VARIANCE */}
                      <div className="p-4 rounded-2xl bg-[#0f1722]/80 border border-[#1f2d40] flex flex-col">
                        <span className="text-[10px] font-brand font-bold text-white/45 uppercase tracking-wider">
                          JITTER VARIANCE
                        </span>
                        <div className="flex items-baseline gap-1.5 mt-1">
                          <span className="font-mono font-black text-5xl text-white tracking-tight">
                            {isBoosting && jitter > 0 ? `±${jitter}` : "—"}
                          </span>
                          <span className="font-mono text-sm text-white/50 font-bold">ms</span>
                        </div>
                        <span className="text-[10px] font-mono text-[#00F0FF] mt-1 font-semibold">
                          {isBoosting && jitter > 0 ? "Mean RTT variation" : "Connect to measure"}
                        </span>
                      </div>
                    </div>

                    {/* ACTION BUTTONS (BOOST NOW / LAUNCH GAME) */}
                    <div className="flex items-center gap-3 pt-2">
                      <button
                        onClick={handleToggleBoost}
                        disabled={isConnecting || (!isBoosting && (!selectedGame.profileAvailable || !selectedRegion || selectedRelay.id === "none"))}
                        className={`h-[50px] px-8 rounded-2xl flex items-center gap-3 font-brand font-extrabold text-sm uppercase tracking-wider cursor-pointer border-none transition-all duration-300 ease-[cubic-bezier(0.16,1,0.3,1)] shadow-xl ${
                          isBoosting
                            ? "bg-[#ef4444] hover:bg-[#dc2626] text-white shadow-red-500/20 active:scale-[0.98]"
                            : isConnecting
                            ? "bg-[#00F0FF]/70 text-black cursor-wait"
                            : !selectedGame.profileAvailable || !selectedRegion || selectedRelay.id === "none"
                            ? "bg-white/10 text-white/40 cursor-not-allowed"
                            : "bg-[#00F0FF] hover:bg-[#33f3ff] text-[#051119] shadow-[#00F0FF]/30 hover:scale-[1.02] active:scale-[0.98]"
                        }`}
                      >
                        <IconBolt className={`w-5 h-5 ${isConnecting ? "animate-spin" : isBoosting ? "animate-glow-pulse text-white" : ""}`} />
                        <span>
                          {isBoosting
                            ? "STOP ACCELERATION"
                            : isConnecting
                            ? "CONNECTING..."
                            : !selectedGame.profileAvailable || !selectedRegion
                            ? "PROFILE REQUIRED"
                            : selectedRelay.id === "none"
                            ? "RELAY REQUIRED"
                            : "BOOST NOW"}
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

                    {connectionError && (
                      <div className="p-3 rounded-xl bg-[#ef4444]/10 border border-[#ef4444]/30 text-[11px] text-[#fca5a5] font-mono leading-relaxed">
                        Tunnel error: {connectionError}
                      </div>
                    )}
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
                      <span className="text-[11px] font-brand font-bold text-white">{gamerNickname} (PC)</span>
                      <span className="text-[9px] font-mono text-white/40">127.0.0.1</span>
                      <span
                        className={`text-[9px] font-mono ${detectedGames[selectedGame.id]?.running ? "text-[#10b981]" : "text-white/30"}`}
                        title={
                          detectedGames[selectedGame.id]?.running
                            ? `Process scan found ${detectedGames[selectedGame.id]?.matchedProcess || "a game process"}`
                            : "The local process scan has not found this game's process yet"
                        }
                      >
                        {detectedGames[selectedGame.id]?.running
                          ? `${detectedGames[selectedGame.id]?.matchedProcess || "game"} running`
                          : "game process not detected"}
                      </span>
                    </div>

                    {/* Connecting Hop 1 -> 2 */}
                    <div className="flex-1 flex flex-col items-center px-3 relative">
                      <div className="w-full h-1 bg-[#182333] rounded relative overflow-hidden">
                        {isBoosting && (
                          <div className="absolute inset-y-0 left-0 w-1/3 bg-[#00F0FF] rounded animate-[pulse_1.5s_infinite]" />
                        )}
                      </div>
                      <span className="text-[10px] font-mono text-[#00F0FF] font-bold mt-1">
                        {networkDiag && networkDiag.lanPingMs > 0 ? `${networkDiag.lanPingMs} ms (LAN)` : "—"}
                      </span>
                    </div>

                    {/* Hop 2: Local Accelerator Engine */}
                    <div className="flex flex-col items-center gap-1.5 z-10">
                      <div className="w-11 h-11 rounded-xl bg-[#111925] border border-[#213042] flex items-center justify-center text-[#00F0FF] shadow">
                        <IconBolt className="w-5 h-5" />
                      </div>
                      <span className="text-[11px] font-brand font-bold text-white">Lagvex Core</span>
                      <span className="text-[9px] font-mono text-[#10b981]">{driverName}</span>
                    </div>

                    {/* Connecting Hop 2 -> 3 */}
                    <div className="flex-1 flex flex-col items-center px-3 relative">
                      <div className="w-full h-1 bg-[#182333] rounded relative overflow-hidden">
                        {isBoosting && (
                          <div className="absolute inset-y-0 left-0 w-1/3 bg-[#00F0FF] rounded animate-[pulse_1.2s_infinite]" />
                        )}
                      </div>
                      <span className="text-[10px] font-mono text-[#00F0FF] font-bold mt-1">
                        {isBoosting && livePing > 0
                          ? `${Math.round(livePing)} ms`
                          : selectedRelay.pingMs > 0
                          ? `${selectedRelay.pingMs} ms`
                          : "—"}
                      </span>
                    </div>

                    {/* Hop 3: Selected Super Relay Node */}
                    <div className="flex flex-col items-center gap-1.5 z-10">
                      <div className="w-11 h-11 rounded-xl bg-[#111925] border border-[#00F0FF]/40 flex items-center justify-center text-xl shadow-lg shadow-[#00F0FF]/15">
                        {selectedRelay.flag}
                      </div>
                      <span className="text-[11px] font-brand font-bold text-white">{selectedRelay.cleanName}</span>
                      <span className="text-[9px] font-mono text-[#00F0FF]">Relay Node</span>
                    </div>

                    {/* Connecting Hop 3 -> 4 */}
                    <div className="flex-1 flex flex-col items-center px-3 relative">
                      <div className="w-full h-1 bg-[#182333] rounded relative overflow-hidden">
                        {isBoosting && (
                          <div className="absolute inset-y-0 left-0 w-1/3 bg-[#10b981] rounded animate-[pulse_1s_infinite]" />
                        )}
                      </div>
                      <span className="text-[10px] font-mono text-[#10b981] font-bold mt-1">
                        {isBoosting ? "Relay Active" : "Direct Peering"}
                      </span>
                    </div>

                    {/* Hop 4: Official Game Server */}
                    <div className="flex flex-col items-center gap-1.5 z-10">
                      <div className="w-11 h-11 rounded-xl bg-[#111925] border border-[#213042] flex items-center justify-center text-white shadow">
                        <IconGamepad className="w-5 h-5 text-[#10b981]" />
                      </div>
                      <span className="text-[11px] font-brand font-bold text-white">{selectedGame.name} Server</span>
                      <span className="text-[9px] font-mono text-white/40">
                        {selectedRegion?.name || selectedGame.region}
                      </span>
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
                              <span className="text-[10px] text-white/40">{isBoosting ? "Watched" : "Idle"}</span>
                            </div>
                          ))}
                        </div>
                      </div>

                      <div className="p-4 rounded-xl bg-[#090e15] border border-[#16202e] flex flex-col gap-2">
                        <span className="text-[10px] font-brand font-bold text-[#00F0FF] uppercase tracking-wider">
                          ACTIVE IP CIDR SUBNETS
                        </span>
                        <div className="flex flex-col gap-1.5 font-mono text-xs text-white/80 max-h-[110px] overflow-y-auto">
                          {(selectedRegion?.cidrs || []).map((cidr) => (
                            <div key={cidr} className="flex items-center justify-between p-1.5 rounded bg-white/[0.02]">
                              <span className="flex items-center gap-2">
                                <span className="w-1.5 h-1.5 rounded-full bg-[#00F0FF]" />
                                <span>{cidr}</span>
                              </span>
                              <span className="text-[10px] text-[#00F0FF]">Split-Tunnel</span>
                            </div>
                          ))}
                          {(selectedRegion?.cidrs || []).length === 0 && (
                            <span className="text-[10px] text-[#f59e0b]">
                              No verified CIDRs for this game/region.
                            </span>
                          )}
                        </div>
                      </div>

                      {/* ROUTE DETAIL (MEASURED) */}
                      <div className="col-span-2 p-4 rounded-xl bg-[#090e15] border border-[#16202e] flex flex-col gap-3">
                        <div className="flex items-center justify-between gap-3">
                          <div className="flex flex-col">
                            <span className="text-[10px] font-brand font-bold text-[#00F0FF] uppercase tracking-wider">
                              ROUTE DETAIL (MEASURED)
                            </span>
                            <span className="text-[10px] text-white/40">
                              Every value is measured locally. Segments the client cannot observe say why.
                            </span>
                          </div>
                          <div className="flex items-center gap-2 shrink-0">
                            {routeDetail?.measuredAt && (
                              <span className="text-[10px] font-mono text-white/35">
                                measured {new Date(routeDetail.measuredAt).toLocaleTimeString()}
                              </span>
                            )}
                            <button
                              onClick={() => void handleMeasureRouteDetail()}
                              disabled={isMeasuringRoute}
                              className={`px-3 py-1.5 rounded-lg text-[10px] font-brand font-bold cursor-pointer border-none transition-all ${
                                isMeasuringRoute
                                  ? "bg-white/10 text-white/40 cursor-wait"
                                  : "bg-[#00F0FF]/15 hover:bg-[#00F0FF]/25 text-[#00F0FF]"
                              }`}
                            >
                              {isMeasuringRoute ? "Measuring…" : "Measure route"}
                            </button>
                          </div>
                        </div>

                        {routeDetail && (
                          <>
                            <div className="flex flex-col gap-1.5">
                              {routeDetail.hops.map((hop) => (
                                <div
                                  key={hop.id}
                                  className="flex items-center justify-between gap-3 p-2 rounded-lg bg-white/[0.02]"
                                >
                                  <div className="flex items-center gap-2 min-w-0">
                                    <span
                                      className={`w-2 h-2 rounded-full shrink-0 ${
                                        hop.measured ? "bg-[#10b981]" : "bg-white/20"
                                      }`}
                                    />
                                    <span className="text-xs font-brand font-bold text-white shrink-0">
                                      {hop.label}
                                    </span>
                                    {hop.target && (
                                      <span className="text-[10px] font-mono text-white/35 truncate">
                                        {hop.target}
                                      </span>
                                    )}
                                  </div>
                                  <div className="flex flex-col items-end shrink-0">
                                    <span
                                      className={`font-mono text-xs font-bold ${
                                        hop.measured ? "text-[#00F0FF]" : "text-white/30"
                                      }`}
                                    >
                                      {hop.measured && hop.rttMs ? `${hop.rttMs} ms` : "—"}
                                    </span>
                                    {hop.measured ? (
                                      <span className="text-[9px] text-white/40 font-mono text-right">
                                        {hop.note || ""}
                                        {hop.samples ? ` • ${hop.samples} sample(s)` : ""}
                                        {hop.rttAvgMs
                                          ? ` • avg ${hop.rttAvgMs}±${hop.rttStdMs ?? 0}`
                                          : ""}
                                        {typeof hop.lossPct === "number" && hop.lossPct > 0
                                          ? ` • loss ${hop.lossPct}%`
                                          : ""}
                                      </span>
                                    ) : (
                                      <span className="text-[9px] text-[#f59e0b]/80 font-mono text-right max-w-[340px]">
                                        {hop.reason || "not measured"}
                                      </span>
                                    )}
                                  </div>
                                </div>
                              ))}
                            </div>

                            <div className="flex items-center gap-3 flex-wrap">
                              {routeDetail.overheadMs !== 0 && (
                                <span className="text-[10px] font-mono text-white/60 bg-white/[0.04] px-2 py-1 rounded">
                                  Tunnel overhead vs direct probe:{" "}
                                  {routeDetail.overheadMs > 0 ? "+" : ""}
                                  {routeDetail.overheadMs} ms
                                </span>
                              )}
                              {routeDetail.anchor.target && (
                                <span className="text-[10px] font-mono text-white/60 bg-white/[0.04] px-2 py-1 rounded">
                                  {routeDetail.anchor.label} anchor (direct):{" "}
                                  {routeDetail.anchor.measured && routeDetail.anchor.rttMs
                                    ? `${routeDetail.anchor.rttMs} ms`
                                    : "—"}{" "}
                                  — {routeDetail.anchor.note}
                                </span>
                              )}
                              {routeDetail.anchor.relayMeasured && (
                                <span className="text-[10px] font-mono text-white/60 bg-white/[0.04] px-2 py-1 rounded">
                                  Relay &rarr; {routeDetail.anchor.label} anchor:{" "}
                                  {routeDetail.anchor.relayRttMs} ms (measured by the relay)
                                </span>
                              )}
                              {routeDetail.anchor.improvementMeasured && (
                                <span
                                  className={`text-[10px] font-mono px-2 py-1 rounded bg-white/[0.04] ${
                                    (routeDetail.anchor.improvementMs ?? 0) > 0
                                      ? "text-[#10b981]"
                                      : "text-[#f59e0b]"
                                  }`}
                                >
                                  Estimated via-relay improvement:{" "}
                                  {(routeDetail.anchor.improvementMs ?? 0) > 0 ? "+" : ""}
                                  {routeDetail.anchor.improvementMs} ms (region anchor estimate)
                                </span>
                              )}
                            </div>

                            {routeDetail.notes.map((note, idx) => (
                              <span key={idx} className="text-[9px] text-white/30">
                                {note}
                              </span>
                            ))}
                          </>
                        )}
                      </div>

                      {/* PEER ROUTING (OBSERVED PROCESS TRAFFIC) */}
                      <div className="col-span-2 p-4 rounded-xl bg-[#090e15] border border-[#16202e] flex flex-col gap-3">
                        <div className="flex items-center justify-between gap-3">
                          <div className="flex flex-col">
                            <span className="text-[10px] font-brand font-bold text-[#00F0FF] uppercase tracking-wider">
                              PEER ROUTING (OBSERVED PROCESS TRAFFIC)
                            </span>
                            <span className="text-[10px] text-white/40">
                              Routes only the public IPs the detected game process actually talks to, read from the OS
                              socket tables. Off by default; stale peers are removed automatically.
                            </span>
                          </div>
                          <button
                            onClick={() => void handleTogglePeerRouting()}
                            className={`h-[34px] px-4 rounded-xl text-[10px] font-brand font-bold cursor-pointer border-none transition-all shrink-0 ${
                              peerRouting
                                ? "bg-[#10b981] text-black shadow-md shadow-[#10b981]/20"
                                : "bg-white/[0.08] text-white/60 hover:bg-white/15"
                            }`}
                          >
                            {peerRouting ? "PEER ROUTING: ON" : "PEER ROUTING: OFF"}
                          </button>
                        </div>

                        <div className="flex items-center gap-3 flex-wrap">
                          <span className="text-[10px] font-mono text-white/60 bg-white/[0.04] px-2 py-1 rounded">
                            Active peer routes: {peerRouteCount}
                          </span>
                          {peerRoutes.length > 0 && (
                            <span className="text-[10px] font-mono text-white/45 truncate max-w-[560px]">
                              {peerRoutes.slice(0, 8).join(", ")}
                              {peerRoutes.length > 8 ? ` +${peerRoutes.length - 8} more` : ""}
                            </span>
                          )}
                        </div>

                        <span className="text-[9px] text-white/30">
                          On Windows only TCP peers are visible to the OS; UDP peers are not observed there, so peer
                          routing may need the region CIDRs as well.
                        </span>
                      </div>

                      {/* STANDBY RELAY (PRE-MEASURED) */}
                      <div className="col-span-2 p-4 rounded-xl bg-[#090e15] border border-[#16202e] flex flex-col gap-2">
                        <div className="flex items-center justify-between gap-3">
                          <span className="text-[10px] font-brand font-bold text-[#00F0FF] uppercase tracking-wider">
                            STANDBY RELAY (PRE-MEASURED)
                          </span>
                          {standby ? (
                            <span className="text-[10px] font-mono text-[#10b981]">
                              {standby.name} — {standby.rttMs} ms
                              {standby.totalMs > 0 ? ` (via relay ${standby.totalMs} ms)` : ""} •{" "}
                              {standby.ageSec}s ago
                            </span>
                          ) : (
                            <span className="text-[10px] font-mono text-white/35">No standby measured yet</span>
                          )}
                        </div>
                        <span className="text-[9px] text-white/30">
                          While a tunnel is up, one alternative relay is probed in the background. Automatic failover
                          promotes it when the active path degrades, so the switch does not wait for a fresh sweep.
                        </span>
                      </div>

                      {/* FULL TUNNEL (OPT-IN) */}
                      <div className="col-span-2 p-4 rounded-xl bg-[#090e15] border border-[#16202e] flex flex-col gap-3">
                        <div className="flex items-center justify-between gap-3">
                          <div className="flex flex-col">
                            <span className="text-[10px] font-brand font-bold text-[#f59e0b] uppercase tracking-wider">
                              FULL TUNNEL (ALL IPv4 TRAFFIC)
                            </span>
                            <span className="text-[10px] text-white/40">
                              Sends everything — browser, Discord, updates — through the relay, not just game traffic.
                              Off by default.
                            </span>
                          </div>
                          <button
                            onClick={() => void handleToggleFullTunnel()}
                            className={`h-[34px] px-4 rounded-xl text-[10px] font-brand font-bold cursor-pointer border-none transition-all shrink-0 ${
                              fullTunnel
                                ? "bg-[#f59e0b] text-black shadow-md shadow-[#f59e0b]/20"
                                : "bg-white/[0.08] text-white/60 hover:bg-white/15"
                            }`}
                          >
                            {fullTunnel ? "FULL TUNNEL: ON" : "FULL TUNNEL: OFF"}
                          </button>
                        </div>

                        <span className="text-[9px] text-[#f59e0b]/80">
                          If the relay is slow or unreachable while this is on, all normal browsing slows down or stops
                          too. The relay route stays pinned and DNS still uses your local resolver.
                        </span>
                      </div>
                    </div>
                  ) : (
                    <>
                    <div className="mt-4 p-4 rounded-xl bg-[#090e15] border border-[#16202e] flex flex-col gap-2">
                      <div className="flex items-center justify-between">
                        <span className="text-[10px] font-brand font-bold text-[#00F0FF] uppercase tracking-wider">
                          LIVE ENGINE RTT SAMPLES
                        </span>
                        <span className="text-[10px] font-mono text-white/45">
                          Latest: {livePing > 0 ? `${Math.round(livePing)}ms` : "—"} &bull; Jitter: {isBoosting && jitter > 0 ? `±${jitter}ms` : "—"}
                        </span>
                      </div>
                      <div className="h-[120px] w-full rounded-lg overflow-hidden bg-black/40">
                        <canvas ref={historyCanvasRef} width={900} height={120} className="w-full h-full" />
                      </div>
                    </div>

                    <div className="mt-4 p-4 rounded-xl bg-[#090e15] border border-[#16202e] flex flex-col gap-2">
                      <div className="flex items-center justify-between">
                        <span className="text-[10px] font-brand font-bold text-[#00F0FF] uppercase tracking-wider">
                          TUNNEL THROUGHPUT (MEASURED)
                        </span>
                        <span className="text-[10px] font-mono text-white/45">
                          <span className="text-[#10b981]">↓</span>{" "}
                          {formatRate(rateHistory.length > 0 ? rateHistory[rateHistory.length - 1].down : 0)} &bull;{" "}
                          <span className="text-[#00F0FF]">↑</span>{" "}
                          {formatRate(rateHistory.length > 0 ? rateHistory[rateHistory.length - 1].up : 0)}
                        </span>
                      </div>
                      <div className="h-[100px] w-full rounded-lg overflow-hidden bg-black/40">
                        <canvas ref={trafficCanvasRef} width={900} height={100} className="w-full h-full" />
                      </div>
                      <div className="flex items-center justify-between text-[10px] font-mono text-white/45">
                        <span>
                          Session <span className="text-[#10b981]">↓</span> {formatBytes(sessionBytes.down)} &bull;{" "}
                          <span className="text-[#00F0FF]">↑</span> {formatBytes(sessionBytes.up)}
                        </span>
                        <span>
                          {rateHistory.length < 2
                            ? "Waiting for measured samples…"
                            : `${rateHistory.length} samples @ 2s`}
                        </span>
                      </div>
                    </div>

                    <div className="mt-4 p-4 rounded-xl bg-[#090e15] border border-[#16202e] flex flex-col gap-2">
                      <div className="flex items-center justify-between">
                        <span className="text-[10px] font-brand font-bold text-[#00F0FF] uppercase tracking-wider">
                          RELAY FAILOVERS (THIS SESSION)
                        </span>
                        <span className="text-[10px] font-mono text-white/45">
                          {failoverEvents.length === 0
                            ? "No failovers yet"
                            : `${failoverEvents.length} handover${failoverEvents.length === 1 ? "" : "s"}`}
                        </span>
                      </div>
                      {failoverEvents.length === 0 ? (
                        <span className="text-[10px] text-white/35">
                          Automatic failover only switches when the measured path degrades; nothing has switched this
                          session.
                        </span>
                      ) : (
                        <div className="flex flex-col gap-1.5 font-mono text-[10px] text-white/70 max-h-[120px] overflow-y-auto">
                          {failoverEvents
                            .slice(-3)
                            .reverse()
                            .map((ev, idx) => (
                              <div
                                key={idx}
                                className="flex items-center justify-between gap-3 p-1.5 rounded bg-white/[0.02]"
                              >
                                <span className="truncate">
                                  {new Date(ev.timestamp).toLocaleTimeString()} &bull;{" "}
                                  {ev.oldRelayName || ev.oldRelayId} &rarr; {ev.newRelayName || ev.newRelayId}
                                </span>
                                <span className="shrink-0 text-white/50">
                                  {Math.round(ev.oldPingMs)} &rarr; {Math.round(ev.newPingMs)} ms
                                  {ev.handoverOk ? "" : " (failed)"}
                                </span>
                              </div>
                            ))}
                        </div>
                      )}
                    </div>
                    </>
                  )}

                  {/* GAME SERVER CONNECTIONS (MEASURED, READ-ONLY) */}
                  <div className="mt-4 p-4 rounded-xl bg-[#090e15] border border-[#16202e] flex flex-col gap-3">
                    <div className="flex items-center justify-between gap-3">
                      <div className="flex flex-col">
                        <span className="text-[10px] font-brand font-bold text-[#00F0FF] uppercase tracking-wider">
                          GAME SERVER CONNECTIONS
                        </span>
                        <span className="text-[10px] text-white/40">
                          Read from the local OS socket tables. Nothing is injected into {selectedGame.name}.
                        </span>
                      </div>
                      <div className="flex items-center gap-2 shrink-0">
                        {gameConnections?.scannedAt && (
                          <span className="text-[10px] font-mono text-white/35">
                            scanned {new Date(gameConnections.scannedAt).toLocaleTimeString()}
                          </span>
                        )}
                        <button
                          onClick={() => void handleScanGameConnections()}
                          disabled={isScanningConnections}
                          className={`px-3 py-1.5 rounded-lg text-[10px] font-brand font-bold cursor-pointer border-none transition-all ${
                            isScanningConnections
                              ? "bg-white/10 text-white/40 cursor-wait"
                              : "bg-[#00F0FF]/15 hover:bg-[#00F0FF]/25 text-[#00F0FF]"
                          }`}
                        >
                          {isScanningConnections ? "Scanning…" : "Scan now"}
                        </button>
                      </div>
                    </div>

                    {gameConnections && !gameConnections.supported && (
                      <span className="text-[11px] text-[#f59e0b]">
                        {gameConnections.note || "Socket enumeration is unavailable on this system."}
                      </span>
                    )}

                    {gameConnections?.supported && (!selectedGameConnections || !selectedGameConnections.running) && (
                      <span className="text-[11px] text-white/40">
                        Game process not detected. Launch the game or join a match, then scan again.
                      </span>
                    )}

                    {gameConnections?.supported && selectedGameConnections?.running && (
                      <>
                        <span className="text-[10px] font-mono text-white/40">
                          {selectedGameConnections.process} (PID {selectedGameConnections.pid}) &bull;{" "}
                          {selectedGameConnections.connections.length} unique endpoint(s)
                        </span>
                        {selectedGameConnections.connections.length === 0 ? (
                          <span className="text-[11px] text-white/40">
                            No connected sockets seen yet. Enter a match and scan again.
                          </span>
                        ) : (
                          <div className="flex flex-col gap-1.5 font-mono text-xs text-white/80 max-h-[140px] overflow-y-auto">
                            {selectedGameConnections.connections.map((conn, idx) => (
                              <div
                                key={`${conn.protocol}-${conn.remote || conn.local}-${idx}`}
                                className="flex items-center justify-between gap-3 p-1.5 rounded bg-white/[0.02]"
                              >
                                <span className="flex items-center gap-2 min-w-0">
                                  <span
                                    className={`text-[9px] font-brand font-bold px-1.5 py-0.5 rounded ${
                                      conn.protocol === "udp"
                                        ? "bg-[#00F0FF]/15 text-[#00F0FF]"
                                        : "bg-[#7E42FF]/15 text-[#a78bfa]"
                                    }`}
                                  >
                                    {conn.protocol.toUpperCase()}
                                  </span>
                                  <span className="truncate">
                                    {conn.remote || `${conn.local} (local only)`}
                                  </span>
                                  {conn.state && <span className="text-[9px] text-white/30">{conn.state}</span>}
                                </span>
                                {conn.remote && (
                                  <button
                                    onClick={() => {
                                      navigator.clipboard?.writeText(conn.remote);
                                      notify(`Copied ${conn.remote}`);
                                    }}
                                    className="text-[10px] text-white/40 hover:text-white bg-transparent border-none cursor-pointer shrink-0"
                                  >
                                    Copy
                                  </button>
                                )}
                              </div>
                            ))}
                          </div>
                        )}
                        {recentPeers.length > 0 && (
                          <span className="text-[9px] font-mono text-white/35">
                            Recent endpoints this session: {recentPeers.map((p) => p.endpoint).join(", ")}
                          </span>
                        )}
                        {gameConnections.note && (
                          <span className="text-[9px] text-white/30">{gameConnections.note}</span>
                        )}
                      </>
                    )}
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* ========================================================= */}
          {/* TAB 2: GAME LIBRARY                                       */}
          {/* ========================================================= */}
          {activeTab === "library" && (
            <div className="flex flex-col gap-6 max-w-[1240px] mx-auto w-full tab-enter">
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
                    placeholder="Search games, publishers, or process names..."
                    className="w-full h-full pl-11 pr-4 rounded-2xl bg-[#0f1722] border border-[#1c293a] text-xs text-white outline-none focus:border-[#00F0FF]/60 placeholder:text-white/30"
                  />
                </div>

                {/* Categories & Add Custom Game */}
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
                  <button
                    onClick={() => setIsAddingGame(true)}
                    className="px-4 py-2 rounded-xl text-xs font-brand font-bold bg-[#00F0FF]/15 hover:bg-[#00F0FF]/25 text-[#00F0FF] border border-[#00F0FF]/30 cursor-pointer transition-all flex items-center gap-1.5"
                  >
                    <span>+ Add Custom Game</span>
                  </button>
                </div>
              </div>

              {/* Game Cards Grid */}
              <div className="grid grid-cols-4 gap-5">
                {filteredGames.map((game) => {
                  const originalIdx = games.findIndex((g) => g.id === game.id);
                  const isCurrent = selectedGame.id === game.id;
                  const isCurrentActive = isBoosting && isCurrent;
                  const protocolRoute = game.platforms.replace(/^PC\s*\((.*)\)$/, "$1");
                  return (
                    <div
                      key={game.id}
                      onClick={() => handleSelectGameAndBoost(originalIdx)}
                      className={`relative rounded-2xl overflow-hidden bg-[#0a0f16] border cursor-pointer card-smooth group flex flex-col shadow-lg ${
                        isCurrentActive
                          ? "border-[#10b981] shadow-[0_0_24px_rgba(16,185,129,0.25)]"
                          : isCurrent
                          ? "border-[#00F0FF] shadow-[0_0_24px_rgba(0,240,255,0.25)]"
                          : "border-[#192433] hover:border-[#00F0FF]/50"
                      }`}
                    >
                      {/* Game Image Banner */}
                      <div className="relative h-[160px] overflow-hidden bg-[#05080c]">
                        <img
                          src={game.coverImg}
                          alt={game.name}
                          className="w-full h-full object-cover transition-transform duration-700 ease-[cubic-bezier(0.16,1,0.3,1)] group-hover:scale-105"
                        />
                        <div className="absolute inset-0 bg-gradient-to-t from-[#0a0f16] via-transparent to-transparent" />
                        <span className="absolute top-3 right-3 px-2 py-0.5 rounded bg-black/70 backdrop-blur text-[10px] font-brand font-bold text-[#00F0FF] border border-white/10">
                          {game.tag}
                        </span>
                        {detectedGames[game.id]?.running && (
                          <span className="absolute top-3 left-3 px-2 py-0.5 rounded bg-[#10b981] text-[10px] font-brand font-bold text-black">
                            RUNNING
                          </span>
                        )}
                      </div>

                      {/* Game Info */}
                      <div className="p-4 flex flex-col gap-2 flex-1">
                        <div>
                          <h3 className="font-brand font-bold text-sm text-white truncate">{game.name}</h3>
                          <p className="text-[11px] text-white/40 truncate">{game.genre} &bull; {game.region}</p>
                        </div>

                        {/* Real Network Routing Pill */}
                        <div className="mt-auto pt-2.5 border-t border-white/5 flex items-center justify-between text-xs">
                          {isCurrentActive ? (
                            <>
                              <div className="flex items-center gap-2 font-mono text-[#10b981] font-bold text-xs">
                                <span className="w-2 h-2 rounded-full bg-[#10b981] animate-pulse" />
                                <span>{livePing > 0 ? `${Math.round(livePing)} ms` : "—"}</span>
                              </div>
                              <span className="text-[10px] font-brand font-bold text-[#10b981] bg-[#10b981]/10 border border-[#10b981]/30 px-2 py-0.5 rounded-full">
                                TUNNEL LIVE
                              </span>
                            </>
                          ) : (
                            <>
                              <div className="flex items-center gap-1.5 text-white/50 text-[11px] truncate max-w-[140px]">
                                <span className="text-[#00F0FF]/85 font-medium truncate">{protocolRoute}</span>
                              </div>
                              <span className="text-[10px] font-brand font-semibold text-white/60 bg-white/[0.04] border border-white/10 px-2 py-0.5 rounded-md group-hover:border-[#00F0FF]/40 group-hover:text-[#00F0FF] transition-colors duration-200">
                                {game.profileAvailable ? "Profile Ready" : "Profile Pending"}
                              </span>
                            </>
                          )}
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {/* ========================================================= */}
          {/* TAB 3: SQUAD SYNC (REALISTIC SQUAD FLOW - NO FAKE DATA)  */}
          {/* ========================================================= */}
          {activeTab === "squad" && (
            <div className="flex flex-col gap-6 max-w-[1240px] mx-auto w-full tab-enter">
              {/* Informative Explanation Banner */}
              <div className="p-6 rounded-3xl bg-gradient-to-r from-[#0a121e] to-[#081724] border border-[#1b2b3d] flex items-center justify-between shadow-xl">
                <div className="max-w-[720px] flex flex-col gap-2">
                  <div className="flex items-center gap-2 text-xs font-brand font-bold text-[#00F0FF] uppercase tracking-wider">
                    <IconSquadUsers className="w-4 h-4" />
                    <span>SQUAD ROSTER</span>
                  </div>
                  <h2 className="font-brand font-extrabold text-2xl text-white">
                    Keep Your Party Together
                  </h2>
                  <p className="text-xs text-white/60 leading-relaxed">
                    Squad rooms share a live roster and each member's reported RTT. They do not change relay selection or game routing. Rooms are relayed through the connected relay; without one, a room only exists on this machine.
                  </p>
                </div>

                <div className="flex flex-col items-end gap-2 shrink-0">
                  <span className={`px-3 py-1 rounded-full text-xs font-brand font-bold border ${
                    currentSquadRoom
                      ? "bg-[#10b981]/15 text-[#10b981] border-[#10b981]/30"
                      : "bg-white/[0.05] text-white/40 border-white/10"
                  }`}>
                    {currentSquadRoom ? `ROOM ACTIVE (${squadMembers.length}/5)` : "NOT IN A ROOM"}
                  </span>
                  <span className="text-[10px] font-mono text-white/40">Gamer: {gamerNickname}</span>
                </div>
              </div>

              {/* LOBBY / ROOM STATE SWITCHER */}
              {!currentSquadRoom ? (
                /* STATE A: NOT IN ROOM (LOBBY CONTROLS) */
                <div className="grid grid-cols-2 gap-6">
                  {/* Card 1: Create Room */}
                  <div className="p-6 rounded-3xl bg-[#0b1018] border border-[#1c2738] flex flex-col justify-between shadow-lg">
                    <div className="flex flex-col gap-2">
                      <span className="text-xs font-brand font-bold text-[#00F0FF] uppercase tracking-wider">
                        HOST NEW PARTY
                      </span>
                      <h3 className="font-brand font-extrabold text-xl text-white">
                        Create a Squad Room
                      </h3>
                      <p className="text-xs text-white/50 leading-relaxed">
                        Start a party as Host. You get a Room Code to share and everyone who joins appears in the same roster. Squad members still connect through their own relay.
                      </p>
                    </div>

                    <div className="pt-6">
                      <button
                        onClick={handleCreateSquadRoom}
                        className="h-[48px] w-full rounded-2xl bg-[#00F0FF] hover:bg-[#33f3ff] text-black font-brand font-bold text-xs uppercase tracking-wider cursor-pointer border-none transition-all shadow-lg shadow-[#00F0FF]/20"
                      >
                        CREATE SQUAD ROOM
                      </button>
                    </div>
                  </div>

                  {/* Card 2: Join Room */}
                  <div className="p-6 rounded-3xl bg-[#0b1018] border border-[#1c2738] flex flex-col justify-between shadow-lg">
                    <div className="flex flex-col gap-2">
                      <span className="text-xs font-brand font-bold text-white/40 uppercase tracking-wider">
                        JOIN EXISTING PARTY
                      </span>
                      <h3 className="font-brand font-extrabold text-xl text-white">
                        Join Teammate's Squad
                      </h3>
                      <p className="text-xs text-white/50 leading-relaxed">
                        Enter the Room Code from your party leader to join their roster. Joining does not change your relay or game route.
                      </p>
                    </div>

                    <div className="pt-4 flex items-center gap-3">
                      <input
                        type="text"
                        value={joinInputCode}
                        onChange={(e) => setJoinInputCode(e.target.value.toUpperCase())}
                        placeholder="e.g. LGVX-9821"
                        className="flex-1 h-[48px] px-4 rounded-xl bg-[#121b27] border border-[#233347] text-sm font-mono font-bold text-white outline-none focus:border-[#00F0FF] placeholder:text-white/30"
                      />
                      <button
                        onClick={handleJoinSquadRoom}
                        className="h-[48px] px-6 rounded-xl bg-white/[0.08] hover:bg-white/15 text-white font-brand font-bold text-xs uppercase tracking-wider cursor-pointer border border-white/10 transition-all shrink-0"
                      >
                        JOIN
                      </button>
                    </div>
                  </div>
                </div>
              ) : (
                /* STATE B: IN ACTIVE SQUAD ROOM */
                <div className="flex flex-col gap-6">
                  {/* Room Status Card */}
                  <div className="p-6 rounded-3xl bg-[#0b1018] border border-[#1c2738] flex items-center justify-between shadow-lg">
                    <div className="flex items-center gap-5">
                      <div className="flex flex-col">
                        <span className="text-[10px] font-brand text-white/40 uppercase">ROOM CODE</span>
                        <span className="font-mono font-black text-2xl text-[#00F0FF] tracking-wider">
                          {currentSquadRoom}
                        </span>
                      </div>
                      <button
                        onClick={() => {
                          navigator.clipboard?.writeText(currentSquadRoom);
                          notify("Copied Squad Room Code to clipboard!");
                        }}
                        className="px-3 py-2 rounded-xl bg-white/[0.05] hover:bg-white/10 text-white font-brand text-xs font-bold flex items-center gap-1.5 cursor-pointer border-none transition-all"
                      >
                        <IconCopy className="w-4 h-4" />
                        <span>Copy Code</span>
                      </button>
                    </div>

                    <div className="flex items-center gap-3">
                      <button
                        onClick={handleLeaveSquadRoom}
                        className="px-4 py-2 rounded-xl bg-red-500/10 hover:bg-red-500/20 text-red-400 font-brand text-xs font-bold cursor-pointer border border-red-500/20 transition-all"
                      >
                        Leave Squad
                      </button>
                    </div>
                  </div>

                  {/* Members Roster */}
                  <div className="p-6 rounded-3xl bg-[#0b1018] border border-[#1c2738] flex flex-col gap-4 shadow-lg">
                    <div className="flex items-center justify-between">
                      <h3 className="font-brand font-bold text-base text-white">
                        Squad Party Members ({squadMembers.length}/5)
                      </h3>
                      <span className="text-[10px] font-mono text-white/40">
                        Reported by each member
                      </span>
                    </div>

                    <div className="flex flex-col gap-2.5">
                      {squadMembers.map((member, idx) => (
                        <div
                          key={idx}
                          className={`flex items-center justify-between p-3.5 rounded-2xl border transition-all ${
                            member.isSelf
                              ? "bg-[#00F0FF]/10 border-[#00F0FF]/40 shadow-[0_0_15px_rgba(0,240,255,0.1)]"
                              : "bg-[#101622] border-[#1b2738]"
                          }`}
                        >
                          <div className="flex items-center gap-3">
                            <div className="w-9 h-9 rounded-xl bg-gradient-to-tr from-[#00F0FF]/20 to-[#7E42FF]/20 border border-[#00F0FF]/30 flex items-center justify-center font-brand font-bold text-xs text-[#00F0FF]">
                              {member.name.charAt(0).toUpperCase()}
                            </div>
                            <div>
                              <div className="flex items-center gap-2">
                                <span className="font-brand font-bold text-xs text-white">{member.name}</span>
                                <span className="text-[9px] font-brand font-bold px-1.5 py-0.2 rounded bg-white/[0.06] text-white/60">
                                  {member.role}
                                </span>
                                {member.isSelf && (
                                  <span className="text-[9px] font-brand font-bold px-1.5 py-0.2 rounded bg-[#00F0FF]/20 text-[#00F0FF]">
                                    YOU
                                  </span>
                                )}
                              </div>
                              <span className="text-[10px] text-white/40 font-mono">
                                {member.game || "Game not reported"}
                                {member.isp ? ` \u2022 Gateway ${member.isp}` : ""}
                              </span>
                            </div>
                          </div>

                          <div className="flex items-center gap-6">
                            <div className="flex flex-col items-end">
                              <span className="font-mono font-bold text-xs text-[#10b981]">
                                {member.ping > 0 ? `${member.ping} ms` : "—"}
                              </span>
                              <span className="text-[9px] text-white/40">Reported RTT</span>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>

                    {squadMembers.length === 1 && (
                      <div className="p-4 rounded-2xl bg-white/[0.02] border border-white/5 flex items-center justify-between text-xs text-white/50 font-medium">
                        <span>Waiting for teammates to join room <strong className="text-white">{currentSquadRoom}</strong>...</span>
                        <span className="text-[11px] text-[#00F0FF]">Share Room Code via Discord / Steam / Party Chat</span>
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>
          )}

          {/* ========================================================= */}
          {/* TAB 4: GLOBAL RELAY NODES                                 */}
          {/* ========================================================= */}
          {activeTab === "nodes" && (
            <div className="flex flex-col gap-6 max-w-[1240px] mx-auto w-full tab-enter">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="font-brand font-extrabold text-2xl text-white">Relay Nodes</h2>
                  <p className="text-xs text-white/50">
                    Relays configured in this client. RTT and loss come from the local prober; no public nodes are bundled. Add your own relay endpoint and PSK to begin.
                  </p>
                </div>
                <div className="flex items-center gap-3">
                  <button
                    onClick={() => setIsAddingNode(true)}
                    className="px-4 py-2 rounded-xl bg-[#00F0FF]/15 hover:bg-[#00F0FF]/25 border border-[#00F0FF]/30 text-xs font-brand font-bold text-[#00F0FF] flex items-center gap-1.5 cursor-pointer transition-all"
                  >
                    <span>+ Add Custom Node</span>
                  </button>
                  <button
                    onClick={async () => {
                      notify(
                        selectedRegion?.id
                          ? "⚡ Probing relays (client RTT + relay → region anchor)..."
                          : "⚡ Probing global routing nodes in real-time..."
                      );
                      try {
                        const params = new URLSearchParams();
                        if (selectedGame.profileAvailable) params.set("gameId", selectedGame.id);
                        if (selectedRegion?.id) params.set("regionId", selectedRegion.id);
                        const query = params.toString();
                        const res = await fetch(`/api/probe-relays${query ? `?${query}` : ""}`);
                        if (res.ok) {
                          const data = await res.json();
                          const mapped = mapProbeRelays(data);
                          setRelays(mapped);
                          if (mapped.length === 0) {
                            notify("No relays are configured. Deploy a relay or add a custom node first.");
                          } else if (selectedRegion?.id) {
                            notify("✅ Relays ranked by measured client + relay → region anchor RTT.");
                          } else {
                            notify("✅ Relays reprobed with measured RTT and loss.");
                          }
                        }
                      } catch {
                        notify("Node probe sweep completed.");
                      }
                    }}
                    className="px-4 py-2 rounded-xl bg-[#101722] hover:bg-[#162130] border border-[#223042] text-xs font-brand font-bold text-white flex items-center gap-2 cursor-pointer transition-all"
                  >
                    <IconRefresh className="w-4 h-4 text-[#00F0FF]" />
                    <span>Refresh All Latencies</span>
                  </button>
                </div>
              </div>

              {/* Node Search and Continent Filter Bar */}
              <div className="flex items-center justify-between gap-4">
                <div className="flex-1 relative h-[44px] max-w-[420px]">
                  <div className="absolute left-4 top-1/2 -translate-y-1/2 text-white/35">
                    <IconSearch className="w-4 h-4" />
                  </div>
                  <input
                    type="text"
                    value={nodeSearchQuery}
                    onChange={(e) => setNodeSearchQuery(e.target.value)}
                    placeholder="Search node name, city, or datacenter..."
                    className="w-full h-full pl-11 pr-4 rounded-2xl bg-[#0f1722] border border-[#1c293a] text-xs text-white outline-none focus:border-[#00F0FF]/60 placeholder:text-white/30"
                  />
                </div>

                {/* Continent Chips */}
                <div className="flex items-center gap-2">
                  {(["ALL", "ASIA", "NORTH AMERICA", "EUROPE", "OCEANIA", "GLOBAL"] as const).map((cont) => (
                    <button
                      key={cont}
                      onClick={() => setSelectedNodeContinent(cont)}
                      className={`px-3.5 py-1.5 rounded-xl text-xs font-brand font-bold cursor-pointer border-none transition-all ${
                        selectedNodeContinent === cont
                          ? "bg-[#00F0FF] text-black shadow-md shadow-[#00F0FF]/20"
                          : "bg-[#0f1722] text-white/50 hover:text-white hover:bg-[#162233]"
                      }`}
                    >
                      {cont}
                    </button>
                  ))}
                </div>
              </div>

              {filteredRelays.length === 0 && (
                <div className="p-6 rounded-2xl bg-[#0a0f16] border border-[#1b2636] flex flex-col gap-2">
                  <h3 className="font-brand font-bold text-sm text-white">No relay nodes configured</h3>
                  <p className="text-xs text-white/50 leading-relaxed">
                    Lagvex does not ship public community relays. Deploy your own relay on a Linux VPS, then add its endpoint and PSK with <strong className="text-white">+ Add Custom Node</strong>. On Linux you can also add <strong className="text-white">127.0.0.1:4433</strong> to start the embedded local relay with your own PSK.
                  </p>
                </div>
              )}

              <div className="grid grid-cols-2 gap-4">
                {filteredRelays.map((relay) => {
                  const originalIdx = relays.findIndex((r) => r.id === relay.id);
                  const isSelected = selectedRelayIdx === originalIdx;
                  return (
                    <div
                      key={relay.id}
                      onClick={() => {
                        if (originalIdx >= 0) setSelectedRelayIdx(originalIdx);
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
                            {!isSelected && relay.isOptimal && relay.reachable && (
                              <span
                                className="text-[9px] font-brand font-bold px-1.5 py-0.5 rounded bg-[#10b981]/20 text-[#10b981]"
                                title="Top-ranked by the last measured sweep"
                              >
                                BEST ROUTE
                              </span>
                            )}
                            {relay.recent && (
                              <span
                                className="text-[9px] font-brand font-bold px-1.5 py-0.5 rounded bg-white/[0.08] text-white/60"
                                title="You connected through this relay recently"
                              >
                                RECENT
                              </span>
                            )}
                            {relay.unstable && (
                              <span
                                className="text-[9px] font-brand font-bold px-1.5 py-0.5 rounded bg-[#f59e0b]/20 text-[#f59e0b]"
                                title="Retained probe history shows high loss or a large RTT spread"
                              >
                                UNSTABLE
                              </span>
                            )}
                          </div>
                          <p className="text-xs text-white/40 font-mono mt-0.5">{relay.location || "No location set"}</p>
                          <p className="text-[10px] text-white/30 font-mono mt-1">{relay.address}</p>
                        </div>
                      </div>

                      <div className="flex items-center gap-5">
                        <div className="flex flex-col items-end gap-1">
                          <span className={`font-mono font-black text-xl ${relay.reachable ? "text-[#00F0FF]" : "text-white/30"}`}>
                            {relay.reachable && relay.pingMs > 0 ? `${relay.pingMs} ms` : "—"}
                          </span>
                          <span className={`text-[10px] font-semibold ${relay.reachable ? "text-[#10b981]" : "text-[#ef4444]"}`}>
                            {relay.reachable
                              ? `Loss: ${relay.packetLoss.toFixed(1)}%${relay.historySamples > 1 && relay.lossAvgPct > 0 ? ` (avg ${relay.lossAvgPct.toFixed(1)}%)` : ""}`
                              : "Unreachable"}
                          </span>
                          <span className="text-[9px] font-mono text-white/35">
                            {relay.historySamples > 0 ? `${relay.historySamples} probe${relay.historySamples === 1 ? "" : "s"} kept` : "No probe history yet"}
                          </span>
                          {relay.icmpRttMs > 0 && (
                            <span
                              className="text-[9px] font-mono text-white/40"
                              title="ICMP echo measured locally, when the OS allows unprivileged ICMP"
                            >
                              ICMP {relay.icmpRttMs} ms
                            </span>
                          )}
                          {relay.totalEstimateMs > 0 && (
                            <span className="text-[9px] font-mono text-[#10b981]">
                              via relay: {relay.totalEstimateMs} ms total (relay → region {relay.relayAnchorRttMs} ms)
                            </span>
                          )}
                        </div>

                        <div className="flex flex-col items-end gap-1 w-[110px]">
                          <span className="text-[9px] font-brand font-bold text-white/30 uppercase tracking-wider">
                            RTT history
                          </span>
                          <RttSparkline samples={relay.rttTrend} />
                          <span className="text-[9px] font-mono text-white/40">
                            {relay.historySamples > 1 && relay.rttAvgMs > 0
                              ? `avg ${relay.rttAvgMs} ms \u00b1${relay.rttStdMs}`
                              : relay.historySamples === 1
                                ? "1 sample"
                                : "\u2014"}
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
          {/* TAB 5: SYSTEM TWEAKER                                     */}
          {/* ========================================================= */}
          {activeTab === "tweaker" && (
            <div className="flex flex-col gap-6 max-w-[960px] mx-auto w-full tab-enter">
              {/* Bufferbloat Diagnostic Test Card */}
              <div className="p-6 rounded-3xl bg-gradient-to-r from-[#0d1624] via-[#101a29] to-[#0a121e] border border-[#1d2b3d] shadow-xl flex flex-col gap-4">
                <div className="flex items-center justify-between">
                  <div className="flex flex-col gap-1">
                    <div className="flex items-center gap-2 text-xs font-brand font-bold text-[#00F0FF] uppercase tracking-wider">
                      <IconActivity className="w-4 h-4" />
                      <span>BUFFERBLOAT & LATENCY UNDER LOAD DIAGNOSTIC</span>
                    </div>
                    <h3 className="font-brand font-extrabold text-xl text-white">
                      Router Queue Saturation & Bufferbloat Analysis
                    </h3>
                    <p className="text-xs text-white/60 max-w-[620px]">
                      Bufferbloat causes sudden ping spikes when background tasks download or upload data. Lagvex compares idle RTT with RTT during a TCP download burst; the result is reported only when both measurements succeed.
                    </p>
                  </div>

                  <button
                    onClick={handleRunBufferbloatTest}
                    disabled={isTestingBufferbloat}
                    className={`h-[46px] px-6 rounded-2xl font-brand font-bold text-xs uppercase tracking-wider cursor-pointer border-none transition-all shadow-lg shrink-0 flex items-center gap-2 ${
                      isTestingBufferbloat
                        ? "bg-[#00F0FF]/30 text-white/50 cursor-wait"
                        : "bg-[#00F0FF] hover:bg-[#33f3ff] text-black shadow-[#00F0FF]/20"
                    }`}
                  >
                    {isTestingBufferbloat ? (
                      <>
                        <span className="w-3.5 h-3.5 border-2 border-black border-t-transparent rounded-full animate-spin" />
                        <span>Testing Queue...</span>
                      </>
                    ) : (
                      <>
                        <IconBolt className="w-4 h-4 text-black" />
                        <span>Run Bufferbloat Test</span>
                      </>
                    )}
                  </button>
                </div>

                {bufferbloatResult && (
                  <div className="p-5 rounded-2xl bg-[#080d14] border border-[#1a2535] grid grid-cols-4 gap-4 items-center mt-1">
                    <div className="flex flex-col items-center justify-center p-3 rounded-xl bg-white/[0.03] border border-white/5">
                      <span className="text-[10px] font-brand text-white/40 uppercase">BUFFERBLOAT GRADE</span>
                      <span className={`text-3xl font-brand font-black mt-1 ${
                        bufferbloatResult.grade.startsWith("A") ? "text-[#10b981]" :
                        bufferbloatResult.grade === "B" ? "text-[#00F0FF]" :
                        bufferbloatResult.grade === "C" ? "text-[#f59e0b]" : "text-red-400"
                      }`}>
                        {bufferbloatResult.grade}
                      </span>
                      <span className="text-[10px] text-white/60 font-medium">{bufferbloatResult.rating}</span>
                    </div>

                    <div className="flex flex-col">
                      <span className="text-[10px] font-brand text-white/40 uppercase">IDLE LATENCY</span>
                      <span className="font-mono font-bold text-lg text-white mt-0.5">{bufferbloatResult.baselineMs} ms</span>
                      <span className="text-[10px] text-white/40 font-mono">Unloaded Socket RTT</span>
                    </div>

                    <div className="flex flex-col">
                      <span className="text-[10px] font-brand text-white/40 uppercase">LOADED LATENCY</span>
                      <span className="font-mono font-bold text-lg text-[#f59e0b] mt-0.5">{bufferbloatResult.loadedMs} ms</span>
                      <span className="text-[10px] text-white/40 font-mono">+{bufferbloatResult.diffMs} ms Delta</span>
                    </div>

                    <div className="flex flex-col">
                      <span className="text-[10px] font-brand text-white/40 uppercase">RECOMMENDATION</span>
                      <span className="text-xs text-white/80 font-medium leading-relaxed mt-0.5">{bufferbloatResult.advice}</span>
                    </div>
                  </div>
                )}
              </div>

              <div className="p-6 rounded-3xl bg-gradient-to-r from-[#0d1624] to-[#0a121e] border border-[#1d2b3d] shadow-xl flex items-center justify-between">
                <div>
                  <h2 className="font-brand font-extrabold text-2xl text-white">Windows Network Tweaks</h2>
                  <p className="text-xs text-white/60 mt-1 max-w-[600px]">
                    Runs supported Windows network commands: DNS cache flush, TCP auto-tuning, and optional MMCSS registry values. The result list shows exactly which commands succeeded. TCP_NODELAY and MTU preferences apply only to Lagvex sockets and its virtual adapter.
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
                    desc: "Saved as a Lagvex preference for TCP sockets owned by the client. It does not change game or system-wide TCP settings.",
                    state: disableNagle,
                    setter: setDisableNagle
                  },
                  {
                    title: "Multimedia Class Scheduler (MMCSS) Network Tuning",
                    desc: "Sets the MMCSS NetworkThrottlingIndex and SystemResponsiveness registry values on Windows (requires administrator). It does not reserve CPU cores for games.",
                    state: mmcssPriority,
                    setter: setMmcssPriority
                  },
                  {
                    title: "Kernel Socket TCP_NODELAY Flag",
                    desc: "Saved as a preference; TCP_NODELAY is not a global Windows switch and is applied only to Lagvex-owned sockets.",
                    state: tcpNoDelay,
                    setter: setTcpNoDelay
                  },
                  {
                    title: "MTU Clamping Preference",
                    desc: "Saved for the virtual adapter. On connect the adapter uses the MTU negotiated with the relay instead of a fixed value.",
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
                  <span>DNS resolver cache flushed by the local service.</span>
                </div>
              )}

              {/* Diagnostics Snapshot Card */}
              <div className="p-6 rounded-3xl bg-[#0b1018] border border-[#1c2738] flex flex-col gap-4 shadow-lg">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex flex-col gap-1">
                    <div className="flex items-center gap-2 text-xs font-brand font-bold text-[#00F0FF] uppercase tracking-wider">
                      <IconShieldCheck className="w-4 h-4" />
                      <span>DIAGNOSTICS SNAPSHOT</span>
                    </div>
                    <h3 className="font-brand font-extrabold text-lg text-white">
                      Export Measured State, Import Settings
                    </h3>
                    <p className="text-xs text-white/60 max-w-[620px]">
                      The snapshot contains only measured values (tunnel state, relay probe history) plus your non-secret
                      settings. Relay pre-shared keys are never written to a snapshot and relays are therefore skipped on import.
                    </p>
                  </div>

                  <div className="flex items-center gap-2 shrink-0">
                    <button
                      onClick={handleExportDiagnostics}
                      className="h-[42px] px-5 rounded-2xl bg-[#00F0FF]/15 hover:bg-[#00F0FF]/25 border border-[#00F0FF]/30 text-xs font-brand font-bold text-[#00F0FF] cursor-pointer transition-all"
                    >
                      Export Snapshot
                    </button>
                    <button
                      onClick={() => snapshotFileRef.current?.click()}
                      className="h-[42px] px-5 rounded-2xl bg-[#101722] hover:bg-[#162130] border border-[#223042] text-xs font-brand font-bold text-white cursor-pointer transition-all"
                    >
                      Import Snapshot
                    </button>
                    <input
                      ref={snapshotFileRef}
                      type="file"
                      accept="application/json,.json"
                      className="hidden"
                      onChange={(e) => {
                        const file = e.target.files?.[0];
                        e.target.value = "";
                        if (file) void handleImportDiagnostics(file);
                      }}
                    />
                  </div>
                </div>

                {snapshotResult && (
                  <div className="p-4 rounded-2xl bg-[#080d14] border border-[#1a2535] flex flex-col gap-2">
                    <span className="text-[10px] font-brand font-bold text-white/40 uppercase tracking-wider">
                      Last import
                    </span>
                    <span className="text-xs text-[#10b981]">
                      Applied: {snapshotResult.applied.length > 0 ? snapshotResult.applied.join(", ") : "nothing"}
                    </span>
                    {snapshotResult.warnings.map((warning, idx) => (
                      <span key={idx} className="text-xs text-[#f59e0b]">
                        {warning}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </main>

      <FloatingHUD />
    </div>
  );
}
