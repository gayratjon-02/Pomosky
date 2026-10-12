import { maybeById } from "../core/dom";
import { ageText, describeWeather, type WeatherData } from "../sky/weather-map";
import { t } from "../core/i18n";

/** Shows "Seoul · 20°C · Mostly clear · updated 3 min ago" and credits Open-Meteo (CC-BY 4.0). */
export function createWeatherChip(): (w: WeatherData | null, city: string) => void {
  const chip = maybeById("weather");
  const text = maybeById("weather-text");
  let current: WeatherData | null = null;
  let city = "";

  const paint = () => {
    if (!chip || !text) return;
    chip.hidden = current === null;
    if (!current) return;
    const place = city ? `${city} · ` : "";
    text.textContent = `${place}${Math.round(current.temperatureC)}°C · ${describeWeather(current.code)} · ${t("updated {age}", { age: ageText(current.fetchedAtMs, Date.now()) })}`;
  };
  window.setInterval(paint, 30_000);

  return (w, c) => {
    current = w;
    city = c;
    paint();
  };
}
