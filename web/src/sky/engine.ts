import { clamp, clamp01 } from "./color";
import { backgroundLayer } from "./layers/background";
import { cloudsLayer } from "./layers/clouds";
import { fogLayer } from "./layers/fog";
import { horizonLayer } from "./layers/horizon";
import { lightningLayer } from "./layers/lightning";
import { moonLayer } from "./layers/moon";
import { rainLayer } from "./layers/rain";
import { snowLayer } from "./layers/snow";
import { starsLayer } from "./layers/stars";
import { sunLayer } from "./layers/sun";
import { vignetteLayer } from "./layers/vignette";
import { onVisibilityChange, prefersReducedMotion } from "./perf";
import { DEFAULT_INPUTS, type Frame, type Layer, type SkyController, type SkyInputs } from "./types";

export interface SkyOptions {
  /** Highest device pixel ratio to render at (default 2). */
  maxDpr?: number;
  /** Start without any animation loop, rendering single frames on demand. */
  still?: boolean;
}

const MAX_PIXELS = 4_500_000;

/** Moves `current` towards `target`; angles take the short way round. */
const approach = (current: number, target: number, k: number): number => current + (target - current) * k;
const approachAngle = (current: number, target: number, k: number): number => {
  const delta = ((target - current + 540) % 360) - 180;
  return (current + delta * k + 360) % 360;
};

/** Cloud cover and precipitation both make the sky look overcast. */
export function overcastOf(i: SkyInputs): number {
  const cover = clamp01(i.cloudCoverPct / 100);
  return clamp01(Math.pow(cover, 1.2) + 0.2 * clamp01(i.precipitationMmh / 6) + 0.15 * clamp01(i.snowfallCmh / 2));
}

/**
 * The live sky: a stack of canvas layers (gradient, stars, moon, sun, clouds,
 * hills, fog, rain, snow, lightning, vignette) drawn each frame.
 *
 * - Inputs are blended in smoothly, so a weather update never makes the sky jump.
 * - The loop pauses while the tab is hidden and draws a single frame when the
 *   user prefers reduced motion.
 * - If frames are slow the particle quality is lowered automatically.
 */
export function createSky(canvas: HTMLCanvasElement, options: SkyOptions = {}): SkyController {
  const ctx = canvas.getContext("2d", { alpha: false });
  const layers: Layer[] = [
    backgroundLayer(),
    starsLayer(),
    moonLayer(),
    sunLayer(),
    cloudsLayer(),
    horizonLayer(),
    fogLayer(),
    rainLayer(),
    snowLayer(),
    lightningLayer(),
    vignetteLayer(),
  ];

  const target: SkyInputs = { ...DEFAULT_INPUTS };
  const current: SkyInputs = { ...DEFAULT_INPUTS };
  const frame: Frame = {
    ctx: ctx as CanvasRenderingContext2D,
    w: 1,
    h: 1,
    dpr: 1,
    timeMs: 0,
    dt: 0,
    inputs: current,
    quality: 1,
    horizonY: 1,
    overcast: 0,
  };

  const reduceMotion = options.still === true || prefersReducedMotion();
  let userQuality = 1;
  let quality = 1;
  let slowEma = 0.016;
  let frames = 0;
  let raf = 0;
  let last = 0;
  let running = false;
  let destroyed = false;

  const resize = () => {
    let dpr = Math.min(window.devicePixelRatio || 1, options.maxDpr ?? 2);
    const cssW = Math.max(1, canvas.clientWidth);
    const cssH = Math.max(1, canvas.clientHeight);
    while (cssW * cssH * dpr * dpr > MAX_PIXELS && dpr > 1) dpr -= 0.25;
    const w = Math.max(1, Math.round(cssW * dpr));
    const h = Math.max(1, Math.round(cssH * dpr));
    if (canvas.width !== w || canvas.height !== h) {
      canvas.width = w;
      canvas.height = h;
    }
    frame.w = w;
    frame.h = h;
    frame.dpr = dpr;
    frame.horizonY = h * 0.8;
    for (const layer of layers) layer.resize?.(w, h, dpr);
    if (!running) render(performance.now(), 0);
  };

  const step = (dt: number) => {
    const k = 1 - Math.exp(-dt * 2.5); // blends over roughly a second
    current.sunElevationDeg = approach(current.sunElevationDeg, target.sunElevationDeg, k);
    current.sunAzimuthDeg = approachAngle(current.sunAzimuthDeg, target.sunAzimuthDeg, k);
    current.moonElevationDeg = approach(current.moonElevationDeg, target.moonElevationDeg, k);
    current.moonAzimuthDeg = approachAngle(current.moonAzimuthDeg, target.moonAzimuthDeg, k);
    current.moonPhase = target.moonPhase;
    current.moonFraction = approach(current.moonFraction, target.moonFraction, k);
    current.cloudCoverPct = approach(current.cloudCoverPct, target.cloudCoverPct, k);
    current.precipitationMmh = approach(current.precipitationMmh, target.precipitationMmh, k);
    current.snowfallCmh = approach(current.snowfallCmh, target.snowfallCmh, k);
    current.windSpeedKmh = approach(current.windSpeedKmh, target.windSpeedKmh, k);
    current.windDirDeg = approachAngle(current.windDirDeg, target.windDirDeg, k);
    current.fogDensity = approach(current.fogDensity, target.fogDensity, k);
    current.thunder = target.thunder;
  };

  const render = (timeMs: number, dt: number) => {
    if (!ctx) return;
    frame.timeMs = timeMs;
    frame.dt = dt;
    frame.quality = quality;
    frame.overcast = overcastOf(current);
    ctx.globalAlpha = 1;
    ctx.globalCompositeOperation = "source-over";
    for (const layer of layers) layer.draw(frame);
  };

  const adaptQuality = (dt: number) => {
    slowEma = slowEma * 0.95 + dt * 0.05;
    if (++frames % 60 !== 0) return;
    if (slowEma > 0.024 && quality > 0.35) quality = Math.max(0.35, quality - 0.15);
    else if (slowEma < 0.014 && quality < userQuality) quality = Math.min(userQuality, quality + 0.05);
  };

  const loop = (now: number) => {
    if (!running) return;
    const dt = clamp((now - last) / 1000, 0, 0.1);
    last = now;
    step(dt);
    render(now, dt);
    adaptQuality(dt);
    raf = requestAnimationFrame(loop);
  };

  const observer = new ResizeObserver(resize);
  observer.observe(canvas);
  const stopWatching = onVisibilityChange((visible) => {
    if (visible) controller.start();
    else pause();
  });

  const pause = () => {
    running = false;
    cancelAnimationFrame(raf);
  };

  const controller: SkyController = {
    start() {
      if (running || destroyed) return;
      resize();
      if (reduceMotion) {
        Object.assign(current, target);
        render(performance.now(), 0);
        return;
      }
      running = true;
      last = performance.now();
      raf = requestAnimationFrame(loop);
    },
    stop: pause,
    update(next) {
      Object.assign(target, next);
      if (!running) {
        Object.assign(current, target);
        render(performance.now(), 0);
      }
    },
    setQuality(q) {
      userQuality = clamp(q, 0.3, 1);
      quality = Math.min(quality, userQuality);
    },
    destroy() {
      destroyed = true;
      pause();
      observer.disconnect();
      stopWatching();
    },
  };
  return controller;
}
