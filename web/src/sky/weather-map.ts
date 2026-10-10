import type { SkyInputs } from "./types";

/** The JSON of GET /api/weather (see docs/api.md). */
export interface WeatherData {
  cloudCoverPct: number;
  precipitationMmh: number;
  snowfallCmh: number;
  windSpeedKmh: number;
  windDirDeg: number;
  fogDensity: number;
  thunder: boolean;
  isDay: boolean;
  temperatureC: number;
  code: number;
  observedAtMs: number;
  fetchedAtMs: number;
}

/** Sky inputs that come from the weather. Sun and moon are computed separately. */
export function skyFromWeather(w: WeatherData): Partial<SkyInputs> {
  return {
    cloudCoverPct: w.cloudCoverPct,
    precipitationMmh: w.precipitationMmh,
    snowfallCmh: w.snowfallCmh,
    windSpeedKmh: w.windSpeedKmh,
    windDirDeg: w.windDirDeg,
    fogDensity: w.fogDensity,
    thunder: w.thunder,
  };
}

/** Plain-words description of a WMO weather code (Open-Meteo's table). */
export function describeWeather(code: number): string {
  if (code === 0) return "Clear";
  if (code === 1) return "Mostly clear";
  if (code === 2) return "Partly cloudy";
  if (code === 3) return "Overcast";
  if (code === 45 || code === 48) return "Fog";
  if (code >= 51 && code <= 57) return "Drizzle";
  if ((code >= 61 && code <= 67) || (code >= 80 && code <= 82)) return "Rain";
  if ((code >= 71 && code <= 77) || code === 85 || code === 86) return "Snow";
  if (code >= 95 && code <= 99) return "Thunderstorm";
  return "Weather";
}

/** "just now", "5 min ago", "2 h ago". */
export function ageText(sinceMs: number, nowMs: number): string {
  const min = Math.floor(Math.max(0, nowMs - sinceMs) / 60_000);
  if (min < 1) return "just now";
  if (min < 60) return `${min} min ago`;
  return `${Math.floor(min / 60)} h ago`;
}
