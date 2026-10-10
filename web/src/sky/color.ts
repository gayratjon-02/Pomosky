// Small colour and easing helpers. Colours are [r, g, b] with 0..255 channels.

export type RGB = [number, number, number];

export const clamp = (v: number, min: number, max: number): number => Math.min(max, Math.max(min, v));
export const clamp01 = (v: number): number => clamp(v, 0, 1);
export const lerp = (a: number, b: number, t: number): number => a + (b - a) * t;

/** 0 below a, 1 above b, smooth in between. */
export function smoothstep(a: number, b: number, x: number): number {
  const t = clamp01((x - a) / (b - a));
  return t * t * (3 - 2 * t);
}

/** Parses "#rgb" or "#rrggbb". */
export function hex(value: string): RGB {
  let s = value.replace("#", "");
  if (s.length === 3) s = s.split("").map((c) => c + c).join("");
  const n = parseInt(s, 16);
  return [(n >> 16) & 255, (n >> 8) & 255, n & 255];
}

export function mix(a: RGB, b: RGB, t: number): RGB {
  return [lerp(a[0], b[0], t), lerp(a[1], b[1], t), lerp(a[2], b[2], t)];
}

export function luminance([r, g, b]: RGB): number {
  return (0.2126 * r + 0.7152 * g + 0.0722 * b) / 255;
}

/** Moves a colour towards its own grey (amount 0..1) and scales its brightness. */
export function desaturate(c: RGB, amount: number, brightness = 1): RGB {
  const grey = luminance(c) * 255;
  const mixed = mix(c, [grey, grey, grey], clamp01(amount));
  return [mixed[0] * brightness, mixed[1] * brightness, mixed[2] * brightness];
}

export function css(c: RGB, alpha = 1): string {
  const [r, g, b] = c.map((v) => Math.round(clamp(v, 0, 255)));
  return alpha >= 1 ? `rgb(${r}, ${g}, ${b})` : `rgba(${r}, ${g}, ${b}, ${alpha})`;
}
