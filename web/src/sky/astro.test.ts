import { describe, expect, it } from "vitest";
import { moonLight, moonPosition, sunPosition } from "./astro";

const SEOUL = { lat: 37.5665, lon: 126.978 };

// 2026-10-10 in Seoul (KST = UTC+9).
const kst = (h: number, m = 0): Date => new Date(Date.UTC(2026, 9, 10, h - 9, m));

describe("sunPosition", () => {
  it("is high above the horizon around noon in Seoul in October", () => {
    // Solar noon in Seoul is roughly 12:20 KST; the sun then stands about 45 degrees high.
    const noon = sunPosition(kst(12, 20), SEOUL.lat, SEOUL.lon);
    expect(noon.elevationDeg).toBeGreaterThan(43);
    expect(noon.elevationDeg).toBeLessThan(48);
    expect(noon.azimuthDeg).toBeGreaterThan(170);
    expect(noon.azimuthDeg).toBeLessThan(190);
  });

  it("rises in the east and sets in the west", () => {
    const morning = sunPosition(kst(7), SEOUL.lat, SEOUL.lon);
    const evening = sunPosition(kst(17), SEOUL.lat, SEOUL.lon);
    expect(morning.azimuthDeg).toBeGreaterThan(80);
    expect(morning.azimuthDeg).toBeLessThan(140);
    expect(evening.azimuthDeg).toBeGreaterThan(220);
    expect(evening.azimuthDeg).toBeLessThan(280);
  });

  it("is below the horizon before sunrise and after sunset", () => {
    expect(sunPosition(kst(5), SEOUL.lat, SEOUL.lon).elevationDeg).toBeLessThan(-5);
    expect(sunPosition(kst(6), SEOUL.lat, SEOUL.lon).elevationDeg).toBeLessThan(0); // sunrise is ~06:25
    expect(sunPosition(kst(7), SEOUL.lat, SEOUL.lon).elevationDeg).toBeGreaterThan(3);
    expect(sunPosition(kst(17, 30), SEOUL.lat, SEOUL.lon).elevationDeg).toBeGreaterThan(0); // sunset is ~17:55
    expect(sunPosition(kst(18, 30), SEOUL.lat, SEOUL.lon).elevationDeg).toBeLessThan(0);
    expect(sunPosition(kst(0), SEOUL.lat, SEOUL.lon).elevationDeg).toBeLessThan(-30);
  });

  it("is straight overhead at the equator on the equinox", () => {
    const p = sunPosition(new Date(Date.UTC(2026, 2, 20, 12, 0)), 0, 0);
    expect(p.elevationDeg).toBeGreaterThan(84);
  });

  it("is mirrored in the southern hemisphere", () => {
    // Sydney, noon in December (summer): the sun is high and to the north.
    const p = sunPosition(new Date(Date.UTC(2026, 11, 21, 1, 30)), -33.87, 151.21);
    expect(p.elevationDeg).toBeGreaterThan(70);
  });
});

describe("moonLight", () => {
  it("is new moon at the reference new moon (2000-01-06 18:14 UTC)", () => {
    const m = moonLight(new Date(Date.UTC(2000, 0, 6, 18, 14)));
    expect(m.fraction).toBeLessThan(0.03);
  });

  it("is full moon at the reference full moon (2000-01-21 04:40 UTC)", () => {
    const m = moonLight(new Date(Date.UTC(2000, 0, 21, 4, 40)));
    expect(m.fraction).toBeGreaterThan(0.97);
    expect(Math.abs(m.phase - 0.5)).toBeLessThan(0.03);
  });

  it("waxes through the first quarter after a new moon", () => {
    const m = moonLight(new Date(Date.UTC(2000, 0, 14, 0, 0)));
    expect(m.phase).toBeGreaterThan(0.2);
    expect(m.phase).toBeLessThan(0.3);
    expect(m.fraction).toBeGreaterThan(0.4);
    expect(m.fraction).toBeLessThan(0.6);
  });

  it("keeps its values in range", () => {
    for (let day = 0; day < 60; day++) {
      const m = moonLight(new Date(Date.UTC(2026, 9, 1 + day)));
      expect(m.phase).toBeGreaterThanOrEqual(0);
      expect(m.phase).toBeLessThanOrEqual(1);
      expect(m.fraction).toBeGreaterThanOrEqual(0);
      expect(m.fraction).toBeLessThanOrEqual(1);
    }
  });
});

describe("moonPosition", () => {
  it("returns a finite position that moves across the sky", () => {
    const a = moonPosition(kst(20), SEOUL.lat, SEOUL.lon);
    const b = moonPosition(kst(23), SEOUL.lat, SEOUL.lon);
    expect(Number.isFinite(a.elevationDeg)).toBe(true);
    expect(Math.abs(a.azimuthDeg - b.azimuthDeg)).toBeGreaterThan(5);
  });
});
