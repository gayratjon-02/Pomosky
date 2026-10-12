import type { SkyInputs } from "./types";

// Moods: ready-made looks the visitor can pick instead of the live sky. A mood fixes the
// sky (time of day and weather) and, through html[data-mood] in styles/moods.css, the
// colours of the whole interface. The choice is kept in a cookie so it is back on the
// next visit, and the server writes it into the page (data-mood) before the first paint.

export type MoodId =
  | "live"
  | "dawn"
  | "clear"
  | "golden"
  | "sunset"
  | "blue"
  | "night"
  | "moon"
  | "rain"
  | "snow"
  | "storm";

export interface Mood {
  id: MoodId;
  name: string;
  /** One short line under the name. */
  hint: string;
  /** The fixed sky. null: follow the real time and weather. */
  sky: Partial<SkyInputs> | null;
  /** Where the sun or moon sits in the little preview (0..1 of width and height), if visible. */
  orb?: { x: number; y: number; color: string };
}

const CALM = { windSpeedKmh: 8, windDirDeg: 270, precipitationMmh: 0, snowfallCmh: 0, fogDensity: 0, thunder: false };
const NO_MOON = { moonElevationDeg: -40, moonAzimuthDeg: 90, moonPhase: 0.5, moonFraction: 1 };

export const MOODS: readonly Mood[] = [
  { id: "live", name: "Live sky", hint: "The real sky above the host", sky: null },
  {
    id: "dawn",
    name: "Dawn",
    hint: "Soft pink light before sunrise",
    sky: { ...CALM, ...NO_MOON, sunElevationDeg: -2.5, sunAzimuthDeg: 96, cloudCoverPct: 32, windSpeedKmh: 6 },
    orb: { x: 0.22, y: 0.78, color: "#ffd1b0" },
  },
  {
    id: "clear",
    name: "Clear day",
    hint: "Bright blue, a few clouds",
    sky: { ...CALM, ...NO_MOON, sunElevationDeg: 52, sunAzimuthDeg: 178, cloudCoverPct: 8 },
    orb: { x: 0.5, y: 0.22, color: "#fffbe8" },
  },
  {
    id: "golden",
    name: "Golden hour",
    hint: "Warm low sun",
    sky: { ...CALM, ...NO_MOON, sunElevationDeg: 5, sunAzimuthDeg: 252, cloudCoverPct: 28 },
    orb: { x: 0.78, y: 0.66, color: "#ffe2a8" },
  },
  {
    id: "sunset",
    name: "Sunset",
    hint: "Pink and violet clouds",
    sky: { ...CALM, ...NO_MOON, sunElevationDeg: -3, sunAzimuthDeg: 262, cloudCoverPct: 40, windSpeedKmh: 10 },
    orb: { x: 0.8, y: 0.8, color: "#ffb08a" },
  },
  {
    id: "blue",
    name: "Blue hour",
    hint: "Deep indigo after sunset",
    sky: { ...CALM, ...NO_MOON, sunElevationDeg: -9, sunAzimuthDeg: 270, cloudCoverPct: 22 },
  },
  {
    id: "night",
    name: "Starry night",
    hint: "Dark sky full of stars",
    sky: { ...CALM, ...NO_MOON, sunElevationDeg: -45, sunAzimuthDeg: 0, cloudCoverPct: 0, windSpeedKmh: 3 },
  },
  {
    id: "moon",
    name: "Full moon",
    hint: "A silver night",
    sky: {
      ...CALM,
      sunElevationDeg: -45,
      sunAzimuthDeg: 0,
      moonElevationDeg: 38,
      moonAzimuthDeg: 126,
      moonPhase: 0.5,
      moonFraction: 1,
      cloudCoverPct: 12,
      windSpeedKmh: 5,
    },
    orb: { x: 0.5, y: 0.3, color: "#f4f6ff" },
  },
  {
    id: "rain",
    name: "Cozy rain",
    hint: "Grey sky and falling rain",
    sky: {
      ...NO_MOON,
      sunElevationDeg: 15,
      sunAzimuthDeg: 180,
      cloudCoverPct: 92,
      precipitationMmh: 4.5,
      snowfallCmh: 0,
      fogDensity: 0.1,
      thunder: false,
      windSpeedKmh: 16,
      windDirDeg: 250,
    },
  },
  {
    id: "snow",
    name: "Snowfall",
    hint: "Quiet, white and slow",
    sky: {
      ...NO_MOON,
      sunElevationDeg: 13,
      sunAzimuthDeg: 180,
      cloudCoverPct: 86,
      precipitationMmh: 0,
      snowfallCmh: 1.4,
      fogDensity: 0.18,
      thunder: false,
      windSpeedKmh: 7,
      windDirDeg: 320,
    },
  },
  {
    id: "storm",
    name: "Thunderstorm",
    hint: "Dark clouds and lightning",
    sky: {
      ...NO_MOON,
      sunElevationDeg: -8,
      sunAzimuthDeg: 200,
      cloudCoverPct: 100,
      precipitationMmh: 9,
      snowfallCmh: 0,
      fogDensity: 0.05,
      thunder: true,
      windSpeedKmh: 38,
      windDirDeg: 235,
    },
  },
];

const COOKIE = "pomosky_mood"; // the server reads the same cookie (internal/views/mood.go)
const YEAR = 60 * 60 * 24 * 365;

const byId = (id: string): Mood | undefined => MOODS.find((m) => m.id === id);

function readCookie(): string | null {
  try {
    const match = document.cookie.split("; ").find((c) => c.startsWith(`${COOKIE}=`));
    return match ? decodeURIComponent(match.slice(COOKIE.length + 1)) : null;
  } catch {
    return null;
  }
}

function writeCookie(id: MoodId): void {
  try {
    const secure = location.protocol === "https:" ? "; Secure" : "";
    document.cookie = `${COOKIE}=${encodeURIComponent(id)}; Max-Age=${YEAR}; Path=/; SameSite=Lax${secure}`;
  } catch {
    // cookies blocked: the mood still applies until the page closes
  }
}

const hasDom = typeof document !== "undefined";
let current: Mood = (hasDom && byId(readCookie() ?? document.documentElement.dataset["mood"] ?? "")) || MOODS[0]!;
const listeners = new Set<() => void>();

function paint(): void {
  if (!hasDom) return;
  if (current.id === "live") delete document.documentElement.dataset["mood"];
  else document.documentElement.dataset["mood"] = current.id;
}
paint();

export const currentMood = (): Mood => current;

/** The sky values of the chosen mood, or null while the live sky is on. */
export const moodSky = (): Partial<SkyInputs> | null => current.sky;

/** Chooses a mood: repaints the interface, remembers the choice and tells the sky. */
export function setMood(id: MoodId): void {
  const next = byId(id);
  if (!next || next.id === current.id) return;
  current = next;
  paint();
  writeCookie(next.id);
  listeners.forEach((fn) => fn());
}

/** Calls fn whenever the mood changes; returns a function that stops listening. */
export function onMoodChange(fn: () => void): () => void {
  listeners.add(fn);
  return () => listeners.delete(fn);
}
