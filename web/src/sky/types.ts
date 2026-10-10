export type Condition = "unknown" | "clear" | "cloudy" | "rain" | "snow" | "fog" | "storm";

/** Everything the renderer needs. Sun and moon values are computed locally; the rest comes from /api/weather. */
export interface SkyInputs {
  sunElevationDeg: number; // -90..90, negative = below the horizon
  sunAzimuthDeg: number;
  moonPhase: number; // 0..1
  cloudCoverPct: number; // 0..100
  precipitationMmh: number;
  snowfallCmh: number;
  windSpeedKmh: number;
  windDirDeg: number;
  condition: Condition;
}

export interface SkyController {
  start(): void;
  stop(): void;
  update(inputs: Partial<SkyInputs>): void;
}
