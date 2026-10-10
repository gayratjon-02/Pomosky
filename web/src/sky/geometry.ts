import type { Frame } from "./types";

type Size = Pick<Frame, "w" | "h" | "horizonY">;

/**
 * Maps a compass bearing and an elevation to a point on the canvas. The view
 * looks south: east is on the left, west on the right, and the horizon sits at
 * `horizonY`. Points can lie outside the canvas (a sun in the north-east).
 */
export function skyPoint(size: Size, azimuthDeg: number, elevationDeg: number): { x: number; y: number } {
  const relative = ((azimuthDeg - 180 + 540) % 360) - 180; // -180..180, 0 = due south
  return {
    x: size.w * (0.5 + relative / 200),
    y: size.horizonY - (elevationDeg / 70) * size.horizonY * 0.9,
  };
}

/** Number of particles for the area, relative to a 1920x1080 screen. */
export function areaScale(w: number, h: number): number {
  return (w * h) / (1920 * 1080);
}
