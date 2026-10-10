import { moonLight, moonPosition, sunPosition } from "./astro";
import type { SkyInputs } from "./types";

/** Sun and moon part of the sky inputs for a moment and a place. Weather is added separately. */
export function astroInputs(at: Date, lat: number, lon: number): Partial<SkyInputs> {
  const sun = sunPosition(at, lat, lon);
  const moon = moonPosition(at, lat, lon);
  const light = moonLight(at);
  return {
    sunElevationDeg: sun.elevationDeg,
    sunAzimuthDeg: sun.azimuthDeg,
    moonElevationDeg: moon.elevationDeg,
    moonAzimuthDeg: moon.azimuthDeg,
    moonPhase: light.phase,
    moonFraction: light.fraction,
  };
}
