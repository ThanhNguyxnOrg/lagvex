import { useState } from "react";
import { apiFetch } from "../lib/api";

export type RelaySummary = {
  id: string;
  cleanName: string;
  endpoint: string;
  tier?: string;
};

type Props = {
  relays: RelaySummary[];
  notify: (msg: string) => void;
  onRelaysChanged: () => Promise<void>;
};

type Mode = "closed" | "uri" | "wireguard" | "subscription";

export default function FreeBackbonePanel({ relays, notify, onRelaysChanged }: Props) {
  const [mode, setMode] = useState<Mode>("closed");
  const [busy, setBusy] = useState(false);

  // Import URI state
  const [shareURI, setShareURI] = useState("");
  // WireGuard .conf state
  const [wgName, setWgName] = useState("");
  const [wgConfig, setWgConfig] = useState("");
  // Subscription URL state
  const [subUrl, setSubUrl] = useState("");

  const handleRegisterWARP = async () => {
    setBusy(true);
    notify("Registering free Cloudflare WARP WireGuard split-tunnel profile...");
    try {
      const res = await apiFetch("/api/relay/register-warp", { method: "POST" });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        notify(`WARP setup failed: ${data.error || `HTTP ${res.status}`}`);
        return;
      }
      notify(`Cloudflare WARP (Free) ready at ${data.relay?.endpoint || "Anycast"} — select it to boost!`);
      await onRelaysChanged();
    } catch {
      notify("Could not reach the local Lagvex client to register WARP.");
    } finally {
      setBusy(false);
    }
  };

  const handleImportURI = async () => {
    if (!shareURI.trim()) {
      notify("Paste a lagvex:// share URI first.");
      return;
    }
    setBusy(true);
    try {
      const res = await apiFetch("/api/relay/import-uri", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ uri: shareURI.trim() }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        notify(`Import failed: ${data.error || `HTTP ${res.status}`}`);
        return;
      }
      setShareURI("");
      setMode("closed");
      notify(`Imported shared relay "${data.relay?.name || data.relay?.endpoint}".`);
      await onRelaysChanged();
    } catch {
      notify("Failed to import share URI.");
    } finally {
      setBusy(false);
    }
  };

  const handleImportWireGuard = async () => {
    if (!wgConfig.trim()) {
      notify("Paste a WireGuard .conf file first.");
      return;
    }
    setBusy(true);
    try {
      const res = await apiFetch("/api/relay/import-wireguard", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: wgName.trim(),
          config: wgConfig,
        }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        notify(`WireGuard import failed: ${data.error || `HTTP ${res.status}`}`);
        return;
      }
      setWgName("");
      setWgConfig("");
      setMode("closed");
      notify(`Imported WireGuard endpoint "${data.relay?.name || data.relay?.endpoint}".`);
      await onRelaysChanged();
    } catch {
      notify("Failed to import WireGuard configuration.");
    } finally {
      setBusy(false);
    }
  };

  const handleImportSubscription = async () => {
    if (!subUrl.trim()) {
      notify("Enter an https:// relay list URL first.");
      return;
    }
    setBusy(true);
    try {
      const res = await apiFetch("/api/relay/import-subscription", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ url: subUrl.trim() }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        notify(`Subscription import failed: ${data.error || `HTTP ${res.status}`}`);
        return;
      }
      setSubUrl("");
      setMode("closed");
      notify(`Imported ${data.imported ?? 0} relay(s) from subscription URL.`);
      await onRelaysChanged();
    } catch {
      notify("Failed to fetch subscription URL.");
    } finally {
      setBusy(false);
    }
  };

  const handleCopyShareURI = async (relay: RelaySummary) => {
    try {
      const res = await apiFetch("/api/relay/export-uri", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ id: relay.id, endpoint: relay.endpoint }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok || !data.uri) {
        notify(data.error || "Only native Lagvex relays with a PSK can be exported as lagvex:// URIs.");
        return;
      }
      await navigator.clipboard.writeText(data.uri);
      notify(`Copied lagvex:// share URI for "${relay.cleanName}" to clipboard.`);
    } catch {
      notify("Could not copy share URI.");
    }
  };

  const handleDeleteRelay = async (relay: RelaySummary) => {
    try {
      const res = await apiFetch("/api/relay/delete", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ id: relay.id }),
      });
      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        notify(data.error || "Could not remove relay.");
        return;
      }
      notify(`Removed relay "${relay.cleanName}".`);
      await onRelaysChanged();
    } catch {
      notify("Failed to remove relay.");
    }
  };

  const handleReadWGFile = (file: File | null) => {
    if (!file) return;
    const reader = new FileReader();
    reader.onload = () => {
      if (typeof reader.result === "string") {
        setWgConfig(reader.result);
        if (!wgName) {
          setWgName(file.name.replace(/\.conf$/i, ""));
        }
      }
    };
    reader.readAsText(file);
  };

  return (
    <div className="p-5 rounded-2xl bg-gradient-to-r from-[#0c1624] via-[#101a2c] to-[#0c1421] border border-[#1f314a] flex flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div className="flex flex-col gap-1 min-w-[280px] flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <span className="px-2.5 py-0.5 rounded-full bg-[#10b981]/15 border border-[#10b981]/35 text-[10px] font-brand font-bold text-[#34d399] uppercase tracking-wider">
              Free Mode &amp; Quick Import
            </span>
            <span className="text-[11px] font-mono text-white/60">
              Zero-VPS Cloudflare WARP &middot; WireGuard .conf &middot; Squad Share Links
            </span>
          </div>
          <h3 className="font-brand font-extrabold text-base text-white">
            Start Free Without a VPS — or Import a Shared / WireGuard Node
          </h3>
          <p className="text-xs text-white/70 max-w-[680px] leading-relaxed">
            Use Cloudflare WARP&apos;s free Anycast WireGuard backbone or any standard <code className="text-[#00F0FF]">.conf</code> file with Lagvex&apos;s route-based game split-tunneling, or paste a teammate&apos;s <code className="text-[#00F0FF]">lagvex://</code> link for full Reed-Solomon FEC + Dual-Path acceleration.
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2 shrink-0">
          <button
            type="button"
            disabled={busy}
            onClick={() => void handleRegisterWARP()}
            className="h-9 px-4 rounded-xl bg-[#10b981] hover:bg-[#34d399] border border-transparent text-black font-brand font-extrabold text-xs whitespace-nowrap inline-flex items-center gap-1.5 cursor-pointer transition-all disabled:opacity-50"
          >
            <svg className="w-3.5 h-3.5 shrink-0" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
              <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" />
            </svg>
            <span>{busy ? "Setting up..." : "Enable Free WARP"}</span>
          </button>
          <button
            type="button"
            onClick={() => setMode(mode === "wireguard" ? "closed" : "wireguard")}
            className={`h-9 px-3.5 rounded-xl text-xs font-brand font-bold whitespace-nowrap inline-flex items-center gap-1.5 cursor-pointer border transition-all ${
              mode === "wireguard"
                ? "bg-[#7E42FF]/25 border-[#7E42FF] text-white"
                : "bg-[#111a28] hover:bg-[#182438] border-[#25354b] text-white/90"
            }`}
          >
            <svg className="w-3.5 h-3.5 shrink-0 text-[#A366FF]" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            </svg>
            <span>Import WireGuard .conf</span>
          </button>
          <button
            type="button"
            onClick={() => setMode(mode === "uri" ? "closed" : "uri")}
            className={`h-9 px-3.5 rounded-xl text-xs font-brand font-bold whitespace-nowrap inline-flex items-center gap-1.5 cursor-pointer border transition-all ${
              mode === "uri"
                ? "bg-[#00F0FF]/20 border-[#00F0FF] text-[#00F0FF]"
                : "bg-[#111a28] hover:bg-[#182438] border-[#25354b] text-white/90"
            }`}
          >
            <svg className="w-3.5 h-3.5 shrink-0 text-[#00F0FF]" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
              <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
              <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
            </svg>
            <span>Paste lagvex:// Link</span>
          </button>
          <button
            type="button"
            onClick={() => setMode(mode === "subscription" ? "closed" : "subscription")}
            className={`h-9 px-3.5 rounded-xl text-xs font-brand font-bold whitespace-nowrap inline-flex items-center gap-1.5 cursor-pointer border transition-all ${
              mode === "subscription"
                ? "bg-[#00F0FF]/20 border-[#00F0FF] text-[#00F0FF]"
                : "bg-[#111a28] hover:bg-[#182438] border-[#25354b] text-white/90"
            }`}
          >
            <svg className="w-3.5 h-3.5 shrink-0 text-[#00F0FF]" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
              <line x1="8" y1="6" x2="21" y2="6" />
              <line x1="8" y1="12" x2="21" y2="12" />
              <line x1="8" y1="18" x2="21" y2="18" />
              <line x1="3" y1="6" x2="3.01" y2="6" />
              <line x1="3" y1="12" x2="3.01" y2="12" />
              <line x1="3" y1="18" x2="3.01" y2="18" />
            </svg>
            <span>Community List URL</span>
          </button>
        </div>
      </div>

      {mode === "uri" && (
        <div className="p-4 rounded-2xl bg-[#080e17] border border-[#1d2d42] flex flex-col gap-3">
          <div className="text-xs font-brand font-bold text-white">
            Import Shared Relay URI (<code className="text-[#00F0FF]">lagvex://host:port?psk=...</code>)
          </div>
          <div className="flex gap-2">
            <input
              type="text"
              value={shareURI}
              onChange={(e) => setShareURI(e.target.value)}
              placeholder="lagvex://203.0.113.10:51820?psk=...&name=SG+Squad&loc=Singapore"
              className="flex-1 h-10 px-3 rounded-xl bg-[#0f1724] border border-[#223247] text-xs font-mono text-white outline-none focus:border-[#00F0FF]"
            />
            <button
              type="button"
              disabled={busy}
              onClick={() => void handleImportURI()}
              className="px-4 h-10 rounded-xl bg-[#00F0FF] hover:bg-[#4df5ff] text-black font-brand font-extrabold text-xs cursor-pointer border-none"
            >
              Import Node
            </button>
          </div>
        </div>
      )}

      {mode === "wireguard" && (
        <div className="p-4 rounded-2xl bg-[#080e17] border border-[#29224a] flex flex-col gap-3">
          <div className="flex items-center justify-between gap-2">
            <div className="text-xs font-brand font-bold text-white">
              Import Standard WireGuard / WARP / Mullvad <code className="text-[#A366FF]">.conf</code>
            </div>
            <label className="px-3 py-1 rounded-lg bg-white/5 hover:bg-white/10 border border-white/15 text-[11px] font-brand text-white/80 cursor-pointer">
              Load .conf file
              <input
                type="file"
                accept=".conf,.txt"
                className="hidden"
                onChange={(e) => handleReadWGFile(e.target.files?.[0] ?? null)}
              />
            </label>
          </div>
          <input
            type="text"
            value={wgName}
            onChange={(e) => setWgName(e.target.value)}
            placeholder="Label (optional, e.g. Cloudflare WARP SG or My WireGuard VPS)"
            className="h-9 px-3 rounded-xl bg-[#0f1724] border border-[#223247] text-xs text-white outline-none focus:border-[#7E42FF]"
          />
          <textarea
            rows={8}
            value={wgConfig}
            onChange={(e) => setWgConfig(e.target.value)}
            placeholder={`[Interface]\nPrivateKey = ...\nAddress = 172.16.0.2/32\nMTU = 1280\n\n[Peer]\nPublicKey = ...\nEndpoint = 162.159.192.1:2408`}
            className="min-h-[156px] p-3 rounded-xl bg-[#0f1724] border border-[#223247] text-xs font-mono leading-relaxed text-white outline-none focus:border-[#7E42FF]"
          />
          <div className="flex justify-end">
            <button
              type="button"
              disabled={busy}
              onClick={() => void handleImportWireGuard()}
              className="px-4 h-9 rounded-xl bg-[#7E42FF] hover:bg-[#9360ff] text-white font-brand font-bold text-xs cursor-pointer border-none"
            >
              Save WireGuard Endpoint
            </button>
          </div>
        </div>
      )}

      {mode === "subscription" && (
        <div className="p-4 rounded-2xl bg-[#080e17] border border-[#1d2d42] flex flex-col gap-3">
          <div className="text-xs font-brand font-bold text-white">
            Import Relay List from HTTPS URL (JSON array or newline <code className="text-[#00F0FF]">lagvex://</code> URIs)
          </div>
          <div className="flex gap-2">
            <input
              type="url"
              value={subUrl}
              onChange={(e) => setSubUrl(e.target.value)}
              placeholder="https://example.com/squad-relays.json"
              className="flex-1 h-10 px-3 rounded-xl bg-[#0f1724] border border-[#223247] text-xs font-mono text-white outline-none focus:border-[#00F0FF]"
            />
            <button
              type="button"
              disabled={busy}
              onClick={() => void handleImportSubscription()}
              className="px-4 h-10 rounded-xl bg-[#00F0FF] hover:bg-[#4df5ff] text-black font-brand font-extrabold text-xs cursor-pointer border-none"
            >
              Fetch &amp; Import
            </button>
          </div>
        </div>
      )}

      {relays.length > 0 && (
        <div className="pt-2 border-t border-white/5 flex flex-wrap items-center gap-2">
          <span className="text-[10px] font-brand uppercase tracking-wider text-white/40 mr-1">
            Quick Share / Manage:
          </span>
          {relays.map((rel) => (
            <div
              key={rel.id}
              className="flex items-center gap-1.5 pl-2.5 pr-1.5 py-1 rounded-lg bg-white/[0.04] border border-white/10 text-[11px] font-mono text-white/80"
            >
              <span className="truncate max-w-[160px]">{rel.cleanName}</span>
              {rel.tier === "Free" || rel.tier === "WireGuard" ? (
                <span className="px-1.5 py-0.5 rounded bg-[#10b981]/15 text-[#34d399] text-[9px] font-brand font-bold">
                  WG
                </span>
              ) : (
                <button
                  type="button"
                  onClick={() => void handleCopyShareURI(rel)}
                  title="Copy lagvex:// share URI (includes PSK)"
                  className="px-1.5 py-0.5 rounded bg-[#00F0FF]/15 hover:bg-[#00F0FF]/25 text-[#00F0FF] text-[9px] font-brand font-bold cursor-pointer border-none"
                >
                  SHARE
                </button>
              )}
              <button
                type="button"
                onClick={() => void handleDeleteRelay(rel)}
                title="Remove relay from catalog"
                aria-label={`Remove ${rel.cleanName}`}
                className="w-5 h-5 inline-flex items-center justify-center rounded bg-red-500/15 hover:bg-red-500/30 text-red-300 cursor-pointer border-none"
              >
                <svg className="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" aria-hidden="true">
                  <line x1="18" y1="6" x2="6" y2="18" />
                  <line x1="6" y1="6" x2="18" y2="18" />
                </svg>
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
