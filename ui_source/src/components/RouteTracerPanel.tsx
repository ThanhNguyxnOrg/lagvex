import { useState } from "react";
import { apiFetch } from "../lib/api";

type TraceHop = {
  hop: number;
  ip?: string;
  rttsMs?: number[];
  avgMs?: number;
  minMs?: number;
  maxMs?: number;
  lossPct: number;
  timeout: boolean;
};

type TracerouteResult = {
  host: string;
  maxHops: number;
  tool: string;
  hops: TraceHop[];
  measuredAt: string;
  durationMs: number;
  error?: string;
};

type DNSBenchmarkEntry = {
  id: string;
  name: string;
  address: string;
  category: string;
  medianMs?: number;
  minMs?: number;
  maxMs?: number;
  lossPct: number;
  samples: number;
  reachable: boolean;
  isFastest: boolean;
};

type MTUDiscoveryResult = {
  host: string;
  pathMtu: number;
  recommendedMtu: number;
  wireguardMtu: number;
  probes: number;
  measured: boolean;
  durationMs: number;
  note?: string;
};

type Props = {
  notify: (msg: string) => void;
  obfuscated?: boolean;
  onToggleObfuscation?: (enabled: boolean) => Promise<void>;
};

export default function RouteTracerPanel({ notify, obfuscated = false, onToggleObfuscation }: Props) {
  const [traceHost, setTraceHost] = useState("1.1.1.1");
  const [isTracing, setIsTracing] = useState(false);
  const [traceResult, setTraceResult] = useState<TracerouteResult | null>(null);

  const [isBenchingDNS, setIsBenchingDNS] = useState(false);
  const [dnsResults, setDnsResults] = useState<DNSBenchmarkEntry[] | null>(null);

  const [isProbingMTU, setIsProbingMTU] = useState(false);
  const [mtuResult, setMtuResult] = useState<MTUDiscoveryResult | null>(null);

  const handleRunTraceroute = async (targetOverride?: string) => {
    const host = (targetOverride ?? traceHost).trim();
    if (!host) {
      notify("Enter a public IPv4 literal to trace.");
      return;
    }
    if (targetOverride) setTraceHost(targetOverride);
    setIsTracing(true);
    notify(`Tracing hops to ${host} (PingPlotter-style hop inspection)...`);
    try {
      const res = await apiFetch("/api/traceroute", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ host, maxHops: 18 }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        notify(`Traceroute failed: ${data.error || `HTTP ${res.status}`}`);
        return;
      }
      setTraceResult(data as TracerouteResult);
      if (data.error) {
        notify(data.error);
      } else {
        notify(`Traceroute to ${host} completed (${(data.hops || []).length} hops in ${data.durationMs} ms).`);
      }
    } catch {
      notify("Failed to run traceroute.");
    } finally {
      setIsTracing(false);
    }
  };

  const handleRunDNSBenchmark = async () => {
    setIsBenchingDNS(true);
    notify("Benchmarking public DNS resolvers via real UDP queries...");
    try {
      const res = await apiFetch("/api/dns-benchmark", { method: "POST" });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        notify(`DNS benchmark failed: ${data.error || `HTTP ${res.status}`}`);
        return;
      }
      const list = (data.resolvers || []) as DNSBenchmarkEntry[];
      setDnsResults(list);
      const fastest = list.find((r) => r.isFastest);
      if (fastest && typeof fastest.medianMs === "number") {
        notify(`Fastest DNS: ${fastest.name} (${fastest.address}) at ${fastest.medianMs.toFixed(1)} ms.`);
      } else {
        notify("DNS benchmark finished.");
      }
    } catch {
      notify("Failed to run DNS benchmark.");
    } finally {
      setIsBenchingDNS(false);
    }
  };

  const handleRunMTUProbe = async () => {
    setIsProbingMTU(true);
    notify("Probing exact non-fragmenting Path MTU (DF-bit binary search)...");
    try {
      const res = await apiFetch("/api/diagnose-mtu", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ host: "1.1.1.1" }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        notify(`MTU probe failed: ${data.error || `HTTP ${res.status}`}`);
        return;
      }
      setMtuResult(data as MTUDiscoveryResult);
      if (data.measured) {
        notify(`Path MTU ${data.pathMtu}B detected — Lagvex MTU ${data.recommendedMtu}B / WireGuard ${data.wireguardMtu}B.`);
      } else {
        notify(data.note || "Path MTU could not be measured on this network.");
      }
    } catch {
      notify("Failed to run Path-MTU discovery.");
    } finally {
      setIsProbingMTU(false);
    }
  };

  const maxHopRTT = Math.max(
    50,
    ...(traceResult?.hops || []).map((h) => h.maxMs || h.avgMs || 0),
  );

  return (
    <div className="flex flex-col gap-6">
      {/* PingPlotter-style Hop-by-Hop Route Tracer */}
      <div className="p-6 rounded-3xl bg-gradient-to-r from-[#0d1624] via-[#101a29] to-[#0a121e] border border-[#1d2b3d] shadow-xl flex flex-col gap-4">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div className="flex flex-col gap-1">
            <span className="text-xs font-brand font-bold text-[#00F0FF] uppercase tracking-wider">
              HOP-BY-HOP ROUTE TRACER (FREE DIAGNOSTIC)
            </span>
            <h3 className="font-brand font-extrabold text-xl text-white">
              Pinpoint Exactly Which Hop Causes Lag Spikes or Packet Loss
            </h3>
            <p className="text-xs text-white/55 max-w-[620px]">
              Traces every router hop from your PC to a public IPv4 target (game server or backbone IP) to reveal whether lag comes from your home Wi-Fi, ISP gateway, or international transit.
            </p>
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <input
            type="text"
            value={traceHost}
            onChange={(e) => setTraceHost(e.target.value)}
            placeholder="Public IPv4 (e.g. 1.1.1.1 or game server IP)"
            className="flex-1 min-w-[220px] h-10 px-3.5 rounded-xl bg-[#080d14] border border-[#1f2e42] text-xs font-mono text-white outline-none focus:border-[#00F0FF]"
          />
          {["1.1.1.1", "8.8.8.8", "162.159.192.1"].map((preset) => (
            <button
              key={preset}
              type="button"
              disabled={isTracing}
              onClick={() => void handleRunTraceroute(preset)}
              className="px-3 h-10 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] border border-white/10 text-[11px] font-mono text-white/75 cursor-pointer"
            >
              {preset}
            </button>
          ))}
          <button
            type="button"
            disabled={isTracing}
            onClick={() => void handleRunTraceroute()}
            className="px-5 h-10 rounded-xl bg-[#00F0FF] hover:bg-[#47f4ff] text-black font-brand font-extrabold text-xs uppercase tracking-wider cursor-pointer border-none disabled:opacity-50"
          >
            {isTracing ? "Tracing..." : "Run Trace"}
          </button>
        </div>

        {traceResult && (
          <div className="rounded-2xl bg-[#080d14] border border-[#1a2535] overflow-hidden">
            <div className="px-4 py-2.5 border-b border-white/5 flex items-center justify-between text-[11px] font-mono text-white/50">
              <span>
                Target: <strong className="text-white">{traceResult.host}</strong> ({traceResult.tool})
              </span>
              <span>
                {traceResult.hops.length} hops • {traceResult.durationMs} ms
              </span>
            </div>
            {traceResult.error && traceResult.hops.length === 0 ? (
              <div className="p-4 text-xs font-mono text-amber-300">{traceResult.error}</div>
            ) : (
              <div className="divide-y divide-white/5 overflow-x-auto">
                {traceResult.hops.map((hop) => {
                  const pct = hop.avgMs ? Math.min(100, (hop.avgMs / maxHopRTT) * 100) : 0;
                  const barColor =
                    hop.timeout || hop.lossPct >= 50
                      ? "bg-red-500"
                      : (hop.avgMs || 0) > 100
                        ? "bg-amber-400"
                        : "bg-[#00F0FF]";
                  return (
                    <div
                      key={hop.hop}
                      className="px-4 py-2 min-w-[520px] grid grid-cols-[44px_140px_1fr_92px_72px] items-center gap-3 text-xs font-mono"
                    >
                      <span className="text-white/40">#{hop.hop}</span>
                      <span className={hop.timeout ? "text-white/30" : "text-white/90"}>
                        {hop.ip || "* * *"}
                      </span>
                      <div className="h-2 rounded-full bg-white/[0.06] overflow-hidden">
                        <div className={`h-full rounded-full ${barColor}`} style={{ width: `${pct}%` }} />
                      </div>
                      <span className="text-right tabular-nums text-white/85">
                        {hop.timeout || !hop.avgMs ? "—" : `${hop.avgMs.toFixed(1)} ms`}
                      </span>
                      <span
                        className={`text-right tabular-nums ${
                          hop.lossPct > 0 ? "text-amber-300" : "text-[#10b981]"
                        }`}
                      >
                        {hop.lossPct.toFixed(0)}% loss
                      </span>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Free Public DNS Latency Benchmark */}
      <div className="p-6 rounded-3xl bg-gradient-to-r from-[#0d1624] via-[#101a29] to-[#0a121e] border border-[#1d2b3d] shadow-xl flex flex-col gap-4">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div className="flex flex-col gap-1">
            <span className="text-xs font-brand font-bold text-[#A366FF] uppercase tracking-wider">
              DNS RESOLVER BENCHMARK (FREE LOCAL OPTIMIZER)
            </span>
            <h3 className="font-brand font-extrabold text-xl text-white">
              Find the Lowest-Latency DNS Resolver for Your ISP
            </h3>
            <p className="text-xs text-white/55 max-w-[600px]">
              Measures real UDP lookup RTT across Cloudflare, Google, Quad9, OpenDNS and AdGuard to speed up matchmaking and server discovery.
            </p>
          </div>
          <button
            type="button"
            disabled={isBenchingDNS}
            onClick={() => void handleRunDNSBenchmark()}
            className="h-[42px] px-5 rounded-2xl bg-[#7E42FF] hover:bg-[#9360ff] text-white font-brand font-bold text-xs uppercase tracking-wider cursor-pointer border-none transition-all disabled:opacity-50"
          >
            {isBenchingDNS ? "Benchmarking..." : "Benchmark DNS"}
          </button>
        </div>

        {dnsResults && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            {dnsResults.map((entry) => (
              <div
                key={entry.id}
                className={`p-4 rounded-2xl border flex flex-col gap-1.5 ${
                  entry.isFastest
                    ? "bg-[#10b981]/10 border-[#10b981]/40"
                    : "bg-[#080d14] border-[#1a2535]"
                }`}
              >
                <div className="flex items-center justify-between">
                  <span className="font-brand font-bold text-sm text-white">{entry.name}</span>
                  {entry.isFastest && (
                    <span className="px-2 py-0.5 rounded-full bg-[#10b981]/20 text-[#34d399] text-[9px] font-brand font-extrabold uppercase">
                      FASTEST
                    </span>
                  )}
                </div>
                <div className="flex items-center justify-between text-xs font-mono">
                  <span className="text-white/50">{entry.address}</span>
                  <span className={entry.reachable ? "text-[#00F0FF] font-bold tabular-nums" : "text-red-400"}>
                    {entry.reachable && typeof entry.medianMs === "number"
                      ? `${entry.medianMs.toFixed(1)} ms`
                      : "Unreachable"}
                  </span>
                </div>
                <div className="text-[10px] font-mono text-white/40 flex justify-between">
                  <span>{entry.category}</span>
                  <span>Loss {entry.lossPct.toFixed(0)}%</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Path-MTU Discovery & Anti-DPI UDP Obfuscation Controls */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="p-5 rounded-3xl bg-gradient-to-r from-[#0d1624] to-[#0a121e] border border-[#1d2b3d] flex flex-col justify-between gap-3">
          <div className="flex flex-col gap-1">
            <span className="text-[10px] font-brand font-bold text-[#00F0FF] uppercase tracking-wider">
              AUTO PATH-MTU DISCOVERY (PMTUD)
            </span>
            <h4 className="font-brand font-extrabold text-base text-white">
              Detect PPPoE / Cellular MTU Bottlenecks
            </h4>
            <p className="text-xs text-white/55">
              Binary-searches non-fragmenting DF-bit packets to calculate the exact safe tunnel MTU for Lagvex and WireGuard.
            </p>
          </div>
          {mtuResult && (
            <div className="p-3 rounded-xl bg-[#080d14] border border-white/10 text-xs font-mono flex flex-col gap-1">
              {mtuResult.measured ? (
                <div className="flex items-center justify-between">
                  <span className="text-white/70">Path MTU: <strong className="text-[#00F0FF]">{mtuResult.pathMtu} B</strong></span>
                  <span className="text-[#10b981]">Lagvex {mtuResult.recommendedMtu} · WG {mtuResult.wireguardMtu}</span>
                </div>
              ) : (
                <span className="text-amber-300">Not measured ({mtuResult.note})</span>
              )}
              {mtuResult.measured && mtuResult.note && (
                <span className="text-[10px] text-white/45">{mtuResult.note}</span>
              )}
            </div>
          )}
          <div>
            <button
              type="button"
              disabled={isProbingMTU}
              onClick={() => void handleRunMTUProbe()}
              className="px-4 h-9 rounded-xl bg-[#00F0FF]/15 hover:bg-[#00F0FF]/25 border border-[#00F0FF]/35 text-[#00F0FF] font-brand font-bold text-xs cursor-pointer disabled:opacity-50"
            >
              {isProbingMTU ? "Probing MTU..." : "Discover Optimal MTU"}
            </button>
          </div>
        </div>

        <div className="p-5 rounded-3xl bg-gradient-to-r from-[#0d1624] to-[#0a121e] border border-[#1d2b3d] flex flex-col justify-between gap-3">
          <div className="flex flex-col gap-1">
            <span className="text-[10px] font-brand font-bold text-[#A366FF] uppercase tracking-wider">
              ANTI-DPI UDP OBFUSCATION (ZERO OVERHEAD)
            </span>
            <h4 className="font-brand font-extrabold text-base text-white">
              Bypass Peak-Hour ISP UDP Throttling
            </h4>
            <p className="text-xs text-white/55">
              Scrambles the 17-byte header using the random Poly1305 tag + variable handshake padding so ISP DPI cannot fingerprint or throttle Lagvex UDP flows.
            </p>
          </div>
          <div className="flex items-center justify-between pt-1">
            <span className="text-xs font-mono text-white/60">
              Status:{" "}
              <strong className={obfuscated ? "text-[#10b981]" : "text-white/45"}>
                {obfuscated ? "ENABLED (Next Connect)" : "STANDARD"}
              </strong>
            </span>
            {onToggleObfuscation && (
              <button
                type="button"
                onClick={() => void onToggleObfuscation(!obfuscated)}
                className={`px-4 h-9 rounded-xl font-brand font-bold text-xs cursor-pointer border transition-all ${
                  obfuscated
                    ? "bg-[#10b981]/20 border-[#10b981]/50 text-[#34d399]"
                    : "bg-white/[0.06] hover:bg-white/[0.12] border-white/15 text-white"
                }`}
              >
                {obfuscated ? "Disable Obfuscation" : "Enable Anti-DPI"}
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
