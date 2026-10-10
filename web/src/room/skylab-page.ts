import { byId } from "../core/dom";
import { createSky, moonLight, moonPosition, sunPosition, type SkyInputs } from "../sky";

// Development playground for the sky. Sun and moon come from the same astronomy
// code the real room uses; the weather values come from the sliders.

interface Place {
  lat: number;
  lon: number;
  tz: number; // UTC offset in hours
}

const PLACES: Record<string, Place> = {
  seoul: { lat: 37.5665, lon: 126.978, tz: 9 },
  tashkent: { lat: 41.2995, lon: 69.2401, tz: 5 },
  reykjavik: { lat: 64.1466, lon: -21.9426, tz: 0 },
  sydney: { lat: -33.8688, lon: 151.2093, tz: 11 },
  newyork: { lat: 40.7128, lon: -74.006, tz: -4 },
};

interface Preset {
  time: number; // minutes since local midnight
  cloud: number;
  rain: number;
  snow: number;
  wind: number;
  winddir: number;
  fog: number;
  thunder: boolean;
}

const PRESETS: Record<string, Preset> = {
  noon: { time: 12 * 60 + 20, cloud: 12, rain: 0, snow: 0, wind: 8, winddir: 270, fog: 0, thunder: false },
  golden: { time: 17 * 60 + 25, cloud: 35, rain: 0, snow: 0, wind: 10, winddir: 250, fog: 0, thunder: false },
  dusk: { time: 18 * 60 + 5, cloud: 45, rain: 0, snow: 0, wind: 12, winddir: 260, fog: 0, thunder: false },
  night: { time: 23 * 60 + 10, cloud: 6, rain: 0, snow: 0, wind: 4, winddir: 270, fog: 0, thunder: false },
  cloudy: { time: 14 * 60, cloud: 92, rain: 0, snow: 0, wind: 18, winddir: 300, fog: 0, thunder: false },
  rain: { time: 15 * 60, cloud: 96, rain: 4.5, snow: 0, wind: 22, winddir: 300, fog: 0.1, thunder: false },
  storm: { time: 19 * 60, cloud: 100, rain: 9, snow: 0, wind: 45, winddir: 310, fog: 0.05, thunder: true },
  snow: { time: 11 * 60, cloud: 88, rain: 0, snow: 1.8, wind: 14, winddir: 330, fog: 0.15, thunder: false },
  fog: { time: 8 * 60, cloud: 70, rain: 0, snow: 0, wind: 3, winddir: 200, fog: 0.85, thunder: false },
};

const pad = (n: number): string => String(n).padStart(2, "0");
const clock = (minutes: number): string => `${pad(Math.floor(minutes / 60))}:${pad(minutes % 60)}`;

export function initSkyLab(): void {
  const canvas = byId<HTMLCanvasElement>("sky");
  const sky = createSky(canvas);
  sky.start();

  const place = byId<HTMLSelectElement>("place");
  const date = byId<HTMLInputElement>("date");
  const time = byId<HTMLInputElement>("time");
  const readout = byId("readout");
  const sliders = ["cloud", "rain", "snow", "wind", "winddir", "fog", "quality"].map((id) => byId<HTMLInputElement>(id));
  const thunder = byId<HTMLInputElement>("thunder");

  const apply = () => {
    const p = PLACES[place.value] ?? PLACES["seoul"]!;
    const [y, m, d] = date.value.split("-").map(Number) as [number, number, number];
    const minutes = Number(time.value);
    const moment = new Date(Date.UTC(y, m - 1, d, 0, minutes) - p.tz * 3_600_000);

    const sun = sunPosition(moment, p.lat, p.lon);
    const moon = moonPosition(moment, p.lat, p.lon);
    const light = moonLight(moment);
    const v = (id: string): number => Number(byId<HTMLInputElement>(id).value);

    const inputs: Partial<SkyInputs> = {
      sunElevationDeg: sun.elevationDeg,
      sunAzimuthDeg: sun.azimuthDeg,
      moonElevationDeg: moon.elevationDeg,
      moonAzimuthDeg: moon.azimuthDeg,
      moonPhase: light.phase,
      moonFraction: light.fraction,
      cloudCoverPct: v("cloud"),
      precipitationMmh: v("rain"),
      snowfallCmh: v("snow"),
      windSpeedKmh: v("wind"),
      windDirDeg: v("winddir"),
      fogDensity: v("fog"),
      thunder: thunder.checked,
    };
    sky.update(inputs);
    sky.setQuality(v("quality"));

    byId("time-out").textContent = clock(minutes);
    for (const s of sliders) byId(`${s.id}-out`).textContent = s.value;
    readout.textContent =
      `Sun ${sun.elevationDeg.toFixed(1)}° elevation, ${sun.azimuthDeg.toFixed(0)}° bearing · ` +
      `Moon ${moon.elevationDeg.toFixed(1)}°, ${(light.fraction * 100).toFixed(0)}% lit`;
  };

  for (const el of [place, date, time, thunder, ...sliders]) el.addEventListener("input", apply);

  byId("presets").addEventListener("click", (e) => {
    const button = (e.target as HTMLElement).closest<HTMLButtonElement>("button[data-preset]");
    const preset = button && PRESETS[button.dataset["preset"] ?? ""];
    if (!preset) return;
    time.value = String(preset.time);
    byId<HTMLInputElement>("cloud").value = String(preset.cloud);
    byId<HTMLInputElement>("rain").value = String(preset.rain);
    byId<HTMLInputElement>("snow").value = String(preset.snow);
    byId<HTMLInputElement>("wind").value = String(preset.wind);
    byId<HTMLInputElement>("winddir").value = String(preset.winddir);
    byId<HTMLInputElement>("fog").value = String(preset.fog);
    thunder.checked = preset.thunder;
    apply();
  });

  // "Play the day": runs the clock through 24 hours in 40 seconds.
  const play = byId<HTMLButtonElement>("play");
  let raf = 0;
  let playing = false;
  const stopPlaying = () => {
    playing = false;
    cancelAnimationFrame(raf);
    play.textContent = "Play the day";
  };
  play.addEventListener("click", () => {
    if (playing) return stopPlaying();
    playing = true;
    play.textContent = "Stop";
    let last = performance.now();
    const tick = (now: number) => {
      if (!playing) return;
      const dtSeconds = (now - last) / 1000;
      last = now;
      time.value = String((Number(time.value) + (dtSeconds * 1440) / 40) % 1440);
      apply();
      raf = requestAnimationFrame(tick);
    };
    raf = requestAnimationFrame(tick);
  });
  time.addEventListener("pointerdown", stopPlaying);

  const toggle = byId<HTMLButtonElement>("lab-toggle");
  const body = byId("lab-body");
  toggle.addEventListener("click", () => {
    const open = toggle.getAttribute("aria-expanded") !== "true";
    toggle.setAttribute("aria-expanded", String(open));
    toggle.textContent = open ? "Hide" : "Show";
    body.hidden = !open;
  });

  apply();
}
