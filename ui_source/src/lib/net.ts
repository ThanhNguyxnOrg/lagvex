/**
 * Pure network helpers shared by the dashboard.
 *
 * These live outside the React component tree so they can be unit-tested
 * directly (see *.test.ts) — the coverage readout in the UI depends on them
 * being correct, and a silent regression there would fabricate route data.
 */

/** Parse an IPv4 literal into an unsigned 32-bit value, or null. */
export function ipv4ToUint(host: string): number | null {
  const m = /^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$/.exec(host);
  if (!m) return null;
  const oct = m.slice(1).map(Number);
  if (oct.some((o) => o > 255)) return null;
  return ((oct[0] << 24) | (oct[1] << 16) | (oct[2] << 8) | oct[3]) >>> 0;
}

/** True for routable public IPv4 literals (the TCP ping tool refuses private ones). */
export function isPublicIPv4(host: string): boolean {
  const m = /^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$/.exec(host);
  if (!m) return false;
  const a = Number(m[1]);
  const b = Number(m[2]);
  if (a === 0 || a === 10 || a === 127) return false;
  if (a === 169 && b === 254) return false;
  if (a === 172 && b >= 16 && b <= 31) return false;
  if (a === 192 && b === 168) return false;
  if (a === 100 && b >= 64 && b <= 127) return false;
  return a < 224;
}

/** True when the IPv4 host falls inside the CIDR block. */
export function isIPv4InCIDR(host: string, cidr: string): boolean {
  const slash = cidr.indexOf("/");
  if (slash < 0) return false;
  const bits = Number(cidr.slice(slash + 1));
  const ip = ipv4ToUint(host);
  const net = ipv4ToUint(cidr.slice(0, slash));
  if (ip === null || net === null || !Number.isFinite(bits) || bits < 0 || bits > 32) return false;
  const mask = bits === 0 ? 0 : (0xffffffff << (32 - bits)) >>> 0;
  return (ip & mask) === (net & mask);
}

/**
 * Extract the host part from a socket-table endpoint.
 *
 * Accepts "1.2.3.4:443", "[2001:db8::1]:443" and bare hosts; returns an
 * empty string for anything unusable.
 */
export function hostFromEndpoint(endpoint: string): string {
  const value = (endpoint || "").trim();
  if (!value) return "";
  if (value.includes("]")) {
    const end = value.indexOf("]");
    return end > 0 ? value.slice(1, end) : "";
  }
  const colon = value.indexOf(":");
  if (colon >= 0) return value.slice(0, colon);
  return value;
}

/**
 * Count how many observed endpoints fall inside the given CIDR blocks.
 *
 * Only public IPv4 literals are counted — private or malformed entries are
 * ignored so the coverage badge never overstates route coverage.
 */
export function countEndpointsInCIDRs(hosts: Iterable<string>, cidrs: string[]): {
  total: number;
  inside: number;
} {
  const unique = new Set<string>();
  for (const host of hosts) {
    if (isPublicIPv4(host)) unique.add(host);
  }
  let inside = 0;
  for (const host of unique) {
    if (cidrs.some((cidr) => isIPv4InCIDR(host, cidr))) inside++;
  }
  return { total: unique.size, inside };
}
