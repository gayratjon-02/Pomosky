import type { SkyInputs } from "./types";

// A tiny bridge from the sky to anything that wants to follow it, such as the
// ambient sound: the sky announces the values it is heading for (live or mood)
// and every lightning flash.

let latest: SkyInputs | null = null;
const skyListeners = new Set<(i: SkyInputs) => void>();
const boltListeners = new Set<() => void>();

export const latestSky = (): SkyInputs | null => latest;

export function publishSky(inputs: SkyInputs): void {
  latest = { ...inputs };
  skyListeners.forEach((fn) => fn(latest!));
}

/** Calls fn with the target sky now (if known) and on every change. Returns a stop function. */
export function onSky(fn: (i: SkyInputs) => void): () => void {
  skyListeners.add(fn);
  if (latest) fn(latest);
  return () => skyListeners.delete(fn);
}

export function publishBolt(): void {
  boltListeners.forEach((fn) => fn());
}

export function onBolt(fn: () => void): () => void {
  boltListeners.add(fn);
  return () => boltListeners.delete(fn);
}
