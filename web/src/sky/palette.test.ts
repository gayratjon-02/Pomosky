import { describe, expect, it } from "vitest";
import { clamp01, css, desaturate, hex, luminance, mix, smoothstep } from "./color";
import { cloudLit, skyGradient, sunColor } from "./palette";

describe("color helpers", () => {
  it("parses short and long hex colours", () => {
    expect(hex("#fff")).toEqual([255, 255, 255]);
    expect(hex("#1566cc")).toEqual([21, 102, 204]);
  });

  it("mixes and formats colours", () => {
    expect(mix([0, 0, 0], [100, 200, 50], 0.5)).toEqual([50, 100, 25]);
    expect(css([300, -5, 12.4])).toBe("rgb(255, 0, 12)");
    expect(css([10, 20, 30], 0.5)).toBe("rgba(10, 20, 30, 0.5)");
  });

  it("clamps and smooths", () => {
    expect(clamp01(2)).toBe(1);
    expect(smoothstep(0, 10, -5)).toBe(0);
    expect(smoothstep(0, 10, 15)).toBe(1);
    expect(smoothstep(0, 10, 5)).toBeCloseTo(0.5, 5);
  });

  it("desaturates towards grey", () => {
    const [r, g, b] = desaturate([255, 0, 0], 1);
    expect(Math.abs(r - g)).toBeLessThan(1);
    expect(Math.abs(g - b)).toBeLessThan(1);
  });
});

describe("skyGradient", () => {
  const brightness = (e: number) => luminance(skyGradient(e).bottom) + luminance(skyGradient(e).top);

  it("gets brighter from night up to golden hour", () => {
    const elevations = [-30, -18, -12, -6, 0, 6, 12];
    for (let i = 1; i < elevations.length; i++) {
      expect(brightness(elevations[i]!)).toBeGreaterThanOrEqual(brightness(elevations[i - 1]!) - 1e-9);
    }
  });

  it("is brighter by day than at dusk (the golden-hour horizon glow may peak earlier)", () => {
    expect(brightness(40)).toBeGreaterThan(brightness(-6));
    expect(brightness(40)).toBeGreaterThan(brightness(-12));
  });

  it("is very dark at night and bright by day", () => {
    expect(luminance(skyGradient(-30).top)).toBeLessThan(0.05);
    expect(luminance(skyGradient(45).bottom)).toBeGreaterThan(0.6);
  });

  it("clamps outside the keyframe range", () => {
    expect(skyGradient(-90)).toEqual(skyGradient(-18));
    expect(skyGradient(90)).toEqual(skyGradient(60));
  });

  it("is warm at the horizon during sunset", () => {
    const [r, , b] = skyGradient(0).bottom;
    expect(r).toBeGreaterThan(b);
  });

  it("is greyer and darker when overcast", () => {
    const clear = skyGradient(40, 0);
    const grey = skyGradient(40, 1);
    const spread = (c: number[]) => Math.max(...c) - Math.min(...c);
    expect(spread(grey.top)).toBeLessThan(spread(clear.top));
    expect(luminance(grey.top)).toBeLessThanOrEqual(luminance(clear.top) + 0.05);
  });
});

describe("sun and cloud colours", () => {
  it("makes the sun orange near the horizon and pale when high", () => {
    const low = sunColor(0);
    const high = sunColor(50);
    expect(low[0]).toBeGreaterThan(low[2]);
    expect(high[2]).toBeGreaterThan(low[2]);
  });

  it("lights clouds white by day and dim at night", () => {
    expect(luminance(cloudLit(45))).toBeGreaterThan(0.9);
    expect(luminance(cloudLit(-30))).toBeLessThan(0.4);
  });
});
