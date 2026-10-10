import type { RoomClient } from "../core/room-client";
import { WeatherFeed } from "../core/weather";
import { createWeatherChip } from "../ui/weather-chip";
import { createSky } from "./index";
import { skyPlace } from "./place";
import { astroInputs } from "./inputs";
import { skyFromWeather, type WeatherData } from "./weather-map";

/**
 * The room's living sky: sun and moon for the host's place at the current
 * time, plus the weather there. Returns a function that redraws it.
 */
export function startLiveSky(canvas: HTMLCanvasElement | null, client: RoomClient, roomId: string): () => void {
  const sky = canvas ? createSky(canvas) : null;
  const chip = createWeatherChip();
  const feed = new WeatherFeed(roomId);
  let weather: WeatherData | null = null;

  const refresh = () => {
    const p = skyPlace(client.state);
    sky?.update({ ...astroInputs(new Date(), p.lat, p.lon), ...(weather ? skyFromWeather(weather) : {}) });
  };

  feed.onWeather = (w) => {
    weather = w;
    chip(w, client.state?.city ?? "");
    refresh();
  };
  client.on("state", (s) => {
    feed.setPlace(s.location?.lat, s.location?.lon);
    chip(weather, s.city);
    refresh();
  });

  refresh();
  sky?.start();
  window.setInterval(refresh, 30_000);
  return refresh;
}
