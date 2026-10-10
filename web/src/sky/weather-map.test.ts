import { describe, expect, it } from "vitest";
import { ageText, describeWeather, skyFromWeather } from "./weather-map";

describe("describeWeather", () => {
  it("covers the WMO groups", () => {
    expect(describeWeather(0)).toBe("Clear");
    expect(describeWeather(3)).toBe("Overcast");
    expect(describeWeather(45)).toBe("Fog");
    expect(describeWeather(55)).toBe("Drizzle");
    expect(describeWeather(63)).toBe("Rain");
    expect(describeWeather(81)).toBe("Rain");
    expect(describeWeather(73)).toBe("Snow");
    expect(describeWeather(95)).toBe("Thunderstorm");
    expect(describeWeather(999)).toBe("Weather");
  });
});

describe("ageText", () => {
  it("formats minutes and hours", () => {
    expect(ageText(1000, 1000 + 20_000)).toBe("just now");
    expect(ageText(0, 5 * 60_000)).toBe("5 min ago");
    expect(ageText(0, 130 * 60_000)).toBe("2 h ago");
    expect(ageText(10_000, 0)).toBe("just now");
  });
});

describe("skyFromWeather", () => {
  it("passes the numbers through", () => {
    const sky = skyFromWeather({
      cloudCoverPct: 80, precipitationMmh: 2, snowfallCmh: 0, windSpeedKmh: 15, windDirDeg: 200,
      fogDensity: 0.1, thunder: true, isDay: true, temperatureC: 12, code: 95, observedAtMs: 0, fetchedAtMs: 0,
    });
    expect(sky).toMatchObject({ cloudCoverPct: 80, precipitationMmh: 2, windDirDeg: 200, thunder: true, fogDensity: 0.1 });
  });
});
