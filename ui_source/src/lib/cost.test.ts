import test from "node:test";
import assert from "node:assert/strict";

import {
  MIN_MEASURED_SAMPLES,
  TYPICAL_BOOSTER_PRICE_PER_MONTH,
  TYPICAL_BYTES_PER_SECOND,
  projectMonthlyCost,
} from "./cost.ts";

test("the typical throughput matches the documented 30 MB per hour", () => {
  assert.equal(Math.round((TYPICAL_BYTES_PER_SECOND * 3600) / 1e6), 30);
});

test("without a measurement the projection uses the documented typical", () => {
  const result = projectMonthlyCost({
    measuredBytesPerSecond: 0,
    hoursPerDay: 3,
    squadSize: 1,
    vpsPricePerMonth: 3.5,
  });
  assert.equal(result.measured, false);
  assert.equal(result.source, "typical");
  assert.ok(Math.abs(result.monthlyGb - 2.7) < 0.01, `expected ~2.7 GB, got ${result.monthlyGb}`);
  assert.equal(result.transferLabel, "2.7 GB");
});

test("a measured session rate drives the projection", () => {
  const measured = 1_000_000; // 1 MB/s
  const result = projectMonthlyCost({
    measuredBytesPerSecond: measured,
    hoursPerDay: 2,
    squadSize: 1,
    vpsPricePerMonth: 3.5,
  });
  assert.equal(result.measured, true);
  assert.equal(result.source, "measured");
  const expected = (measured * 2 * 30 * 3600) / 1e9;
  assert.ok(Math.abs(result.monthlyGb - expected) < 1e-6);
});

test("squad sharing splits the monthly price", () => {
  const result = projectMonthlyCost({
    measuredBytesPerSecond: 0,
    hoursPerDay: 1,
    squadSize: 4,
    vpsPricePerMonth: 4,
  });
  assert.equal(result.perPlayer, 1);
  assert.equal(result.savingsVsBooster, TYPICAL_BOOSTER_PRICE_PER_MONTH - 1);
});

test("a free tier yields a zero price and the full saving", () => {
  const result = projectMonthlyCost({
    measuredBytesPerSecond: 0,
    hoursPerDay: 1,
    squadSize: 1,
    vpsPricePerMonth: 0,
  });
  assert.equal(result.perPlayer, 0);
  assert.equal(result.savingsVsBooster, TYPICAL_BOOSTER_PRICE_PER_MONTH);
});

test("invalid inputs are clamped instead of producing NaN", () => {
  const result = projectMonthlyCost({
    measuredBytesPerSecond: -5,
    hoursPerDay: -2,
    squadSize: 0,
    vpsPricePerMonth: -10,
  });
  assert.equal(result.monthlyGb, 0);
  assert.equal(result.perPlayer, 0, "squad size below 1 is treated as a single player");
  assert.equal(result.measured, false);
  assert.equal(result.transferLabel, "0 MB");
  assert.equal(result.savingsVsBooster, TYPICAL_BOOSTER_PRICE_PER_MONTH);
});

test("the minimum sample threshold is the one the dashboard uses", () => {
  assert.equal(MIN_MEASURED_SAMPLES, 5);
});
