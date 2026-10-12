import { clamp01, css, smoothstep, type RGB } from "../color";
import { cloudLit, cloudShade, dayAmount } from "../palette";
import { createRandom } from "../random";
import type { Layer } from "../types";

const SPRITE_W = 640;
const SPRITE_H = 280;
const SPRITE_VARIANTS = 5;
const MAX_CLOUDS = 18;

interface Cloud {
  variant: number;
  depth: number; // 0 far .. 1 near
  x: number; // 0..1 of the canvas width, wraps around
  yRand: number;
  scale: number;
  speed: number;
  alpha: number;
  presence: number; // eases towards 1 when the cloud is wanted, 0 when not
}

/** A soft, flat-bottomed cloud made of overlapping translucent puffs. */
function buildSprite(seed: number): HTMLCanvasElement {
  const canvas = document.createElement("canvas");
  canvas.width = SPRITE_W;
  canvas.height = SPRITE_H;
  const ctx = canvas.getContext("2d")!;
  const rnd = createRandom(seed);

  for (let i = 0; i < 44; i++) {
    const t = rnd();
    const hump = Math.sin(t * Math.PI); // taller in the middle
    const r = (26 + rnd() * 48) * (0.5 + 0.8 * hump);
    const x = SPRITE_W * 0.1 + t * SPRITE_W * 0.8;
    const y = SPRITE_H * 0.7 - hump * SPRITE_H * 0.3 * rnd() - r * 0.2;
    const g = ctx.createRadialGradient(x, y, 0, x, y, r);
    g.addColorStop(0, "rgba(255,255,255,0.6)");
    g.addColorStop(0.55, "rgba(255,255,255,0.26)");
    g.addColorStop(1, "rgba(255,255,255,0)");
    ctx.fillStyle = g;
    ctx.beginPath();
    ctx.arc(x, y, r, 0, Math.PI * 2);
    ctx.fill();
  }
  // Bright domes on the sun side give the top a rounded, billowy look.
  for (let i = 0; i < 12; i++) {
    const t = 0.15 + rnd() * 0.7;
    const hump = Math.sin(t * Math.PI);
    const r = (22 + rnd() * 34) * (0.5 + 0.7 * hump);
    const x = SPRITE_W * 0.1 + t * SPRITE_W * 0.8;
    const y = SPRITE_H * 0.5 - hump * SPRITE_H * 0.22 * rnd();
    const g = ctx.createRadialGradient(x - r * 0.2, y - r * 0.25, 0, x, y, r);
    g.addColorStop(0, "rgba(255,255,255,0.7)");
    g.addColorStop(0.6, "rgba(255,255,255,0.22)");
    g.addColorStop(1, "rgba(255,255,255,0)");
    ctx.fillStyle = g;
    ctx.beginPath();
    ctx.arc(x, y, r, 0, Math.PI * 2);
    ctx.fill();
  }
  // Flatten the base: fade everything below a soft line.
  ctx.globalCompositeOperation = "destination-out";
  const fade = ctx.createLinearGradient(0, SPRITE_H * 0.7, 0, SPRITE_H);
  fade.addColorStop(0, "rgba(0,0,0,0)");
  fade.addColorStop(1, "rgba(0,0,0,1)");
  ctx.fillStyle = fade;
  ctx.fillRect(0, SPRITE_H * 0.7, SPRITE_W, SPRITE_H * 0.3);
  return canvas;
}

/** Copies a sprite and paints it with a lit top and a shaded underside. */
function tint(base: HTMLCanvasElement, lit: RGB, shade: RGB): HTMLCanvasElement {
  const canvas = document.createElement("canvas");
  canvas.width = base.width;
  canvas.height = base.height;
  const ctx = canvas.getContext("2d")!;
  ctx.drawImage(base, 0, 0);
  ctx.globalCompositeOperation = "source-atop";
  const g = ctx.createLinearGradient(0, SPRITE_H * 0.15, 0, SPRITE_H * 0.8);
  g.addColorStop(0, css(lit));
  g.addColorStop(0.5, css(lit));
  g.addColorStop(1, css(shade));
  ctx.fillStyle = g;
  ctx.fillRect(0, 0, SPRITE_W, SPRITE_H);
  return canvas;
}

const quantize = (c: RGB): string => c.map((v) => Math.round(v / 8) * 8).join(",");

/**
 * Drifting clouds. How many there are follows the cloud cover; they move with
 * the wind (direction and speed) and are tinted by the time of day.
 */
export function cloudsLayer(seed = 11): Layer {
  const rnd = createRandom(seed);
  const bases: HTMLCanvasElement[] = [];
  const clouds: Cloud[] = [];
  for (let i = 0; i < MAX_CLOUDS; i++) {
    const depth = rnd();
    clouds.push({
      variant: Math.floor(rnd() * SPRITE_VARIANTS),
      depth,
      x: rnd(),
      yRand: rnd(),
      scale: 0.55 + depth * 0.95 + rnd() * 0.25,
      speed: 0.6 + depth * 0.9,
      alpha: 0.55 + rnd() * 0.35,
      presence: 0,
    });
  }
  clouds.sort((a, b) => a.depth - b.depth);

  let tinted: HTMLCanvasElement[] = [];
  let tintKey = "";
  let lastRetint = -1e9;

  const ensureSprites = () => {
    if (bases.length === 0) {
      for (let i = 0; i < SPRITE_VARIANTS; i++) bases.push(buildSprite(seed * 31 + i));
    }
  };

  return {
    draw({ ctx, w, dt, timeMs, inputs, overcast, quality, horizonY }) {
      const cover = clamp01(inputs.cloudCoverPct / 100);
      const fromCover = Math.round(Math.pow(cover, 1.25) * MAX_CLOUDS * Math.min(1, 0.55 + quality * 0.6));
      // A blue day sky is never completely bare: a few fair-weather clouds drift by.
      const wanted = Math.max(fromCover, Math.round(4 * dayAmount(inputs.sunElevationDeg) * (1 - 0.5 * overcast)));

      // Even a clear sky is allowed a whisper of cloud; nothing is drawn at 0 and no cloud exists.
      if (wanted === 0 && clouds.every((c) => c.presence < 0.01)) return;
      ensureSprites();

      const lit = cloudLit(inputs.sunElevationDeg);
      const shade = cloudShade(inputs.sunElevationDeg, overcast);
      const key = `${quantize(lit)}|${quantize(shade)}`;
      if ((key !== tintKey && timeMs - lastRetint > 250) || tinted.length === 0) {
        tinted = bases.map((b) => tint(b, lit, shade));
        tintKey = key;
        lastRetint = timeMs;
      }

      // Wind: speed in canvas widths per second; clouds drift with the wind.
      const speedKmh = inputs.windSpeedKmh;
      const dir = Math.sin((inputs.windDirDeg * Math.PI) / 180);
      const base = (0.006 + speedKmh * 0.0009) * (dir === 0 ? 1 : Math.sign(dir)) * Math.max(0.35, Math.abs(dir));

      const sizeBase = w * 0.34;
      const ease = 1 - Math.exp(-dt * 2.2);

      for (let i = 0; i < clouds.length; i++) {
        const c = clouds[i]!;
        const target = i < Math.max(wanted, 0) ? 1 : 0;
        c.presence += (target - c.presence) * ease;
        if (c.presence < 0.01) continue;

        c.x += base * c.speed * dt;
        if (c.x > 1.3) c.x -= 1.6;
        if (c.x < -0.3) c.x += 1.6;

        const sw = sizeBase * c.scale * (0.9 + 0.5 * cover);
        const sh = (sw * SPRITE_H) / SPRITE_W;
        // Far clouds sit low near the horizon, near ones higher and larger.
        const y = horizonY * (0.62 - 0.46 * c.depth + 0.18 * c.yRand) - sh * 0.1;
        const a = c.alpha * (0.5 + 0.5 * c.depth) * (0.55 + 0.45 * cover) * c.presence;
        ctx.globalAlpha = clamp01(a);
        ctx.drawImage(tinted[c.variant]!, c.x * w - sw / 2, y - sh / 2, sw, sh);
      }
      ctx.globalAlpha = 1;

      // A thick deck dims the top of the sky.
      const deck = smoothstep(0.7, 1, cover);
      if (deck > 0.01) {
        const g = ctx.createLinearGradient(0, 0, 0, horizonY);
        g.addColorStop(0, css(shade, 0.5 * deck));
        g.addColorStop(1, css(shade, 0.08 * deck));
        ctx.fillStyle = g;
        ctx.fillRect(0, 0, w, horizonY);
      }
    },
  };
}
