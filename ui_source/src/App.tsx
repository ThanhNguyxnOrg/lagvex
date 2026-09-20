import { useState } from "react";

const assetPathPrefix = "/assets";

// Icons
const imgSearch = `${assetPathPrefix}/1908f.svg`;
const imgBell = `${assetPathPrefix}/934ba.svg`;
const imgPaperPlane = `${assetPathPrefix}/ec393.svg`;
const imgHome = `${assetPathPrefix}/74870.svg`;
const imgFriends = `${assetPathPrefix}/1f98a.svg`;
const imgChatAdd = `${assetPathPrefix}/71fbf.svg`;
const imgShoppingBag = `${assetPathPrefix}/617de.svg`;
const imgStar = `${assetPathPrefix}/7c528.svg`;
const imgDownload = `${assetPathPrefix}/0a20b.svg`;
const imgGroup3 = `${assetPathPrefix}/4b585.svg`;
const imgChevronRight = `${assetPathPrefix}/27925.svg`;
const imgSidePanel = `${assetPathPrefix}/91305.svg`;
const imgNavActive = `${assetPathPrefix}/3dfd6.svg`;

// Avatar / profile pics
const imgAvatar = `${assetPathPrefix}/cf167.png`;
const imgAvatar2 = `${assetPathPrefix}/9f52e.png`;
const imgAvatar3 = `${assetPathPrefix}/87270.png`;

// Game images
const imgAssassin = `${assetPathPrefix}/bcda7.png`;   // Assassin's creed character
const imgGhost = `${assetPathPrefix}/d2cd1.png`;      // Ghost of Tsushima
const imgMinecraft = `${assetPathPrefix}/c3f83.png`;  // Minecraft
const imgAssassin2 = `${assetPathPrefix}/404fa.png`;  // Assassin creed 2
const imgFallGuys = `${assetPathPrefix}/631a8.png`;   // Fall guys
const imgCallOfDuty = `${assetPathPrefix}/b8a97.png`; // Call of duty
const imgFortnite = `${assetPathPrefix}/50de5.png`;   // Fortnite
const imgHeadset = `${assetPathPrefix}/600fc.png`;    // Gaming headset

// Green dot indicator
const GreenDot = () => (
  <span className="inline-block w-2 h-2 rounded-full bg-[#3dbda7] shrink-0" />
);

// Download / optimize icon button
const DownloadIcon = ({ className }: { className?: string }) => (
  <div className={className ?? "relative w-[38px] h-[38px]"}>
    <div className="absolute inset-[-1.32%_-1.32%_0_0]">
      <img alt="" className="block max-w-none w-full h-full" src={imgDownload} />
    </div>
  </div>
);

// Ping badge
const PingBadge = ({ ms, color }: { ms: number; color: "green" | "yellow" | "red" }) => {
  const bg =
    color === "green"
      ? "bg-[#3dbda7]/20 text-[#3dbda7]"
      : color === "yellow"
      ? "bg-yellow-500/20 text-yellow-400"
      : "bg-red-500/20 text-red-400";
  return (
    <span className={`text-[11px] font-['Poppins:Bold'] px-2 py-0.5 rounded-full ${bg}`}>
      {ms}ms
    </span>
  );
};

type Game = {
  name: string;
  platforms: string;
  ping: number;
  img: string;
  pingColor: "green" | "yellow" | "red";
};

const games: Game[] = [
  { name: "Valorant", platforms: "PC", ping: 18, img: imgCallOfDuty, pingColor: "green" },
  { name: "CS2", platforms: "PC", ping: 24, img: imgAssassin2, pingColor: "green" },
  { name: "Minecraft", platforms: "Ps5 & Xbox & PC", ping: 62, img: imgMinecraft, pingColor: "yellow" },
  { name: "Fall Guys", platforms: "Ps5 & Xbox & PC", ping: 110, img: imgFallGuys, pingColor: "red" },
];

const friends = [
  { name: "Goodboi79", status: "Playing Valorant", img: imgAvatar3 },
  { name: "Doglover96", status: "Playing CS2", img: imgAvatar2 },
  { name: "iamgoat__", status: "Playing Minecraft", img: imgAvatar },
];

const servers = [
  { region: "Singapore", flag: "🇸🇬", ping: 12, load: 42 },
  { region: "Tokyo", flag: "🇯🇵", ping: 28, load: 71 },
  { region: "Sydney", flag: "🇦🇺", ping: 55, load: 33 },
];

type NavItem = { icon: string; id: string };
const navItems: NavItem[] = [
  { icon: imgHome, id: "home" },
  { icon: imgFriends, id: "friends" },
  { icon: imgChatAdd, id: "chat" },
  { icon: imgShoppingBag, id: "shop" },
  { icon: imgStar, id: "star" },
];

export default function App() {
  const [activeNav, setActiveNav] = useState("chat");
  const [optimizing, setOptimizing] = useState(false);
  const [selectedGame, setSelectedGame] = useState(0);
  const [currentPing, setCurrentPing] = useState(18);

  const handleOptimize = () => {
    setOptimizing(true);
    const interval = setInterval(() => {
      setCurrentPing((p) => {
        const next = Math.max(8, p - Math.floor(Math.random() * 3 + 1));
        if (next <= 8) clearInterval(interval);
        return next;
      });
    }, 200);
    setTimeout(() => {
      clearInterval(interval);
      setOptimizing(false);
    }, 3000);
  };

  return (
    <div
      className="w-full min-h-screen bg-[#1d1e22] relative overflow-hidden"
      style={{ fontFamily: "'Poppins:Regular', sans-serif" }}
    >
      {/* ── Left Sidebar ── */}
      <aside className="fixed top-0 left-0 h-full w-[125px] bg-[#1d1e22] shadow-[0px_4px_40px_0px_rgba(0,0,0,0.3)] z-30 flex flex-col items-center pt-[41px]">
        {/* Logo */}
        <div className="w-[45px] h-[45px] mb-10">
          <svg viewBox="0 0 45 45" fill="none" xmlns="http://www.w3.org/2000/svg">
            <circle cx="22.5" cy="22.5" r="22.5" fill="#3dbda7" opacity="0.15" />
            <path
              d="M22.5 10C15.6 10 10 15.6 10 22.5S15.6 35 22.5 35 35 29.4 35 22.5 29.4 10 22.5 10zm-1 18.5l-6-6 1.5-1.5 4.5 4.5 8.5-8.5 1.5 1.5-10 10z"
              fill="#3dbda7"
            />
          </svg>
        </div>

        {/* Nav buttons */}
        <nav className="flex flex-col gap-3 items-center">
          {navItems.map(({ icon, id }) => {
            const isActive = activeNav === id;
            return (
              <button
                key={id}
                onClick={() => setActiveNav(id)}
                className="relative w-[70px] h-[55px] flex items-center justify-center cursor-pointer"
                style={{ background: "none", border: "none" }}
              >
                {isActive && (
                  <img
                    alt=""
                    className="absolute inset-0 w-full h-full"
                    src={imgNavActive}
                  />
                )}
                <img
                  alt=""
                  className="relative w-6 h-6 z-10"
                  style={{ filter: isActive ? "brightness(10)" : "brightness(3) opacity(0.5)" }}
                  src={icon}
                />
              </button>
            );
          })}
        </nav>
      </aside>

      {/* ── Main Content ── */}
      <main className="ml-[125px] pr-[360px] min-h-screen flex flex-col">
        {/* ── Top Header ── */}
        <header className="flex items-center gap-4 pt-[39px] pb-6 px-6">
          {/* Search */}
          <div className="flex-1 relative h-[50px]">
            <div className="absolute inset-0 rounded-[30px] bg-[rgba(100,110,109,0.2)]" />
            <img
              alt=""
              className="absolute left-[6%] top-1/2 -translate-y-1/2 w-6 h-6"
              src={imgSearch}
            />
            <span className="absolute left-[13%] top-1/2 -translate-y-1/2 font-['Poppins:Medium'] text-[12px] text-white opacity-70">
              Search games or servers...
            </span>
          </div>
          {/* Bell */}
          <button className="w-[55px] h-[50px] rounded-[30px] bg-gradient-to-b from-[#393939] to-[#1d1e22] flex items-center justify-center cursor-pointer border-none">
            <img alt="" className="w-6 h-6" src={imgBell} />
          </button>
          {/* Send */}
          <button className="w-[55px] h-[50px] rounded-[30px] bg-[#3dbda7] flex items-center justify-center cursor-pointer border-none">
            <img alt="" className="w-6 h-6" src={imgPaperPlane} />
          </button>
          {/* Avatar */}
          <div className="w-[55px] h-[50px] rounded-[30px] overflow-hidden bg-gradient-to-b from-[#393939] to-[#1d1e22]">
            <img alt="" className="w-full h-full object-cover" src={imgAvatar} />
          </div>
        </header>

        <div className="px-6 flex flex-col gap-6 pb-10">
          {/* ── Hero Ping Card ── */}
          <div className="relative h-[320px] rounded-[50px] overflow-hidden bg-[#071b24] border border-black/20">
            {/* Background glow ellipse */}
            <div
              className="absolute inset-0 rounded-[50px]"
              style={{
                background:
                  "radial-gradient(ellipse 120% 160% at 50% 110%, #067d71 0%, transparent 60%)",
              }}
            />
            {/* Hero game art */}
            <img
              alt=""
              className="absolute right-0 bottom-0 h-[110%] object-contain pointer-events-none opacity-80"
              style={{ transform: "scaleX(-1)" }}
              src={games[selectedGame]?.img ?? imgAssassin}
            />
            {/* Year badge */}
            <div
              className="absolute top-9 left-11 h-[23px] w-[34px] rounded-[6px] flex items-center justify-center"
              style={{
                background: "linear-gradient(123deg, #faa525 3%, #c13509 98%)",
              }}
            >
              <span className="font-['Poppins:Bold'] text-[9px] text-white">LIVE</span>
            </div>
            {/* Title */}
            <div className="absolute top-[20%] left-[5.3%]">
              <p className="font-['Poppins:SemiBold'] text-[40px] text-white leading-[45px]">
                {games[selectedGame]?.name ?? "Valorant"}
              </p>
            </div>
            {/* Ping stats row */}
            <div className="absolute bottom-[60px] left-11 flex gap-6 items-center">
              <div className="flex flex-col">
                <span className="font-['Poppins:Bold'] text-[32px] text-[#3dbda7] leading-none">
                  {currentPing}ms
                </span>
                <span className="font-['Poppins:Medium'] text-[11px] text-white/50 mt-0.5">
                  Ping
                </span>
              </div>
              <div className="flex flex-col">
                <span className="font-['Poppins:Bold'] text-[32px] text-white leading-none">2%</span>
                <span className="font-['Poppins:Medium'] text-[11px] text-white/50 mt-0.5">
                  Packet Loss
                </span>
              </div>
              <div className="flex flex-col">
                <span className="font-['Poppins:Bold'] text-[32px] text-white leading-none">
                  4ms
                </span>
                <span className="font-['Poppins:Medium'] text-[11px] text-white/50 mt-0.5">
                  Jitter
                </span>
              </div>
            </div>
            {/* Bottom row: server + button */}
            <div className="absolute bottom-[14px] left-11 flex items-center gap-4">
              {/* Ghost price row repurposed as server label */}
              <div
                className="h-[56px] w-[239px] rounded-[73px] flex items-center px-5"
                style={{
                  background:
                    "linear-gradient(94deg, rgb(7,27,36) 2.6%, rgba(13,32,40,0.99) 23%, rgba(87,87,87,0.84) 95.6%)",
                }}
              >
                <span className="font-['Poppins:Regular'] text-[13px] text-white/80">
                  🇸🇬 Singapore · 12ms
                </span>
              </div>
              {/* Optimize button */}
              <button
                onClick={handleOptimize}
                className="h-[42px] w-[150px] rounded-[73px] bg-[#3dbda7] flex items-center justify-center gap-2 cursor-pointer border-none transition-opacity"
                style={{ opacity: optimizing ? 0.7 : 1 }}
              >
                <img alt="" className="w-4 h-4" src={imgGroup3} />
                <span className="font-['Poppins:Bold'] text-[12px] text-white">
                  {optimizing ? "Optimizing..." : "Optimize"}
                </span>
              </button>
            </div>
            {/* Slider dots */}
            <div className="absolute bottom-4 right-6 flex gap-2">
              {games.map((_, i) => (
                <button
                  key={i}
                  onClick={() => setSelectedGame(i)}
                  className="border-none cursor-pointer rounded-[3px] transition-all"
                  style={{
                    width: i === selectedGame ? 27 : 11,
                    height: 4,
                    background: i === selectedGame ? "#3dbda7" : "rgba(255,255,255,0.1)",
                  }}
                />
              ))}
            </div>
          </div>

          {/* ── Bottom row: Server picker + Recently Played ── */}
          <div className="flex gap-6">
            {/* Server Picker card */}
            <div
              className="relative w-[357px] h-[441px] rounded-[40px] shrink-0 overflow-hidden bg-[#151515] flex flex-col"
            >
              {/* heading */}
              <div className="flex items-center justify-between px-7 pt-8 pb-4">
                <span className="font-['Poppins:SemiBold'] text-[20px] text-white">
                  Best Servers
                </span>
                <button className="flex items-center justify-center w-6 h-6 cursor-pointer border-none bg-transparent">
                  <img
                    alt=""
                    className="w-6 h-6"
                    src={imgChevronRight}
                    style={{ transform: "rotate(180deg)" }}
                  />
                </button>
              </div>
              {/* Server list */}
              <div className="flex flex-col gap-2 px-5 overflow-auto">
                {servers.map((s) => (
                  <div
                    key={s.region}
                    className="flex items-center gap-3 bg-[#1d1e22] rounded-2xl px-4 py-3 cursor-pointer hover:bg-[#252629] transition-colors"
                    onClick={() => {}}
                  >
                    <span className="text-2xl">{s.flag}</span>
                    <div className="flex-1">
                      <p className="font-['Poppins:Regular'] text-[14px] text-white leading-tight">
                        {s.region}
                      </p>
                      <p className="font-['Poppins:Medium'] text-[11px] text-[#969696] mt-0.5">
                        Load {s.load}%
                      </p>
                    </div>
                    <PingBadge
                      ms={s.ping}
                      color={s.ping < 30 ? "green" : s.ping < 70 ? "yellow" : "red"}
                    />
                  </div>
                ))}
              </div>
              {/* Headset art */}
              <div className="flex-1 relative mt-2">
                <img
                  alt=""
                  className="absolute bottom-0 left-1/2 -translate-x-1/2 w-[260px] object-contain pointer-events-none"
                  src={imgHeadset}
                />
              </div>
            </div>

            {/* Recently Played */}
            <div className="flex-1 flex flex-col">
              <p className="font-['Poppins:SemiBold'] text-[20px] text-white mb-4">
                Recently Played
              </p>
              <div className="flex flex-col gap-3 flex-1">
                {[
                  { name: "Call of Duty", img: imgCallOfDuty, ping: 22 },
                  { name: "Minecraft", img: imgMinecraft, ping: 55 },
                  { name: "Fortnite", img: imgFortnite, ping: 35 },
                ].map((g) => (
                  <div
                    key={g.name}
                    className="relative h-[115px] rounded-2xl overflow-hidden cursor-pointer hover:scale-[1.01] transition-transform"
                  >
                    <img
                      alt=""
                      className="absolute inset-0 w-full h-full object-cover opacity-70"
                      src={g.img}
                    />
                    <div className="absolute inset-0 bg-gradient-to-r from-black/60 to-transparent" />
                    <div className="absolute bottom-3 left-4 flex items-center gap-3">
                      <span className="font-['Poppins:SemiBold'] text-[14px] text-white">
                        {g.name}
                      </span>
                      <PingBadge
                        ms={g.ping}
                        color={g.ping < 30 ? "green" : g.ping < 70 ? "yellow" : "red"}
                      />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </main>

      {/* ── Right Sidebar ── */}
      <aside
        className="fixed top-0 right-0 h-full w-[334px] z-20 flex flex-col pt-[137px]"
        style={{ background: "none" }}
      >
        {/* Background panel */}
        <div
          className="absolute inset-0 rounded-l-[50px]"
          style={{
            background:
              "linear-gradient(251deg, rgb(31,31,31) 1%, rgb(44,44,44) 49%, rgb(37,37,37) 98%)",
          }}
        />
        <img
          alt=""
          className="absolute inset-0 w-full h-full pointer-events-none"
          src={imgSidePanel}
          style={{ opacity: 0.12 }}
        />

        <div className="relative flex flex-col h-full px-5 pb-8 overflow-y-auto">
          {/* Library / My Games */}
          <p className="font-['Poppins:Regular'] text-[20px] text-white mb-3">My Games</p>
          <div className="flex flex-col gap-1 mb-6">
            {games.map((g, i) => (
              <div
                key={g.name}
                className="flex items-center gap-3 py-2 px-2 rounded-xl cursor-pointer hover:bg-white/5 transition-colors"
                onClick={() => setSelectedGame(i)}
              >
                <div className="w-[55px] h-[50px] rounded-[12px] overflow-hidden bg-[#2a2a2a] shrink-0">
                  <img alt="" className="w-full h-full object-cover" src={g.img} />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="font-['Poppins:Regular'] text-[14px] text-white truncate">
                    {g.name}
                  </p>
                  <p className="font-['Poppins:Medium'] text-[11px] text-[#969696] truncate">
                    {g.platforms}
                  </p>
                </div>
                <PingBadge ms={g.ping} color={g.pingColor} />
                <DownloadIcon className="relative w-[28px] h-[28px] shrink-0" />
              </div>
            ))}
          </div>

          {/* Status */}
          <div
            className="rounded-[20px] px-4 py-3 mb-6"
            style={{
              background:
                "linear-gradient(102deg, rgba(29,30,34,0.65) 7%, rgba(53,53,53,0.58) 152%)",
            }}
          >
            <p className="font-['Poppins:Regular'] text-[18px] text-white mb-1">Status</p>
            <p className="font-['Poppins:Regular'] text-[11px] text-[rgba(196,196,196,0.28)]">
              VPN Tunnel active · SG-01
            </p>
          </div>

          {/* Online Friends */}
          <p className="font-['Poppins:Regular'] text-[18px] text-white mb-3">Online Friends</p>
          <div className="flex flex-col gap-2">
            {friends.map((f) => (
              <div key={f.name} className="flex items-center gap-3">
                <div className="relative w-[55px] h-[50px] rounded-[30px] overflow-hidden bg-gradient-to-b from-[#393939] to-[#1d1e22] shrink-0">
                  <img alt="" className="w-full h-full object-cover" src={f.img} />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="font-['Poppins:Medium'] text-[14px] text-white truncate">
                    {f.name}
                  </p>
                  <p className="font-['Poppins:Medium'] text-[11px] text-[#969696] truncate">
                    {f.status}
                  </p>
                </div>
                <GreenDot />
              </div>
            ))}
          </div>
        </div>
      </aside>
    </div>
  );
}
