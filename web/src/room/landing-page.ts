import { maybeById } from "../core/dom";
import { fetchCommunity } from "../stats/community";
import { formatTotal } from "../stats/format";
import { createSky, moonLight, moonPosition, sunPosition } from "../sky";

// Landing page: a time-lapse of one day over Seoul loops behind the hero so the
// sky feels alive. The speed is 24 hours in 90 seconds; reduced motion gets a still dusk.
const LAT = 37.5665;
const LON = 126.978;
const DAY_SECONDS = 90;
const START_HOUR_UTC = 8; // late afternoon in Seoul, so the first frame is warm

export function initLanding(): void {
  const canvas = maybeById<HTMLCanvasElement>("sky");
  if (!canvas) return;
  const reduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  const sky = createSky(canvas, { still: reduced });

  const apply = (at: Date) => {
    const sun = sunPosition(at, LAT, LON);
    const moon = moonPosition(at, LAT, LON);
    const light = moonLight(at);
    sky.update({
      sunElevationDeg: sun.elevationDeg,
      sunAzimuthDeg: sun.azimuthDeg,
      moonElevationDeg: moon.elevationDeg,
      moonAzimuthDeg: moon.azimuthDeg,
      moonPhase: light.phase,
      moonFraction: light.fraction,
      cloudCoverPct: 30,
      windSpeedKmh: 9,
      windDirDeg: 270,
    });
  };

  const base = new Date();
  base.setUTCHours(START_HOUR_UTC, 0, 0, 0);
  apply(base);
  sky.start();
  if (reduced) return;

  const t0 = performance.now();
  window.setInterval(() => {
    const elapsed = (performance.now() - t0) / 1000;
    apply(new Date(base.getTime() + (elapsed / DAY_SECONDS) * 86_400_000));
  }, 500);
}

/** "1,234 h focused together so far", from the community totals; hidden while there is nothing to show. */
export async function showCommunityLine(): Promise<void> {
  const line = maybeById("community-line");
  const hours = maybeById("community-hours");
  if (!line || !hours) return;
  const stats = await fetchCommunity();
  if (!stats || stats.all.seconds < 60) return;
  hours.textContent = formatTotal(stats.all.seconds);
  line.hidden = false;
}
