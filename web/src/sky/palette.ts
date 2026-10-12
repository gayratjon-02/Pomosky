import { desaturate, hex, mix, smoothstep, type RGB } from "./color";

// The sky colours follow the sun's elevation above the horizon (degrees).
// Between the keyframes the colours are interpolated smoothly.

interface Key {
  e: number;
  top: RGB;
  mid: RGB;
  bottom: RGB;
}

const key = (e: number, top: string, mid: string, bottom: string): Key => ({
  e,
  top: hex(top),
  mid: hex(mid),
  bottom: hex(bottom),
});

const KEYS: Key[] = [
  key(-18, "#02030f", "#050a22", "#0a1236"), // night
  key(-12, "#060b2a", "#101a4d", "#27306b"), // nautical twilight
  key(-6, "#17236a", "#4a3f86", "#c4607a"), // dusk / dawn
  key(-1, "#2d3f94", "#a15f9d", "#ff9a62"), // sunrise / sunset
  key(4, "#3f62b8", "#c98fb0", "#ffc27a"),
  key(12, "#3f7fd0", "#8fb9e8", "#ffe1b0"), // golden hour
  key(28, "#2b7ad8", "#6fb2ef", "#b4dbff"), // day
  key(60, "#1566cc", "#4ea1ee", "#a3d1ff"), // high noon
];

export interface SkyGradient {
  top: RGB;
  mid: RGB;
  bottom: RGB;
}

/**
 * Sky gradient for a sun elevation. `overcast` (0 clear .. 1 fully overcast)
 * desaturates and darkens the sky the way a cloud deck does.
 */
export function skyGradient(sunElevationDeg: number, overcast = 0): SkyGradient {
  const e = Math.min(KEYS[KEYS.length - 1]!.e, Math.max(KEYS[0]!.e, sunElevationDeg));
  let i = 0;
  while (i < KEYS.length - 2 && e > KEYS[i + 1]!.e) i++;
  const a = KEYS[i]!;
  const b = KEYS[i + 1]!;
  const t = smoothstep(a.e, b.e, e);

  const g: SkyGradient = {
    top: mix(a.top, b.top, t),
    mid: mix(a.mid, b.mid, t),
    bottom: mix(a.bottom, b.bottom, t),
  };
  if (overcast <= 0) return g;

  // Overcast skies are greyer and a little darker, but never go flat.
  const amount = 0.75 * overcast;
  const dim = 1 - 0.18 * overcast;
  return {
    top: desaturate(g.top, amount, dim),
    mid: desaturate(g.mid, amount, dim),
    bottom: desaturate(g.bottom, amount * 0.9, dim),
  };
}

/** Colour of the sun disc: white-yellow when high, deep orange near the horizon. */
export function sunColor(sunElevationDeg: number): RGB {
  return mix(hex("#ff7a2e"), hex("#fff6dc"), smoothstep(-2, 25, sunElevationDeg));
}

/** Colour of the lit side of clouds: pink-gold at sunrise/sunset, white by day, deep blue at night. */
export function cloudLit(sunElevationDeg: number): RGB {
  const night = hex("#3a4778");
  const twilight = hex("#ffb899");
  const day = hex("#ffffff");
  const toTwilight = smoothstep(-14, 0, sunElevationDeg);
  const toDay = smoothstep(2, 22, sunElevationDeg);
  return mix(mix(night, twilight, toTwilight), day, toDay);
}

/** Colour of the shaded underside of clouds. */
export function cloudShade(sunElevationDeg: number, overcast: number): RGB {
  const lit = cloudLit(sunElevationDeg);
  const dark = desaturate(mix(lit, hex("#1c2440"), 0.55), 0.35);
  return mix(mix(lit, dark, 0.45), mix(dark, hex("#0e1426"), 0.35), overcast);
}

/** How much daylight there is, 0 at night and in twilight, 1 once the sun is well up. */
export const dayAmount = (sunElevationDeg: number): number => smoothstep(0, 20, sunElevationDeg);

// By day the hills are seen through air: the far ridge is pale and blue, the near one deep green.
const DAY_HILLS = [hex("#9dbce0"), hex("#5f93b4"), hex("#2f6f68")];

/**
 * Colour of the horizon silhouette (hills), tinted by the sky above it. At night
 * they are almost black; by day they are hazy blue far away and green up close.
 */
export function horizonColor(sunElevationDeg: number, depth: number): RGB {
  const sky = skyGradient(sunElevationDeg).bottom;
  const night = mix(sky, hex("#05070f"), 0.6 + 0.3 * depth);
  const day = dayAmount(sunElevationDeg);
  if (day <= 0) return night;
  const i = Math.min(DAY_HILLS.length - 2, Math.floor(depth * (DAY_HILLS.length - 1)));
  const local = depth * (DAY_HILLS.length - 1) - i;
  const lit = mix(DAY_HILLS[i]!, DAY_HILLS[i + 1]!, local);
  // Low sun: let the warm horizon light stain the hills a little.
  const warm = 0.18 * (1 - smoothstep(8, 30, sunElevationDeg));
  return mix(night, mix(lit, sky, warm), day);
}
