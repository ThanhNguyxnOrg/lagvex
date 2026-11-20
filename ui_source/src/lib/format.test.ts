import test from "node:test";
import assert from "node:assert/strict";

import { formatAgeShort, formatTransferVolume, probeMethodShort, routeHintMsLabel } from "./format.ts";

test("probeMethodShort maps known probe techniques to tags", () => {
  assert.equal(probeMethodShort("tunnel-udp"), "TUN");
  assert.equal(probeMethodShort("icmp-echo"), "ICMP");
  assert.equal(probeMethodShort("tcp-connect"), "TCP");
  assert.equal(probeMethodShort("udp-echo"), "UDP");
  assert.equal(probeMethodShort("mystery"), "");
  assert.equal(probeMethodShort(""), "");
});

test("routeHintMsLabel prefers the total estimate and never invents a value", () => {
  assert.equal(routeHintMsLabel({ totalEstimateMs: 42.44, pingMs: 10 }), "42.4 ms");
  assert.equal(routeHintMsLabel({ totalEstimateMs: 0, pingMs: 12.34 }), "12.3 ms");
  assert.equal(routeHintMsLabel({ totalEstimateMs: 0, pingMs: 0 }), "—");
  assert.equal(routeHintMsLabel({ totalEstimateMs: -1, pingMs: -2 }), "—");
});

test("formatAgeShort renders relative ages deterministically", () => {
  const now = Date.parse("2026-11-19T12:00:00Z");
  assert.equal(formatAgeShort("2026-11-19T11:59:30Z", now), "30s ago");
  assert.equal(formatAgeShort("2026-11-19T11:30:00Z", now), "30m ago");
  assert.equal(formatAgeShort("2026-11-19T09:00:00Z", now), "3h ago");
  assert.equal(formatAgeShort("2026-11-16T12:00:00Z", now), "3d ago");
  assert.equal(formatAgeShort("not-a-date", now), "unknown");
  assert.equal(formatAgeShort("2026-11-19T12:00:00Z", now - 60_000), "0s ago", "future timestamps clamp to 0s");
});

test("formatTransferVolume switches units under 1 GB", () => {
  assert.equal(formatTransferVolume(0.0027), "3 MB");
  assert.equal(formatTransferVolume(2.74), "2.7 GB");
  assert.equal(formatTransferVolume(0), "0 MB");
  assert.equal(formatTransferVolume(-1), "—");
});
