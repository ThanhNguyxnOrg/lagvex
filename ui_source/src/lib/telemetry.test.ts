import { test } from "node:test";
import assert from "node:assert/strict";

import {
  RTT_MAX_AGE_MS,
  cpuPercentLabel,
  isRttFresh,
} from "./telemetry.ts";

test("fresh sampled RTT is live", () => {
  assert.equal(isRttFresh(true, 0), true);
  assert.equal(isRttFresh(true, RTT_MAX_AGE_MS - 1), true);
});

test("an RTT older than three keepalives is stale, not live", () => {
  assert.equal(isRttFresh(true, RTT_MAX_AGE_MS + 1), false);
  // A far-past sample must never be shown as the current RTT.
  assert.equal(isRttFresh(true, 600_000), false);
});

test("an unsampled or malformed age is never live", () => {
  assert.equal(isRttFresh(false, 0), false);
  assert.equal(isRttFresh(false, 10), false);
  assert.equal(isRttFresh(true, Number.NaN), false);
  assert.equal(isRttFresh(true, -1), false);
});

test("cpu label renders a measured percentage", () => {
  assert.equal(cpuPercentLabel(0), "0.0%");
  assert.equal(cpuPercentLabel(12.34), "12.3%");
});

test("cpu label is null when the value was not measured", () => {
  assert.equal(cpuPercentLabel(null), null);
  assert.equal(cpuPercentLabel(undefined), null);
  assert.equal(cpuPercentLabel(-0.5), null);
  assert.equal(cpuPercentLabel(Number.NaN), null);
});