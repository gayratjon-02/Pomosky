import { maybeById } from "../core/dom";
import { createSky } from "../sky";
import { astroInputs } from "../sky/inputs";

/** A calm, still-weather sky behind simple pages (login, errors). Follows the visitor's own clock. */
export function initAmbientSky(): void {
  const canvas = maybeById<HTMLCanvasElement>("sky");
  if (!canvas) return;
  const sky = createSky(canvas);
  const refresh = () => sky.update({ ...astroInputs(new Date(), 37.57, 126.98), cloudCoverPct: 35, windSpeedKmh: 8 });
  refresh();
  sky.start();
  window.setInterval(refresh, 60_000);
}
