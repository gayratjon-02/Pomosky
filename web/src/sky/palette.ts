// Maps the sun elevation to sky gradient colours.
// TODO(stage 6): smooth interpolation through dawn, day, golden hour, dusk, night.

export interface SkyGradient {
  top: string;
  bottom: string;
}

export function skyGradient(sunElevationDeg: number): SkyGradient {
  if (sunElevationDeg > 10) return { top: "#3b82f6", bottom: "#bfdbfe" }; // day
  if (sunElevationDeg > -6) return { top: "#7c3aed", bottom: "#fb923c" }; // twilight / golden hour
  return { top: "#050816", bottom: "#111b3a" }; // night
}
