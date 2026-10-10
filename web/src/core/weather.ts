import type { WeatherData } from "../sky/weather-map";

const REFRESH_MS = 5 * 60_000;

/**
 * Keeps the weather of a room fresh: fetches it when the host's place changes
 * and every few minutes. The server caches, so this is cheap for everyone.
 * A room without a place (404) or a failing upstream just means "no weather":
 * the sky keeps its clear-sky defaults.
 */
export class WeatherFeed {
  onWeather: (w: WeatherData | null) => void = () => {};

  private placeKey = "";
  private timer: number | undefined;
  private generation = 0;

  constructor(private readonly roomId: string) {}

  /** Call with the host's place whenever room state arrives; it only reacts to changes. */
  setPlace(lat: number | undefined, lon: number | undefined): void {
    const key = lat === undefined || lon === undefined ? "" : `${lat},${lon}`;
    if (key === this.placeKey) return;
    this.placeKey = key;
    window.clearInterval(this.timer);
    this.generation++;
    if (!key) {
      this.onWeather(null);
      return;
    }
    void this.load(this.generation);
    this.timer = window.setInterval(() => void this.load(this.generation), REFRESH_MS);
  }

  stop(): void {
    window.clearInterval(this.timer);
  }

  private async load(generation: number): Promise<void> {
    try {
      const res = await fetch(`/api/weather?room=${encodeURIComponent(this.roomId)}`, { headers: { Accept: "application/json" } });
      if (generation !== this.generation) return; // the place changed while waiting
      if (!res.ok) {
        if (res.status === 404) this.onWeather(null);
        return; // a failed refresh keeps the weather that is already shown
      }
      this.onWeather((await res.json()) as WeatherData);
    } catch {
      // offline: keep what we have
    }
  }
}
