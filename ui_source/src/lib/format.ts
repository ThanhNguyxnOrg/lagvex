/**
 * Pure display helpers shared by the dashboard.
 *
 * Kept free of React so they can be unit-tested directly; every one of these
 * renders a number that must never be fabricated.
 */

/** Short tag for the measurement technique behind a node's RTT value. */
export function probeMethodShort(method: string): string {
  switch (method) {
    case "tunnel-udp":
      return "TUN";
    case "icmp-echo":
      return "ICMP";
    case "tcp-connect":
      return "TCP";
    case "udp-echo":
      return "UDP";
    default:
      return "";
  }
}

/** Minimal relay shape needed for the route hint (avoids a type import cycle). */
export type RouteHintRelay = {
  totalEstimateMs: number;
  pingMs: number;
};

/** Best measured metric for a relay, formatted for the route hint. */
export function routeHintMsLabel(relay: RouteHintRelay): string {
  const v = relay.totalEstimateMs > 0 ? relay.totalEstimateMs : relay.pingMs;
  return v > 0 ? `${Math.round(v * 10) / 10} ms` : "—";
}

/**
 * Short relative age for a persisted measurement timestamp.
 *
 * `nowMs` is injectable so the output is deterministic under test.
 */
export function formatAgeShort(iso: string, nowMs: number = Date.now()): string {
  const at = Date.parse(iso);
  if (!Number.isFinite(at)) return "unknown";
  const seconds = Math.max(0, Math.round((nowMs - at) / 1000));
  if (seconds < 60) return `${seconds}s ago`;
  if (seconds < 3600) return `${Math.round(seconds / 60)}m ago`;
  if (seconds < 86400) return `${Math.round(seconds / 3600)}h ago`;
  return `${Math.round(seconds / 86400)}d ago`;
}

/** Human label for a projected monthly transfer volume. */
export function formatTransferVolume(gigabytes: number): string {
  if (!Number.isFinite(gigabytes) || gigabytes < 0) return "—";
  if (gigabytes < 1) return `${Math.round(gigabytes * 1000)} MB`;
  return `${gigabytes.toFixed(1)} GB`;
}
