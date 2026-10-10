export type Condition = "unknown" | "clear" | "cloudy" | "rain" | "snow" | "fog" | "storm";

/**
 * Everything the renderer needs. Sun and moon values are computed locally
 * (see astro.ts); the rest comes from the weather data (or from the sky lab).
 */
export interface SkyInputs {
  sunElevationDeg: number; // -90..90, negative = below the horizon
  sunAzimuthDeg: number; // compass bearing, 180 = south
  moonElevationDeg: number;
  moonAzimuthDeg: number;
  moonPhase: number; // 0 new, 0.5 full
  moonFraction: number; // lit fraction of the disc, 0..1
  cloudCoverPct: number; // 0..100
  precipitationMmh: number; // rain, mm per hour
  snowfallCmh: number; // snow, cm per hour
  windSpeedKmh: number;
  windDirDeg: number; // direction the wind blows from
  fogDensity: number; // 0..1
  thunder: boolean;
}

export const DEFAULT_INPUTS: SkyInputs = {
  sunElevationDeg: 30,
  sunAzimuthDeg: 180,
  moonElevationDeg: -30,
  moonAzimuthDeg: 90,
  moonPhase: 0.5,
  moonFraction: 1,
  cloudCoverPct: 20,
  precipitationMmh: 0,
  snowfallCmh: 0,
  windSpeedKmh: 10,
  windDirDeg: 270,
  fogDensity: 0,
  thunder: false,
};

export interface SkyController {
  /** Starts the animation loop (a single static frame when the user prefers reduced motion). */
  start(): void;
  stop(): void;
  /** Merges new inputs; they are blended in smoothly. */
  update(inputs: Partial<SkyInputs>): void;
  /** Particle density multiplier, 0.3..1. The engine lowers it by itself when frames are slow. */
  setQuality(quality: number): void;
  destroy(): void;
}

/** What a layer receives for every frame. */
export interface Frame {
  ctx: CanvasRenderingContext2D;
  w: number; // canvas size in device pixels
  h: number;
  dpr: number;
  timeMs: number;
  dt: number; // seconds since the previous frame, clamped
  inputs: SkyInputs; // smoothed values
  quality: number;
  horizonY: number;
  overcast: number; // 0..1, derived from cloud cover and precipitation
}

export interface Layer {
  /** Called when the canvas size changes. */
  resize?(w: number, h: number, dpr: number): void;
  draw(frame: Frame): void;
}
