import { clamp01, smoothstep } from "../sky/color";
import type { SkyInputs } from "../sky/types";

/**
 * What the ambient sound should contain, each layer from 0 (silent) to 1 (full).
 * It is worked out from the sky itself, so every mood and the live sky get a
 * fitting soundscape: birds at dawn, crickets after sunset, rain, wind, thunder.
 */
export interface Scene {
  wind: number; // moving air
  rain: number; // steady rain
  drips: number; // single drops
  hush: number; // muffled snowy air
  chimes: number; // rare, soft bells in the snow
  birds: number;
  crickets: number;
  owl: number;
  drone: number; // deep night air
  thunder: number; // rolls after lightning
}

export const SILENT: Scene = { wind: 0, rain: 0, drips: 0, hush: 0, chimes: 0, birds: 0, crickets: 0, owl: 0, drone: 0, thunder: 0 };

export function sceneFor(i: SkyInputs): Scene {
  const e = i.sunElevationDeg;
  const day = smoothstep(-2, 12, e);
  const night = 1 - smoothstep(-14, -2, e);
  const dusk = 1 - smoothstep(-4, 6, e); // sky still dim: crickets time
  const cover = clamp01(i.cloudCoverPct / 100);
  const wet = clamp01(i.precipitationMmh / 5);
  const rainy = smoothstep(0.1, 0.6, i.precipitationMmh);
  const snowy = clamp01(i.snowfallCmh / 1.5);
  const snowing = smoothstep(0.05, 0.4, i.snowfallCmh);
  const calm = (1 - rainy) * (1 - snowing); // dry and not snowing
  const windN = clamp01(i.windSpeedKmh / 45);
  const east = clamp01((200 - i.sunAzimuthDeg) / 60); // 1 in the morning sky, 0 in the evening one

  const birdTime = smoothstep(-9, -3, e) * (1 - 0.5 * smoothstep(15, 45, e)) * (0.25 + 0.75 * east);
  const moonUp = i.moonElevationDeg > 0 ? 1 : 0.4;

  return {
    wind: clamp01((0.1 + 0.75 * windN + 0.1 * cover) * (1 - 0.55 * snowing)),
    rain: rainy * (0.3 + 0.7 * wet),
    drips: rainy * (0.55 - 0.3 * wet),
    hush: snowing * (0.4 + 0.6 * snowy),
    chimes: snowing * 0.4,
    birds: calm * (1 - 0.7 * cover) * Math.max(birdTime, 0.45 * day),
    crickets: calm * dusk * (1 - 0.3 * cover) * (0.55 + 0.45 * (1 - east)),
    owl: calm * night * moonUp * 0.5,
    drone: night * (0.45 + 0.35 * (1 - rainy)),
    thunder: i.thunder ? 1 : 0,
  };
}
