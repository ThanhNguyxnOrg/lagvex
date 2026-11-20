/**
 * Pure telemetry-honesty helpers shared by the dashboard.
 *
 * The dashboard only shows values that were measured. These helpers gate
 * those values so a stale RTT sample or an unsupported CPU counter renders
 * "not measured" instead of a plausible-looking number.
 */

/** The engine's authenticated keepalive period, in milliseconds. */
export const KEEPALIVE_INTERVAL_MS = 2000;

/**
 * An RTT sample counts as live only while a Pong arrived within three
 * keepalive periods; older numbers describe the past, not the current route.
 */
export const RTT_MAX_AGE_MS = 3 * KEEPALIVE_INTERVAL_MS;

/** True when a sampled RTT is recent enough to present as live. */
export function isRttFresh(sampled: boolean, ageMs: number): boolean {
  return sampled && Number.isFinite(ageMs) && ageMs >= 0 && ageMs <= RTT_MAX_AGE_MS;
}

/**
 * Label for a per-core CPU percentage from the process scan, or null when the
 * value was not measured: either the OS exposes no counter (the API reports
 * cpuSupported = false) or a second sample has not been taken yet.
 */
export function cpuPercentLabel(cpuPercent: number | null | undefined): string | null {
  if (typeof cpuPercent !== "number" || !Number.isFinite(cpuPercent) || cpuPercent < 0) {
    return null;
  }
  return `${cpuPercent.toFixed(1)}%`;
}