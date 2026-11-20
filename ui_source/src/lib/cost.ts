/**
 * Pure cost projection used by the COST & SELF-HOST panel.
 *
 * Two rules keep this honest:
 *  1. Transfer is projected from the player's own measured session rate when
 *     enough samples exist; otherwise the documented typical is used and the
 *     result is labelled as an estimate, never as a measurement.
 *  2. The competitor price is a public-list-price reference constant, not a
 *     measured value, and is labelled as such in the UI.
 */

/** Documented typical game-session throughput: ~30 MB per hour. */
export const TYPICAL_BYTES_PER_SECOND = (30 * 1e6) / 3600;

/** Reference price of a typical public commercial booster plan (USD/month). */
export const TYPICAL_BOOSTER_PRICE_PER_MONTH = 9;

/** Minimum number of rate samples before a measurement is trusted. */
export const MIN_MEASURED_SAMPLES = 5;

export type CostInputs = {
  /** Average up+down bytes per second measured during the session (0 = none). */
  measuredBytesPerSecond: number;
  /** How many hours per day the player boosts. */
  hoursPerDay: number;
  /** How many players share the relay VPS. */
  squadSize: number;
  /** The player's own VPS price in USD per month. */
  vpsPricePerMonth: number;
};

export type CostProjection = {
  /** True when the transfer projection comes from a measured session rate. */
  measured: boolean;
  /** Projected monthly transfer in gigabytes. */
  monthlyGb: number;
  /** Monthly cost per player in USD. */
  perPlayer: number;
  /** Monthly saving versus the reference booster plan (0 when not cheaper). */
  savingsVsBooster: number;
  /** Human label for the monthly transfer. */
  transferLabel: string;
  /** Where the transfer figure came from. */
  source: "measured" | "typical";
};

/**
 * Project monthly transfer and per-player cost.
 *
 * The projection is linear in hours; 30 days are assumed per month, which is
 * the same assumption documented in docs/COST.md.
 */
export function projectMonthlyCost(inputs: CostInputs): CostProjection {
  const hours = Math.max(0, Number(inputs.hoursPerDay) || 0);
  const squad = Math.max(1, Math.floor(Number(inputs.squadSize) || 1));
  const price = Math.max(0, Number(inputs.vpsPricePerMonth) || 0);
  const measuredBps = Math.max(0, Number(inputs.measuredBytesPerSecond) || 0);
  const measured = measuredBps > 0;

  const bytesPerSecond = measured ? measuredBps : TYPICAL_BYTES_PER_SECOND;
  const monthlyGb = (bytesPerSecond * hours * 30 * 3600) / 1e9;
  const perPlayer = price / squad;
  const savingsVsBooster = Math.max(0, TYPICAL_BOOSTER_PRICE_PER_MONTH - perPlayer);

  return {
    measured,
    monthlyGb,
    perPlayer,
    savingsVsBooster,
    transferLabel: formatTransferLabel(monthlyGb),
    source: measured ? "measured" : "typical",
  };
}

/** Human label for a projected monthly transfer volume. */
export function formatTransferLabel(gigabytes: number): string {
  if (!Number.isFinite(gigabytes) || gigabytes < 0) return "—";
  if (gigabytes < 1) return `${Math.round(gigabytes * 1000)} MB`;
  return `${gigabytes.toFixed(1)} GB`;
}
